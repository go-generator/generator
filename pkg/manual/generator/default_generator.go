package generator

import (
	"context"
	"github.com/go-generator/core"
	"go-generator/pkg/manual"
)

type DefaultGenerator struct {
}

func (t *DefaultGenerator) Generate(ctx context.Context, project metadata.Project, templates map[string]string) []metadata.File {
	var outputFile []metadata.File
	//buffer := new(bytes.Buffer)
	//err1 := json.NewEncoder(buffer).Encode(&project)
	//if err1 != nil {
	//	panic(err1)
	//}
	//err1 = ioutil.WriteFile("./ts.json", buffer.Bytes(), 0664)
	//if err1 != nil {
	//	panic(err1)
	//}
	//b, err := ioutil.ReadFile("./ts.json")
	//if err != nil {
	//	panic(err)
	//}
	//json.NewDecoder(bytes.NewBuffer(b)).Decode(&project)
	for _, v := range project.Statics {
		//COMPLETE THE FILENAME
		var defaultStaticTemplate manual.DefaultStaticTemplate
		v.File = defaultStaticTemplate.Generate(context.Background(), v.File, project.Env)
		//READ THE TEMPLATE FILES
		s := templates[v.Name]
		//CREATE TEXT
		text := defaultStaticTemplate.Generate(context.Background(), s, project.Env)
		outputFile = append(outputFile, metadata.File{Name: "/" + v.File, Content: text})
	}
	for _, v := range project.Arrays { //
		//COMPLETE THE FILENAME
		var t1 manual.DefaultStaticTemplate
		v.File = t1.Generate(context.Background(), v.File, project.Env)
		//READ THE TEMPLATE FILES
		s := templates[v.Name]
		//CREATE TEXT
		arrayTemplate := manual.NewArrayTemplate("${begin}", "${end}")
		text := arrayTemplate.Array(context.Background(), s, project.Env, nil, InitMapArray(project.Collection))
		outputFile = append(outputFile, metadata.File{Name: "/" + v.File, Content: text})
	}
	for _, v := range project.Entities {
		for _, c := range project.Collection {
			buildNamesMap := manual.BuildNames(c)
			//COMPLETE THE FILENAME
			var entityTemplate manual.EntityTemplate
			entityTemplate = &manual.DefaultEntityTemplate{}
			tmpFile := entityTemplate.Merge(context.Background(), v.File, project.Env, buildNamesMap, nil)

			//READ THE TEMPLATE FILES
			s := templates[v.Name]
			//CREATE TEXT
			text := entityTemplate.Merge(context.Background(), s, project.Env, buildNamesMap, nil)
			outputFile = append(outputFile, metadata.File{Name: "/" + tmpFile, Content: text})
		}
	}
	for _, v := range project.Models {
		//READ THE TEMPLATE
		s := templates["model"]
		//CREATE TEXT
		modelTemplate := manual.NewModelTemplate("${begin}", "${end}", "${case ", "${endcase}")
		text := modelTemplate.Model(context.Background(), s, project.Env, manual.BuildNames(v.Name), InitFields(v.Fields), nil, nil, nil)
		outputFile = append(outputFile, metadata.File{Name: project.Env["model_raw"] + "/" + v.Name + ".go_bk", Content: text})
	}
	return outputFile
}
