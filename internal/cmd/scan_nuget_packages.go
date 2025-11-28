package cmd

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"

	"github.com/davidalpert/go-printers/v1"
	"github.com/davidalpert/selanger/internal/diagnostics"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type ScanNugetPackagesOptions struct {
	*printers.PrinterOptions
	SolutionFileOrFolder string
	FS                   afero.Fs
	ShowAbsolutePaths    bool
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
	// Reuse the solution scanning logic to get projects
	scanModulesOpts := &ScanModulesOptions{
		PrinterOptions:       o.PrinterOptions,
		SolutionFileOrFolder: o.SolutionFileOrFolder,
		FS:                   o.FS,
		ShowAbsolutePaths:    o.ShowAbsolutePaths,
	}

	slnFiles, err := scanModulesOpts.findSolutionFiles()
	if err != nil {
		return fmt.Errorf("error finding solution files: %w", err)
	}

	if len(slnFiles) == 0 {
		return fmt.Errorf("no .sln files found in %s", o.SolutionFileOrFolder)
	}

	// Parse all solution files and collect projects
	var allProjects []ProjectReference
	for _, slnFile := range slnFiles {
		diagnostics.Log.WithField("solutionFile", slnFile).Debug("parsing solution file for nuget packages")
		projects, err := scanModulesOpts.parseSolutionFile(slnFile)
		if err != nil {
			return fmt.Errorf("error parsing %s: %w", slnFile, err)
		}
		allProjects = append(allProjects, projects...)
	}

	// Analyze packages for each project
	var projectValidations []ProjectPackageValidation
	allPackages := make(map[string]map[string][]string) // packageName -> version -> []projectNames

	for _, proj := range allProjects {
		validation, err := o.validateProjectPackages(proj)
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
	solutionValidation := o.validateSolutionPackages(allPackages)

	// Output results
	return o.outputResults(projectValidations, solutionValidation)
}

// PackageReference represents a NuGet package reference
type PackageReference struct {
	Name    string
	Version string
	Source  string // "packages.config" or "project"
}

// ProjectPackageValidation holds validation results for a single project
type ProjectPackageValidation struct {
	ProjectName              string
	AllPackages              []PackageReference
	OrphanedConfigReferences []PackageReference // in packages.config but not in project
	BrokenProjectReferences  []PackageReference // in project but not in packages.config
	VersionMismatches        []VersionMismatch
	IsValid                  bool
}

// VersionMismatch represents a package with different versions in packages.config and project file
type VersionMismatch struct {
	PackageName    string
	ConfigVersion  string
	ProjectVersion string
}

// SolutionPackageValidation holds validation results for the entire solution
type SolutionPackageValidation struct {
	PackageVersionConflicts []PackageVersionConflict
	IsValid                 bool
}

// PackageVersionConflict represents a package installed at different versions across projects
type PackageVersionConflict struct {
	PackageName     string
	VersionProjects map[string][]string // version -> []projectNames
}

func (o *ScanNugetPackagesOptions) validateProjectPackages(proj ProjectReference) (ProjectPackageValidation, error) {
	validation := ProjectPackageValidation{
		ProjectName: proj.Name,
	}

	// Determine base directory: if SolutionFileOrFolder is a .sln file, use its directory
	baseDir := o.SolutionFileOrFolder
	if filepath.Ext(o.SolutionFileOrFolder) == ".sln" {
		baseDir = filepath.Dir(o.SolutionFileOrFolder)
	}

	// Get the directory containing the project file
	projectDir := filepath.Dir(proj.Path)
	if !filepath.IsAbs(projectDir) {
		projectDir = filepath.Join(baseDir, projectDir)
	}

	// Read packages.config if it exists
	packagesConfigPath := filepath.Join(projectDir, "packages.config")
	configPackages, err := o.readPackagesConfig(packagesConfigPath)
	if err != nil && !os.IsNotExist(err) {
		return validation, err
	}

	// Read PackageReference nodes from project file
	projectPackages, err := o.readProjectPackages(proj.Path)
	if err != nil {
		return validation, err
	}

	// Create maps for easier lookup
	configMap := make(map[string]string) // packageName -> version
	for _, pkg := range configPackages {
		configMap[pkg.Name] = pkg.Version
	}

	projectMap := make(map[string]string) // packageName -> version
	for _, pkg := range projectPackages {
		projectMap[pkg.Name] = pkg.Version
	}

	// Collect all unique packages
	allPackagesMap := make(map[string]PackageReference)
	for _, pkg := range configPackages {
		allPackagesMap[pkg.Name] = pkg
	}
	for _, pkg := range projectPackages {
		if existing, exists := allPackagesMap[pkg.Name]; !exists || existing.Source == "packages.config" {
			allPackagesMap[pkg.Name] = pkg
		}
	}
	for _, pkg := range allPackagesMap {
		validation.AllPackages = append(validation.AllPackages, pkg)
	}

	// Check for orphaned packages.config references
	for _, pkg := range configPackages {
		if _, exists := projectMap[pkg.Name]; !exists {
			validation.OrphanedConfigReferences = append(validation.OrphanedConfigReferences, pkg)
		}
	}

	// Check for broken project references
	for _, pkg := range projectPackages {
		if _, exists := configMap[pkg.Name]; !exists {
			validation.BrokenProjectReferences = append(validation.BrokenProjectReferences, pkg)
		}
	}

	// Check for version mismatches
	for pkgName, configVersion := range configMap {
		if projectVersion, exists := projectMap[pkgName]; exists {
			if configVersion != projectVersion {
				validation.VersionMismatches = append(validation.VersionMismatches, VersionMismatch{
					PackageName:    pkgName,
					ConfigVersion:  configVersion,
					ProjectVersion: projectVersion,
				})
			}
		}
	}

	// Validation passes only if all counts are zero
	validation.IsValid = len(validation.OrphanedConfigReferences) == 0 &&
		len(validation.BrokenProjectReferences) == 0 &&
		len(validation.VersionMismatches) == 0

	return validation, nil
}

func (o *ScanNugetPackagesOptions) validateSolutionPackages(allPackages map[string]map[string][]string) SolutionPackageValidation {
	validation := SolutionPackageValidation{
		IsValid: true,
	}

	for pkgName, versions := range allPackages {
		if len(versions) > 1 {
			conflict := PackageVersionConflict{
				PackageName:     pkgName,
				VersionProjects: versions,
			}
			validation.PackageVersionConflicts = append(validation.PackageVersionConflicts, conflict)
			validation.IsValid = false
		}
	}

	return validation
}

// PackagesConfig represents the packages.config XML structure
type PackagesConfig struct {
	XMLName  xml.Name `xml:"packages"`
	Packages []struct {
		ID      string `xml:"id,attr"`
		Version string `xml:"version,attr"`
	} `xml:"package"`
}

func (o *ScanNugetPackagesOptions) readPackagesConfig(path string) ([]PackageReference, error) {
	var packages []PackageReference

	data, err := afero.ReadFile(o.FS, path)
	if err != nil {
		return packages, err
	}

	var config PackagesConfig
	if err := xml.Unmarshal(data, &config); err != nil {
		return packages, fmt.Errorf("error parsing packages.config: %w", err)
	}

	for _, pkg := range config.Packages {
		packages = append(packages, PackageReference{
			Name:    pkg.ID,
			Version: pkg.Version,
			Source:  "packages.config",
		})
	}

	return packages, nil
}

// ProjectFile represents the relevant parts of a .csproj XML structure
type ProjectFile struct {
	XMLName           xml.Name `xml:"Project"`
	ItemGroups        []ItemGroup
	PackageReferences []PackageReferenceXML `xml:"ItemGroup>PackageReference"`
}

type ItemGroup struct {
	PackageReferences []PackageReferenceXML `xml:"PackageReference"`
}

type PackageReferenceXML struct {
	Include string `xml:"Include,attr"`
	Version string `xml:"Version,attr"`
}

func (o *ScanNugetPackagesOptions) readProjectPackages(projectPath string) ([]PackageReference, error) {
	var packages []PackageReference

	// Determine base directory: if SolutionFileOrFolder is a .sln file, use its directory
	baseDir := o.SolutionFileOrFolder
	if filepath.Ext(o.SolutionFileOrFolder) == ".sln" {
		baseDir = filepath.Dir(o.SolutionFileOrFolder)
	}

	// Resolve full path
	fullPath := projectPath
	if !filepath.IsAbs(projectPath) {
		fullPath = filepath.Join(baseDir, projectPath)
	}

	data, err := afero.ReadFile(o.FS, fullPath)
	if err != nil {
		return packages, err
	}

	var project ProjectFile
	if err := xml.Unmarshal(data, &project); err != nil {
		return packages, fmt.Errorf("error parsing project file: %w", err)
	}

	// Extract PackageReference elements
	for _, pkg := range project.PackageReferences {
		if pkg.Include != "" && pkg.Version != "" {
			packages = append(packages, PackageReference{
				Name:    pkg.Include,
				Version: pkg.Version,
				Source:  "project",
			})
		}
	}

	return packages, nil
}

func (o *ScanNugetPackagesOptions) outputResults(projectValidations []ProjectPackageValidation, solutionValidation SolutionPackageValidation) error {
	// Output per-project validation results
	fmt.Fprintln(o.Out, "\n=== Per-Project Validation ===")
	for _, pv := range projectValidations {
		status := "PASS"
		if !pv.IsValid {
			status = "FAIL"
		}
		fmt.Fprintf(o.Out, "\nProject: %s [%s]\n", pv.ProjectName, status)

		if len(pv.OrphanedConfigReferences) > 0 {
			fmt.Fprintf(o.Out, "  Orphaned packages.config references (%d):\n", len(pv.OrphanedConfigReferences))
			for _, pkg := range pv.OrphanedConfigReferences {
				fmt.Fprintf(o.Out, "    - %s %s\n", pkg.Name, pkg.Version)
			}
		}

		if len(pv.BrokenProjectReferences) > 0 {
			fmt.Fprintf(o.Out, "  Broken project references (%d):\n", len(pv.BrokenProjectReferences))
			for _, pkg := range pv.BrokenProjectReferences {
				fmt.Fprintf(o.Out, "    - %s %s\n", pkg.Name, pkg.Version)
			}
		}

		if len(pv.VersionMismatches) > 0 {
			fmt.Fprintf(o.Out, "  Version mismatches (%d):\n", len(pv.VersionMismatches))
			for _, mm := range pv.VersionMismatches {
				fmt.Fprintf(o.Out, "    - %s: packages.config=%s, project=%s\n", mm.PackageName, mm.ConfigVersion, mm.ProjectVersion)
			}
		}

		if pv.IsValid {
			fmt.Fprintf(o.Out, "  All package references are consistent.\n")
		}
	}

	// Output solution-wide validation results
	fmt.Fprintln(o.Out, "\n=== Solution-Wide Validation ===")
	if solutionValidation.IsValid {
		fmt.Fprintln(o.Out, "PASS: All packages are at consistent versions across projects.")
	} else {
		fmt.Fprintf(o.Out, "FAIL: Found %d package(s) with version conflicts:\n\n", len(solutionValidation.PackageVersionConflicts))
		for _, conflict := range solutionValidation.PackageVersionConflicts {
			fmt.Fprintf(o.Out, "Package: %s\n", conflict.PackageName)
			for version, projects := range conflict.VersionProjects {
				fmt.Fprintf(o.Out, "  Version %s used by:\n", version)
				for _, proj := range projects {
					fmt.Fprintf(o.Out, "    - %s\n", proj)
				}
			}
			fmt.Fprintln(o.Out)
		}
	}

	// Return error if validation failed
	if !solutionValidation.IsValid {
		return fmt.Errorf("solution-wide validation failed")
	}

	for _, pv := range projectValidations {
		if !pv.IsValid {
			return fmt.Errorf("per-project validation failed for one or more projects")
		}
	}

	return nil
}
