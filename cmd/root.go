package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/nuonco/nuon-ext-ctx/internal/ctx"
	"github.com/spf13/cobra"
)

var BuildVersion = "dev"

func NewRootCmd() *cobra.Command {
	var (
		showCurrent bool
		unset       bool
		deleteMode  bool
		save        string
		showVersion bool
	)

	cmd := &cobra.Command{
		Use:   "nuon-ext-ctx [NAME | NEW_NAME=NAME | -]",
		Short: "Switch between nuon CLI configurations",
		Long: `Manage multiple nuon CLI configurations by switching ~/.nuon between
named contexts stored in ~/.config/nuon/contexts/.

Each context is a separate nuon config file. The active context is a
symlink at ~/.nuon pointing to one of these files.`,
		Example: strings.Join([]string{
			"  nuon ctx                       : list the contexts",
			"  nuon ctx <NAME>                : switch to context <NAME>",
			"  nuon ctx -                     : switch to the previous context",
			"  nuon ctx -c, --current         : show the current context name",
			"  nuon ctx <NEW_NAME>=<NAME>     : rename context <NAME> to <NEW_NAME>",
			"  nuon ctx <NEW_NAME>=.          : rename current-context to <NEW_NAME>",
			"  nuon ctx -u, --unset           : unset the current context",
			"  nuon ctx -d <NAME> [<NAME...>] : delete context(s) ('.' for current-context)",
			"  nuon ctx -s, --save <NAME>          : save current ~/.nuon as a named context",
			"  nuon ctx -s, --save <NAME> <FILE>   : save an existing config file as a named context",
		}, "\n"),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := ctx.NewManager()
			if err != nil {
				return err
			}

			if showVersion {
				fmt.Println(BuildVersion)
				return nil
			}

			if showCurrent {
				return runCurrent(m)
			}

			if unset {
				return m.Unset()
			}

			if deleteMode {
				if len(args) == 0 {
					return fmt.Errorf("usage: nuon ctx -d <NAME> [<NAME...>]")
				}
				return m.Delete(args)
			}

			if save != "" {
				var srcFile string
				if len(args) > 0 {
					srcFile = args[0]
				}
				if err := m.Save(save, srcFile); err != nil {
					return err
				}
				if srcFile != "" {
					fmt.Fprintf(os.Stderr, "Context %q saved from %s\n", save, srcFile)
				} else {
					fmt.Fprintf(os.Stderr, "Context %q saved and activated\n", save)
				}
				return nil
			}

			if len(args) == 0 {
				return runList(m)
			}

			arg := args[0]

			// Switch to previous context.
			if arg == "-" {
				name, err := m.SwitchPrevious()
				if err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "Switched to context %q\n", name)
				return nil
			}

			// Rename: NEW=OLD or NEW=.
			if strings.Contains(arg, "=") {
				parts := strings.SplitN(arg, "=", 2)
				newName, oldName := parts[0], parts[1]
				if newName == "" || oldName == "" {
					return fmt.Errorf("usage: nuon ctx <NEW_NAME>=<NAME>")
				}
				if err := m.Rename(oldName, newName); err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "Context %q renamed to %q\n", oldName, newName)
				return nil
			}

			// Switch context.
			if err := m.Switch(arg); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Switched to context %q\n", arg)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&showCurrent, "current", "c", false, "Show the current context name")
	cmd.Flags().BoolVarP(&unset, "unset", "u", false, "Unset the current context")
	cmd.Flags().BoolVarP(&deleteMode, "delete", "d", false, "Delete context(s)")
	cmd.Flags().StringVarP(&save, "save", "s", "", "Save current ~/.nuon as a named context")
	cmd.Flags().BoolVarP(&showVersion, "version", "V", false, "Print the extension version")

	return cmd
}

func runList(m *ctx.Manager) error {
	if err := m.EnsureDir(); err != nil {
		return err
	}

	names, err := m.List()
	if err != nil {
		return err
	}

	if len(names) == 0 {
		fmt.Fprintln(os.Stderr, "No contexts found. Use 'nuon ctx -s <name>' to save the current config as a context.")
		return nil
	}

	current, _ := m.Current()
	for _, name := range names {
		if name == current {
			fmt.Printf("* %s\n", name)
		} else {
			fmt.Printf("  %s\n", name)
		}
	}
	return nil
}

func runCurrent(m *ctx.Manager) error {
	current, err := m.Current()
	if err != nil {
		return err
	}
	if current == "" {
		return fmt.Errorf("no current context is set")
	}
	fmt.Println(current)
	return nil
}
