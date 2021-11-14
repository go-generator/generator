package manual

import (
	"context"
	"strings"
)

type DefaultEntityTemplate struct {
}

func (t *DefaultEntityTemplate) Merge(ctx context.Context, template string, share map[string]string, parent map[string]string, fields []map[string]string) string {
	s := template
	for k, v := range share {
		s = strings.ReplaceAll(s, "${env:"+k+"}", v)
	}
	for k, v := range parent {
		s = strings.ReplaceAll(s, "${"+k+"}", v)
	}
	return s
}
