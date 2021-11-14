package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/go-generator/core"
	"github.com/go-generator/core/loader"
	"io/ioutil"
	"log"
)

func SaveMetadataJson(projectStruct metadata.Project, filePath string) error { //s *TableInfo, conn *gorm.DB, tables []string, packageName, output string) {
	data, err := json.MarshalIndent(&projectStruct, "", " ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filePath, data, 0644) // Create and write files
	if err != nil {
		return err
	}
	return err
}

func SaveModelMetadataJson(projectModel []metadata.Model, filePath string) error { //s *TableInfo, conn *gorm.DB, tables []string, packageName, output string) {
	data, err := json.MarshalIndent(&projectModel, "", " ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filePath, data, 0644) // Create and write files
	if err != nil {
		return err
	}
	return err
}

func JsonRead(filename string) (metadata.DatabaseConfig, error) {
	var input metadata.DatabaseConfig
	byteValue, err := ioutil.ReadFile(filename)
	if err != nil {
		return input, err
	}
	err = json.NewDecoder(bytes.NewBuffer(byteValue)).Decode(&input)
	if err != nil {
		return input, err
	}
	return input, nil
}

func WriteMetadata(dc metadata.DatabaseConfig, separateFile bool, inFile, outFile, modelFile string) error {
	projectStruct, err := loader.LoadProject(inFile)
	if err != nil {
		return err
	}
	if separateFile {
		projectStruct.ModelsFile = modelFile
		err = SaveModelMetadataJson(projectStruct.Models, modelFile)
		if err != nil {
			return err
		}
		projectStruct.Models = nil
		err = SaveMetadataJson(projectStruct, outFile)
		if err != nil {
			return err
		}
	} else {
		err = SaveMetadataJson(projectStruct, outFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	var i = flag.String("i", "go_file.json", "input json")
	var o = flag.String("o", "metadata.json", "output json")
	var c = flag.String("c", "dbConfig.json", "configs json")
	var m = flag.String("m", "", "separated metadata output")

	flag.Parse()
	dbConfig, err := JsonRead(*c)
	if err != nil {
		log.Panicln(err)
	}
	err = WriteMetadata(dbConfig, *m != "", *i, *o, *m)
	if err != nil {
		log.Panicln(err)
	}
	log.Print("Successful")

}
