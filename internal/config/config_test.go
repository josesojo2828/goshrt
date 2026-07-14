package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	cfg := Load()
	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.ClickSyncInterval != 5*time.Second {
		t.Errorf("expected default interval 5s, got %v", cfg.ClickSyncInterval)
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("CLICK_SYNC_INTERVAL", "10s")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("CLICK_SYNC_INTERVAL")
	}()

	cfg := Load()
	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.ClickSyncInterval != 10*time.Second {
		t.Errorf("expected interval 10s, got %v", cfg.ClickSyncInterval)
	}
}
