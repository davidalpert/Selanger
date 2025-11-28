package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apex/log"
	"github.com/davidalpert/go-printers/v1"
	"github.com/davidalpert/selanger/internal/diagnostics"
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
		t.SetHeader([]string{"solution", "aligned", "project name", "folder name", "file name", "project file"})
		t.SetAutoWrapText(false)
		for _, proj := range allProjects {
			aligned := "false"
			if proj.IsAligned() {
				aligned = "true"
			}
			t.Append([]string{proj.SolutionFile, aligned, proj.Name, proj.ParentFolderName, proj.FileNameWithoutExt, proj.Path})
		}
	}).WriteOutput(allProjects)
}

// ProjectReference represents a project reference in a solution file
type ProjectReference struct {
	SolutionFile       string
	Name               string
	Path               string // full path to the project file
	ProjectFolder      string // folder path to the directory containing the project file (deprecated, use ParentFolderName)
	ParentFolderName   string // name of the immediate parent directory containing the project file
	FileNameWithoutExt string // project file name without extension
	ProjectID          string // project GUID
}

// IsAligned returns true if ProjectName, FolderName, and FileNameWithoutExt are all the same
func (p *ProjectReference) IsAligned() bool {
	return p.Name == p.ParentFolderName && p.Name == p.FileNameWithoutExt
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
	// Format: Project("{TYPE-GUID}") = "ProjectName", "Path\To\Project.csproj", "{PROJECT-GUID}"
	projectRegex := regexp.MustCompile(`Project\("\{([^}]+)\}"\)\s*=\s*"([^"]+)"\s*,\s*"([^"]+)"\s*,\s*"\{([^}]+)\}"`)

	// Solution folder type GUID - we want to filter these out
	const solutionFolderTypeGUID = "2150E333-8FDC-42A3-9474-1A3956D46DE8"

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := projectRegex.FindStringSubmatch(line)
		if len(matches) == 5 {
			projectTypeGUID := strings.ToUpper(matches[1])
			projectName := matches[2]
			projectPath := matches[3]
			projectID := strings.ToUpper(matches[4])

			// Skip solution folders
			if projectTypeGUID == solutionFolderTypeGUID {
				diagnostics.Log.WithFields(log.Fields{
					"line":        line,
					"projectName": projectName,
					"projectType": projectTypeGUID,
				}).Debug("skipping solution folder")
				continue
			}
			diagnostics.Log.WithFields(log.Fields{
				"line":        line,
				"projectName": projectName,
				"projectPath": projectPath,
			}).Debug("parsed project reference")

			// .sln files always use backslashes, convert to OS-specific separator
			projectPath = strings.ReplaceAll(projectPath, "\\", string(filepath.Separator))
			diagnostics.Log.WithFields(log.Fields{
				"line":        line,
				"projectName": projectName,
				"projectPath": projectPath,
			}).Debug("parsed project reference: normalized path")

			// Extract path components using filepath package (uses OS-specific separators)
			projectFileName := filepath.Base(projectPath)
			projectFolder := filepath.Dir(projectPath)
			parentFolderName := filepath.Base(projectFolder)

			// Extract filename without extension
			fileNameWithoutExt := strings.TrimSuffix(projectFileName, filepath.Ext(projectFileName))

			diagnostics.Log.WithFields(log.Fields{
				"line":                         line,
				"projectName":                  projectName,
				"projectPath":                  projectPath,
				"projectFolder":                projectFolder,
				"filepath.Base(projectPath)":   projectFileName,
				"filepath.Dir(projectPath)":    projectFolder,
				"filepath.Base(projectFolder)": parentFolderName,
				"fileNameWithoutExt":           fileNameWithoutExt,
			}).Debug("parsed project reference: extracted components")

			// Determine the path to display
			displayPath := projectPath
			displayFolder := projectFolder
			if o.ShowAbsolutePaths {
				slnDir := filepath.Dir(fullPath)
				// Join paths using OS-specific separators
				displayPath = filepath.Join(slnDir, projectPath)
				displayFolder = filepath.Join(slnDir, projectFolder)
			}

			projects = append(projects, ProjectReference{
				SolutionFile:       slnPath,
				Name:               projectName,
				Path:               displayPath,
				ProjectFolder:      displayFolder,
				ParentFolderName:   parentFolderName,
				FileNameWithoutExt: fileNameWithoutExt,
				ProjectID:          projectID,
			})
		}
	}

	return projects, scanner.Err()
}
