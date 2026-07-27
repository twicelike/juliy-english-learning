package config

import (
	"os"
	"testing"
)

func TestLoad_RequiredFieldsMissing(t *testing.T) {
	os.Unsetenv("MONGO_URI")
	os.Unsetenv("MONGO_DB")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when MONGO_URI is missing, got nil")
	}
}

func TestLoad_Success(t *testing.T) {
	t.Setenv("MONGO_URI", "mongodb://localhost:27017")
	t.Setenv("MONGO_DB", "test_db")
	t.Setenv("PORT", "9000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "9000" {
		t.Errorf("expected port 9000, got %s", cfg.Port)
	}
	if cfg.MongoURI != "mongodb://localhost:27017" {
		t.Errorf("unexpected MongoURI: %s", cfg.MongoURI)
	}
}

func TestLoad_DefaultPort(t *testing.T) {
	t.Setenv("MONGO_URI", "mongodb://localhost:27017")
	t.Setenv("MONGO_DB", "test_db")
	os.Unsetenv("PORT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Port)
	}
}
