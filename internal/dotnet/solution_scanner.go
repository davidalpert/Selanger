package dotnet

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apex/log"
	"github.com/davidalpert/selanger/internal/diagnostics"
	"github.com/spf13/afero"
)

// SolutionScanner provides functionality to scan .NET solution files
type SolutionScanner struct {
	FS                   afero.Fs
	SolutionFileOrFolder string
	ShowAbsolutePaths    bool
}

// NewSolutionScanner creates a new SolutionScanner
func NewSolutionScanner(fs afero.Fs, solutionFileOrFolder string, showAbsolutePaths bool) *SolutionScanner {
	return &SolutionScanner{
		FS:                   fs,
		SolutionFileOrFolder: solutionFileOrFolder,
		ShowAbsolutePaths:    showAbsolutePaths,
	}
}

// FindSolutionFiles finds all .sln files - handles both a direct .sln file path or a folder
func (s *SolutionScanner) FindSolutionFiles() ([]string, error) {
	var slnFiles []string

	// Check if the path is a file or directory
	info, err := s.FS.Stat(s.SolutionFileOrFolder)
	if err != nil {
		return nil, fmt.Errorf("cannot access path: %w", err)
	}

	// If it's a file, verify it's a .sln file and return it
	if !info.IsDir() {
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".sln") {
			return nil, fmt.Errorf("file %s is not a .sln file", s.SolutionFileOrFolder)
		}

		// Return the file path (absolute or relative based on flag)
		if s.ShowAbsolutePaths {
			slnFiles = append(slnFiles, s.SolutionFileOrFolder)
		} else {
			// For a direct file, use just the filename
			slnFiles = append(slnFiles, filepath.Base(s.SolutionFileOrFolder))
		}
		return slnFiles, nil
	}

	// If it's a directory, walk it to find all .sln files
	err = afero.Walk(s.FS, s.SolutionFileOrFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".sln") {
			if s.ShowAbsolutePaths {
				slnFiles = append(slnFiles, path)
			} else {
				relPath, err := filepath.Rel(s.SolutionFileOrFolder, path)
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

// ParseSolutionFile parses a .sln file and extracts project references
func (s *SolutionScanner) ParseSolutionFile(slnPath string) ([]ProjectReference, error) {
	var projects []ProjectReference

	// Resolve the full path for reading
	fullPath := slnPath
	if !filepath.IsAbs(slnPath) {
		if filepath.Ext(slnPath) == ".sln" {
			fullPath = filepath.Join(filepath.Dir(s.SolutionFileOrFolder), slnPath)
		} else {
			fullPath = filepath.Join(s.SolutionFileOrFolder, slnPath)
		}
	}

	file, err := s.FS.Open(fullPath)
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
			if s.ShowAbsolutePaths {
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

// GetAllProjects finds all solution files and parses them to get all project references
func (s *SolutionScanner) GetAllProjects() ([]ProjectReference, error) {
	slnFiles, err := s.FindSolutionFiles()
	if err != nil {
		return nil, fmt.Errorf("error finding solution files: %w", err)
	}

	if len(slnFiles) == 0 {
		return nil, fmt.Errorf("no .sln files found in %s", s.SolutionFileOrFolder)
	}

	var allProjects []ProjectReference
	for _, slnFile := range slnFiles {
		diagnostics.Log.WithField("solutionFile", slnFile).Debug("parsing solution file")
		projects, err := s.ParseSolutionFile(slnFile)
		if err != nil {
			return nil, fmt.Errorf("error parsing %s: %w", slnFile, err)
		}
		allProjects = append(allProjects, projects...)
	}

	return allProjects, nil
}
