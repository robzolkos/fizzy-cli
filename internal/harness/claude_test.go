package harness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectClaude(t *testing.T) {
	t.Run("detects via .claude directory", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
			t.Fatal(err)
		}
		if !DetectClaude() {
			t.Error("expected DetectClaude() to return true when ~/.claude exists")
		}
	})

	t.Run("returns false when nothing present", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("PATH", home) // no claude binary
		if DetectClaude() {
			t.Error("expected DetectClaude() to return false")
		}
	})
}

func TestCheckClaudePlugin(t *testing.T) {
	t.Run("pass when plugin found in v2 format", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		pluginsDir := filepath.Join(home, ".claude", "plugins")
		os.MkdirAll(pluginsDir, 0o755)

		data, _ := json.Marshal(map[string]any{
			"version": 2,
			"plugins": map[string]any{
				"fizzy@marketplace": []any{},
			},
		})
		os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), data, 0o644)

		check := CheckClaudePlugin()
		if check.Status != "pass" {
			t.Errorf("expected pass, got %s: %s", check.Status, check.Message)
		}
	})

	t.Run("pass when plugin found in array format", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		pluginsDir := filepath.Join(home, ".claude", "plugins")
		os.MkdirAll(pluginsDir, 0o755)

		data, _ := json.Marshal([]map[string]any{
			{"name": "fizzy", "version": "0.1.0"},
		})
		os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), data, 0o644)

		check := CheckClaudePlugin()
		if check.Status != "pass" {
			t.Errorf("expected pass, got %s: %s", check.Status, check.Message)
		}
	})

	t.Run("fail when no plugins file", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)

		check := CheckClaudePlugin()
		if check.Status != "fail" {
			t.Errorf("expected fail, got %s: %s", check.Status, check.Message)
		}
	})

	t.Run("fail when plugin not in list", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		pluginsDir := filepath.Join(home, ".claude", "plugins")
		os.MkdirAll(pluginsDir, 0o755)

		data, _ := json.Marshal([]map[string]any{
			{"name": "other-plugin"},
		})
		os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), data, 0o644)

		check := CheckClaudePlugin()
		if check.Status != "fail" {
			t.Errorf("expected fail, got %s: %s", check.Status, check.Message)
		}
	})
}

func TestCheckClaudeSkillLink(t *testing.T) {
	t.Run("pass when skill file exists", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		skillDir := filepath.Join(home, ".claude", "skills", "fizzy")
		os.MkdirAll(skillDir, 0o755)
		os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("content"), 0o644)

		check := CheckClaudeSkillLink()
		if check.Status != "pass" {
			t.Errorf("expected pass, got %s: %s", check.Status, check.Message)
		}
	})

	t.Run("fail when skill file missing", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)

		check := CheckClaudeSkillLink()
		if check.Status != "fail" {
			t.Errorf("expected fail, got %s: %s", check.Status, check.Message)
		}
	})
}

func TestPluginInstalled(t *testing.T) {
	tests := []struct {
		name string
		data string
		want bool
	}{
		{"v1 flat map", `{"fizzy@fizzy": {}}`, true},
		{"v2 envelope", `{"version":2,"plugins":{"fizzy@marketplace":[]}}`, true},
		{"array with name", `[{"name":"fizzy"}]`, true},
		{"array with package", `[{"package":"fizzy@latest"}]`, true},
		{"bare key", `{"fizzy":{}}`, true},
		{"empty array", `[]`, false},
		{"empty object", `{}`, false},
		{"other plugin only", `[{"name":"other"}]`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pluginInstalled([]byte(tt.data))
			if got != tt.want {
				t.Errorf("pluginInstalled(%s) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestMatchesPluginKey(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"fizzy", true},
		{"fizzy@marketplace", true},
		{"fizzy@fizzy", true},
		{"other", false},
		{"fizzybuzz", false},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := matchesPluginKey(tt.key); got != tt.want {
				t.Errorf("matchesPluginKey(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}
