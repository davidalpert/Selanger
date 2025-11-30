package cmd

import (
	"github.com/davidalpert/selanger/internal/dotnet"
	"path/filepath"
	"testing"

	"github.com/davidalpert/go-printers/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadPackagesConfig(t *testing.T) {
	packagesConfigContent := `<?xml version="1.0" encoding="utf-8"?>
<packages>
  <package id="Newtonsoft.Json" version="12.0.3" targetFramework="net472" />
  <package id="Microsoft.AspNet.WebApi.Client" version="5.2.7" targetFramework="net472" />
</packages>`

	fs := afero.NewMemMapFs()
	configPath := "/project/packages.config"
	err := afero.WriteFile(fs, configPath, []byte(packagesConfigContent), 0644)
	require.NoError(t, err)

	o := &ScanNugetPackagesOptions{
		FS: fs,
	}

	packages, err := o.readPackagesConfig(configPath)
	require.NoError(t, err)
	assert.Len(t, packages, 2)
	assert.Equal(t, "Newtonsoft.Json", packages[0].Name)
	assert.Equal(t, "12.0.3", packages[0].Version)
	assert.Equal(t, "packages.config", packages[0].Source)
	assert.Equal(t, "Microsoft.AspNet.WebApi.Client", packages[1].Name)
	assert.Equal(t, "5.2.7", packages[1].Version)
	assert.Equal(t, "packages.config", packages[1].Source)
}

func TestReadProjectPackages(t *testing.T) {
	projectContent := `<?xml version="1.0" encoding="utf-8"?>
<Project ToolsVersion="15.0" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">
  <ItemGroup>
    <PackageReference Include="Newtonsoft.Json" Version="12.0.3" />
    <PackageReference Include="Microsoft.AspNet.WebApi.Client" Version="5.2.7" />
  </ItemGroup>
</Project>`

	fs := afero.NewMemMapFs()
	projectPath := "/project/MyProject.csproj"
	err := afero.WriteFile(fs, projectPath, []byte(projectContent), 0644)
	require.NoError(t, err)

	o := &ScanNugetPackagesOptions{
		PrinterOptions:       printers.NewPrinterOptions().WithDefaultTableWriter(),
		SolutionFileOrFolder: "/",
		FS:                   fs,
	}

	packages, err := o.readProjectPackages(projectPath)
	require.NoError(t, err)
	assert.Len(t, packages, 2)
	assert.Equal(t, "Newtonsoft.Json", packages[0].Name)
	assert.Equal(t, "12.0.3", packages[0].Version)
	assert.Equal(t, "project", packages[0].Source)

	assert.Equal(t, "Microsoft.AspNet.WebApi.Client", packages[1].Name)
	assert.Equal(t, "5.2.7", packages[1].Version)
	assert.Equal(t, "project", packages[1].Source)
}

func TestValidateProjectPackages_AllAligned(t *testing.T) {
	packagesConfigContent := `<?xml version="1.0" encoding="utf-8"?>
<packages>
  <package id="Newtonsoft.Json" version="12.0.3" targetFramework="net472" />
</packages>`

	projectContent := `<?xml version="1.0" encoding="utf-8"?>
<Project ToolsVersion="15.0" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">
  <ItemGroup>
    <PackageReference Include="Newtonsoft.Json" Version="12.0.3" />
  </ItemGroup>
</Project>`

	fs := afero.NewMemMapFs()
	projectDir := "/solution/MyProject"
	projectPath := filepath.Join(projectDir, "MyProject.csproj")
	configPath := filepath.Join(projectDir, "packages.config")

	err := afero.WriteFile(fs, projectPath, []byte(projectContent), 0644)
	require.NoError(t, err)
	err = afero.WriteFile(fs, configPath, []byte(packagesConfigContent), 0644)
	require.NoError(t, err)

	o := &ScanNugetPackagesOptions{
		PrinterOptions:       printers.NewPrinterOptions().WithDefaultTableWriter(),
		SolutionFileOrFolder: "/solution",
		FS:                   fs,
	}

	proj := dotnet.ProjectReference{
		Name: "MyProject",
		Path: "MyProject/MyProject.csproj",
	}

	validation, err := o.validateProjectPackages(proj)
	require.NoError(t, err)
	assert.True(t, validation.IsValid)
	assert.Len(t, validation.OrphanedConfigReferences, 0)
	assert.Len(t, validation.BrokenProjectReferences, 0)
	assert.Len(t, validation.VersionMismatches, 0)
}

