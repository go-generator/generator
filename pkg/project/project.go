package project

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/core-go/cipher"
	s "github.com/core-go/sql"
	"github.com/go-generator/core"
	"github.com/go-generator/core/display"
	"github.com/go-generator/core/export"
	edb "github.com/go-generator/core/export/db"
	"github.com/go-generator/core/export/relationship"
	"github.com/go-generator/core/generator"
	"github.com/go-generator/core/io"
	"github.com/sqweek/dialog"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	BasicAuth  = "Basic"
	Datasource = "Datasource"
)

var (
	cacheFile, _ = filepath.Abs("./configs/cache.yaml")
	dbFile, _    = filepath.Abs("./configs/database.yaml")
)

const (
	ErrInvDialect = "invalid dialect"
	ErrInvUser    = "invalid user"
	ErrInvPass    = "invalid password"
	ErrInvAddr    = "invalid host address"
	ErrInvPort    = "invalid port"
	ErrInvDBName  = "invalid relationship name"
)

func BuildDataSourceName(c metadata.DatabaseConfig) string {
	if c.Driver == "postgres" {
		uri := fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%d sslmode=disable", c.User, c.Database, c.Password, c.Host, c.Port)
		return uri
	} else if c.Driver == "mysql" {
		uri := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", c.User, c.Password, c.Host, c.Port, c.Database)
		return uri
	} else if c.Driver == "mssql" { // mssql
		uri := fmt.Sprintf("sqlserver://%s:%s@%s:%d?Database=%s", c.User, c.Password, c.Host, c.Port, c.Database)
		return uri
	} else if c.Driver == "godror" || c.Driver == "oracle" {
		return fmt.Sprintf("user=\"%s\" password=\"%s\" connectString=\"%s:%d/%s\"", c.User, c.Password, c.Host, c.Port, c.Database)
	} else { //sqlite
		return c.Host // return sql.Open("sqlite3", c.Host)
	}
}
func ConnectToDatabase(c metadata.DatabaseConfig, driver string, dbCache metadata.Database, authType string) (*sql.DB, error) { //use data source
	switch authType {
	case "Basic":
		dataSource := BuildDataSourceName(c)
		return sql.Open(c.Driver, dataSource)
	default:
		switch driver {
		case "mysql":
			return sql.Open("mysql", dbCache.MySql)
		case "postgres":
			return sql.Open("postgres", dbCache.Postgres)
		case "mssql":
			return sql.Open("mssql", dbCache.Mssql)
		case "sqlite3":
			return sql.Open("sqlite3", dbCache.Sqlite3)
		case "godror":
			return sql.Open("godror", dbCache.Oracle)
		case "oracle":
			return sql.Open("godror", dbCache.Oracle)
		default:
			return nil, errors.New(s.DriverNotSupport)
		}
	}
}

func ValidateDatabaseConfig(dc metadata.DatabaseConfig) error {
	var err error
	if dc.Driver == "" {
		err = errors.New(ErrInvDialect)
	}
	if dc.User == "" {
		err = errors.New(ErrInvUser)
	}
	if dc.Password == "" {
		err = errors.New(ErrInvPass)
	}
	if dc.Host == "" {
		err = errors.New(ErrInvAddr)
	}
	if dc.Database == "" {
		err = errors.New(ErrInvDBName)
	}
	return err
}

