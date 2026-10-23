package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("BIFROST_ADDR", "")
	t.Setenv("BIFROST_DATA_DIR", "")

	c := Load()
	if c.Addr != ":8080" || c.DataDir != "/data" {
		t.Fatalf("unexpected defaults: %+v", c)
	}
	if got := c.DBPath(); got != "/data/bifrost.db" {
		t.Errorf("DBPath = %q", got)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("BIFROST_ADDR", "0.0.0.0:9000")
	t.Setenv("BIFROST_DATA_DIR", "/tmp/bifrost")

	c := Load()
	if c.Addr != "0.0.0.0:9000" || c.DataDir != "/tmp/bifrost" {
		t.Fatalf("unexpected config: %+v", c)
	}
}

func TestHealthURL(t *testing.T) {
	cases := map[string]string{
		":8080":          "http://127.0.0.1:8080/healthz",
		"0.0.0.0:9000":   "http://127.0.0.1:9000/healthz",
		"192.168.1.5:80": "http://127.0.0.1:80/healthz",
		"garbage":        "http://127.0.0.1:8080/healthz",
	}
	for addr, want := range cases {
		if got := (Config{Addr: addr}).HealthURL(); got != want {
			t.Errorf("HealthURL(%q) = %q, want %q", addr, got, want)
		}
	}
}
