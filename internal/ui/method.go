package ui

import (
	"fyne.io/fyne/v2/data/binding"
	metadata "github.com/go-generator/core"
	"github.com/go-generator/core/build"
	"github.com/go-generator/core/generator"
	"github.com/go-generator/core/io"
	"github.com/go-generator/core/types"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
)

func loadAllLanguageTemplates(directory string) (map[string]map[string]string, error) {
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

func generateFilesList(data binding.ExternalStringList, dataSt binding.Struct, files []metadata.File) error {
	err := data.Set(nil)
	if err != nil {
		return err
	}
	for i := range files {
		filename := strconv.Itoa(i) + ". " + filepath.Base(files[i].Name)
		err = data.Append(filename)
		if err != nil {
			return err
		}
		err = dataSt.SetValue(filename, files[i].Content)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateProjectFiles(projectName, projectJson string, funcMap template.FuncMap, langTmpl map[string]map[string]string) ([]metadata.File, error) {
	prj, err := generator.DecodeProject([]byte(projectJson), projectName, generator.InitEnv)
	if err != nil {
		return nil, err
	}
	_, ok := prj.Env["go_module"]
	if ok && projectName != "" {
		prj.Env["go_module"] = projectName
	}
	prj.Types = types.Types[prj.Language]
	return generator.Generate(prj, langTmpl[prj.Language], funcMap, build.BuildModel)
}
