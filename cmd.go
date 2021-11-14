package main

import (
	"flag"
	"github.com/go-generator/core/build"
	"github.com/go-generator/core/generator"
	"github.com/go-generator/core/io"
	"github.com/go-generator/core/loader"
	"io/ioutil"
	"log"
	"path/filepath"
)

func main() {
	var t = flag.String("t", "template", "template folder")
	var i = flag.String("i", "go_sql", "input json")
	var p = flag.String("p", "projectName", "project name")
	flag.Parse()
	output, err := generator.GenerateFromFile(*t, *p, *i, loader.LoadProject, io.Load, generator.InitEnv, build.BuildModel)
	if err != nil {
		log.Fatal(err)
	}
	result, err := generator.ToString(output.Files)
	if err != nil {
		log.Fatal(err)
	}
	path, err := filepath.Abs("./" + *p + ".txt")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(path, []byte(result), 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("generated successful")
}
