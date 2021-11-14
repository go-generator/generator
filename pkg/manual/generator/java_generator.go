package generator

import (
	"context"
	"github.com/go-generator/core"
	"go-generator/pkg/manual"
)

type JavaGenerator struct {
}

func (t *JavaGenerator) Generate(ctx context.Context, project metadata.Project, templates map[string]string) []metadata.File {
	var outputFile []metadata.File
	for _, v := range project.Models {
		//CREATE PK
		count := 0
		var vTmp []metadata.Field
		for _, v1 := range v.Fields {
			if v1.Key {
				count += 1
				vTmp = append(vTmp, v1)
			}
		}
		t := manual.NewJavaModelTemplate("${begin}", "${end}", "${case ", "${endcase}")
		selfMap := manual.BuildNames(v.Name)
		text := ""
		if count > 1 {
			text = t.Model(context.Background(), templates["pk"], project.Env, selfMap, InitJavaFields(vTmp), nil, nil, nil)
			outputFile = append(outputFile, metadata.File{Name: project.Env["model_Name"] + "/" + selfMap["Table"] + "PK.java", Content: text})
			selfMap["IdClassJavaAnnotation"] = "@IdClass(" + selfMap["Table"] + "PK.class)"
		} else {
			tmpMap := manual.BuildNames(vTmp[0].Type)
			selfMap["IdClass"] = tmpMap["Table"]
			selfMap["idClass"] = tmpMap["name"]
			selfMap["IdClassJavaAnnotation"] = ""
		}

		//READ THE TEMPLATE
		template := templates["normal"]
		if v.Arrays != nil {
			template = templates["array"]
		}
		//CREATE TEXT
		text = t.Model(context.Background(), template, project.Env, selfMap, InitJavaFields(v.Fields), nil, v.Arrays, nil)
		outputFile = append(outputFile, metadata.File{Name: project.Env["model_Name"] + "/" + selfMap["Table"] + ".java", Content: text})
	}

	return outputFile
}
