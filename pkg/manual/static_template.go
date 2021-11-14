package manual

import "context"

type StaticTemplate interface {
	Generate(ctx context.Context, template string, fullMap map[string]string) string
	GenerateTS(ctx context.Context, template string, m map[string]interface{}) string
}