func RunWithUI(ctx context.Context, app fyne.App, size fyne.Size, types map[string]string, projectTmpl, prTmplName, projectName string, dbCache metadata.Database, authType string) error {
	var dbConfig metadata.DatabaseConfig
	outer := make(map[string]string, 0)
	encryptField := "password"
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) { // Check if configs file exists, if not then create one
		err = ioutil.WriteFile(cacheFile, nil, 0666)
		if err != nil {
			return err
		}
	}
	fs, err := os.Stat(cacheFile)
	if err != nil {
		return err
	}
	if fs.Size() != 0 { // Check if configs file is empty
		err = cipher.Read(cacheFile, outer, encryptField, "")
		if err != nil {
			display.ShowErrorWindows(app, err, size)
		}
	}
	port, err := strconv.ParseInt(outer["port"], 10, 64)
	if err != nil {
		return err
	}
	dbConfig = metadata.DatabaseConfig{
		Driver:   outer["driver"],
		Host:     outer["host"],
		Port:     port,
		Database: outer["database"],
		User:     outer["user"],
		Password: outer["password"],
	}
	err = DriverInputUI(ctx, dbConfig, app, projectTmpl, prTmplName, projectName, encryptField, size, types, dbCache, authType)
	return err
}

func DriverInputUI(ctx context.Context, dc metadata.DatabaseConfig, app fyne.App, projectTmpl, prTmplName, projectName, encryptField string, size fyne.Size, types map[string]string, dbCache metadata.Database, authType string) error {
	oldCache := dbCache
	driverEntry := widget.NewRadioGroup([]string{s.DriverMysql, s.DriverPostgres, s.DriverMssql, s.DriverSqlite3}, func(s string) {
		dc.Driver = s
	})
	driverEntry.Selected = s.DriverMysql

	codeEntry := widget.NewMultiLineEntry()
	codeEntry.Wrapping = fyne.TextWrapWord

	outer := make(map[string]string, 0)
	inner := make(map[string]string, 0)

	err := cipher.Read(cacheFile, outer, encryptField, "")
	if err != nil {
		return err
	}
	port, err := strconv.ParseInt(outer["port"], 10, 64)
	if err != nil {
		return err
	}
	databaseConfig := metadata.DatabaseConfig{
		Driver:   outer["driver"],
		Host:     outer["host"],
		Port:     port,
		Database: outer["database"],
		User:     outer["user"],
		Password: outer["password"],
	}

	optimize := true
	optimizeEntry := widget.NewCheck("Models Only", func(b bool) {
		optimize = b
	})
	optimizeEntry.Checked = true
	usernameEntry := widget.NewEntry()
	usernameEntry.OnChanged = func(u string) {
		dc.User = u
	}
	usernameEntry.Text = dc.User
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.OnChanged = func(p string) {
		dc.Password = p
	}
	passwordEntry.Text = dc.Password
	passwordEntry.Hidden = true
	hostEntry := widget.NewEntry()
	hostEntry.OnChanged = func(h string) {
		dc.Host = h
	}
	hostEntry.Text = dc.Host
	portEntry := widget.NewEntry()
	portEntry.Text = strconv.FormatInt(dc.Port, 10)
	portEntry.OnChanged = func(s string) {
		if s == "" {
			return
		}
		tmp, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			display.ShowErrorWindows(app, err, size)
			portEntry.SetText(strconv.FormatInt(dc.Port, 10))
			return
		}
		if tmp < 1 {
			display.ShowErrorWindows(app, errors.New(ErrInvPort), size)
			portEntry.SetText(strconv.FormatInt(dc.Port, 100))
			return
		}
		dc.Port = tmp
		portEntry.SetText(strconv.FormatInt(dc.Port, 10))
	}
	databaseEntry := widget.NewEntry()
	databaseEntry.OnChanged = func(d string) {
		dc.Database = d
	}
	databaseEntry.Text = dc.Database

	window := app.NewWindow("Database Metadata JSON Generator")
	window.Resize(fyne.Size{
		Width: 640,
	})

	basicAuth := container.NewVBox(
		widget.NewLabel("Database:"),
		databaseEntry,
		widget.NewLabel("Driver:"),
		driverEntry,
		widget.NewLabel("User:"),
		usernameEntry,
		widget.NewLabel("Password:"),
		passwordEntry,
		widget.NewLabel("Host:"),
		hostEntry,
		widget.NewLabel("Port:"),
		portEntry,
	)

	var (
		toModels []metadata.Model
		prj      *metadata.Project
		pData    bytes.Buffer
	)
	enc := json.NewEncoder(&pData)
	enc.SetIndent("", "    ")

	executeButton := widget.NewButton("Generate Database JSON Description", func() {
		db, err := ConnectToDatabase(dc, driverEntry.Selected, dbCache, authType)
		if err != nil {
			display.ShowErrorWindows(app, err, size)
			return
		}
		defer func() {
			err = db.Close()
			if err != nil {
				display.ShowErrorWindows(app, err, size)
			}
		}()
		err = ValidateDatabaseConfig(dc)
		if err != nil {
			display.ShowErrorWindows(app, err, size)
			return
		}
		dbName := ""
		switch authType {
		case BasicAuth:
			dbName = dc.Database
		case Datasource:
			dbName, err = getDatabaseName(dbCache, driverEntry.Selected)
			if err != nil {
				display.ShowErrorWindows(app, err, size)
				return
			}
		}
		rt, _, err := relationship.FindRelationships(ctx, db, dbName)
		if err != nil {
			display.ShowErrorWindows(app, err, size)
			return
		}
		tables, err := edb.ListTables(ctx, db, dbName)
		if err != nil {
			display.ShowErrorWindows(app, err, size)
			return
		}
		toModels, err = export.ToModels(ctx, db, dbName, tables, rt, types)
		if err != nil {
			display.ShowErrorWindows(app, err, size)
			return
		}
		if optimize {
			err = enc.Encode(&toModels)
			if err != nil {
				display.ShowErrorWindows(app, err, size)
				return
			}
		} else {
			prj, err = generator.ExportProject(projectTmpl, io.Load, prTmplName, projectName, toModels, generator.InitEnv)
			if err != nil {
				display.ShowErrorWindows(app, err, size)
				return
			}
			err = enc.Encode(&prj)
			if err != nil {
				display.ShowErrorWindows(app, err, size)
				return
			}
		}
		codeEntry.SetText(pData.String())
		codeEntry.Refresh()
		switch authType {
		case BasicAuth:
			if databaseConfig != dc {
				mapConfig(inner, dc)
				err = cipher.Write(cacheFile, inner, encryptField, "")
				if err != nil {
					display.ShowErrorWindows(app, err, size)
					return
				}
			}
		case Datasource:
			if oldCache != dbCache {
				data, err := yaml.Marshal(&dbCache)
				if err != nil {
					display.ShowErrorWindows(app, err, size)
					return
				}
				err = ioutil.WriteFile(dbFile, data, 0664)
				if err != nil {
					display.ShowErrorWindows(app, err, size)
					return
				}
			}
		}
	})

	saveProject := widget.NewButton("Save Codes", func() {
		outFile, err := dialog.File().Filter("json", ".json").Title("Save As").Save()
		if err != nil {
			display.ShowErrorWindows(app, err, size)
			return
		}
		err = io.SaveContent(outFile, codeEntry.Text)
		if err != nil {
			display.ShowErrorWindows(app, err, size)
			return
		}
		//if optimize {
		//	err = project.SaveModels(toModels, outFile)
		//	if err != nil {
		//		display.ShowErrorWindows(app, err, size)
		//		return
		//	}
		//} else {
		//	pr, err := export.ToProject(projectTmpl, prTmplName, projectName, toModels)
		//	if err != nil {
		//		display.ShowErrorWindows(app, err, size)
		//		return
		//	}
		//	err = project.SaveProject(*pr, outFile)
		//	if err != nil {
		//		display.ShowErrorWindows(app, err, size)
		//		return
		//	}
		//}
	})
	codeWidows := app.NewWindow("Generated Code")
	codeWidows.Resize(display.ResizeWindows(15, 20, size))
	codeWidows.SetContent(container.NewBorder(nil, saveProject, nil, nil, container.NewScroll(codeEntry)))
	codeWidows.Show()

	auth := container.NewVBox()
	if authType == BasicAuth {
		auth = basicAuth
	}
	if authType == Datasource {
		auth = datasource(dbCache, driverEntry)
	}
	window.SetContent(container.NewBorder(optimizeEntry, executeButton, nil, nil,
		auth,
	))
	window.SetOnClosed(func() {
		codeWidows.Close()
	})
	window.CenterOnScreen()
	window.Show()
	return err
}

