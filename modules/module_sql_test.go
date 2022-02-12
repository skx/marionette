package modules

import (
	"strings"
	"testing"
)

func TestSqlArgs(t *testing.T) {

	s := &SQLModule{}

	args := make(map[string]interface{})

	// Missing 'driver'
	err := s.Check(args)
	if err == nil {
		t.Fatalf("expected error due to missing driver")
	}
	if !strings.Contains(err.Error(), "missing 'driver'") {
		t.Fatalf("got error - but wrong one : %s", err)
	}

	// setup a driver
	args["driver"] = "mysql"

	// Missing 'dsn'
	err = s.Check(args)
	if err == nil {
		t.Fatalf("expected error due to missing dsn")
	}
	if !strings.Contains(err.Error(), "missing 'dsn'") {
		t.Fatalf("got error - but wrong one : %s", err)
	}

	// setup a dsn
	args["dsn"] = "root@blah"

	err = s.Check(args)
	if err == nil {
		t.Fatalf("expected error due to missing sql")
	}
	if !strings.Contains(err.Error(), "must specify one of") {
		t.Fatalf("got error - but wrong one : %s", err)
	}

	// setup sql
	args["sql"] = "SELECT 1"
	err = s.Check(args)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	// setup sql_file
	args["sql_file"] = "/etc/passwd"

	err = s.Check(args)
	if err == nil {
		t.Fatalf("expected error due to setting sql AND sql_file")
	}
	if !strings.Contains(err.Error(), "must specify one of") {
		t.Fatalf("got error - but wrong one : %s", err)
	}

}
