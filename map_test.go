package godac

import (
	"database/sql"
	"encoding/json"
	"log"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("mysql", "root@/zils?charset=utf8mb4&multiStatements=true")
	if err != nil {
		log.Fatal(err)
	}
}
func TestMapQuery(t *testing.T) {
	result, err := MapQuery(db, "SELECT * FROM Libraries")
	if err != nil {
		t.Fatal(err)
	}

	js, err := json.MarshalIndent(result, "", strings.Repeat(" ", 2))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("\n%s\n", js)
}
