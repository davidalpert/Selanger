package dotnet

import "encoding/xml"

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

// NugetValidationResults combines all validation results for output
type NugetValidationResults struct {
	ProjectValidations []ProjectPackageValidation
	SolutionValidation SolutionPackageValidation
	OverallValid       bool
}

// PackagesConfig represents the packages.config XML structure
type PackagesConfig struct {
	XMLName  xml.Name `xml:"packages"`
	Packages []struct {
		ID      string `xml:"id,attr"`
		Version string `xml:"version,attr"`
	} `xml:"package"`
}

// ProjectFile represents the relevant parts of a .csproj XML structure
type ProjectFile struct {
	XMLName           xml.Name `xml:"Project"`
	ItemGroups        []ItemGroup
	PackageReferences []PackageReferenceXML `xml:"ItemGroup>PackageReference"`
}

// ItemGroup represents an ItemGroup element in a project file
type ItemGroup struct {
	PackageReferences []PackageReferenceXML `xml:"PackageReference"`
}

// PackageReferenceXML represents a PackageReference XML element
type PackageReferenceXML struct {
	Include string `xml:"Include,attr"`
	Version string `xml:"Version,attr"`
}
