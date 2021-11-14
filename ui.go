// Package main provides various examples of Fyne API capabilities
package main

import (
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"github.com/core-go/config"
	"github.com/go-generator/core"
	"github.com/go-generator/core/display"
	"github.com/go-generator/core/io"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"go-generator/internal/ui"
	"log"
	"os"
	"path/filepath"
)

func setIcon(path string) (fyne.Resource, error) {
	settingIcon, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	r2, err1 := fyne.LoadResourceFromPath(settingIcon)
	if err1 != nil {
		return nil, err1
	}
	return r2, nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	//project.RunWithCommandLine()
	var root metadata.Config
	var dbCache metadata.Database
	err := config.Load(&root, "configs/config")
	if err != nil {
		panic(err)
	}
	err = config.Load(&dbCache, "configs/database")
	if err != nil {
		panic(err)
	}
	types, err := initTypes("type")
	if err != nil {
		panic(err)
	}
	//project.RunWithCommandLine(types["go"].(map[string]string))
	a := app.New()
	w := a.NewWindow("Metadata and Code Generator")
	sWidth, sHeight, err := display.GetActiveDisplaySize(0)
	if err != nil {
		log.Fatal(err)
	}
	size := fyne.NewSize(float32(sWidth), float32(sHeight))
	w.Resize(display.ResizeWindows(30, 25, size))
	r, err1 := setIcon("./icons/icon.png")
	if err1 != nil {
		log.Fatal(err1)
	}
	settingsItem := fyne.NewMenuItem("Settings", func() {
		wi := a.NewWindow("App Settings")
		r1, err1 := setIcon("./icons/app.jpg")
		if err1 != nil {
			display.ShowErrorWindows(a, err1, size)
			return
		}
		wi.SetIcon(r1)
		wi.SetContent(settings.NewSettings().LoadAppearanceScreen(wi))
		wi.Resize(display.ResizeWindows(25, 25, size))
		wi.Show()
	})
	w.SetIcon(r)
	w.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu("Setting", settingsItem)))
	t := types[filepath.Base(root.Template)].(map[string]string)
	wContent := ui.WidgetScreen(context.TODO(), a, size, root, t, dbCache)
	w.SetContent(wContent)
	//w.SetFullScreen(true)
	w.SetMaster()
	w.CenterOnScreen()
	w.ShowAndRun()
}

func initTypes(folder string) (map[string]interface{}, error) {
	types := make(map[string]interface{}, 0)
	names, err := io.List(folder)
	if err != nil {
		return nil, err
	}
	for _, name := range names {
		filename := removeExt(name)
		t, err := config.LoadMap(folder + string(os.PathSeparator) + name)
		if err != nil {
			return nil, err
		}
		types[filename] = t
	}
	return types, err
}

func removeExt(name string) string {
	ext := filepath.Ext(name)
	return name[0 : len(name)-len(ext)]
}
