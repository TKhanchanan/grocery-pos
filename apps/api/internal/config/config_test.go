package config

import "testing"

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
