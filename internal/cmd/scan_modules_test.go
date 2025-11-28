package cmd

import (
	"path/filepath"
	"testing"

	"github.com/davidalpert/go-printers/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSolutionFile(t *testing.T) {
	tests := []struct {
		name              string
		slnContent        string
		slnPath           string
		showAbsolutePaths bool
		expected          []ProjectReference
		expectError       bool
	}{
		{
			name:    "parse solution with multiple projects",
			slnPath: "TestSolution.sln",
			slnContent: `Microsoft Visual Studio Solution File, Format Version 12.00
# Visual Studio Version 17
VisualStudioVersion = 17.0.31903.59
MinimumVisualStudioVersion = 10.0.40219.1
Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "MyApp", "src\MyApp\MyApp.csproj", "{12345678-1234-1234-1234-123456789012}"
EndProject
Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "MyApp.Core", "src\MyApp.Core\MyApp.Core.csproj", "{87654321-4321-4321-4321-210987654321}"
EndProject
Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "MyApp.Tests", "test\MyApp.Tests\MyApp.Tests.csproj", "{ABCDEF12-ABCD-ABCD-ABCD-123456ABCDEF}"
EndProject
Global
	GlobalSection(SolutionConfigurationPlatforms) = preSolution
		Debug|Any CPU = Debug|Any CPU
		Release|Any CPU = Release|Any CPU
	EndGlobalSection
EndGlobal
`,
			showAbsolutePaths: false,
			expected: []ProjectReference{
				{
					SolutionFile:       "TestSolution.sln",
					Name:               "MyApp",
					Path:               filepath.FromSlash("src/MyApp/MyApp.csproj"),
					ProjectFolder:      filepath.FromSlash("src/MyApp"),
					ParentFolderName:   "MyApp",
					FileNameWithoutExt: "MyApp",
					ProjectID:          "12345678-1234-1234-1234-123456789012",
				},
				{
					SolutionFile:       "TestSolution.sln",
					Name:               "MyApp.Core",
					Path:               filepath.FromSlash("src/MyApp.Core/MyApp.Core.csproj"),
					ProjectFolder:      filepath.FromSlash("src/MyApp.Core"),
					ParentFolderName:   "MyApp.Core",
					FileNameWithoutExt: "MyApp.Core",
					ProjectID:          "87654321-4321-4321-4321-210987654321",
				},
				{
					SolutionFile:       "TestSolution.sln",
					Name:               "MyApp.Tests",
					Path:               filepath.FromSlash("test/MyApp.Tests/MyApp.Tests.csproj"),
					ProjectFolder:      filepath.FromSlash("test/MyApp.Tests"),
					ParentFolderName:   "MyApp.Tests",
					FileNameWithoutExt: "MyApp.Tests",
					ProjectID:          "ABCDEF12-ABCD-ABCD-ABCD-123456ABCDEF",
				},
			},
			expectError: false,
		},
		{
			name:    "parse solution with nested folder structure",
			slnPath: "NestedSolution.sln",
			slnContent: `Microsoft Visual Studio Solution File, Format Version 12.00
Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "DeepProject", "level1\level2\level3\DeepProject\DeepProject.csproj", "{11111111-1111-1111-1111-111111111111}"
EndProject
`,
			showAbsolutePaths: false,
			expected: []ProjectReference{
				{
					SolutionFile:       "NestedSolution.sln",
					Name:               "DeepProject",
					Path:               filepath.FromSlash("level1/level2/level3/DeepProject/DeepProject.csproj"),
					ProjectFolder:      filepath.FromSlash("level1/level2/level3/DeepProject"),
					ParentFolderName:   "DeepProject",
					FileNameWithoutExt: "DeepProject",
					ProjectID:          "11111111-1111-1111-1111-111111111111",
				},
			},
			expectError: false,
		},
		{
			name:    "parse solution with single project in root",
			slnPath: "RootProject.sln",
			slnContent: `Microsoft Visual Studio Solution File, Format Version 12.00
Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "RootApp", "RootApp.csproj", "{22222222-2222-2222-2222-222222222222}"
EndProject
`,
			showAbsolutePaths: false,
			expected: []ProjectReference{
				{
					SolutionFile:       "RootProject.sln",
					Name:               "RootApp",
					Path:               "RootApp.csproj",
					ProjectFolder:      ".",
					ParentFolderName:   ".",
					FileNameWithoutExt: "RootApp",
					ProjectID:          "22222222-2222-2222-2222-222222222222",
				},
			},
			expectError: false,
		},
		{
			name:    "filter out solution folders",
			slnPath: "WithFolders.sln",
			slnContent: `Microsoft Visual Studio Solution File, Format Version 12.00
Project("{2150E333-8FDC-42A3-9474-1A3956D46DE8}") = "SolutionItems", "SolutionItems", "{AAAAAAAA-AAAA-AAAA-AAAA-AAAAAAAAAAAA}"
EndProject
Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "RealProject", "RealProject\RealProject.csproj", "{BBBBBBBB-BBBB-BBBB-BBBB-BBBBBBBBBBBB}"
EndProject
Project("{2150E333-8FDC-42A3-9474-1A3956D46DE8}") = "Tests", "Tests", "{CCCCCCCC-CCCC-CCCC-CCCC-CCCCCCCCCCCC}"
EndProject
`,
			showAbsolutePaths: false,
			expected: []ProjectReference{
				{
					SolutionFile:       "WithFolders.sln",
					Name:               "RealProject",
					Path:               filepath.FromSlash("RealProject/RealProject.csproj"),
					ProjectFolder:      "RealProject",
					ParentFolderName:   "RealProject",
					FileNameWithoutExt: "RealProject",
					ProjectID:          "BBBBBBBB-BBBB-BBBB-BBBB-BBBBBBBBBBBB",
				},
			},
			expectError: false,
		},
		{
			name:              "parse empty solution file",
			slnPath:           "Empty.sln",
			slnContent:        `Microsoft Visual Studio Solution File, Format Version 12.00`,
			showAbsolutePaths: false,
			expected:          []ProjectReference{},
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up in-memory filesystem
			fs := afero.NewMemMapFs()

			// Create the solution file
			err := afero.WriteFile(fs, tt.slnPath, []byte(tt.slnContent), 0644)
			require.NoError(t, err, "failed to create test solution file")

			// Create the options
			o := &ScanModulesOptions{
				PrinterOptions:       printers.NewPrinterOptions().WithDefaultTableWriter(),
				SolutionFileOrFolder: tt.slnPath,
				FS:                   fs,
				ShowAbsolutePaths:    tt.showAbsolutePaths,
			}

			// Parse the solution file
			result, err := o.parseSolutionFile(tt.slnPath)

			// Check error expectations
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, len(tt.expected), len(result), "number of projects should match")

			// Compare each project reference
			for i, expected := range tt.expected {
				assert.Equal(t, expected.SolutionFile, result[i].SolutionFile, "solution file should match for project %d", i)
				assert.Equal(t, expected.Name, result[i].Name, "project name should match for project %d", i)
				assert.Equal(t, expected.Path, result[i].Path, "project path should match for project %d", i)
				assert.Equal(t, expected.ProjectFolder, result[i].ProjectFolder, "project folder should match for project %d", i)
				assert.Equal(t, expected.ParentFolderName, result[i].ParentFolderName, "parent folder name should match for project %d", i)
				assert.Equal(t, expected.FileNameWithoutExt, result[i].FileNameWithoutExt, "file name without extension should match for project %d", i)
				assert.Equal(t, expected.ProjectID, result[i].ProjectID, "project ID should match for project %d", i)
			}
		})
	}
}

