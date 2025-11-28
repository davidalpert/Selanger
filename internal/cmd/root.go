package cmd

import (
	"fmt"
	"github.com/davidalpert/go-printers/v1"
	"github.com/davidalpert/selanger/internal/diagnostics"
	"github.com/spf13/cobra"
	"os"
	"runtime"
)

// cfgFile is an optional path to a configuration file used to initialize viper
var cfgFile string

// Execute builds the default root command and invokes it with os.Args
func Execute() {
	s := printers.DefaultOSStreams()
	// configure the logger here in the outer scope so that we can defer
	// any cleanup such as writing/flushing the stream
	logCleanupFn := diagnostics.ConfigureLogger(s)
	defer logCleanupFn()

	// disable "You need to open cmd.exe and run it from there." warning when running on Windows;
	// we do not need to run this app from the console since by default now it opens a web server
	// and terminates when all the signalr connections have closed.
	if runtime.GOOS == "windows" {
		cobra.MousetrapHelpText = ""
	}

	rootCmd := NewRootCmd(s)

	rootCmd.SetArgs(os.Args[1:]) // without program

	// look for matching subcommand
	var cmdFound bool
	for _, a := range rootCmd.Commands() {
		for _, b := range os.Args[1:] {
			if a.Name() == b {
				cmdFound = true
				break
			} else {
				for _, alias := range a.Aliases {
					if alias == b {
						cmdFound = true
						break
					}
				}
			}
		}
	}
	if cmdFound == false {
		// found no matching subcommand; run the default command
		//args := append([]string{"version"}, os.Args[1:]...)
		//rootCmd.SetArgs(args)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(s.ErrOut, err)
		os.Exit(1)
	}
}

// RootCmdOptions is a struct containing options for this command
type RootCmdOptions struct {
	printers.IOStreams
	ModulesDir string
	Verbose    bool
}

// NewRootCmdOptions returns initialized command options
func NewRootCmdOptions(ioStreams printers.IOStreams) *RootCmdOptions {
	return &RootCmdOptions{
		IOStreams: ioStreams,
	}
}

// NewRootCmd creates the 'root' command and configures it's nested children
func NewRootCmd(s printers.IOStreams) *cobra.Command {
	o := NewRootCmdOptions(s)
	rootCmd := &cobra.Command{
		Use:           "selange",
		Short:         "A command-line tool for managing and investigating a drupal codebase",
		Long:          "",
		Args:          cobra.RangeArgs(0, 1),
		SilenceUsage:  true,
		SilenceErrors: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   false,
			DisableNoDescFlag:   false,
			DisableDescriptions: false,
			HiddenDefaultCmd:    true,
		},
		// Uncomment the following line if your bare application
		// has an action associated with it:
		//	Run: func(cmd *cobra.Command, args []string) { },
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(cmd, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			return o.Run()
		},
		Aliases: []string{},
	}

	// Register subcommands
	rootCmd.AddCommand(NewCmdScan(s))
	rootCmd.AddCommand(NewCmdVersion(s))

	//rootCmd.PersistentFlags().BoolP("verbose", "vv", false, "enable verbose output")

	return rootCmd
}

// Complete the options
func (o *RootCmdOptions) Complete(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		o.ModulesDir = args[0]
	}
	verbose, _ := cmd.PersistentFlags().GetBool("verbose")
	o.Verbose = verbose
	return nil
}

// Validate the options
func (o *RootCmdOptions) Validate() error {
	return nil
}

// Run the command
func (o *RootCmdOptions) Run() error {
	return NewCmdVersion(o.IOStreams).Execute()
}

func init() {
}
