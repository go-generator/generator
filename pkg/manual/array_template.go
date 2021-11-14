package manual

import "context"

type ArrayTemplate interface {
	Array(ctx context.Context, template string, share map[string]string, selfMap map[string]string, arr []map[string]string) string
}
