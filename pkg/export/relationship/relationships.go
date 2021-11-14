package relationship

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	s "github.com/core-go/sql"
	edb "github.com/go-generator/core/export/db"
	"strings"
)

// FindRelationships
// 1-1 -> both fields are unique
// 1-n -> only one field is unique
// n-n -> both fields are not unique
// self reference will be in the same table with the same datatype
func FindRelationships(ctx context.Context, db *sql.DB, database string) ([]RelTables, []string, error) {
	rt, err := newRelationshipTables(ctx, db, database)
	if err != nil {
		return nil, nil, err
	}
	jt, err := listAllJoinTablesWithCompositeKey(ctx, db, database, rt)
	for i := range rt {
		rt[i].Relationship, err = findRelationShip(ctx, db, database, jt, &rt[i])
		if err != nil {
			return nil, nil, err
		}
	}
	joinTable, err := listAllJoinTablesWithCompositeKey(ctx, db, database, rt)
	return rt, joinTable, err
}

func GetRelationship(column string, rt []RelTables) *RelTables {
	for _, v := range rt {
		if column == v.ReferencedColumn {
			return &v
		}
	}
	return nil
}

func checkRelation(ctx context.Context, db *sql.DB, check, checkReference bool, database, driver string, rt *RelTables) (string, error) {
	// Already cover the ManyToMany case where a joined table consists of two or more primary key tags that are all foreign keys
	isPrimaryTag, err := checkPrimaryTag(ctx, db, database, driver, rt.Table, rt.Column)
	if err != nil {
		return "", err
	}
	isReferencedPrimaryTag, err := checkPrimaryTag(ctx, db, database, driver, rt.ReferencedTable, rt.ReferencedColumn)
	if err != nil {
		return "", err
	}
	if !checkReference {
		return Unsupported, err
	}
	if check {
		count, err := compositeKeyToString(ctx, db, database, driver, rt.Table)
		if len(count) == 1 { // Only one column has Primary Tag
			if isPrimaryTag && isReferencedPrimaryTag { // Both are Primary key
				return OneToOne, err
			}
			if !isPrimaryTag && isReferencedPrimaryTag { // Column is only a foreign key referenced to other primary key
				return ManyToOne, err
			}
		}
		if len(count) > 1 { // Consist of at least one column that has primary key tag and not referenced to other table
			return OneToMany, err
		}
	}
	if !check {
		return ManyToOne, err
	}
	if !check && !checkReference {
		return Unsupported, err
	}
	return Unknown, err
}

func findRelationShip(ctx context.Context, db *sql.DB, database string, joinedTable []string, rt *RelTables) (string, error) { //TODO: switch gorm to core-go/sql
	driver := s.GetDriver(db)
	check, err := checkUnique(ctx, db, database, driver, rt.Table, rt.Column)
	if err != nil {
		return "", err
	}
	checkReference, err := checkUnique(ctx, db, database, driver, rt.ReferencedTable, rt.ReferencedColumn)
	if err != nil {
		return "", err
	}
	for _, v := range joinedTable {
		if rt.Table == v {
			return ManyToMany, err
		}
	}
	return checkRelation(ctx, db, check, checkReference, database, driver, rt)
}

func checkForeignKey(table, column string, rt []RelTables) bool {
	for _, v := range rt {
		if v.Table == table && v.Column == column {
			return true
		}
	}
	return false
} // Check if the column of the table is a foreign key

func isJoinedTable(table string, columns []string, rt []RelTables) bool {
	for _, v := range columns {
		if checkForeignKey(table, v, rt) == false {
			return false
		}
	}
	return true
}

func newRelationshipTables(ctx context.Context, db *sql.DB, database string) ([]RelTables, error) { // mysql only for now
	driver := s.GetDriver(db)
	var res []RelTables
	query, err := buildQuery(database, driver)
	if err != nil {
		return nil, err
	}
	err = s.Query(ctx, db, nil, &res, query)
	if err != nil {
		return nil, err
	}
	return res, err
} // Find all columns, table and its referenced columns, tables

func listAllJoinTablesWithCompositeKey(ctx context.Context, db *sql.DB, database string, rt []RelTables) ([]string, error) {
	var joinTable []string
	tables, err := edb.ListTables(ctx, db, database)
	if err != nil {
		return nil, err
	}
	for _, v := range tables {
		columns, err := containCompositeKey(ctx, db, database, v)
		if err != nil {
			return nil, err
		}
		if len(columns) > 1 && isJoinedTable(v, getCKeyName(columns), rt) {
			joinTable = append(joinTable, v)
		}
	}
	return joinTable, err
}

func checkUnique(ctx context.Context, db *sql.DB, database, driver, table, column string) (bool, error) {
	var (
		mySqlIndex []MySqlUnique
		pgIndex    []PgUnique
	)
	query, err := buildCheckUnique(database, driver, table)
	if err != nil {
		return false, err
	}
	switch driver {
	case s.DriverPostgres:
		err = s.Query(ctx, db, nil, &pgIndex, query)
		if err != nil {
			return false, err
		}
		for _, v := range pgIndex {
			if strings.Contains(v.IndexName, "unq") {
				tokens := strings.Split(v.IndexName, "_")
				for i := range tokens {
					if tokens[i] == "unq" {
						columnName := strings.Join(tokens[i:], "_")
						if column == columnName {
							return true, err
						}
					}
				}
			}
		}
	case s.DriverMysql:
		err = s.Query(ctx, db, nil, &mySqlIndex, query)
		if err != nil {
			return false, err
		}
		for _, v := range mySqlIndex {
			if v.Column == column {
				if v.NonUnique == false {
					return true, err
				}
			}
		}
	default:
		return false, errors.New(s.DriverNotSupport)
	}
	return false, err
} // Check if a column is unique

