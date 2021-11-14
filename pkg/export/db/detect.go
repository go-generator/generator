package db

import "strings"

func DetectDriver(s string) string {
	if strings.Index(s, "sqlserver:") == 0 {
		return "mssql"
	} else {
		if strings.Index(s, "user=") >= 0 && strings.Index(s, "password=") >= 0 {
			if strings.Index(s, "dbname=") >= 0 || strings.Index(s, "host=") >= 0 || strings.Index(s, "port=") >= 0{
				return "postgres"
			} else {
				return "godror"
			}
		} else {
			if strings.Index(s, "@tcp(") >= 0 || strings.Index(s, "charset=") > 0 || strings.Index(s, "parseTime=") > 0 || strings.Index(s, "loc=") > 0 || strings.Index(s, "@") >= 0 || strings.Index(s, ":") >= 0 {
				return "mysql"
			} else {
				return "sqlite3"
			}
		}
	}
}
