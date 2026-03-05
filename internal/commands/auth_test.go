package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/basecamp/cli/credstore"
	"github.com/basecamp/fizzy-cli/internal/config"
	"gopkg.in/yaml.v3"
)

func TestAuthLogin(t *testing.T) {
	t.Run("saves token to config file", func(t *testing.T) {
		// Create temp directory for config
		tempDir, err := os.MkdirTemp("", "fizzy-test-*")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		config.SetTestConfigDir(tempDir)
		defer config.ResetTestConfigDir()

		mock := NewMockClient()
		result := SetTestMode(mock)
		defer ResetTestMode()

		err = authLoginCmd.RunE(authLoginCmd, []string{"test-token-123"})
		assertExitCode(t, err, 0)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Response.OK {
			t.Error("expected success response")
		}

		// Verify config file was created with correct token
		configPath := filepath.Join(tempDir, "config.yaml")
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("config file not created: %v", err)
		}

		var savedConfig config.Config
		if err := yaml.Unmarshal(data, &savedConfig); err != nil {
			t.Fatalf("failed to parse config: %v", err)
		}

		if savedConfig.Token != "test-token-123" {
			t.Errorf("expected token 'test-token-123', got '%s'", savedConfig.Token)
		}
	})

	t.Run("saves token to credstore when available", func(t *testing.T) {
		tempDir := t.TempDir()
		configDir := t.TempDir()

		config.SetTestConfigDir(configDir)
		defer config.ResetTestConfigDir()

		// Create a file-based credstore
		os.Setenv("FIZZY_TEST_NO_KR", "1")
		defer os.Unsetenv("FIZZY_TEST_NO_KR")
		store := credstore.NewStore(credstore.StoreOptions{
			ServiceName:   "fizzy-test",
			DisableEnvVar: "FIZZY_TEST_NO_KR",
			FallbackDir:   tempDir,
		})

		mock := NewMockClient()
		result := SetTestMode(mock)
		SetTestCreds(store)
		defer ResetTestMode()

		err := authLoginCmd.RunE(authLoginCmd, []string{"cred-token-456"})
		assertExitCode(t, err, 0)

		if !result.Response.OK {
			t.Error("expected success response")
		}

		// Token should be in credstore (stored as JSON-encoded string)
		loaded, err := store.Load("token")
		if err != nil {
			t.Fatalf("expected token in credstore: %v", err)
		}
		var tokenStr string
		if err := json.Unmarshal(loaded, &tokenStr); err != nil {
			t.Fatalf("expected JSON-encoded token, got %q: %v", string(loaded), err)
		}
		if tokenStr != "cred-token-456" {
			t.Errorf("expected 'cred-token-456', got '%s'", tokenStr)
		}

		// Token should NOT be in YAML config
		configPath := filepath.Join(configDir, "config.yaml")
		if data, err := os.ReadFile(configPath); err == nil {
			var savedConfig config.Config
			yaml.Unmarshal(data, &savedConfig)
			if savedConfig.Token != "" {
				t.Errorf("expected empty token in YAML, got '%s'", savedConfig.Token)
			}
		}
	})

	t.Run("preserves existing config values", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "fizzy-test-*")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		config.SetTestConfigDir(tempDir)
		defer config.ResetTestConfigDir()

		// Create existing config with account
		existingConfig := &config.Config{
			Account: "existing-account",
			APIURL:  "https://custom.api.com",
		}
		existingData, _ := yaml.Marshal(existingConfig)
		os.WriteFile(filepath.Join(tempDir, "config.yaml"), existingData, 0600)

		mock := NewMockClient()
		SetTestMode(mock)
		defer ResetTestMode()

		err = authLoginCmd.RunE(authLoginCmd, []string{"new-token"})
		assertExitCode(t, err, 0)

		// Verify existing values preserved
		data, _ := os.ReadFile(filepath.Join(tempDir, "config.yaml"))
		var savedConfig config.Config
		yaml.Unmarshal(data, &savedConfig)

		if savedConfig.Token != "new-token" {
			t.Errorf("expected token 'new-token', got '%s'", savedConfig.Token)
		}
		if savedConfig.Account != "existing-account" {
			t.Errorf("expected account 'existing-account', got '%s'", savedConfig.Account)
		}
	})
}

