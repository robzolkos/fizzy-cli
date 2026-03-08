package harness

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func init() {
	RegisterAgent(AgentInfo{
		Name:   "Claude Code",
		ID:     "claude",
		Detect: DetectClaude,
		Checks: func() []*StatusCheck {
			checks := []*StatusCheck{CheckClaudePlugin()}
			// Only check the skill link if ~/.claude exists (i.e. Claude is dir-detected)
			home, err := os.UserHomeDir()
			if err == nil {
				if info, statErr := os.Stat(filepath.Join(home, ".claude")); statErr == nil && info.IsDir() {
					checks = append(checks, CheckClaudeSkillLink())
				}
			}
			return checks
		},
	})
}

// ClaudeMarketplaceSource is the marketplace repository for the Fizzy plugin.
const ClaudeMarketplaceSource = "basecamp/claude-plugins"

// ClaudePluginName is the plugin identifier to install.
const ClaudePluginName = "fizzy"

// DetectClaude returns true if Claude Code is installed.
// Checks ~/.claude/ directory first, then falls back to binary on PATH.
func DetectClaude() bool {
	home, err := os.UserHomeDir()
	if err == nil {
		home = filepath.Clean(home)
		info, statErr := os.Stat(filepath.Join(home, ".claude"))
		if statErr == nil && info.IsDir() {
			return true
		}
	}
	return FindClaudeBinary() != ""
}

// FindClaudeBinary returns the path to the claude binary, or "" if not found.
func FindClaudeBinary() string {
	if p, err := exec.LookPath("claude"); err == nil {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	candidate := filepath.Join(filepath.Clean(home), ".local", "bin", "claude")
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return ""
}

// CheckClaudePlugin checks whether the fizzy plugin is installed in Claude Code.
func CheckClaudePlugin() *StatusCheck {
	home, err := os.UserHomeDir()
	if err != nil {
		return &StatusCheck{
			Name:    "Claude Code Plugin",
			Status:  "warn",
			Message: "Cannot determine home directory",
		}
	}

	pluginsPath := filepath.Join(filepath.Clean(home), ".claude", "plugins", "installed_plugins.json")
	data, err := os.ReadFile(pluginsPath) //nolint:gosec // G304: trusted path
	if err != nil {
		if os.IsNotExist(err) {
			return &StatusCheck{
				Name:    "Claude Code Plugin",
				Status:  "fail",
				Message: "Plugin not installed",
				Hint:    "Run: fizzy setup claude",
			}
		}
		return &StatusCheck{
			Name:    "Claude Code Plugin",
			Status:  "warn",
			Message: "Cannot check Claude Code integration",
			Hint:    "Unable to read " + pluginsPath,
		}
	}

	if pluginInstalled(data) {
		return &StatusCheck{
			Name:    "Claude Code Plugin",
			Status:  "pass",
			Message: "Installed",
		}
	}

	return &StatusCheck{
		Name:    "Claude Code Plugin",
		Status:  "fail",
		Message: "Plugin not installed",
		Hint:    "Run: fizzy setup claude",
	}
}

// CheckClaudeSkillLink checks whether ~/.claude/skills/fizzy contains a valid SKILL.md.
func CheckClaudeSkillLink() *StatusCheck {
	home, err := os.UserHomeDir()
	if err != nil {
		return &StatusCheck{
			Name:    "Claude Code Skill",
			Status:  "warn",
			Message: "Cannot determine home directory",
		}
	}

	skillPath := filepath.Join(filepath.Clean(home), ".claude", "skills", "fizzy", "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		if os.IsNotExist(err) {
			return &StatusCheck{
				Name:    "Claude Code Skill",
				Status:  "fail",
				Message: "Skill not linked",
				Hint:    "Run: fizzy setup claude",
			}
		}
		return &StatusCheck{
			Name:    "Claude Code Skill",
			Status:  "warn",
			Message: "Cannot check skill link",
			Hint:    "Unable to stat " + skillPath,
		}
	}

	return &StatusCheck{
		Name:    "Claude Code Skill",
		Status:  "pass",
		Message: "Linked",
	}
}

// pluginInstalled checks if "fizzy" appears as an installed plugin.
func pluginInstalled(data []byte) bool {
	// Try as array of objects
	var plugins []map[string]any
	if err := json.Unmarshal(data, &plugins); err == nil {
		for _, p := range plugins {
			if matchesFizzy(p) {
				return true
			}
		}
		return false
	}

	// Try as map (key = plugin identifier, or v2 envelope with "plugins" key)
	var pluginMap map[string]any
	if err := json.Unmarshal(data, &pluginMap); err == nil {
		// v2 format: {"version": 2, "plugins": {"fizzy@marketplace": [...]}}
		if inner, ok := pluginMap["plugins"]; ok {
			if innerMap, ok := inner.(map[string]any); ok {
				for key := range innerMap {
					if key == "fizzy" || matchesPluginKey(key) {
						return true
					}
				}
				return false
			}
		}
		// v1 flat map: {"fizzy@fizzy": {...}}
		for key := range pluginMap {
			if key == "fizzy" || matchesPluginKey(key) {
				return true
			}
		}
		return false
	}

	// Fallback: raw string search
	return jsonContainsFizzy(data)
}

func matchesFizzy(p map[string]any) bool {
	for _, field := range []string{"name", "package", "id"} {
		if v, ok := p[field]; ok {
			if s, ok := v.(string); ok {
				if matchesPluginKey(s) {
					return true
				}
			}
		}
	}
	return false
}

// matchesPluginKey returns true if the key identifies the fizzy plugin.
func matchesPluginKey(key string) bool {
	return key == "fizzy" || strings.HasPrefix(key, "fizzy@")
}

func jsonContainsFizzy(data []byte) bool {
	s := string(data)
	return strings.Contains(s, `"fizzy"`) || strings.Contains(s, `"fizzy@`)
}
