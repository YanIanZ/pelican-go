package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load("/nonexistent/path.yml")
	if err != nil {
		t.Fatalf("expected no error for missing config file, got: %v", err)
	}
	if cfg.App.Name != "Pelican" {
		t.Errorf("expected app name 'Pelican', got: %s", cfg.App.Name)
	}
	if cfg.Database.Driver != "mysql" {
		t.Errorf("expected database driver 'mysql', got: %s", cfg.Database.Driver)
	}
	if cfg.Panel.UseBinaryPrefix != true {
		t.Errorf("expected use_binary_prefix true, got: %v", cfg.Panel.UseBinaryPrefix)
	}
	if cfg.API.KeyLimit != 25 {
		t.Errorf("expected key_limit 25, got: %d", cfg.API.KeyLimit)
	}
}

func TestEnvOverride(t *testing.T) {
	os.Setenv("PANEL_APP_NAME", "TestApp")
	defer os.Unsetenv("PANEL_APP_NAME")

	cfg, err := Load("/nonexistent/path.yml")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.App.Name != "TestApp" {
		t.Errorf("expected env override 'TestApp', got: %s", cfg.App.Name)
	}
}
