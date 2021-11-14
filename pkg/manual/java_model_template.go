package manual

import (
	"context"
	"github.com/go-generator/core"
	"strings"
)

type JavaModelTemplate struct {
	Begin   string
	End     string
	Case    string
	EndCase string
}

func NewJavaModelTemplate(begin string, end string, Case string, endCase string) *JavaModelTemplate {
	return &JavaModelTemplate{begin, end, Case, endCase}
}

//func NestedBegin(be, en, ca, enca, text string, arr []map[string]string, oneToOneArr, manyToOneArr, oneToManyArr []metadata.Relationship) string {
//	//for strings.Contains(text, be) {
//	//	begin := strings.Index(text, be)
//	//	end := strings.Index(text, en)
//	//	subString := text[begin+len(be) : end-1]
//	//	count := strings.Count(subString, be)
//	//	for i := 0; i < count; i++ {
//	//		end = strings.Index(text[end+len(en):], en)
//	//	}
//	//	subString = text[begin+len(be) : end-1]
//	//	result := NestedBegin(be, en, ca, enca, subString, arr, oneToOneArr, manyToOneArr, oneToManyArr)
//	//	text = text[:begin] + result + text[end+len(en):]
//	//}
//	for strings.Contains(text, be) {
//		begin := strings.Index(text, be)
//		end := strings.Index(text, en)
//		subString := text[begin+len(be) : end-1]
//		caseList := buildCases(subString, ca, enca)
//		tmpText := text[:begin]
//		//caseList := buildCases(text, ca, enca)
//		//tmpText := ""
//		if len(caseList) == 0 {
//			for j := 0; j < len(arr); j++ {
//				tmp := text
//				for k, v := range arr[j] {
//					x := "${" + k + "}"
//					tmp = strings.ReplaceAll(tmp, x, v)
//				}
//				tmpText += tmp
//			}
//			//tmpText = strings.TrimSuffix(tmpText, ",")
//		} else {
//			for j := 0; j < len(arr); j++ {
//				for _, v := range caseList {
//					if arr[j][v.Field] == v.Value {
//						tmp := v.Content
//						for k, v1 := range arr[j] {
//							x := "${" + k + "}"
//							tmp = strings.ReplaceAll(tmp, x, v1)
//						}
//						tmpText += tmp
//					}
//				}
//			}
//		}
//
//		head := ""
//		middle := ""
//		foot := ""
//		for _, v := range caseList {
//			if v.Field == "head" {
//				head = v.Content
//			}
//			if v.Field == "middle" {
//				middle = v.Content
//			}
//			if v.Field == "foot" {
//				foot = v.Content
//			}
//
//		}
//		if head != "" {
//			for _, v := range manyToOneArr {
//				tmp2 := head
//				tableMap := BuildNames(v.Ref)
//				for _, v1 := range v.Fields {
//					tmp := middle
//					tmp = strings.ReplaceAll(tmp, "${refCol}", v1.Column)
//					tmp = strings.ReplaceAll(tmp, "${refToCol}", v1.To)
//					tmp2 += tmp
//				}
//				tmp := foot
//				tableMap = BuildNames(v.Ref)
//				tmp = strings.ReplaceAll(tmp, "${refTable_Name}", tableMap["Table"])
//				tmp = strings.ReplaceAll(tmp, "${refTable_raw}", tableMap["raw"])
//				tmp2 += tmp
//				tmpText += tmp2
//			}
//		}
//		tmpText = strings.TrimSuffix(tmpText, ",")
//		text = tmpText + text[end+len(en):]
//	}
//
//	//text = tmpText + text[end+len(t.End):]
//
//	return text
//
//}

func (t *JavaModelTemplate) Model(ctx context.Context, template string, share map[string]string, selfMap map[string]string, arr []map[string]string, oneToOneArr, manyToOneArr, oneToManyArr []metadata.Relationship) string {
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
			head := ""
			middle := ""
			foot := ""
			for _, v := range caseList {
				if v.Field == "head" {
					head = v.Content
				}
				if v.Field == "middle" {
					middle = v.Content
				}
				if v.Field == "foot" {
					foot = v.Content
				}

			}
			if head != "" {
				for _, v := range manyToOneArr {
					tmp2 := head
					tableMap := BuildNames(v.Ref)
					for _, v1 := range v.Fields {
						tmp := middle
						tmp = strings.ReplaceAll(tmp, "${refCol}", v1.Column)
						tmp = strings.ReplaceAll(tmp, "${refToCol}", v1.To)
						tmp2 += tmp
					}
					tmp := foot
					tableMap = BuildNames(v.Ref)
					tmp = strings.ReplaceAll(tmp, "${refTable_Name}", tableMap["Table"])
					tmp = strings.ReplaceAll(tmp, "${refTable_raw}", tableMap["raw"])
					tmp2 += tmp
					tmpText += tmp2
				}
			}
			tmpText = strings.TrimSuffix(tmpText, ",")
		}
		text = tmpText + text[end+len(t.End):]
	}
	//text = NestedBegin(t.Begin, t.End, t.Case, t.EndCase, text, arr, oneToOneArr, manyToOneArr, oneToManyArr)
	return text
}
