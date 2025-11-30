package dotnet

// ProjectReference represents a project reference in a solution file
type ProjectReference struct {
	SolutionFile       string // path to the solution file containing this reference
	Name               string // from the reference in the solution file
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
