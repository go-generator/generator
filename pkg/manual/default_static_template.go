package manual

import (
	"context"
	"strings"
)

type DefaultStaticTemplate struct {
}

func (t *DefaultStaticTemplate) Generate(ctx context.Context, template string, m map[string]string) string {
	s := template
	for k, v := range m {
		x := "${env:" + k + "}"
		s = strings.ReplaceAll(s, x, v)
	}
	return s
}
