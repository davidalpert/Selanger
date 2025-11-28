package cmd

import (
	"github.com/davidalpert/go-printers/v1"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"path/filepath"
)

type ScanModulesOptions struct {
	*printers.PrinterOptions
	ModulesRootFolder string
	FS                afero.Fs
	ShowAbsolutePaths bool
}

func NewScanModulesOptions(s printers.IOStreams) *ScanModulesOptions {
	return &ScanModulesOptions{
		PrinterOptions: printers.NewPrinterOptions().WithStreams(s).WithDefaultTableWriter(),
	}
}

func NewCmdScanModules(s printers.IOStreams) *cobra.Command {
	o := NewScanModulesOptions(s)
	var cmd = &cobra.Command{
		Use:     "modules <module-root-path>",
		Aliases: []string{"m"},
		Short:   "scan the modules of a drupal codebase",
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
		o.ModulesRootFolder = "."
	} else if len(args) > 0 {
		o.ModulesRootFolder = args[0]
	}

	if absPath, err := filepath.Abs(o.ModulesRootFolder); err != nil {
		return err
	} else {
		o.ModulesRootFolder = absPath
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
	//opt := drupal.FileScanOptions{
	//	FS:               o.FS,
	//	RootFolder:       o.ModulesRootFolder,
	//	UseAbsolutePaths: o.ShowAbsolutePaths,
	//}
	//
	//result, err := drupal.ScanInfoFiles(opt)
	//if err != nil {
	//	return err
	//}
	//
	//return o.WithTableWriter(fmt.Sprintf("modules from %s", o.ModulesRootFolder), func(t *tablewriter.Table) {
	//	t.SetHeader([]string{"path", "name", "description", "hidden"})
	//	t.SetAutoWrapText(false)
	//	for _, mi := range result {
	//		isHidden := "false"
	//		if mi.Hidden {
	//			isHidden = "true"
	//		}
	//		t.Append([]string{mi.FilePath, mi.Name, mi.Description, isHidden})
	//	}
	//
	//}).WriteOutput(result)
	return nil
}
