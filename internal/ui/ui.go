package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	s "github.com/core-go/sql"
	"github.com/go-generator/core"
	"github.com/go-generator/core/build"
	"github.com/go-generator/core/display"
	"github.com/go-generator/core/export"
	edb "github.com/go-generator/core/export/db"
	"github.com/go-generator/core/export/relationship"
	uni "github.com/go-generator/core/export/types"
	"github.com/go-generator/core/generator"
	"github.com/go-generator/core/io"
	"github.com/go-generator/core/list"
	"github.com/go-generator/core/project"
	"github.com/go-generator/core/template"
	"github.com/skratchdot/open-golang/open"
	"github.com/sqweek/dialog"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// WidgetScreen shows a panel containing widget
func WidgetScreen(ctx context.Context, canvas fyne.Canvas, r metadata.Config, dbCache metadata.Database) fyne.CanvasObject {
	var files []metadata.File

	funcMap := template.MakeFuncMap()

	tmplPath, err := filepath.Abs(filepath.Join(".", r.Template))
	if err != nil {
		log.Fatal(err)
	}
	langTmpl, err := io.LoadAll(tmplPath)
	if err != nil {
		log.Fatal(err)
	}
	pPath, err := filepath.Abs(filepath.Join(".", r.ProjectPath))
	if err != nil {
		log.Fatal(err)
	}
	projTmpl, err := io.Load(pPath)
	if err != nil {
		log.Fatal(err)
	}
	prjTypes, err := io.List(pPath)
	if err != nil {
		log.Fatal(err)
	}

	if len(prjTypes) < 1 {
		log.Fatal("no project type found")
	}
	prjTmplName := prjTypes[0]
	prjTmplNameEntry := widget.NewSelect(prjTypes, func(o string) {
		prjTmplName = o
	})

	projectName := widget.NewEntry()
	projectName.SetText(r.ProjectName)
	projectName.SetPlaceHolder("Package name goes here...")

	largeText := widget.NewMultiLineEntry()
	largeText.SetPlaceHolder("Input")
	outputJson := widget.NewMultiLineEntry()
	outputJson.PlaceHolder = "Generated Code Here"

	projectJsonInput := widget.NewMultiLineEntry()
	projectJsonInput.SetPlaceHolder("Input Project JSON here...")

	outputWindows := widget.NewMultiLineEntry()
	outputWindows.SetPlaceHolder("Files Content Goes Here....")
	outputWindows.TextStyle = fyne.TextStyle{
		Bold:      false,
		Italic:    true,
		Monospace: true,
		TabWidth:  0,
	}

	var driverEntry *widget.RadioGroup
	driverEntry = widget.NewRadioGroup([]string{
		s.DriverMysql,
		s.DriverMssql,
		s.DriverSqlite3,
		s.DriverOracle,
		s.DriverPostgres,
	}, func(s string) {
		driverEntry.SetSelected(s)
	})
	driverEntry.Horizontal = true
	driverEntry.SetSelected(s.DriverMysql)

	dsnSourceEntry := widget.NewEntry()
	dsnSourceEntry.OnChanged = func(dsn string) {
		switch driverEntry.Selected {
		case s.DriverMysql:
			dbCache.MySql = dsn
		case s.DriverPostgres:
			dbCache.Postgres = dsn
		case s.DriverMssql:
			dbCache.Mssql = dsn
		case s.DriverSqlite3:
			dbCache.Sqlite3 = dsn
		case s.DriverOracle:
			dbCache.Oracle = dsn
		}
	}
	dsnSourceEntry.SetText(project.SelectDSN(dbCache, driverEntry.Selected))
	driverEntry.OnChanged = func(driver string) {
		dsnSourceEntry.SetText(project.SelectDSN(dbCache, driver))
	}

	openProjectPath := ""
	openButton := widget.NewButtonWithIcon("Open Output Folder", theme.FolderOpenIcon(), func() {
		if openProjectPath == "" {
			display.PopUpWindows("empty project path", canvas)
			return
		}
		err = open.Run(openProjectPath)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
	})

	data := binding.BindStringList(
		&[]string{},
	)
	dataStruct := binding.BindStruct(&metadata.File{})
	listWithData := widget.NewListWithData(data,
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		})
	listWithData.OnSelected = func(id widget.ListItemID) {
		keys := dataStruct.Keys()
		for i := range keys {
			if strings.HasPrefix(keys[i], strconv.Itoa(id)+".") {
				v, err := dataStruct.GetValue(keys[i])
				if err != nil {
					display.PopUpWindows(err.Error(), canvas)
					return
				}
				if content, ok := v.(string); ok {
					if runtime.GOOS == "windows" {
						content = strings.ReplaceAll(content, "\r", "")
					} else {
						content = strings.ReplaceAll(content, "\n", "")
					}
					outputWindows.SetText(content)
				} else {
					display.PopUpWindows("error setting content", canvas)
				}
			}
		}
	}

	reloadAllTemplate := widget.NewButtonWithIcon("Reload Template", theme.ContentRedoIcon(), func() {
		projTmpl, err = io.Load(pPath)
		if err != nil {
			log.Fatal(err)
		}
		prjTypes, err = io.List(pPath)
		if err != nil {
			log.Fatal(err)
		}
		langTmpl, err = io.LoadAll(tmplPath)
		if err != nil {
			log.Fatal(err)
		}
		outputWindows.SetText("")
		listWithData.UnselectAll()
	})

	loadPrJson := widget.NewButtonWithIcon("Generate From File", theme.DocumentCreateIcon(), func() {
		var (
			prj   metadata.Project
			pData bytes.Buffer
		)
		enc := json.NewEncoder(&pData)
		enc.SetIndent("", "   ")
		filename, err := dialog.File().Title("Load Metadata File").Filter("All Files").Load()
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		} else {
			dt, err := ioutil.ReadFile(filename)
			if err != nil {
				display.PopUpWindows(err.Error(), canvas)
				return
			}
			prj, err = generator.DecodeProject(dt, projectName.Text, build.InitEnv)
			if err != nil {
				display.PopUpWindows(err.Error(), canvas)
				return
			}
			_, ok := prj.Env["go_module"]
			if ok && projectName.Text != "" {
				prj.Env["go_module"] = projectName.Text
			}
			err = enc.Encode(&prj)
			if err != nil {
				display.PopUpWindows(err.Error(), canvas)
				return
			}
			projectJsonInput.SetText(pData.String())
		}
	})

	var optimizeEntry *widget.Check
	optimizeEntry = widget.NewCheck("Models Only", func(b bool) {
		optimizeEntry.SetChecked(b)
	})
	optimizeEntry.SetChecked(false)

	connectGenerate := widget.NewButtonWithIcon("Generate Project JSON", theme.DocumentCreateIcon(), func() {
		var (
			toModels []metadata.Model
			prj      *metadata.Project
			pData    bytes.Buffer
		)
		sqlDB, err := project.ConnectDB(dbCache, driverEntry.Selected)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
		}
		defer func() {
			err = sqlDB.Close()
			if err != nil {
				display.PopUpWindows(err.Error(), canvas)
			}
		}()
		enc := json.NewEncoder(&pData)
		enc.SetIndent("", "   ")
		dbName, err := project.GetDatabaseName(dbCache, driverEntry.Selected)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}

		rt, _, err := relationship.FindRelationships(ctx, sqlDB, dbName)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		tables, err := edb.ListTables(ctx, sqlDB, dbName)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		toUniTypes := uni.Types[driverEntry.Selected]
		toModels, err = export.ToModels(ctx, sqlDB, dbName, tables, rt, toUniTypes)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		if optimizeEntry.Checked {
			err = enc.Encode(&toModels)
			if err != nil {
				display.PopUpWindows(err.Error(), canvas)
				return
			}
		} else {
			prj, err = generator.ExportProject(prjTmplName, projectName.Text, projTmpl, toModels, build.InitEnv)
			if err != nil {
				display.PopUpWindows(err.Error(), canvas)
				return
			}
			_, ok := prj.Env["go_module"]
			if ok && projectName.Text != "" {
				prj.Env["go_module"] = projectName.Text
			}
			err = enc.Encode(&prj)
			if err != nil {
				display.PopUpWindows(err.Error(), canvas)
				return
			}
		}
		cache, err := yaml.Marshal(dbCache)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		err = io.Save(filepath.Join(".", r.DBCache), cache)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		projectJsonInput.SetText(pData.String())
		display.Notify("Success", "Generate Project JSON")
	})

	saveProjectJson := widget.NewButtonWithIcon("Save JSON", theme.DocumentSaveIcon(), func() {
		outFile, err := dialog.File().Filter("json", ".json").Title("Save As").Save()
		if err == dialog.ErrCancelled {
			display.PopUpWindows(dialog.ErrCancelled.Error(), canvas)
			return
		}
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		if filepath.Ext(outFile) != ".json" {
			outFile += ".json"
		}
		err = io.SaveContent(outFile, projectJsonInput.Text)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		display.Notify("Success", "Save JSON")
	})

	generateButton := widget.NewButtonWithIcon("Generate Output", theme.DocumentCreateIcon(), func() {
		newDataStruct := binding.BindStruct(&metadata.File{})
		files, err = generator.GenerateFiles(projectName.Text, projectJsonInput.Text, funcMap, langTmpl)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		err = data.Set(nil)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		err = list.ShowFiles(data, newDataStruct, files)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		dataStruct = newDataStruct // reload list
		display.Notify("Success", "Generate Output")
	})

	generateProject := widget.NewButtonWithIcon("Save Project", theme.DocumentSaveIcon(), func() {
		if len(files) < 1 {
			display.PopUpWindows("output files list is empty", canvas)
			return
		}
		directory, err := dialog.Directory().Title("Save Project Files In...").Browse()
		directory = directory + string(os.PathSeparator) + projectName.Text
		err = io.MkDir(directory)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		openProjectPath = directory
		err = io.SaveFiles(directory, files)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		display.Notify("Success", "Save Project")
	})

	testDsn := widget.NewButtonWithIcon("Test Connection", theme.CheckButtonCheckedIcon(), func() {
		oldDsn := project.SelectDSN(dbCache, driverEntry.Selected)
		if strings.Compare(oldDsn, dsnSourceEntry.Text) != 0 {
			project.UpdateDBCache(&dbCache, driverEntry.Selected, dsnSourceEntry.Text)
		}
		db, err := project.ConnectDB(dbCache, driverEntry.Selected)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		defer func() {
			err = db.Close()
			if err != nil {
				display.PopUpWindows(err.Error(), canvas)
			}
		}()
		err = db.Ping()
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		display.Notify("Success", "Test Connection")
	})

	prScroll := container.NewScroll(projectJsonInput)
	outScroll := container.NewScroll(outputWindows)

	bottomButtons := container.NewVBox(
		container.NewAdaptiveGrid(2,
			testDsn,
			connectGenerate,
			generateButton,
			loadPrJson,
			saveProjectJson,
			generateProject, reloadAllTemplate, openButton),
	)

	tabs := container.NewVBox(
		container.NewAdaptiveGrid(2, widget.NewLabel("Project:"),
			widget.NewLabel("Package:"),
			prjTmplNameEntry,
			projectName), driverEntry, dsnSourceEntry)

	cb1 := container.NewBorder(tabs, bottomButtons, nil, nil, listWithData)

	cb2 := container.NewHSplit(
		prScroll,
		outScroll,
	)
	return container.NewBorder(nil, nil, cb1, nil, cb2)
}
