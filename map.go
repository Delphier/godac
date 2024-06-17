package godac

import (
	"database/sql"
)

// MapQuery fetching rows to []Map. keys define the Map key of columns.
func MapQuery(keys map[string]string, db DB, query string, args ...interface{}) ([]Map, error) {
	return mapQuery(false, keys, db, query, args...)
}

// MapQueryRow fetching first row to Map. keys define the Map key of columns.
func MapQueryRow(keys map[string]string, db DB, query string, args ...interface{}) (Map, error) {
	maps, err := mapQuery(true, keys, db, query, args...)
	if err != nil || len(maps) == 0 {
		return nil, err
	}
	return maps[0], nil
}

// Set firstOnly is true to return the first row only.
func mapQuery(firstOnly bool, keys map[string]string, db DB, query string, args ...interface{}) ([]Map, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	var result []Map
	dest := makeScanDest(colTypes)
	for rows.Next() {
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		row := make(Map)
		for i := range dest {
			name := colTypes[i].Name()
			key := keys[name]
			if key == "" {
				key = convertName(name)
			}
			value := dest[i]
			switch value.(type) {
			case *sql.NullInt64:
				if v := value.(*sql.NullInt64); v.Valid {
					row[key] = v.Int64
				}
			case *sql.NullFloat64:
				if v := value.(*sql.NullFloat64); v.Valid {
					row[key] = v.Float64
				}
			case *sql.NullBool:
				if v := value.(*sql.NullBool); v.Valid {
					row[key] = v.Bool
				}
			case *sql.NullString:
				if v := value.(*sql.NullString); v.Valid {
					row[key] = v.String
				}
			default:
				row[key] = *value.(*interface{})
			}
		}
		result = append(result, row)
		if firstOnly {
			break
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// 如果用 interface{} 做为目标，驱动端返回的是 []byte，无法判断具体类型；
// 如果用 int32, int64, string 等做为目标，如果有的列为 NULL 值，则会报错；
// VARCHAR 和 BINARY 列的 ScanType 都是 sql.RawBytes，所以不能用 ScanType 来判断。
func makeScanDest(colTypes []*sql.ColumnType) []interface{} {
	dest := make([]interface{}, len(colTypes))
	for i, col := range colTypes {
		//fmt.Println(col.Name(), col.DatabaseTypeName(), col.ScanType())
		switch col.DatabaseTypeName() {
		case "INT", "TINYINT", "SMALLINT", "BIGINT":
			dest[i] = new(sql.NullInt64)
		case "FLOAT", "DOUBLE", "REAL":
			dest[i] = new(sql.NullFloat64)
		case "BOOL", "BOOLEAN":
			dest[i] = new(sql.NullBool)
		case "VARCHAR", "CHAR", "TEXT", "JSON", "DECIMAL",
			"DATETIME", "DATE", "TIME":
			dest[i] = new(sql.NullString)
		default:
			dest[i] = new(interface{})
		}
	}
	return dest
}
