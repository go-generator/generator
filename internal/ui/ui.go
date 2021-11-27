package ui

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
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
	"github.com/go-generator/core/loader"
	"github.com/go-generator/core/project"
	"github.com/go-generator/core/types"
	"github.com/sqweek/dialog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func notifySuccess(content string) {
	fyne.CurrentApp().SendNotification(fyne.NewNotification("Message", content))
}

// WidgetScreen shows a panel containing widget
func WidgetScreen(ctx context.Context, canvas fyne.Canvas, r metadata.Config, dbCache metadata.Database) fyne.CanvasObject {
	tmplPath, err := filepath.Abs(filepath.Join(".", r.TemplatePath))
	if err != nil {
		log.Fatal(err)
	}
	langTmpl, err := LoadAllLanguageTemplates(tmplPath)
	if err != nil {
		log.Fatal(err)
	}
	pPath, err := filepath.Abs(filepath.Join(".", r.PrjTmplPath))
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

	prjTypeName := r.PrjTmplName
	prjTmplNameEntry := widget.NewSelect(prjTypes, func(o string) {
		prjTypeName = o
	})
	prjTmplNameEntry.Selected = r.PrjTmplName

	projectName := widget.NewEntry()
	projectName.SetText(r.ProjectName)
	projectName.SetPlaceHolder("Project Name:")

	largeText := widget.NewMultiLineEntry()
	largeText.SetPlaceHolder("Input")
	outputJson := widget.NewMultiLineEntry()
	outputJson.PlaceHolder = "Generated Code Here"

	outputWindows := widget.NewMultiLineEntry()
	outputWindows.SetPlaceHolder("Files Content Goes Here....")
	outputWindows.TextStyle = fyne.TextStyle{
		Bold:      false,
		Italic:    true,
		Monospace: true,
		TabWidth:  0,
	}
	outputWindows.Wrapping = fyne.TextWrapBreak

	var driverEntry *widget.RadioGroup
	driverEntry = widget.NewRadioGroup([]string{s.DriverMysql, s.DriverMssql, s.DriverSqlite3, s.DriverPostgres, s.DriverOracle}, func(s string) {
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

	var output metadata.Output
	loadPrJson := widget.NewButton("Load Metadata File and Generate", func() {
		filename, err := dialog.File().Title("Load Metadata File").Filter("All Files").Load()
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		} else {
			output, err = generator.GenerateFromFile(langTmpl, projectName.Text, filename, loader.LoadProject, generator.InitEnv, build.BuildModel)
			if err != nil {
				display.PopUpWindows(err.Error(), canvas)
				return
			} else {
				result, err := generator.ToString(output.Files)
				if err != nil {
					display.PopUpWindows(err.Error(), canvas)
					return
				}
				outputWindows.SetText(result)
				notifySuccess("Generate Successfully")
				//directory, err := dialog.Directory().Title("Save Project Files In...").Browse()
				//if output.Directory != "" {
				//	directory = directory + string(os.PathSeparator) + output.Directory
				//	err = io.MkDir(directory)
				//	if err != nil {
				//		return
				//	}
				//}
				//err1 := io.SaveOutput(directory, output)
				//if err1 != nil {
				//	display.PopUpWindows(err1.Error(), canvas)
				//	return
				//}
			}
		}
	})

	projectJsonInput := widget.NewMultiLineEntry()
	projectJsonInput.SetPlaceHolder("Input Project JSON here...")
	projectJsonInput.Wrapping = fyne.TextWrapWord

	var optimizeEntry *widget.Check
	optimizeEntry = widget.NewCheck("Models Only", func(b bool) {
		optimizeEntry.SetChecked(b)
	})
	optimizeEntry.SetChecked(false)

	data := binding.BindStringList(
		&[]string{},
	)

	dataStruct := binding.BindStruct(&metadata.File{})

	list := widget.NewListWithData(data,
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		})

	list.OnSelected = func(id widget.ListItemID) {
		keys := dataStruct.Keys()
		for i := range keys {
			if strings.HasPrefix(keys[i], strconv.Itoa(id)+".") {
				v, err := dataStruct.GetValue(keys[i])
				if err != nil {
					display.PopUpWindows(err.Error(), canvas)
					return
				}
				if content, ok := v.(string); ok {
					var b strings.Builder
					b.WriteString(content)
					outputWindows.SetText(b.String())
				} else {
					display.PopUpWindows("error setting content", canvas)
				}
			}
		}
	}

	connectGenerate := widget.NewButton("Connect and Generate", func() {
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
		t := uni.Types[driverEntry.Selected]
		toModels, err = export.ToModels(ctx, sqlDB, dbName, tables, rt, t)
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
			prj, err = generator.ExportProject(projTmpl, r.PrjTmplName, projectName.Text, toModels, generator.InitEnv)
			if err != nil {
				display.PopUpWindows(err.Error(), canvas)
				return
			}
			exTypes, ok := types.Types[prj.Language]
			if !ok {
				log.Fatal("missing export type for current language")
			}
			for i := range prj.Models {
				for j := range prj.Models[i].Fields {
					if _, ok := exTypes[prj.Models[i].Fields[j].Type]; !ok {
						continue
					}
					prj.Models[i].Fields[j].Type = exTypes[prj.Models[i].Fields[j].Type] // converse from universal time to go type
				}
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
		notifySuccess("Generated Successfully")
	})

	saveProjectJson := widget.NewButton("Save Project JSON Input", func() {
		outFile, err := dialog.File().Filter("json", ".json").Title("Save As").Save()
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		err = io.SaveContent(outFile, projectJsonInput.Text)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		notifySuccess("Saved Successfully")
	})

	prScroll := container.NewScroll(projectJsonInput)
	generateButton := widget.NewButton("Generate From Project JSON Input", func() {
		prj, err := generator.DecodeProject([]byte(projectJsonInput.Text), projectName.Text, generator.InitEnv)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		files, err := generator.Generate(prj, langTmpl[prj.Language], build.BuildModel)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		for i := range files {
			err = data.Append(strconv.Itoa(i) + ". " + files[i].Name)
			if err != nil {
				display.PopUpWindows(err.Error(), canvas)
				return
			}
			err = dataStruct.SetValue(strconv.Itoa(i)+". "+files[i].Name, files[i].Content)
			if err != nil {
				display.PopUpWindows(err.Error(), canvas)
				return
			}
		}
		//result, err := generator.ToString(files)
		//if err != nil {
		//	display.PopUpWindows(err.Error(), canvas)
		//	return
		//}
		//outputWindows.SetText(result)
		notifySuccess("Generated Successfully")
	})

	testDsn := widget.NewButton("Test Connection", func() {
		driver := driverEntry.Selected
		if driver == s.DriverOracle {
			driver = "godror"
		}
		db, err := sql.Open(driver, dsnSourceEntry.Text)
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
		notifySuccess("Connected Successfully")
	})

	bottomButtons := container.NewVBox(
		testDsn,
		connectGenerate,
		generateButton,
		saveProjectJson,
		loadPrJson,
	)
	dirBox := container.NewVBox(
		optimizeEntry,
		widget.NewTextGridFromString("Datasource name:"),
		driverEntry,
		dsnSourceEntry,
		//container.NewHSplit(widget.NewTextGridFromString("Project name:"), projectName),
		//container.NewHSplit(widget.NewTextGridFromString("Project Template:"), prjTmplNameEntry),
		widget.NewTextGridFromString("Project name:"),
		projectName,
		widget.NewTextGridFromString("Project template name:"),
		prjTmplNameEntry,
	)

	cb1 := container.NewBorder(dirBox, bottomButtons, nil, nil, prScroll)

	cb2 := container.NewHSplit(container.NewScroll(outputWindows), list)

	return container.NewHSplit(cb1, cb2)
}

func LoadAllLanguageTemplates(directory string) (map[string]map[string]string, error) {
	projTmpl := make(map[string]map[string]string)
	folders, err := io.List(directory)
	if err != nil {
		log.Fatal(err)
	}
	for _, folder := range folders {
		names, err := io.List(filepath.Join(directory, folder))
		if err != nil {
			return nil, err
		}
		tm := make(map[string]string, 0)
		for _, n := range names {
			content, err := ioutil.ReadFile(directory + string(os.PathSeparator) + folder + string(os.PathSeparator) + n)
			if err != nil {
				return nil, err
			}
			tm[n] = string(content)
		}
		projTmpl[folder] = tm
	}
	return projTmpl, err
}
