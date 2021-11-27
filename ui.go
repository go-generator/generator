package main

import (
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"github.com/core-go/config"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/go-generator/core"
	"github.com/go-generator/core/display"
	"github.com/go-generator/core/export/types"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/godror/godror"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"go-generator/internal/ui"
	"log"
)

func main() {
	ctx := context.TODO()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var root metadata.Config
	var dbCache metadata.Database
	err := config.Load(&root, "configs/config")
	if err != nil {
		panic(err)
	}
	err = config.Load(&dbCache, "configs/datasource")
	if err != nil {
		panic(err)
	}

	a := app.NewWithID("Generator")
	r, err1 := display.SetIcon("./icons/icon.png")
	if err1 != nil {
		log.Fatal(err1)
	}
	a.SetIcon(r)
	w := a.NewWindow("Metadata and Code Generator")
	canvas := w.Canvas()
	sWidth, sHeight, err := display.GetActiveDisplaySize(0)
	if err != nil {
		log.Fatal(err)
	}

	size := fyne.NewSize(float32(sWidth), float32(sHeight))
	w.Resize(display.ResizeWindows(70, 60, size))
	settingsItem := fyne.NewMenuItem("Settings", func() {
		wi := a.NewWindow("App Settings")
		r1, err1 := display.SetIcon("./icons/app.jpg")
		if err1 != nil {
			display.PopUpWindows(err1.Error(), canvas)
			return
		}
		wi.SetIcon(r1)
		wi.SetContent(settings.NewSettings().LoadAppearanceScreen(wi))
		wi.Resize(display.ResizeWindows(25, 25, size))
		wi.Show()
	})
	w.SetIcon(r)
	w.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu("Setting", settingsItem)))

	wContent := ui.WidgetScreen(ctx, canvas, types.Types, root, dbCache)
	w.SetContent(wContent)
	w.SetMaster()
	w.CenterOnScreen()
	w.ShowAndRun()
}
