package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListenAddr(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "Railway PORT takes precedence",
			cfg:  Config{Port: " 9000 ", APIAddr: "127.0.0.1:8080"},
			want: ":9000",
		},
		{
			name: "PORT tolerates a leading colon",
			cfg:  Config{Port: ":9001", APIAddr: ":8080"},
			want: ":9001",
		},
		{
			name: "API_ADDR is the fallback",
			cfg:  Config{APIAddr: " :8081 "},
			want: ":8081",
		},
		{
			name: "default binds all interfaces",
			cfg:  Config{},
			want: ":8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.ListenAddr(); got != tt.want {
				t.Fatalf("ListenAddr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadDefaultCORSOrigins(t *testing.T) {
	t.Setenv("CORS_ORIGINS", "")

	got := Load().CORSOrigins
	want := "https://grocery-pos-front-production.up.railway.app,http://localhost:5173"
	if got != want {
		t.Fatalf("CORSOrigins = %q, want %q", got, want)
	}
}

func TestLoadUploadDir(t *testing.T) {
	t.Run("uses configured Railway volume", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		t.Setenv("UPLOAD_DIR", " /app/storage/uploads/ ")

		if got := Load().UploadDir; got != "/app/storage/uploads" {
			t.Fatalf("UploadDir = %q, want %q", got, "/app/storage/uploads")
		}
	})

	t.Run("falls back for local development", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv("UPLOAD_DIR", "")

		if got := Load().UploadDir; got != defaultUploadDir {
			t.Fatalf("UploadDir = %q, want %q", got, defaultUploadDir)
		}
	})

	t.Run("does not use local fallback in production", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		t.Setenv("UPLOAD_DIR", "")

		cfg := Load()
		if cfg.UploadDir != "" {
			t.Fatalf("UploadDir = %q, want empty", cfg.UploadDir)
		}
		if err := cfg.PrepareUploadDir(); err == nil {
			t.Fatal("PrepareUploadDir() error = nil, want production configuration error")
		}
	})
}

func TestPrepareUploadDir(t *testing.T) {
	root := filepath.Join(t.TempDir(), "uploads")
	cfg := Config{UploadDir: root}

	if err := cfg.PrepareUploadDir(); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{root, filepath.Join(root, "products"), filepath.Join(root, "avatars")} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("stat %s: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("%s is not a directory", dir)
		}
	}
}