func mapConfig(inner map[string]string, databaseConfig metadata.DatabaseConfig) {
	inner["driver"] = databaseConfig.Driver
	inner["user"] = databaseConfig.User
	inner["password"] = databaseConfig.Password
	inner["host"] = databaseConfig.Host
	inner["port"] = strconv.FormatInt(databaseConfig.Port, 64)
	inner["database"] = databaseConfig.Database
}

func datasource(dbCache metadata.Database, driverEntry *widget.RadioGroup) *fyne.Container {
	datasourceEntry := widget.NewEntry()
	switch driverEntry.Selected {
	case s.DriverMysql:
		datasourceEntry.Text = dbCache.MySql
	case s.DriverPostgres:
		datasourceEntry.Text = dbCache.Postgres
	case s.DriverSqlite3:
		datasourceEntry.Text = dbCache.Sqlite3
	case s.DriverMssql:
		datasourceEntry.Text = dbCache.Mssql
	case s.DriverOracle:
		datasourceEntry.Text = dbCache.Oracle
	}

	datasourceEntry.OnChanged = func(source string) {
		datasourceEntry.SetText(source)
		switch driverEntry.Selected {
		case s.DriverMysql:
			dbCache.MySql = datasourceEntry.Text
		case s.DriverPostgres:
			dbCache.Postgres = datasourceEntry.Text
		case s.DriverSqlite3:
			dbCache.Sqlite3 = datasourceEntry.Text
		case s.DriverMssql:
			dbCache.Mssql = datasourceEntry.Text
		case s.DriverOracle:
			dbCache.Oracle = datasourceEntry.Text
		}
	}

	driverEntry.OnChanged = func(opt string) {
		switch opt {
		case s.DriverMysql:
			datasourceEntry.Text = dbCache.MySql
		case s.DriverPostgres:
			datasourceEntry.Text = dbCache.Postgres
		case s.DriverSqlite3:
			datasourceEntry.Text = dbCache.Sqlite3
		case s.DriverMssql:
			datasourceEntry.Text = dbCache.Mssql
		case s.DriverOracle:
			datasourceEntry.Text = dbCache.Oracle
		}
		datasourceEntry.Refresh()
	} //refresh UI

	return container.NewVBox(
		widget.NewLabel("Driver:"),
		driverEntry,
		widget.NewLabel("Datasource:"),
		datasourceEntry,
	)
}

func getDatabaseName(dbCache metadata.Database, driver string) (string, error) {
	switch driver {
	case s.DriverMysql:
		s1 := strings.Split(dbCache.Sqlite3, "/")
		if len(s1) < 2 {
			return "", errors.New("invalid datasource")
		}
		s2 := strings.Split(s1[1], "?")
		return s2[0], nil
	case s.DriverPostgres:
		s1 := strings.Split(dbCache.Sqlite3, "dbname=")
		if len(s1) < 2 {
			return "", errors.New("invalid datasource")
		}
		s2 := strings.Split(s1[1], " ")
		return s2[0], nil
	case s.DriverMssql:
		s1 := strings.Split(dbCache.Sqlite3, "database=")
		if len(s1) < 2 {
			return "", errors.New("invalid datasource")
		}
		s2 := strings.Split(s1[1], "&")
		return s2[0], nil
	case s.DriverSqlite3:
		return filepath.Base(dbCache.Sqlite3), nil
	default:
		return "", errors.New(s.DriverNotSupport)
	}
}
