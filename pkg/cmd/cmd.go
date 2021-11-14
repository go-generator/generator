package cmd

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"github.com/core-go/cipher"
	s "github.com/core-go/sql"
	"github.com/go-generator/core"
	"github.com/go-generator/core/export"
	edb "github.com/go-generator/core/export/db"
	"github.com/go-generator/core/export/relationship"
	"github.com/go-generator/core/io"
	"log"
	"path/filepath"
	"strconv"
)

var (
	cacheFile, _ = filepath.Abs("./configs/cache.yaml")
)

func ConnectToDatabase(dc metadata.DatabaseConfig) (*sql.DB, error) { //use data source
	driver := dc.Driver
	dsn := dc.DSN
	if len(dsn) == 0 {
		var dataSource string
		port := strconv.FormatInt(dc.Port, 64)
		switch driver {
		case s.DriverMysql:
			dataSource = dc.User + ":" + dc.Password + "@(" + dc.Host + ":" + port + ")/" + dc.Database + "?charset=utf8&parseTime=True&loc=Local"
		case s.DriverPostgres:
			dataSource = "user=" + dc.User + " dbname=" + dc.Database + " password=" + dc.Password + " host=" + dc.Host + " port=" + port + " sslmode=disable"
		case s.DriverMssql:
			dataSource = "sqlserver://" + dc.User + ":" + dc.Password + "@" + dc.Host + ":" + port + "?Database=" + dc.Database
		case s.DriverSqlite3:
			dataSource = dc.Host
		default:
			return nil, errors.New(s.DriverNotSupport)
		}
		return sql.Open(dc.Driver, dataSource)
	} else {
		return sql.Open(driver, dsn)
	}
}
func RunWithCommandLine(types map[string]string) {
	var dbConfig metadata.DatabaseConfig
	ctx := context.TODO()
	outer := make(map[string]string)
	inner := make(map[string]string)
	err := cipher.Read(cacheFile, outer, "password", "")
	dbConfig = metadata.DatabaseConfig{
		Driver:   outer["driver"],
		Host:     outer["host"],
		Port:     3306,
		Database: outer["database"],
		User:     outer["user"],
		Password: outer["password"],
	}
	if err != nil {
		dialectPtr := flag.String("driver", "", "input driver")
		userPtr := flag.String("user", "", "input user")
		passPtr := flag.String("password", "", "input password")
		hostPtr := flag.String("host", "", "input host")
		portPtr := flag.Int("port", 0, "input port")
		dbNamePtr := flag.String("database", "", "input database name")
		flag.Parse()
		inner["driver"] = *dialectPtr
		inner["user"] = *userPtr
		inner["password"] = *passPtr
		inner["host"] = *hostPtr
		inner["port"] = strconv.Itoa(*portPtr)
		inner["database"] = *dbNamePtr
		cacheData := metadata.DatabaseConfig{
			Driver:   *dialectPtr,
			User:     *userPtr,
			Password: *passPtr,
			Host:     *hostPtr,
			Port:     int64(*portPtr),
			Database: *dbNamePtr,
		}
		dbConfig = cacheData
		err = cipher.Write(cacheFile, inner, "password", "")
		if err != nil {
			log.Fatal(err)
		}
	}
	db, err := ConnectToDatabase(dbConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	rl, _, err := relationship.FindRelationships(ctx, db, dbConfig.Database)
	if err != nil {
		log.Fatal(err)
	}
	tables, err := edb.ListTables(ctx, db, dbConfig.Database)
	if err != nil {
		log.Fatal(err)
	}
	models, err := export.ToModels(ctx, db, dbConfig.Database, tables, rl, types)
	if err != nil {
		log.Println(err)
	}
	data := bytes.Buffer{}
	encoder := json.NewEncoder(&data)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(models)
	if err != nil {
		log.Println(err)
	}
	err = io.Save("./models.json", data.Bytes())
	if err != nil {
		log.Println(err)
	}
}