func TestParseSolutionFileWithAbsolutePaths(t *testing.T) {
	slnContent := `Microsoft Visual Studio Solution File, Format Version 12.00
Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "MyApp", "src\MyApp\MyApp.csproj", "{12345678-1234-1234-1234-123456789012}"
EndProject
`

	// Set up in-memory filesystem
	fs := afero.NewMemMapFs()
	slnPath := "/projects/TestSolution.sln"

	// Create the solution file
	err := afero.WriteFile(fs, slnPath, []byte(slnContent), 0644)
	require.NoError(t, err)

	// Create the options with absolute paths enabled
	o := &ScanModulesOptions{
		PrinterOptions:       printers.NewPrinterOptions().WithDefaultTableWriter(),
		SolutionFileOrFolder: slnPath,
		FS:                   fs,
		ShowAbsolutePaths:    true,
	}

	// Parse the solution file
	result, err := o.parseSolutionFile(slnPath)

	require.NoError(t, err)
	require.Len(t, result, 1)

	// With absolute paths, the paths should be resolved relative to the solution directory
	assert.Equal(t, slnPath, result[0].SolutionFile)
	assert.Equal(t, "MyApp", result[0].Name)
	assert.Equal(t, filepath.Join("/projects", "src", "MyApp", "MyApp.csproj"), result[0].Path)
	assert.Equal(t, filepath.Join("/projects", "src", "MyApp"), result[0].ProjectFolder)
	assert.Equal(t, "MyApp", result[0].ParentFolderName)
	assert.Equal(t, "MyApp", result[0].FileNameWithoutExt)
	assert.Equal(t, "12345678-1234-1234-1234-123456789012", result[0].ProjectID)
}
