package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"unicode"

	s "github.com/core-go/sql"
)

type Tables struct {
	Name string `gorm:"column:table"`
}

func ToLower(s string) string {
	if len(s) < 0 {
		return ""
	}
	return string(unicode.ToLower(rune(s[0]))) + s[1:]
}

func ListTables(ctx context.Context, db *sql.DB, database string) ([]string, error) {
	driver := s.GetDriver(db)
	var (
		tables []Tables
		res    []string
	)
	query, err := buildTableQuery(database, driver)
	if err != nil {
		return nil, err
	}
	err = s.Query(ctx, db, nil, &tables, query)
	if err != nil {
		return nil, err
	}
	for i := range tables {
		res = append(res, tables[i].Name)
	}
	return res, err
}

func buildTableQuery(database, driver string) (string, error) {
	switch driver {
	case s.DriverMysql:
		query := `
		SELECT 
    		TABLE_NAME AS 'table'
		FROM
    		information_schema.tables
		WHERE
    		table_schema = '%v'`
		return fmt.Sprintf(query, database), nil
	case s.DriverPostgres:
		return `
		SELECT 
    		table_name as table
		FROM
    		information_schema.tables
		WHERE
    		table_schema='public' AND table_type='BASE TABLE'`, nil
	default:
		return "", errors.New("unsupported driver")
	}
}

func ReformatGoName(s string) string {
	var field strings.Builder
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Println(err)
	}
	tokens := strings.Split(s, "_")
	for _, t := range tokens {
		alphanumericString := reg.ReplaceAllString(t, "")
		field.WriteString(strings.Title(alphanumericString))
	}
	return field.String()
}
