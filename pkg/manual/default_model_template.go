package manual

import (
	"context"
	"strings"
)

type DefaultModelTemplate struct {
	Begin   string
	End     string
	Case    string
	EndCase string
}

func NewModelTemplate(begin string, end string, Case string, endCase string) *DefaultModelTemplate {
	return &DefaultModelTemplate{begin, end, Case, endCase}
}

type caseType struct {
	Field   string
	Value   string
	Content string
}

func buildCases(text string, CaseString string, EndCaseString string) []caseType {
	var caseList []caseType
	if CaseString == "" {
		return caseList
	}
	for strings.Contains(text, CaseString) {
		begin := strings.Index(text, CaseString)
		end := strings.Index(text, EndCaseString)
		subString := text[begin+len(CaseString) : end-1]
		closeSymbol := strings.Index(subString, "}")
		equalSymbol := strings.Index(subString, "=")
		caseDefault := strings.Index(subString, "default}")
		content := subString[closeSymbol+1:]
		var field, value string
		if caseDefault != -1 {
			field = "default"
			value = "true"
		} else {
			field = subString[:equalSymbol]
			value = subString[equalSymbol+1 : closeSymbol]
		}
		caseList = append(caseList, caseType{
			Field:   field,
			Value:   value,
			Content: content,
		})
		text = text[end+len(EndCaseString):]
	}
	return caseList
}
func (t *DefaultModelTemplate) Model(ctx context.Context, template string, share map[string]string, selfMap map[string]string, arr []map[string]string, oneToOneArr, manyToOneArr, oneToManyArr []map[string]string) string {
	text := template
	for k, v := range share {
		x := "${env:" + k + "}"
		text = strings.ReplaceAll(text, x, v)
	}
	for k, v := range selfMap {
		x := "${self:" + k + "}"
		text = strings.ReplaceAll(text, x, v)
	}
	for strings.Contains(text, t.Begin) {
		begin := strings.Index(text, t.Begin)
		end := strings.Index(text, t.End)
		subString := text[begin+len(t.Begin) : end-1]
		caseList := buildCases(subString, t.Case, t.EndCase)
		tmpText := text[:begin]
		if len(caseList) == 0 {
			for j := 0; j < len(arr); j++ {
				tmp := subString
				for k, v := range arr[j] {
					x := "${" + k + "}"
					tmp = strings.ReplaceAll(tmp, x, v)
				}
				tmpText += tmp
			}
			tmpText = strings.TrimSuffix(tmpText, ",")
		} else {
			for j := 0; j < len(arr); j++ {
				for _, v := range caseList {
					if arr[j][v.Field] == v.Value {
						tmp := v.Content
						for k, v1 := range arr[j] {
							x := "${" + k + "}"
							tmp = strings.ReplaceAll(tmp, x, v1)
						}
						tmpText += tmp
					}
				}
			}
			tmpText = strings.TrimSuffix(tmpText, ",")
		}
		text = tmpText + text[end+len(t.End):]
	}
	return text
}
