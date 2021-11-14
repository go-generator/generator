// Package db Deprecated
package db

import (
	"context"
	"database/sql"
	"fmt"
	s "github.com/core-go/sql"
)

type TableFields struct {
	Column     string         `gorm:"column:column_name"`
	DataType   string         `gorm:"column:type"`
	IsNullable string         `gorm:"column:is_nullable"`
	ColumnKey  string         `gorm:"column:column_key"`
	Length     sql.NullString `gorm:"column:length"`
}

type TableInfo struct {
	Fields          []TableFields
	HasCompositeKey bool
}

//func StandardizeFieldsName(st *TableInfo) {
//	var count int
//	for _, v := range st.Fields {
//		if v.ColumnKey == "PRI" {
//			count++
//		}
//		st.GoFields = append(st.GoFields, ReformatGoName(v.Column))
//	}
//	if count < 2 {
//		st.HasCompositeKey = false
//	} else {
//		st.HasCompositeKey = true
//	}
//}

func HasCompositeKey(st []TableFields) bool {
	var count int
	for _, v := range st {
		if v.ColumnKey == "PRI" {
			count++
		}
	}
	return count < 2
}

func InitTables(ctx context.Context, db *sql.DB, database, table string, st *TableInfo) error {
	switch s.GetDriver(db) {
	case s.DriverMysql:
		query := `
			SELECT 
				TABLE_NAME AS 'table',
				COLUMN_NAME AS 'column_name',
				DATA_TYPE AS 'type',
				IS_NULLABLE AS 'is_nullable',
				COLUMN_KEY AS 'column_key',
				CHARACTER_MAXIMUM_LENGTH AS 'length'
			FROM
				information_schema.columns
			WHERE
				TABLE_SCHEMA = '%v'
					AND TABLE_NAME = '%v'`
		query = fmt.Sprintf(query, database, table)
		err := s.Query(ctx, db, nil, &st.Fields, query)
		if err != nil {
			return err
		}
		st.HasCompositeKey = HasCompositeKey(st.Fields)
	}
	return nil
}
