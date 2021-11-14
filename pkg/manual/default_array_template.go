package manual

import (
	"context"
	"strings"
)

type DefaultArrayTemplate struct {
	Begin string
	End   string
}

func NewArrayTemplate(begin string, end string) *DefaultArrayTemplate {
	return &DefaultArrayTemplate{begin, end}
}
func (t *DefaultArrayTemplate) Array(ctx context.Context, template string, share map[string]string, selfMap map[string]string, arr []map[string]string) string {
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
		tmpText := text[:begin]

		for j := 0; j < len(arr); j++ {
			tmp := subString
			for k, v := range arr[j] {
				x := "${" + k + "}"
				tmp = strings.ReplaceAll(tmp, x, v)
			}
			tmpText += tmp
		}
		tmpText = strings.TrimSuffix(tmpText, ",")
		text = tmpText + text[end+len(t.End):]
	}
	return text
}
