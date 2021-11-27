package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/core-go/config"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/go-generator/core"
	"github.com/go-generator/core/display"
	"github.com/go-generator/core/export/types"
	uni "github.com/go-generator/core/export/types"
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
	err = project.SetPathEnv(project.UniversalJsonEnv, "./configs/sql_types.json")
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
	allUniversalTypes := make(map[string]map[string]string)

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

	uniJson, err := filepath.Abs(os.Getenv(project.UniversalJsonEnv))
	if err != nil {
		allUniversalTypes = uni.Types
	} else {
		content, err := ioutil.ReadFile(uniJson)
		if err != nil {
			allUniversalTypes = uni.Types
		} else {
			err = json.NewDecoder(bytes.NewBuffer(content)).Decode(&allUniversalTypes)
			if err != nil {
				allUniversalTypes = uni.Types
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
	mainMenu, wContent := ui.AppScreen(ctx, canvas, allTypes, allUniversalTypes, root, dbCache)
	if wContent == nil {
		time.Sleep(time.Duration(5) * time.Second)
		a.Quit()
	}
	w.SetMainMenu(mainMenu)
	w.SetContent(wContent)
	w.ShowAndRun()
}