func TestValidateProjectPackages_OrphanedConfigReference(t *testing.T) {
	packagesConfigContent := `<?xml version="1.0" encoding="utf-8"?>
<packages>
  <package id="Newtonsoft.Json" version="12.0.3" targetFramework="net472" />
  <package id="Obsolete.Package" version="1.0.0" targetFramework="net472" />
</packages>`

	projectContent := `<?xml version="1.0" encoding="utf-8"?>
<Project ToolsVersion="15.0" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">
  <ItemGroup>
    <PackageReference Include="Newtonsoft.Json" Version="12.0.3" />
  </ItemGroup>
</Project>`

	fs := afero.NewMemMapFs()
	projectDir := "/solution/MyProject"
	projectPath := filepath.Join(projectDir, "MyProject.csproj")
	configPath := filepath.Join(projectDir, "packages.config")

	err := afero.WriteFile(fs, projectPath, []byte(projectContent), 0644)
	require.NoError(t, err)
	err = afero.WriteFile(fs, configPath, []byte(packagesConfigContent), 0644)
	require.NoError(t, err)

	o := &ScanNugetPackagesOptions{
		PrinterOptions:       printers.NewPrinterOptions().WithDefaultTableWriter(),
		SolutionFileOrFolder: "/solution",
		FS:                   fs,
	}

	proj := dotnet.ProjectReference{
		Name: "MyProject",
		Path: "MyProject/MyProject.csproj",
	}

	validation, err := o.validateProjectPackages(proj)
	require.NoError(t, err)
	assert.False(t, validation.IsValid)
	assert.Len(t, validation.OrphanedConfigReferences, 1)
	assert.Equal(t, "Obsolete.Package", validation.OrphanedConfigReferences[0].Name)
	assert.Len(t, validation.BrokenProjectReferences, 0)
	assert.Len(t, validation.VersionMismatches, 0)
}

func TestValidateProjectPackages_BrokenProjectReference(t *testing.T) {
	packagesConfigContent := `<?xml version="1.0" encoding="utf-8"?>
<packages>
  <package id="Newtonsoft.Json" version="12.0.3" targetFramework="net472" />
</packages>`

	projectContent := `<?xml version="1.0" encoding="utf-8"?>
<Project ToolsVersion="15.0" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">
  <ItemGroup>
    <PackageReference Include="Newtonsoft.Json" Version="12.0.3" />
    <PackageReference Include="Missing.Package" Version="2.0.0" />
  </ItemGroup>
</Project>`

	fs := afero.NewMemMapFs()
	projectDir := "/solution/MyProject"
	projectPath := filepath.Join(projectDir, "MyProject.csproj")
	configPath := filepath.Join(projectDir, "packages.config")

	err := afero.WriteFile(fs, projectPath, []byte(projectContent), 0644)
	require.NoError(t, err)
	err = afero.WriteFile(fs, configPath, []byte(packagesConfigContent), 0644)
	require.NoError(t, err)

	o := &ScanNugetPackagesOptions{
		PrinterOptions:       printers.NewPrinterOptions().WithDefaultTableWriter(),
		SolutionFileOrFolder: "/solution",
		FS:                   fs,
	}

	proj := dotnet.ProjectReference{
		Name: "MyProject",
		Path: "MyProject/MyProject.csproj",
	}

	validation, err := o.validateProjectPackages(proj)
	require.NoError(t, err)
	assert.False(t, validation.IsValid)
	assert.Len(t, validation.OrphanedConfigReferences, 0)
	assert.Len(t, validation.BrokenProjectReferences, 1)
	assert.Equal(t, "Missing.Package", validation.BrokenProjectReferences[0].Name)
	assert.Len(t, validation.VersionMismatches, 0)
}

func TestValidateProjectPackages_VersionMismatch(t *testing.T) {
	packagesConfigContent := `<?xml version="1.0" encoding="utf-8"?>
<packages>
  <package id="Newtonsoft.Json" version="12.0.3" targetFramework="net472" />
</packages>`

	projectContent := `<?xml version="1.0" encoding="utf-8"?>
<Project ToolsVersion="15.0" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">
  <ItemGroup>
    <PackageReference Include="Newtonsoft.Json" Version="13.0.1" />
  </ItemGroup>
</Project>`

	fs := afero.NewMemMapFs()
	projectDir := "/solution/MyProject"
	projectPath := filepath.Join(projectDir, "MyProject.csproj")
	configPath := filepath.Join(projectDir, "packages.config")

	err := afero.WriteFile(fs, projectPath, []byte(projectContent), 0644)
	require.NoError(t, err)
	err = afero.WriteFile(fs, configPath, []byte(packagesConfigContent), 0644)
	require.NoError(t, err)

	o := &ScanNugetPackagesOptions{
		PrinterOptions:       printers.NewPrinterOptions().WithDefaultTableWriter(),
		SolutionFileOrFolder: "/solution",
		FS:                   fs,
	}

	proj := dotnet.ProjectReference{
		Name: "MyProject",
		Path: "MyProject/MyProject.csproj",
	}

	validation, err := o.validateProjectPackages(proj)
	require.NoError(t, err)
	assert.False(t, validation.IsValid)
	assert.Len(t, validation.OrphanedConfigReferences, 0)
	assert.Len(t, validation.BrokenProjectReferences, 0)
	assert.Len(t, validation.VersionMismatches, 1)
	assert.Equal(t, "Newtonsoft.Json", validation.VersionMismatches[0].PackageName)
	assert.Equal(t, "12.0.3", validation.VersionMismatches[0].ConfigVersion)
	assert.Equal(t, "13.0.1", validation.VersionMismatches[0].ProjectVersion)
}

