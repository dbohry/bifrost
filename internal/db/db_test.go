package db

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func open(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func TestOpenSetsPragmas(t *testing.T) {
	conn := open(t)

	var mode string
	if err := conn.QueryRow(`PRAGMA journal_mode`).Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if mode != "wal" {
		t.Errorf("journal_mode = %q, want wal", mode)
	}

	var fk int
	if err := conn.QueryRow(`PRAGMA foreign_keys`).Scan(&fk); err != nil {
		t.Fatal(err)
	}
	if fk != 1 {
		t.Errorf("foreign_keys = %d, want 1", fk)
	}

	var timeout int
	if err := conn.QueryRow(`PRAGMA busy_timeout`).Scan(&timeout); err != nil {
		t.Fatal(err)
	}
	if timeout != 5000 {
		t.Errorf("busy_timeout = %d, want 5000", timeout)
	}
}

func TestMigrateAppliesOnceAndIsIdempotent(t *testing.T) {
	conn := open(t)

	for i := 0; i < 2; i++ {
		if err := Migrate(conn); err != nil {
			t.Fatalf("run %d: %v", i+1, err)
		}
	}

	var n int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("schema_migrations has %d rows, want 1", n)
	}
	if _, err := conn.Exec(`INSERT INTO meta (key, value) VALUES ('k', 'v')`); err != nil {
		t.Errorf("meta table missing: %v", err)
	}
}

func TestMigrateAppliesInOrderAndRollsBackFailures(t *testing.T) {
	conn := open(t)
	files := fstest.MapFS{
		"migrations/0002_second.sql": {Data: []byte(`INSERT INTO t VALUES (2);`)},
		"migrations/0001_first.sql":  {Data: []byte(`CREATE TABLE t (n INTEGER);`)},
		"migrations/0003_broken.sql": {Data: []byte(`INSERT INTO t VALUES (3); INSERT INTO nope VALUES (1);`)},
	}

	err := migrate(conn, files)
	if err == nil || !strings.Contains(err.Error(), "0003_broken") {
		t.Fatalf("want an error naming 0003_broken, got %v", err)
	}

	var applied, rows int
	conn.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&applied)
	conn.QueryRow(`SELECT COUNT(*) FROM t`).Scan(&rows)
	if applied != 2 {
		t.Errorf("applied = %d, want 2 (broken one rolled back)", applied)
	}
	if rows != 1 {
		t.Errorf("t has %d rows, want 1 (row from the broken migration rolled back)", rows)
	}
}

func TestLoadRejectsBadNames(t *testing.T) {
	cases := map[string]fstest.MapFS{
		"no prefix":   {"migrations/init.sql": {Data: []byte("")}},
		"zero":        {"migrations/0000_x.sql": {Data: []byte("")}},
		"duplicate":   {"migrations/0001_a.sql": {Data: []byte("")}, "migrations/0001_b.sql": {Data: []byte("")}},
		"not numeric": {"migrations/abc_x.sql": {Data: []byte("")}},
	}
	for name, files := range cases {
		if _, err := load(files); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}
