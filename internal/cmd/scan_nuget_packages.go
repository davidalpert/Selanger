package cmd

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/davidalpert/go-printers/v1"
	"github.com/davidalpert/selanger/internal/dotnet"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type ScanNugetPackagesOptions struct {
	*printers.PrinterOptions
	SolutionFileOrFolder string
	FS                   afero.Fs
	ShowAbsolutePaths    bool
	ShowProjectView      bool
}

func NewScanNugetPackagesOptions(s printers.IOStreams) *ScanNugetPackagesOptions {
	return &ScanNugetPackagesOptions{
		PrinterOptions: printers.NewPrinterOptions().WithStreams(s).WithDefaultTableWriter(),
	}
}

func NewCmdScanNugetPackages(s printers.IOStreams) *cobra.Command {
	o := NewScanNugetPackagesOptions(s)
	var cmd = &cobra.Command{
		Use:     "nuget-packages <path to .sln file or a containing folder>",
		Aliases: []string{"n", "nu", "nuget"},
		Short:   "scan and validate NuGet package references in a .NET solution",
		Args:    cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(cmd, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			return o.Run()
		},
	}

	o.AddPrinterFlags(cmd.Flags())
	cmd.Flags().BoolVar(&o.ShowAbsolutePaths, "absolute-paths", false, "show absolute paths in output (default is relative paths)")
	cmd.Flags().BoolVar(&o.ShowProjectView, "project-view", false, "show project view (only relevant when output format is table")

	return cmd
}

// Complete the options
func (o *ScanNugetPackagesOptions) Complete(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		o.SolutionFileOrFolder = "."
	} else if len(args) > 0 {
		o.SolutionFileOrFolder = args[0]
	}

	if absPath, err := filepath.Abs(o.SolutionFileOrFolder); err != nil {
		return err
	} else {
		o.SolutionFileOrFolder = absPath
	}

	o.FS = afero.NewOsFs()

	return nil
}

// Validate the options
func (o *ScanNugetPackagesOptions) Validate() error {
	return o.PrinterOptions.Validate()
}

// Run the command
func (o *ScanNugetPackagesOptions) Run() error {
	// Create scanner and get all projects
	scanner := dotnet.NewSolutionScanner(o.FS, o.SolutionFileOrFolder, o.ShowAbsolutePaths)
	allProjects, err := scanner.GetAllProjects()
	if err != nil {
		return err
	}

	// Analyze packages for each project
	var projectValidations []dotnet.ProjectPackageValidation
	allPackages := make(map[string]map[string][]string) // packageName -> version -> []projectNames

	for _, proj := range allProjects {
		validation, err := scanner.ValidateProjectPackages(proj)
		if err != nil {
			return fmt.Errorf("error validating packages for %s: %w", proj.Name, err)
		}
		projectValidations = append(projectValidations, validation)

		// Collect all packages for solution-wide validation
		for _, pkg := range validation.AllPackages {
			if allPackages[pkg.Name] == nil {
				allPackages[pkg.Name] = make(map[string][]string)
			}
			allPackages[pkg.Name][pkg.Version] = append(allPackages[pkg.Name][pkg.Version], proj.Name)
		}
	}

	// Perform solution-wide validation
	solutionValidation := scanner.ValidateSolutionPackages(allPackages)

	// Output results
	return o.outputResults(projectValidations, solutionValidation)
}

func (o *ScanNugetPackagesOptions) outputResults(projectValidations []dotnet.ProjectPackageValidation, solutionValidation dotnet.SolutionPackageValidation) error {
	// Create combined results structure
	overallValid := solutionValidation.IsValid
	for _, pv := range projectValidations {
		if !pv.IsValid {
			overallValid = false
			break
		}
	}

	results := dotnet.NugetValidationResults{
		ProjectValidations: projectValidations,
		SolutionValidation: solutionValidation,
		OverallValid:       overallValid,
	}

	sort.Slice(results.ProjectValidations, func(i, j int) bool {
		pa := results.ProjectValidations[i]
		pb := results.ProjectValidations[j]
		return strings.Compare(pa.ProjectName, pb.ProjectName) > 0
	})

	sort.Slice(results.SolutionValidation.PackageVersionConflicts, func(i, j int) bool {
		pa := results.SolutionValidation.PackageVersionConflicts[i]
		pb := results.SolutionValidation.PackageVersionConflicts[j]
		return strings.Compare(pa.PackageName, pb.PackageName) > 0
	})

	var renderingErr error
	if o.ShowProjectView {
		// Output per-project validation summary table
		renderingErr = o.WithTableWriter("Per-Project Package Validation", func(t *tablewriter.Table) {
			t.SetHeader([]string{"project", "status", "orphaned", "broken", "mismatched"})
			t.SetAutoWrapText(false)
			for _, pv := range projectValidations {
				status := "PASS"
				if !pv.IsValid {
					status = "FAIL"
				}
				t.Append([]string{
					pv.ProjectName,
					status,
					fmt.Sprintf("%d", len(pv.OrphanedConfigReferences)),
					fmt.Sprintf("%d", len(pv.BrokenProjectReferences)),
					fmt.Sprintf("%d", len(pv.VersionMismatches)),
				})
			}
		}).WriteOutput(results)
	} else {
		// Output solution-wide validation summary table if there are conflicts
		if len(solutionValidation.PackageVersionConflicts) > 0 {
			renderingErr = o.WithTableWriter("Solution-Wide Package Version Conflicts", func(t *tablewriter.Table) {
				t.SetHeader([]string{"package", "versions"})
				t.SetAutoWrapText(false)
				for _, conflict := range solutionValidation.PackageVersionConflicts {
					versions := make([]string, 0, len(conflict.VersionProjects))
					for version := range conflict.VersionProjects {
						versions = append(versions, version)
					}
					t.Append([]string{
						conflict.PackageName,
						fmt.Sprintf("%d", len(versions)),
					})
				}
			}).WriteOutput(results)
		}
	}

	if renderingErr != nil {
		return renderingErr
	}

	if o.FormatCategory() == "json" || o.FormatCategory() == "yaml" {
		// don't want additional error messages in the structured output
		// in case it is being piped somewhere else
		return nil
	}

	// Return error if validation failed
	if !solutionValidation.IsValid {
		return fmt.Errorf("solution-wide validation failed: %d package(s) with version conflicts", len(solutionValidation.PackageVersionConflicts))
	}

	for _, pv := range projectValidations {
		if !pv.IsValid {
			return fmt.Errorf("per-project validation failed for one or more projects")
		}
	}

	return nil
}
