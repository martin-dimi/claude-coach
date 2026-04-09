package cmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
)

func claudeSettingsPath(claudeDir string) string {
	if claudeDir != "" {
		return filepath.Join(claudeDir, "settings.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "settings.json")
}

func resolveCoachBin() string {
	if path, err := exec.LookPath("coach"); err == nil {
		return path
	}
	return "coach"
}

func readJSON(path string) map[string]any {
	m := make(map[string]any)
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &m)
	}
	return m
}

func writeJSON(path string, data map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0644)
}

func ensureMap(m map[string]any, key string) map[string]any {
	if v, ok := m[key].(map[string]any); ok {
		return v
	}
	v := make(map[string]any)
	m[key] = v
	return v
}

func hasHook(hooks map[string]any, event, command string) bool {
	entries, ok := hooks[event].([]any)
	if !ok {
		return false
	}
	for _, e := range entries {
		if entryMap, ok := e.(map[string]any); ok {
			if innerHooks, ok := entryMap["hooks"].([]any); ok {
				for _, h := range innerHooks {
					if hMap, ok := h.(map[string]any); ok {
						if hMap["command"] == command {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func addHook(hooks map[string]any, event, command string) {
	entry := map[string]any{
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": command,
				"timeout": 5,
			},
		},
	}
	if existing, ok := hooks[event].([]any); ok {
		hooks[event] = append(existing, entry)
	} else {
		hooks[event] = []any{entry}
	}
}