func TestValidateSolutionPackages_NoConflicts(t *testing.T) {
	allPackages := map[string]map[string][]string{
		"Newtonsoft.Json": {
			"12.0.3": []string{"Project1", "Project2"},
		},
		"Microsoft.AspNet.WebApi.Client": {
			"5.2.7": []string{"Project1"},
		},
	}

	o := &ScanNugetPackagesOptions{}
	validation := o.validateSolutionPackages(allPackages)

	assert.True(t, validation.IsValid)
	assert.Len(t, validation.PackageVersionConflicts, 0)
}

func TestValidateSolutionPackages_WithConflicts(t *testing.T) {
	allPackages := map[string]map[string][]string{
		"Newtonsoft.Json": {
			"12.0.3": []string{"Project1"},
			"13.0.1": []string{"Project2", "Project3"},
		},
		"Microsoft.AspNet.WebApi.Client": {
			"5.2.7": []string{"Project1"},
		},
	}

	o := &ScanNugetPackagesOptions{}
	validation := o.validateSolutionPackages(allPackages)

	assert.False(t, validation.IsValid)
	assert.Len(t, validation.PackageVersionConflicts, 1)
	assert.Equal(t, "Newtonsoft.Json", validation.PackageVersionConflicts[0].PackageName)
	assert.Len(t, validation.PackageVersionConflicts[0].VersionProjects, 2)
	assert.Contains(t, validation.PackageVersionConflicts[0].VersionProjects["12.0.3"], "Project1")
	assert.Contains(t, validation.PackageVersionConflicts[0].VersionProjects["13.0.1"], "Project2")
	assert.Contains(t, validation.PackageVersionConflicts[0].VersionProjects["13.0.1"], "Project3")
}

func TestValidateProjectPackages_WithSlnFilePath(t *testing.T) {
	// Test that path resolution works correctly when SolutionFileOrFolder is a .sln file
	packagesConfigContent := `<?xml version="1.0" encoding="utf-8"?>
<packages>
  <package id="Newtonsoft.Json" version="12.0.3" targetFramework="net472" />
</packages>`

	projectContent := `<?xml version="1.0" encoding="utf-8"?>
<Project ToolsVersion="15.0" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">
  <ItemGroup>
    <PackageReference Include="Newtonsoft.Json" Version="12.0.3" />
  </ItemGroup>
</Project>`

	fs := afero.NewMemMapFs()
	slnPath := "/solution/MySolution.sln"
	projectDir := "/solution/MyProject"
	projectPath := filepath.Join(projectDir, "MyProject.csproj")
	configPath := filepath.Join(projectDir, "packages.config")

	// Create solution file (content doesn't matter for this test)
	err := afero.WriteFile(fs, slnPath, []byte(""), 0644)
	require.NoError(t, err)

	err = afero.WriteFile(fs, projectPath, []byte(projectContent), 0644)
	require.NoError(t, err)
	err = afero.WriteFile(fs, configPath, []byte(packagesConfigContent), 0644)
	require.NoError(t, err)

	o := &ScanNugetPackagesOptions{
		PrinterOptions:       printers.NewPrinterOptions().WithDefaultTableWriter(),
		SolutionFileOrFolder: slnPath, // This is a .sln file, not a folder
		FS:                   fs,
	}

	proj := dotnet.ProjectReference{
		Name: "MyProject",
		Path: "MyProject/MyProject.csproj",
	}

	validation, err := o.validateProjectPackages(proj)
	require.NoError(t, err)
	assert.True(t, validation.IsValid)
	assert.Len(t, validation.OrphanedConfigReferences, 0)
	assert.Len(t, validation.BrokenProjectReferences, 0)
	assert.Len(t, validation.VersionMismatches, 0)
	assert.Len(t, validation.AllPackages, 1)
}
