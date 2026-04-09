package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var claudeDir string

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Set up coach hook in Claude Code",
	RunE: func(cmd *cobra.Command, args []string) error {
		settingsPath := resolveClaudeSettings()

		coachBin, err := exec.LookPath("coach")
		if err != nil {
			coachBin = "coach"
		}
		hookCommand := fmt.Sprintf("%s check", coachBin)

		settings := make(map[string]any)
		if data, err := os.ReadFile(settingsPath); err == nil {
			json.Unmarshal(data, &settings)
		}

		hooks, ok := settings["hooks"].(map[string]any)
		if !ok {
			hooks = make(map[string]any)
		}

		// Check if already installed
		if existing, ok := hooks["UserPromptSubmit"].([]any); ok {
			for _, h := range existing {
				if entry, ok := h.(map[string]any); ok {
					if innerHooks, ok := entry["hooks"].([]any); ok {
						for _, ih := range innerHooks {
							if ihMap, ok := ih.(map[string]any); ok {
								if ihMap["command"] == hookCommand {
									fmt.Println("  Coach is already installed.")
									return nil
								}
							}
						}
					}
				}
			}
		}

		newHook := map[string]any{
			"hooks": []any{
				map[string]any{
					"type":    "command",
					"command": hookCommand,
					"timeout": 5,
				},
			},
		}

		if existing, ok := hooks["UserPromptSubmit"].([]any); ok {
			hooks["UserPromptSubmit"] = append(existing, newHook)
		} else {
			hooks["UserPromptSubmit"] = []any{newHook}
		}

		settings["hooks"] = hooks

		if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
			return err
		}

		data, err := json.MarshalIndent(settings, "", "  ")
		if err != nil {
			return err
		}

		if err := os.WriteFile(settingsPath, data, 0644); err != nil {
			return err
		}

		fmt.Println("  Coach installed!")
		fmt.Printf("  Hook added to %s\n", settingsPath)
		return nil
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove coach hook from Claude Code",
	RunE: func(cmd *cobra.Command, args []string) error {
		settingsPath := resolveClaudeSettings()

		settings := make(map[string]any)
		data, err := os.ReadFile(settingsPath)
		if err != nil {
			fmt.Println("  Nothing to uninstall.")
			return nil
		}
		json.Unmarshal(data, &settings)

		hooks, ok := settings["hooks"].(map[string]any)
		if !ok {
			fmt.Println("  Nothing to uninstall.")
			return nil
		}

		if existing, ok := hooks["UserPromptSubmit"].([]any); ok {
			var filtered []any
			for _, h := range existing {
				keep := true
				if entry, ok := h.(map[string]any); ok {
					if innerHooks, ok := entry["hooks"].([]any); ok {
						for _, ih := range innerHooks {
							if ihMap, ok := ih.(map[string]any); ok {
								if cmd, ok := ihMap["command"].(string); ok {
									if len(cmd) >= 11 && cmd[len(cmd)-11:] == "coach check" {
										keep = false
									}
								}
							}
						}
					}
				}
				if keep {
					filtered = append(filtered, h)
				}
			}
			if len(filtered) == 0 {
				delete(hooks, "UserPromptSubmit")
			} else {
				hooks["UserPromptSubmit"] = filtered
			}
		}

		settings["hooks"] = hooks
		out, _ := json.MarshalIndent(settings, "", "  ")
		os.WriteFile(settingsPath, out, 0644)

		fmt.Println("  Coach uninstalled. Hook removed.")
		return nil
	},
}

func resolveClaudeSettings() string {
	if claudeDir != "" {
		return filepath.Join(claudeDir, "settings.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "settings.json")
}

func init() {
	installCmd.Flags().StringVar(&claudeDir, "claude-dir", "", "Path to Claude config directory (default ~/.claude)")
	uninstallCmd.Flags().StringVar(&claudeDir, "claude-dir", "", "Path to Claude config directory (default ~/.claude)")
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
}
