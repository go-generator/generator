package generator

import (
	"github.com/go-generator/core"
	"github.com/go-generator/core/io"
	"go-generator/pkg/manual"
	"io/ioutil"
	"os"
)

func InitMapArray(entityList []string) []map[string]string {
	var mapList []map[string]string
	for _, v := range entityList {
		mapList = append(mapList, manual.BuildNames(v))
	}
	return mapList
}
func InitFields(fieldList []metadata.Field) []map[string]string {
	var mapList []map[string]string
	for _, v := range fieldList {
		mapList = append(mapList, manual.BuildFields(v))
	}
	return mapList
}
func InitJavaFields(mo []metadata.Field) []map[string]string {
	var mapList []map[string]string
	for _, v := range mo {
		mapList = append(mapList, manual.JavaBuildFields(v))
	}
	return mapList
}
func InitTemplate(templateFolder string, project metadata.Project) (map[string]string, error) {
	templateMap := make(map[string]string)
	names, err := io.Load(templateFolder)
	if err != nil {
		return nil, err
	}
	for _, tName := range names {
		content, err1 := ioutil.ReadFile(templateFolder + string(os.PathSeparator) + tName)
		if err1 != nil {
			return nil, err1
		}
		templateMap[tName] = string(content)
	}
	//for _, v := range project.Statics {
	//	content, err := ioutil.ReadFile(templateFolder + string(os.PathSeparator) + v.Table)
	//	if err != nil {
	//		return nil, err
	//	}
	//	templateMap[v.Table] = string(content)
	//}
	//for _, v := range project.Arrays {
	//	content, err := ioutil.ReadFile(templateFolder + string(os.PathSeparator) + v.Table)
	//	if err != nil {
	//		return nil, err
	//	}
	//	templateMap[v.Table] = string(content)
	//}
	//for _, v := range project.Entities {
	//	content, err := ioutil.ReadFile(templateFolder + string(os.PathSeparator) + v.Table)
	//	if err != nil {
	//		return nil, err
	//	}
	//	templateMap[v.Table] = string(content)
	//}
	//content, err := ioutil.ReadFile(templateFolder + string(os.PathSeparator) + "model")
	//if err != nil {
	//	return nil, err
	//}
	//templateMap["model"] = string(content)
	return templateMap, nil
}

func InitProject(templateFolder string, project metadata.Project) (map[string]interface{}, error) {
	templateMap := make(map[string]interface{})
	for _, v := range project.Statics {
		content, err := ioutil.ReadFile(templateFolder + string(os.PathSeparator) + v.Name + ".txt")
		if err != nil {
			return nil, err
		}
		templateMap[v.Name] = string(content)
	}
	for _, v := range project.Arrays {
		content, err := ioutil.ReadFile(templateFolder + string(os.PathSeparator) + v.Name + ".txt")
		if err != nil {
			return nil, err
		}
		templateMap[v.Name] = string(content)
	}
	for _, v := range project.Entities {
		content, err := ioutil.ReadFile(templateFolder + string(os.PathSeparator) + v.Name + ".txt")
		if err != nil {
			return nil, err
		}
		templateMap[v.Name] = string(content)
	}
	for _, v := range project.Models {
		content, err := ioutil.ReadFile(templateFolder + string(os.PathSeparator) + v.Name + ".txt")
		if err != nil {
			return nil, err
		}
		templateMap[v.Name] = string(content)
	}
	return templateMap, nil
}