func TestAuthLogout(t *testing.T) {
	t.Run("removes config file", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "fizzy-test-*")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		config.SetTestConfigDir(tempDir)
		defer config.ResetTestConfigDir()

		// Create config file
		configPath := filepath.Join(tempDir, "config.yaml")
		os.WriteFile(configPath, []byte("token: test-token"), 0600)

		mock := NewMockClient()
		SetTestMode(mock)
		defer ResetTestMode()

		err = authLogoutCmd.RunE(authLogoutCmd, []string{})
		assertExitCode(t, err, 0)

		// Verify config file was removed
		if _, err := os.Stat(configPath); !os.IsNotExist(err) {
			t.Error("expected config file to be removed")
		}
	})

	t.Run("succeeds even if no config file exists", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "fizzy-test-*")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		config.SetTestConfigDir(tempDir)
		defer config.ResetTestConfigDir()

		mock := NewMockClient()
		SetTestMode(mock)
		defer ResetTestMode()

		err = authLogoutCmd.RunE(authLogoutCmd, []string{})
		assertExitCode(t, err, 0)
	})
}

func TestAuthStatus(t *testing.T) {
	t.Run("shows authenticated status when token exists", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "fizzy-test-*")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		config.SetTestConfigDir(tempDir)
		defer config.ResetTestConfigDir()

		// Create config with token
		configData := "token: test-token\naccount: test-account"
		os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte(configData), 0600)

		mock := NewMockClient()
		result := SetTestMode(mock)
		defer ResetTestMode()

		err = authStatusCmd.RunE(authStatusCmd, []string{})
		assertExitCode(t, err, 0)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Response.OK {
			t.Error("expected success response")
		}

		// Check response data
		data, ok := result.Response.Data.(map[string]any)
		if !ok {
			t.Fatal("expected map response data")
		}
		if data["authenticated"] != true {
			t.Errorf("expected authenticated=true, got %v", data["authenticated"])
		}
		if data["token_configured"] != true {
			t.Errorf("expected token_configured=true, got %v", data["token_configured"])
		}
		if data["account"] != "test-account" {
			t.Errorf("expected account='test-account', got %v", data["account"])
		}
	})

	t.Run("shows unauthenticated status when no token", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "fizzy-test-*")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		config.SetTestConfigDir(tempDir)
		defer config.ResetTestConfigDir()

		mock := NewMockClient()
		result := SetTestMode(mock)
		defer ResetTestMode()

		err = authStatusCmd.RunE(authStatusCmd, []string{})
		assertExitCode(t, err, 0)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		data, ok := result.Response.Data.(map[string]any)
		if !ok {
			t.Fatal("expected map response data")
		}
		if data["authenticated"] != false {
			t.Errorf("expected authenticated=false, got %v", data["authenticated"])
		}
	})

	t.Run("shows using_keyring field when credstore is set", func(t *testing.T) {
		mock := NewMockClient()
		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")

		// Create a file-based credstore (env var disables keyring probe)
		tempDir := t.TempDir()
		os.Setenv("FIZZY_TEST_NO_KEYRING_ALWAYS", "1")
		defer os.Unsetenv("FIZZY_TEST_NO_KEYRING_ALWAYS")
		store := credstore.NewStore(credstore.StoreOptions{
			ServiceName:   "fizzy-test",
			DisableEnvVar: "FIZZY_TEST_NO_KEYRING_ALWAYS",
			FallbackDir:   tempDir,
		})
		SetTestCreds(store)
		defer ResetTestMode()

		err := authStatusCmd.RunE(authStatusCmd, []string{})
		assertExitCode(t, err, 0)

		data, ok := result.Response.Data.(map[string]any)
		if !ok {
			t.Fatal("expected map response data")
		}
		if _, ok := data["using_keyring"]; !ok {
			t.Error("expected using_keyring field when credstore is set")
		}
	})

	t.Run("shows custom api_url when configured", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "fizzy-test-*")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		config.SetTestConfigDir(tempDir)
		defer config.ResetTestConfigDir()

		// Create config with custom API URL
		configData := "token: test-token\napi_url: https://custom.fizzy.do"
		os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte(configData), 0600)

		mock := NewMockClient()
		result := SetTestMode(mock)
		defer ResetTestMode()

		err = authStatusCmd.RunE(authStatusCmd, []string{})
		assertExitCode(t, err, 0)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		data := result.Response.Data.(map[string]any)
		if data["api_url"] != "https://custom.fizzy.do" {
			t.Errorf("expected api_url='https://custom.fizzy.do', got %v", data["api_url"])
		}
	})
}

