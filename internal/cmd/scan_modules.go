package cmd

import (
	"bufio"
	"fmt"
	"github.com/davidalpert/selanger/internal/diagnostics"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/davidalpert/go-printers/v1"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type ScanModulesOptions struct {
	*printers.PrinterOptions
	SolutionFileOrFolder string
	FS                   afero.Fs
	ShowAbsolutePaths    bool
}

func NewScanModulesOptions(s printers.IOStreams) *ScanModulesOptions {
	return &ScanModulesOptions{
		PrinterOptions: printers.NewPrinterOptions().WithStreams(s).WithDefaultTableWriter(),
	}
}

func NewCmdScanModules(s printers.IOStreams) *cobra.Command {
	o := NewScanModulesOptions(s)
	var cmd = &cobra.Command{
		Use:     "solution <path to .sln file or a containing folder>",
		Aliases: []string{"s", "sln"},
		Short:   "scan the projects in a .NET solution",
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
func (o *ScanModulesOptions) Complete(cmd *cobra.Command, args []string) error {
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
func (o *ScanModulesOptions) Validate() error {
	return o.PrinterOptions.Validate()
}

// Run the command
func (o *ScanModulesOptions) Run() error {
	// Find .sln files in the root folder
	slnFiles, err := o.findSolutionFiles()
	if err != nil {
		return fmt.Errorf("error finding solution files: %w", err)
	}

	if len(slnFiles) == 0 {
		return fmt.Errorf("no .sln files found in %s", o.SolutionFileOrFolder)
	}

	// Parse all solution files and collect projects
	var allProjects []ProjectReference
	for _, slnFile := range slnFiles {
		diagnostics.Log.WithField("solutionFile", slnFile).Debug("parsing solution file")
		projects, err := o.parseSolutionFile(slnFile)
		if err != nil {
			return fmt.Errorf("error parsing %s: %w", slnFile, err)
		}
		allProjects = append(allProjects, projects...)
	}

	// Output results
	return o.WithTableWriter(fmt.Sprintf("projects from %s", o.SolutionFileOrFolder), func(t *tablewriter.Table) {
		t.SetHeader([]string{"solution", "project name", "project path"})
		t.SetAutoWrapText(false)
		for _, proj := range allProjects {
			t.Append([]string{proj.SolutionFile, proj.Name, proj.Path})
		}
	}).WriteOutput(allProjects)
}

// ProjectReference represents a project reference in a solution file
type ProjectReference struct {
	SolutionFile string
	Name         string
	Path         string
}

// findSolutionFiles finds all .sln files - handles both a direct .sln file path or a folder
func (o *ScanModulesOptions) findSolutionFiles() ([]string, error) {
	var slnFiles []string

	// Check if the path is a file or directory
	info, err := o.FS.Stat(o.SolutionFileOrFolder)
	if err != nil {
		return nil, fmt.Errorf("cannot access path: %w", err)
	}

	// If it's a file, verify it's a .sln file and return it
	if !info.IsDir() {
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".sln") {
			return nil, fmt.Errorf("file %s is not a .sln file", o.SolutionFileOrFolder)
		}

		// Return the file path (absolute or relative based on flag)
		if o.ShowAbsolutePaths {
			slnFiles = append(slnFiles, o.SolutionFileOrFolder)
		} else {
			// For a direct file, use just the filename
			slnFiles = append(slnFiles, filepath.Base(o.SolutionFileOrFolder))
		}
		return slnFiles, nil
	}

	// If it's a directory, walk it to find all .sln files
	err = afero.Walk(o.FS, o.SolutionFileOrFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".sln") {
			if o.ShowAbsolutePaths {
				slnFiles = append(slnFiles, path)
			} else {
				relPath, err := filepath.Rel(o.SolutionFileOrFolder, path)
				if err != nil {
					return err
				}
				slnFiles = append(slnFiles, relPath)
			}
		}
		return nil
	})

	return slnFiles, err
}

// parseSolutionFile parses a .sln file and extracts project references
func (o *ScanModulesOptions) parseSolutionFile(slnPath string) ([]ProjectReference, error) {
	var projects []ProjectReference

	// Resolve the full path for reading
	fullPath := slnPath
	if !filepath.IsAbs(slnPath) {
		if filepath.Ext(slnPath) == ".sln" {
			fullPath = filepath.Join(filepath.Dir(o.SolutionFileOrFolder), slnPath)
		} else {
			fullPath = filepath.Join(o.SolutionFileOrFolder, slnPath)
		}
	}

	file, err := o.FS.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Regular expression to match Project lines in .sln files
	// Format: Project("{GUID}") = "ProjectName", "Path\To\Project.csproj", "{GUID}"
	projectRegex := regexp.MustCompile(`Project\("\{[^}]+\}"\)\s*=\s*"([^"]+)"\s*,\s*"([^"]+)"`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := projectRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			projectName := matches[1]
			projectPath := matches[2]

			// Convert backslashes to forward slashes for cross-platform compatibility
			projectPath = filepath.ToSlash(projectPath)

			// Determine the path to display
			displayPath := projectPath
			if o.ShowAbsolutePaths {
				slnDir := filepath.Dir(fullPath)
				displayPath = filepath.Join(slnDir, projectPath)
			}

			projects = append(projects, ProjectReference{
				SolutionFile: slnPath,
				Name:         projectName,
				Path:         displayPath,
			})
		}
	}

	return projects, scanner.Err()
}
