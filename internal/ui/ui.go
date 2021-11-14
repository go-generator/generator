package ui

import (
	"context"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/go-generator/core"
	"github.com/go-generator/core/build"
	"github.com/go-generator/core/display"
	"github.com/go-generator/core/generator"
	"github.com/go-generator/core/io"
	"github.com/go-generator/core/loader"
	"github.com/go-generator/core/project"
	"github.com/sqweek/dialog"
	"log"
	"os"
	"path/filepath"
)

// WidgetScreen shows a panel containing widget
func WidgetScreen(ctx context.Context, app fyne.App, displaySize fyne.Size, r metadata.Config, types map[string]string, dbCache metadata.Database) fyne.CanvasObject {
	templateAbsPath, err := filepath.Abs(r.Template)
	if err != nil {
		log.Fatal(err)
	}
	projectAbsPath, err := filepath.Abs(r.Projects)
	if err != nil {
		log.Fatal(err)
	}

	tmplEntry := widget.NewEntry()
	tmplEntry.SetText(templateAbsPath)

	prjTmplDirEntry := widget.NewEntry()
	prjTmplDirEntry.SetText(projectAbsPath)

	prjTmplNameEntry := widget.NewEntry()
	prjTmplNameEntry.SetText(r.Project)

	projectName := widget.NewEntry()
	projectName.SetText(r.ProjectName)
	projectName.SetPlaceHolder("Project Name:")

	templateBrowse := widget.NewButton("Browse Template Folder", func() {
		di := dialog.Directory()
		di.StartDir = templateAbsPath
		directory, err := di.Title("Browse...").Browse()
		if err == dialog.ErrCancelled {
			return
		}
		if err != nil {
			tmplEntry.SetText(err.Error())
			return
		}
		tmplEntry.SetText(directory)
	})

	projectJson := widget.NewMultiLineEntry()
	projectJson.SetPlaceHolder("Input Project JSON here....")
	pButtonWindows := widget.NewButton("Project JSON Input...", func() {
		oldText := projectJson.Text
		newWindows := app.NewWindow("Project JSON Structure Input")
		saveButton := widget.NewButton("Save", func() {
			oldText = projectJson.Text
			newWindows.Close()
		})
		cancelButton := widget.NewButton("Cancel", func() {
			projectJson.SetText(oldText)
			newWindows.Close()
		})
		scroll := container.NewScroll(projectJson)
		content := container.NewBorder(nil, container.NewVBox(saveButton, cancelButton), nil, nil, scroll)
		newWindows.SetContent(content)
		newWindows.CenterOnScreen()
		newWindows.Show()
	})

	largeText := widget.NewMultiLineEntry()
	largeText.SetPlaceHolder("Input")
	outputJson := widget.NewMultiLineEntry()
	outputJson.PlaceHolder = "Generated Code Here"
	inputButton := widget.NewButton("Generate From JSON Input", func() {
		if projectName.Text == "" {
			projectName.SetText(r.ProjectName)
		}
		err = generator.GenerateFromString(projectName.Text, outputJson.Text, generator.InitEnv)
		if err != nil {
			display.ShowErrorWindows(app, err, displaySize)
		} else {
			display.ShowSuccessWindows(app, "OK", displaySize)
		}
	})
	inputButton.Hidden = true //temporary hidden

	var output metadata.Output
	openFileButton := widget.NewButton("Generate Project From JSON Metadata", func() {
		if !io.IsValidPath(tmplEntry.Text) {
			display.ShowErrorWindows(app, errors.New("template folder path is required"), displaySize)
			return
		}
		filename, err := dialog.File().Title("Load Metadata File").Filter("All Files").Load()
		if err != nil {
			display.ShowErrorWindows(app, err, displaySize)
			return
		} else {
			output, err = generator.GenerateFromFile(tmplEntry.Text, projectName.Text, filename, loader.LoadProject, io.Load, generator.InitEnv, build.BuildModel)
			if err != nil {
				display.ShowErrorWindows(app, err, displaySize)
				return
			} else {
				result, err := generator.ToString(output.Files)
				if err != nil {
					display.ShowErrorWindows(app, err, displaySize)
					return
				}
				size := display.ResizeWindows(40, 30, displaySize)
				wa := app.NewWindow("Generated Code")
				entry := widget.NewMultiLineEntry()
				entry.Wrapping = fyne.TextWrapWord
				entry.SetText(result)
				saveButton := widget.NewButton("Save Project Files", func() {
					directory, err := dialog.Directory().Title("Save Project Files In...").Browse()
					if output.Directory != "" {
						directory = directory + string(os.PathSeparator) + output.Directory
						err = io.MkDir(directory)
						if err != nil {
							return
						}
					}
					err1 := io.SaveOutput(directory, output)
					if err1 != nil {
						display.ShowErrorWindows(app, err1, displaySize)
						wa.Close()
						return
					}
					wa.Close()
				})
				cancelButton := widget.NewButton("Cancel", func() {
					wa.Close()
				})
				scroll := container.NewScroll(entry)
				vBox := container.NewVBox(saveButton, cancelButton)
				wa.SetContent(container.NewBorder(nil, vBox, nil, nil, scroll))
				wa.Resize(size)
				wa.CenterOnScreen()
				wa.Show()
			}
		}
	})

	method := project.BasicAuth
	connMethod := widget.NewSelect([]string{project.BasicAuth, project.Datasource}, func(opt string) {
		method = opt
	})
	connMethod.Selected = project.Datasource
	modelJsonGenerator := widget.NewButton("Connect To Database", func() {
		err = project.RunWithUI(ctx, app, displaySize, types, prjTmplDirEntry.Text, prjTmplNameEntry.Text, projectName.Text, dbCache, connMethod.Selected)
		if err != nil {
			display.ShowErrorWindows(app, err, displaySize)
			return
		}
	})

	vBox := container.NewVBox(
		templateBrowse,
		pButtonWindows,
		inputButton,
		openFileButton,
		modelJsonGenerator,
	)
	dirBox := container.NewVBox(
		widget.NewLabel("Authentication Method:"),
		connMethod,
		widget.NewLabel("Project Name:"),
		projectName,
		widget.NewLabel("Template Dir:"),
		tmplEntry,
		widget.NewLabel("Project Template Dir:"),
		prjTmplDirEntry,
		widget.NewLabel("Project Template Name:"),
		prjTmplNameEntry,
	)
	return container.NewBorder(nil, vBox, nil, nil, dirBox)
}