func checkPrimaryTag(ctx context.Context, db *sql.DB, database, driver, table, column string) (bool, error) {
	//TODO: Add check primary tag for other relationship
	var mySqlIndex []MySqlUnique
	query, err := buildCheckUnique(database, driver, table)
	if err != nil {
		return false, err
	}
	switch driver {
	case s.DriverMysql:
		err := s.Query(ctx, db, nil, &mySqlIndex, query)
		if err != nil {
			return false, err
		}
	}
	for _, v := range mySqlIndex {
		if v.Column == column {
			if v.Key == "PRIMARY" {
				return true, err
			}
		}
	}
	return false, err
} // Check if a column has primary tag

func containCompositeKey(ctx context.Context, db *sql.DB, database, table string) ([]CompositeKey, error) { // Return a slice of Column of the composite key
	driver := s.GetDriver(db)
	var res []CompositeKey
	query, err := buildCompositeKeyQuery(database, driver, table)
	if err != nil {
		return nil, err
	}
	err = s.Query(ctx, db, nil, &res, query)
	if err != nil {
		return nil, err
	}
	return res, err
}

func getCKeyName(cn []CompositeKey) []string { // Get Column Table of the composite key
	var res []string
	for _, v := range cn {
		res = append(res, v.Column)
	}
	return res
}

func buildQuery(database, driver string) (string, error) {
	switch driver {
	case s.DriverMysql:
		query := `
		SELECT
			TABLE_NAME as 'table',
			COLUMN_NAME as 'column',
			REFERENCED_TABLE_NAME as 'referenced_table',
			REFERENCED_COLUMN_NAME as 'referenced_column'
		FROM
    		information_schema.key_column_usage
		WHERE
    		constraint_schema = '%v'
        	AND referenced_table_schema IS NOT NULL
        	AND referenced_table_name IS NOT NULL
        	AND referenced_column_name IS NOT NULL`
		return fmt.Sprintf(query, database), nil
	case s.DriverPostgres:
		return `
		SELECT
			TC.TABLE_NAME AS table,
			KCU.COLUMN_NAME AS column,
			CCU.TABLE_NAME AS referenced_table,
			CCU.COLUMN_NAME AS referenced_column
		FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS AS TC
		JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE AS KCU ON TC.CONSTRAINT_NAME = KCU.CONSTRAINT_NAME
		AND TC.TABLE_SCHEMA = KCU.TABLE_SCHEMA
		JOIN INFORMATION_SCHEMA.CONSTRAINT_COLUMN_USAGE AS CCU ON CCU.CONSTRAINT_NAME = TC.CONSTRAINT_NAME
		AND CCU.TABLE_SCHEMA = TC.TABLE_SCHEMA
		WHERE TC.CONSTRAINT_TYPE = 'FOREIGN KEY';`, nil
	default:
		return "", errors.New(s.DriverNotSupport)
	}
}

func buildCheckUnique(database, driver, table string) (string, error) {
	switch driver {
	case s.DriverMysql:
		query := `show indexes from %v.%v`
		return fmt.Sprintf(query, database, table), nil
	case s.DriverPostgres:
		query := `SELECT * FROM pg_indexes WHERE tablename = '%v'`
		return fmt.Sprintf(query, table), nil
	default:
		return "", errors.New(s.DriverNotSupport)
	}
}

func buildCompositeKeyQuery(database, driver, table string) (string, error) { //TODO: get composite keys for other databases
	switch driver {
	case s.DriverMysql:
		query := `
			SELECT 
    			COLUMN_NAME as 'column'
			FROM
				information_schema.KEY_COLUMN_USAGE
			WHERE
				table_schema = '%v'
					AND table_name = '%v'
					AND constraint_name = 'PRIMARY';`
		return fmt.Sprintf(query, database, table), nil
	default:
		return "", errors.New(s.DriverNotSupport)
	}
}

func compositeKeyToString(ctx context.Context, db *sql.DB, database, driver, table string) ([]string, error) {
	//TODO: Add get composite key for other relationship
	compositeKeys := make([]CompositeKey, 0)
	switch driver {
	case s.DriverMysql:
		query := `
			SELECT 
				K.COLUMN_NAME as 'column'
			FROM
				INFORMATION_SCHEMA.TABLE_CONSTRAINTS AS C
					JOIN
				INFORMATION_SCHEMA.KEY_COLUMN_USAGE AS K ON C.TABLE_NAME = K.TABLE_NAME
					AND C.CONSTRAINT_CATALOG = K.CONSTRAINT_CATALOG
					AND C.CONSTRAINT_SCHEMA = K.CONSTRAINT_SCHEMA
					AND C.CONSTRAINT_NAME = K.CONSTRAINT_NAME
			WHERE
				C.TABLE_SCHEMA = '%v'
					AND K.TABLE_NAME = '%v'
					AND C.CONSTRAINT_TYPE = 'PRIMARY KEY'`
		query = fmt.Sprintf(query, database, table)
		err := s.Query(ctx, db, nil, &compositeKeys, query)
		if err != nil {
			return nil, err
		}
	}
	return getCKeyName(compositeKeys), nil
}
