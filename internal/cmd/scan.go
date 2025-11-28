package cmd

import (
	"github.com/davidalpert/go-printers/v1"
	"github.com/spf13/cobra"
)

type ScanOptions struct {
}

func NewScanOptions(s printers.IOStreams) *ScanOptions {
	return &ScanOptions{}
}

func NewCmdScan(s printers.IOStreams) *cobra.Command {
	//o := NewScanOptions(s)
	var cmd = &cobra.Command{
		Use:     "scan",
		Short:   "scan subcommands",
		Aliases: []string{"s"},
		Args:    cobra.NoArgs,
		//RunE: func(cmd *cobra.Command, args []string) error {
		//	if err := o.Complete(cmd, args); err != nil {
		//		return err
		//	}
		//	if err := o.Validate(); err != nil {
		//		return err
		//	}
		//	return o.Run()
		//},
	}

	cmd.AddCommand(NewCmdScanModules(s))
	cmd.AddCommand(NewCmdScanNugetPackages(s))

	return cmd
}
