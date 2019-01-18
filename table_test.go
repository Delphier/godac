package sqlexpress

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTable(t *testing.T) {
	profile := &TableProfile{
		Name:                    "Libraries",
		PrimaryKey:              "LibraryID",
		AutoIncColumn:           "LibraryID",
		ExcludedColumnsOnInsert: "Inactive, ModifiedOn",
		ExcludedColumnsOnUpdate: "Inactive, CreatedOn",
		OtherColumns:            "LibraryCode, ,,LibraryName, , IsGroup",
	}

	table := NewTable(profile)
	js, err := json.MarshalIndent(table, "", strings.Repeat(" ", 2))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("\n%s\n", js)
}
