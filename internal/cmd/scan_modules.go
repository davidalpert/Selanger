package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/davidalpert/go-printers/v1"
	"github.com/davidalpert/selanger/internal/dotnet"
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
	// Create scanner and get all projects
	scanner := dotnet.NewSolutionScanner(o.FS, o.SolutionFileOrFolder, o.ShowAbsolutePaths)
	allProjects, err := scanner.GetAllProjects()
	if err != nil {
		return err
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
