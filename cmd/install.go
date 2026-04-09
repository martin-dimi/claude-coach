package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newInstallCmd() *cobra.Command {
	var claudeDir string

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Set up coach hook in Claude Code",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := claudeSettingsPath(claudeDir)
			coachBin := resolveCoachBin()

			settings := readJSON(path)
			hooks := ensureMap(settings, "hooks")

			hookCmd := fmt.Sprintf("%s check", coachBin)
			if hasHook(hooks, "UserPromptSubmit", hookCmd) {
				fmt.Println("  Coach is already installed.")
				return nil
			}

			addHook(hooks, "UserPromptSubmit", hookCmd)
			settings["hooks"] = hooks

			if err := writeJSON(path, settings); err != nil {
				return err
			}

			fmt.Println("  Coach installed!")
			fmt.Printf("  Hook added to %s\n", path)
			return nil
		},
	}

	cmd.Flags().StringVar(&claudeDir, "claude-dir", "", "Path to Claude config directory (default ~/.claude)")
	return cmd
}