func TestTokenMigrationToCredstore(t *testing.T) {
	t.Run("migrates YAML token to credstore on first load", func(t *testing.T) {
		configDir := t.TempDir()
		credDir := t.TempDir()

		config.SetTestConfigDir(configDir)
		defer config.ResetTestConfigDir()

		// Write a global config with a token (pre-credstore state)
		os.WriteFile(filepath.Join(configDir, "config.yaml"),
			[]byte("token: migrate-me\naccount: acme"), 0600)

		// Set up credstore (file-based, no keyring)
		os.Setenv("FIZZY_MIGRATE_NO_KR", "1")
		defer os.Unsetenv("FIZZY_MIGRATE_NO_KR")
		store := credstore.NewStore(credstore.StoreOptions{
			ServiceName:   "fizzy-migrate-test",
			DisableEnvVar: "FIZZY_MIGRATE_NO_KR",
			FallbackDir:   credDir,
		})

		mock := NewMockClient()
		SetTestMode(mock)
		SetTestCreds(store)
		SetTestConfig("migrate-me", "acme", "https://app.fizzy.do")
		defer ResetTestMode()

		// resolveToken triggers the migration
		resolveToken()

		// Token should now be in credstore
		var tokenStr string
		loaded, err := store.Load("token")
		if err != nil {
			t.Fatalf("expected token in credstore after migration: %v", err)
		}
		if err := json.Unmarshal(loaded, &tokenStr); err != nil {
			t.Fatalf("expected JSON-encoded token, got %q: %v", string(loaded), err)
		}
		if tokenStr != "migrate-me" {
			t.Errorf("expected 'migrate-me' in credstore, got '%s'", tokenStr)
		}

		// Global YAML config should have token cleared
		data, err := os.ReadFile(filepath.Join(configDir, "config.yaml"))
		if err != nil {
			t.Fatalf("config file missing: %v", err)
		}
		var savedConfig config.Config
		yaml.Unmarshal(data, &savedConfig)
		if savedConfig.Token != "" {
			t.Errorf("expected empty token in YAML after migration, got '%s'", savedConfig.Token)
		}
		// Account should be preserved
		if savedConfig.Account != "acme" {
			t.Errorf("expected account 'acme' preserved, got '%s'", savedConfig.Account)
		}
	})

	t.Run("does not migrate when credstore already has token", func(t *testing.T) {
		configDir := t.TempDir()
		credDir := t.TempDir()

		config.SetTestConfigDir(configDir)
		defer config.ResetTestConfigDir()

		// Write a global config with a stale token
		os.WriteFile(filepath.Join(configDir, "config.yaml"),
			[]byte("token: old-yaml-token\naccount: acme"), 0600)

		os.Setenv("FIZZY_MIGRATE2_NO_KR", "1")
		defer os.Unsetenv("FIZZY_MIGRATE2_NO_KR")
		store := credstore.NewStore(credstore.StoreOptions{
			ServiceName:   "fizzy-migrate2-test",
			DisableEnvVar: "FIZZY_MIGRATE2_NO_KR",
			FallbackDir:   credDir,
		})

		// Pre-populate credstore with a different token
		credToken, _ := json.Marshal("cred-token")
		store.Save("token", credToken)

		mock := NewMockClient()
		SetTestMode(mock)
		SetTestCreds(store)
		SetTestConfig("old-yaml-token", "acme", "https://app.fizzy.do")
		defer ResetTestMode()

		resolveToken()

		// cfg.Token should be the credstore token (it wins over YAML)
		if cfg.Token != "cred-token" {
			t.Errorf("expected 'cred-token' from credstore, got '%s'", cfg.Token)
		}

		// YAML config should be untouched (no migration needed)
		data, _ := os.ReadFile(filepath.Join(configDir, "config.yaml"))
		var savedConfig config.Config
		yaml.Unmarshal(data, &savedConfig)
		if savedConfig.Token != "old-yaml-token" {
			t.Errorf("YAML token should be untouched, got '%s'", savedConfig.Token)
		}
	})

	t.Run("does not migrate env-var token to credstore", func(t *testing.T) {
		configDir := t.TempDir()
		credDir := t.TempDir()

		config.SetTestConfigDir(configDir)
		defer config.ResetTestConfigDir()

		// Global YAML config has NO token — only env var provides one
		os.WriteFile(filepath.Join(configDir, "config.yaml"),
			[]byte("account: acme"), 0600)

		os.Setenv("FIZZY_MIGRATE3_NO_KR", "1")
		defer os.Unsetenv("FIZZY_MIGRATE3_NO_KR")
		store := credstore.NewStore(credstore.StoreOptions{
			ServiceName:   "fizzy-migrate3-test",
			DisableEnvVar: "FIZZY_MIGRATE3_NO_KR",
			FallbackDir:   credDir,
		})

		mock := NewMockClient()
		SetTestMode(mock)
		SetTestCreds(store)
		// cfg.Token set via env-like source (not from global YAML)
		SetTestConfig("env-token", "acme", "https://app.fizzy.do")
		defer ResetTestMode()

		resolveToken()

		// Credstore should remain empty — env tokens must not be persisted
		if _, err := store.Load("token"); err == nil {
			t.Error("env-var token should not be migrated to credstore")
		}
	})

	t.Run("loads legacy raw-string token from credstore", func(t *testing.T) {
		configDir := t.TempDir()
		credDir := t.TempDir()

		config.SetTestConfigDir(configDir)
		defer config.ResetTestConfigDir()

		os.Setenv("FIZZY_MIGRATE4_NO_KR", "1")
		defer os.Unsetenv("FIZZY_MIGRATE4_NO_KR")
		store := credstore.NewStore(credstore.StoreOptions{
			ServiceName:   "fizzy-migrate4-test",
			DisableEnvVar: "FIZZY_MIGRATE4_NO_KR",
			FallbackDir:   credDir,
		})

		// Seed the credentials file with a value that is valid JSON (so the
		// file backend can parse it) but NOT a JSON string — a number.
		// json.Unmarshal into a string will fail, exercising the raw
		// fallback path in credsLoadToken. This simulates what the keyring
		// backend would return for a pre-JSON-encoding token.
		credFile := filepath.Join(credDir, "credentials.json")
		os.MkdirAll(credDir, 0700)
		os.WriteFile(credFile, []byte(`{"token": 12345}`), 0600)

		mock := NewMockClient()
		SetTestMode(mock)
		SetTestCreds(store)
		SetTestConfig("", "acme", "https://app.fizzy.do")
		defer ResetTestMode()

		resolveToken()

		// The number 12345 can't json.Unmarshal into a string, so the
		// fallback path returns the raw bytes as a string.
		if cfg.Token != "12345" {
			t.Errorf("expected legacy raw token '12345', got '%s'", cfg.Token)
		}
	})
}
