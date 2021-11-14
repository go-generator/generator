package manual

import "context"

type EntityTemplate interface {
	Merge(ctx context.Context, template string, share map[string]string, parent map[string]string, fields []map[string]string) string
}
