package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/go-generator/core"
	"github.com/go-generator/core/build"
	"github.com/go-generator/core/display"
	"github.com/go-generator/core/export"
	gdb "github.com/go-generator/core/export/db"
	"github.com/go-generator/core/export/relationship"
	"github.com/go-generator/core/generator"
	"github.com/go-generator/core/io"
	"github.com/go-generator/core/list"
	"github.com/go-generator/core/project"
	"github.com/go-generator/core/template"
	"github.com/skratchdot/open-golang/open"
	"github.com/sqweek/dialog"
	"gopkg.in/yaml.v3"
)

const (
	DriverPostgres = "postgres"
	DriverMysql    = "mysql"
	DriverMssql    = "mssql"
	DriverOracle   = "oracle"
	DriverSqlite3  = "sqlite3"
)

// AppScreen shows a panel containing widget
func AppScreen(ctx context.Context, canvas fyne.Canvas, allTypes, allUniversalTypes map[string]map[string]string, c metadata.Config, dbCache metadata.Database) fyne.CanvasObject {
	var files []metadata.File
	funcMap := template.MakeFuncMap()
	templatePath, err := filepath.Abs(filepath.Join(".", "configs", c.Template))
	if err != nil {
		display.PopUpWindows(fmt.Sprintf("error: %v", err), canvas)
		return nil
	}
	templates, err := io.LoadAll(templatePath)
	if err != nil {
		display.PopUpWindows(fmt.Sprintf("error: %v", err), canvas)
		return nil
	}
	projectPath, err := filepath.Abs(filepath.Join(".", "configs", c.ProjectPath))
	if err != nil {
		display.PopUpWindows(fmt.Sprintf("error: %v", err), canvas)
		return nil
	}
	projectTemplate, err := io.Load(projectPath)
	if err != nil {
		display.PopUpWindows(fmt.Sprintf("error: %v", err), canvas)
		return nil
	}
	projects, err := io.List(projectPath)
	if err != nil {
		display.PopUpWindows(fmt.Sprintf("error: %v", err), canvas)
		return nil
	}

	if len(projects) < 1 {
		display.PopUpWindows("no database types found", canvas)
		return nil
	}

	infinite := widget.NewProgressBarInfinite()
	infinite.Hide()

	projectTemplateName := c.Project
	prjTmplNameEntry := widget.NewSelect(projects, func(o string) {
		c.Project = o
		err = io.SaveConfig(os.Getenv(project.ConfigEnv), c)
		if err != nil {
			display.PopUpWindows(fmt.Sprintf("error: %v", err.Error()), canvas)
			return
		}
		projectTemplateName = o
	})

	prjTmplNameEntry.PlaceHolder = c.Project
	prjTmplNameEntry.SetSelected(c.Project)

	projectName := widget.NewEntry()
	projectName.SetText(c.ProjectName)
	projectName.OnChanged = func(pjName string) {
		c.ProjectName = pjName
		err = io.SaveConfig(os.Getenv(project.ConfigEnv), c)
		if err != nil {
			display.PopUpWindows(fmt.Sprintf("error: %v", err.Error()), canvas)
			return
		}
		projectName.SetText(c.ProjectName)
	}
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
		DriverMysql,
		DriverMssql,
		DriverSqlite3,
		DriverOracle,
		DriverPostgres,
	}, func(s string) {
		driverEntry.SetSelected(s)
	})
	driverEntry.Horizontal = true
	driverEntry.SetSelected(c.DB)

	dsnSourceEntry := widget.NewEntry()
	dsnSourceEntry.OnChanged = func(dsn string) {
		if driverEntry.Selected != "" {
			c.DB = driverEntry.Selected
			err = io.SaveConfig(os.Getenv(project.ConfigEnv), c)
			if err != nil {
				display.PopUpWindows(fmt.Sprintf("error: %v", err.Error()), canvas)
				return
			}
			var (
				prj    metadata.Project
				buffer bytes.Buffer
			)
			prjString := projectTemplate[prjTmplNameEntry.Selected]
			err = json.NewDecoder(bytes.NewBuffer([]byte(prjString))).Decode(&prj)
			if err != nil {
				display.PopUpWindows(fmt.Sprintf("error: %v", err.Error()), canvas)
				return
			}
			dsnSpecial := bytes.Replace([]byte(dsn), []byte(`\u0026`), []byte("&"), -1)
			prj.Env["driver"] = driverEntry.Selected
			if driverEntry.Selected == DriverOracle {
				prj.Env["driver"] = "godror"
			}
			prj.Env["data_source_name"] = string(dsnSpecial)
			encoder := json.NewEncoder(&buffer)
			encoder.SetIndent(" ", "  ")
			err = encoder.Encode(&prj)
			if err != nil {
				display.PopUpWindows(fmt.Sprintf("error: %v", err.Error()), canvas)
				return
			}
			path := filepath.Join(projectPath, prjTmplNameEntry.Selected)
			err = ioutil.WriteFile(path, buffer.Bytes(), os.ModePerm)
			if err != nil {
				display.PopUpWindows(fmt.Sprintf("error: %v", err.Error()), canvas)
				return
			}
		}
		project.UpdateDBCache(&dbCache, driverEntry.Selected, dsn)
	}
	dsnSourceEntry.SetText(project.SelectDSN(dbCache, driverEntry.Selected))
	driverEntry.OnChanged = func(driver string) {
		dsnSourceEntry.SetText(project.SelectDSN(dbCache, driver))
	}

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
			if strings.HasPrefix(keys[i], strconv.Itoa(id+1)+".") {
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

	btnLoadAndGenerate := widget.NewButtonWithIcon("Generate From File", theme.DocumentCreateIcon(), func() {
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

	var modeEntry, showPathEntry *widget.Check
	modeEntry = widget.NewCheck("Models Only", func(b bool) {
		modeEntry.SetChecked(b)
	})
	modeEntry.SetChecked(false)

	showPathEntry = widget.NewCheck("Show Path", func(b bool) {
		showPathEntry.SetChecked(b)
	})
	showPathEntry.SetChecked(false)

	btnGenerateJSON := widget.NewButtonWithIcon("Generate Project JSON", theme.DocumentCreateIcon(), func() {
		infinite.Show()
		var (
			models []metadata.Model
			prj    *metadata.Project
			pData  bytes.Buffer
		)
		sqlDB, err := project.ConnectDB(dbCache, driverEntry.Selected)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			infinite.Hide()
		}
		defer func() {
			err = sqlDB.Close()
			if err != nil {
				display.PopUpWindows(err.Error(), canvas)
				infinite.Hide()
			}
		}()
		enc := json.NewEncoder(&pData)
		enc.SetIndent("", "   ")
		dbName, err := project.GetDatabaseName(dbCache, driverEntry.Selected)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			infinite.Hide()
			return
		}

		tables, err := gdb.ListTables(ctx, sqlDB, dbName)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			infinite.Hide()
			return
		}

		primaryKeys, err := export.GetAllPrimaryKeys(ctx, sqlDB, dbName, driverEntry.Selected, tables)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			infinite.Hide()
			return
		}

		relations, err := relationship.GetRelationshipTable(ctx, sqlDB, dbName, tables, primaryKeys)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			infinite.Hide()
			return
		}

		models, err = export.ToModels(ctx, sqlDB, dbName, tables, relations, allUniversalTypes[driverEntry.Selected], primaryKeys)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			infinite.Hide()
			return
		}

		if modeEntry.Checked {
			err = enc.Encode(&models)
			if err != nil {
				display.PopUpWindows(err.Error(), canvas)
				infinite.Hide()
				return
			}
		} else {
			prj, err = generator.ExportProject(projectTemplateName, projectName.Text, projectTemplate, models, build.InitEnv)
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
				infinite.Hide()
				return
			}
		}
		cache, err := yaml.Marshal(dbCache)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			infinite.Hide()
			return
		}
		err = io.Save(filepath.Join(".", c.DBCache), cache)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			infinite.Hide()
			return
		}
		projectJsonInput.SetText(pData.String())
		infinite.Hide()
		display.Notify("Success", "Generate Project JSON")
	})

	btnSaveProjectJSON := widget.NewButtonWithIcon("Save JSON", theme.DocumentSaveIcon(), func() {
		outFile, err := dialog.File().Filter("json", ".json").Title("Save As").Save()
		if err == dialog.ErrCancelled {
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

	btnGenerate := widget.NewButtonWithIcon("Generate Output", theme.DocumentCreateIcon(), func() {
		newDataStruct := binding.BindStruct(&metadata.File{})
		files, err = generator.GenerateFiles(projectName.Text, projectJsonInput.Text, templates, funcMap, allTypes)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		err = data.Set(nil)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		err = list.ShowFiles(showPathEntry.Checked, data, newDataStruct, files)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
		dataStruct = newDataStruct // reload list
		display.Notify("Success", "Generate Output")
	})

	openProjectPath := ""
	btnOpenOutputDirectory := widget.NewButtonWithIcon("Open Output Folder", theme.FolderOpenIcon(), func() {
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

	btnSave := widget.NewButtonWithIcon("Save Project", theme.DocumentSaveIcon(), func() {
		if len(files) < 1 {
			display.PopUpWindows("output files list is empty", canvas)
			return
		}
		directory, err := dialog.Directory().Title("Save Project Files In...").Browse()
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			return
		}
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

	btnTestDsn := widget.NewButtonWithIcon("Test Connection", theme.CheckButtonCheckedIcon(), func() {
		infinite.Show()
		oldDsn := project.SelectDSN(dbCache, driverEntry.Selected)
		if strings.Compare(oldDsn, dsnSourceEntry.Text) != 0 {
			project.UpdateDBCache(&dbCache, driverEntry.Selected, dsnSourceEntry.Text)
		}
		db, err := project.ConnectDB(dbCache, driverEntry.Selected)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			infinite.Hide()
			return
		}
		defer func() {
			err = db.Close()
			if err != nil {
				display.PopUpWindows(err.Error(), canvas)
				infinite.Hide()
			}
		}()
		err = db.Ping()
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			infinite.Hide()
			return
		}
		cache, err := yaml.Marshal(dbCache)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			infinite.Hide()
			return
		}
		err = io.Save(filepath.Join(".", c.DBCache), cache)
		if err != nil {
			display.PopUpWindows(err.Error(), canvas)
			infinite.Hide()
			return
		}
		infinite.Hide()
		display.Notify("Success", "Test Connection")
	})

	btnReloadTemplates := widget.NewButtonWithIcon("Reload All", theme.ContentRedoIcon(), func() {
		projectTemplate, err = io.Load(projectPath)
		if err != nil {
			display.PopUpWindows(fmt.Sprintf("error: %v", err), canvas)
			return
		}
		projects, err = io.List(projectPath)
		if err != nil {
			display.PopUpWindows(fmt.Sprintf("error: %v", err), canvas)
			return
		}
		templates, err = io.LoadAll(templatePath)
		if err != nil {
			display.PopUpWindows(fmt.Sprintf("error: %v", err), canvas)
			return
		}
		outputWindows.SetText("")
		err = data.Set(nil)
		if err != nil {
			display.PopUpWindows(fmt.Sprintf("error: %v", err), canvas)
			return
		}
		dataStruct = binding.BindStruct(&metadata.File{})
		projectJsonInput.SetText("")
	})

	prScroll := container.NewScroll(projectJsonInput)
	outScroll := container.NewScroll(outputWindows)

	bottomButtons := container.NewVBox(
		container.NewAdaptiveGrid(2,
			btnTestDsn,
			btnGenerateJSON,
			btnGenerate,
			btnLoadAndGenerate,
			btnSaveProjectJSON,
			btnSave, btnReloadTemplates, btnOpenOutputDirectory),
	)

	tabs := container.NewVBox(container.NewAdaptiveGrid(2, modeEntry, showPathEntry),
		container.NewAdaptiveGrid(2, widget.NewLabel("Project:"),
			widget.NewLabel("Package:"),
			prjTmplNameEntry,
			projectName), driverEntry, dsnSourceEntry)

	cb1 := container.NewBorder(tabs, bottomButtons, nil, nil, listWithData)

	cb2 := container.NewHSplit(
		prScroll,
		outScroll,
	)
	return container.NewBorder(nil, infinite, cb1, nil, cb2)
}
