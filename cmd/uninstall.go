package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newUninstallCmd() *cobra.Command {
	var claudeDir string

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove coach hook from Claude Code",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := claudeSettingsPath(claudeDir)
			settings := readJSON(path)

			hooks, ok := settings["hooks"].(map[string]any)
			if !ok {
				fmt.Println("  Nothing to uninstall.")
				return nil
			}

			entries, ok := hooks["UserPromptSubmit"].([]any)
			if !ok {
				fmt.Println("  Nothing to uninstall.")
				return nil
			}

			var filtered []any
			for _, e := range entries {
				if !entryContainsCommand(e, "coach check") {
					filtered = append(filtered, e)
				}
			}

			if len(filtered) == 0 {
				delete(hooks, "UserPromptSubmit")
			} else {
				hooks["UserPromptSubmit"] = filtered
			}
			settings["hooks"] = hooks

			writeJSON(path, settings)
			fmt.Println("  Coach uninstalled. Hook removed.")
			return nil
		},
	}

	cmd.Flags().StringVar(&claudeDir, "claude-dir", "", "Path to Claude config directory (default ~/.claude)")
	return cmd
}

func entryContainsCommand(entry any, needle string) bool {
	e, ok := entry.(map[string]any)
	if !ok {
		return false
	}
	innerHooks, ok := e["hooks"].([]any)
	if !ok {
		return false
	}
	for _, h := range innerHooks {
		if hMap, ok := h.(map[string]any); ok {
			if cmd, ok := hMap["command"].(string); ok && strings.HasSuffix(cmd, needle) {
				return true
			}
		}
	}
	return false
}
