package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"github.com/core-go/config"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/go-generator/core"
	"github.com/go-generator/core/display"
	"github.com/go-generator/core/export/types"
	"github.com/go-generator/core/project"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/godror/godror"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"go-generator/internal/ui"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	err := project.SetPathEnv(project.TypesJsonEnv, "./configs/types.json")
	if err != nil {
		panic(err)
	}
	err = project.SetPathEnv(project.WindowsIconEnv, "./configs/icon/icon.png")
	if err != nil {
		panic(err)
	}
	err = project.SetPathEnv(project.AppIconEnv, "./configs/icon/app.jpg")
	if err != nil {
		panic(err)
	}
	err = project.SetPathEnv(project.ConfigEnv, "./configs/config.yaml")
	if err != nil {
		panic(err)
	}
}

func main() {
	var (
		root    metadata.Config
		dbCache metadata.Database
	)
	err := config.Load(&root, "configs/config")
	if err != nil {
		panic(err)
	}
	err = config.Load(&dbCache, "configs/datasource")
	if err != nil {
		panic(err)
	}
	ctx := context.TODO()
	allTypes := make(map[string]map[string]string)

	tpJson, err := filepath.Abs(os.Getenv(project.TypesJsonEnv))
	if err != nil {
		allTypes = types.Types
	} else {
		content, err := ioutil.ReadFile(tpJson)
		if err != nil {
			allTypes = types.Types
		} else {
			err = json.NewDecoder(bytes.NewBuffer(content)).Decode(&allTypes)
			if err != nil {
				allTypes = types.Types
			}
		}
	}
	a := app.NewWithID("Generator")
	r, err := display.SetIcon(os.Getenv(project.WindowsIconEnv))
	if err != nil {
		log.Fatal(err)
	}
	a.SetIcon(r)
	w := a.NewWindow("Metadata and Code Generator")
	w.SetMaster()
	w.CenterOnScreen()
	canvas := w.Canvas()
	sWidth, sHeight, err := display.GetActiveDisplaySize(0)
	if err != nil {
		log.Fatal(err)
	}

	size := fyne.NewSize(float32(sWidth), float32(sHeight))
	w.Resize(display.ResizeWindows(70, 60, size))
	settingsItem := fyne.NewMenuItem("Settings", func() {
		settingWindows := a.NewWindow("App Settings")
		r1, err1 := display.SetIcon(os.Getenv(project.AppIconEnv))
		if err1 != nil {
			display.PopUpWindows(err1.Error(), canvas)
			return
		}
		settingWindows.SetIcon(r1)
		settingWindows.SetContent(settings.NewSettings().LoadAppearanceScreen(settingWindows))
		settingWindows.Resize(display.ResizeWindows(25, 25, size))
		settingWindows.Show()
	})
	w.SetIcon(r)
	w.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("Setting", settingsItem)))
	wContent := ui.AppScreen(ctx, canvas, allTypes, root, dbCache)
	w.SetContent(wContent)
	w.ShowAndRun()
}
