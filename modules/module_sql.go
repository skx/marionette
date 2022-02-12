package modules

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/skx/marionette/config"
	"github.com/skx/marionette/environment"

	// TODO - more?
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

// SQLModule stores our state.
type SQLModule struct {

	// cfg contains our configuration object.
	cfg *config.Config
}

// Check is part of the module-api, and checks arguments.
func (f *SQLModule) Check(args map[string]interface{}) error {

	// Ensure we have a "driver" and a "dsn"
	for _, arg := range []string{"driver", "dsn"} {

		_, ok := args[arg]
		if !ok {
			return fmt.Errorf("missing '%s' parameter", arg)
		}
	}

	// We must have one of "sql" or "sql_file"
	count := 0

	for _, arg := range []string{"sql", "sql_file"} {
		_, ok := args[arg]
		if ok {
			count++
		}
	}

	if count != 1 {
		return fmt.Errorf("you must specify one of 'sql' or 'sql_file'")
	}

	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *SQLModule) Execute(env *environment.Environment, args map[string]interface{}) (bool, error) {

	// Get our DSN + Driver
	dsn := StringParam(args, "dsn")
	driver := StringParam(args, "driver")

	// One of these will be valid
	sqlText := StringParam(args, "sql")
	sqlFile := StringParam(args, "file")

	// Open the database
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return false, err
	}

	// Avoid leaking the handle.
	defer db.Close()

	// We're either running a query with a literal string,
	// or reading from a file.
	if sqlFile != "" {

		// If reading from a file then do so.
		data, err := ioutil.ReadFile(sqlFile)
		if err != nil {
			return false, err
		}

		sqlText = string(data)
	}

	// Now actually run the SQL
	res, execErr := db.Exec(sqlText)
	if execErr != nil {
		return false, execErr
	}

	// Try to see if we can get a useful output.
	rows, rErr := res.RowsAffected()
	ins, iErr := res.LastInsertId()

	// Show rows-affected, or the appropriate error.
	//
	// NOTE: An error here doesn't break our module invocation.
	if rErr == nil {
		log.Printf("[DEBUG] sql - affected rows  %d", rows)
	} else {
		log.Printf("[DEBUG] sql - affected rows error %s", rErr)
	}

	// Show the last insert-id, or the appropriate error.
	//
	// NOTE: An error here doesn't break our module invocation.
	if iErr == nil {
		log.Printf("[DEBUG] sql - last insert id %d", ins)
	} else {
		log.Printf("[DEBUG] sql - last insert error %s", iErr)
	}

	// Return no error.
	//
	// But since we can't prove different we'll always regard
	// this module as having made a change - just like the
	// shell-execution.
	return true, nil

}

// init is used to dynamically register our module.
func init() {
	Register("sql", func(cfg *config.Config) ModuleAPI {
		return &SQLModule{cfg: cfg}
	})
}
