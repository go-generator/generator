package manual

import "context"

type ModelTemplate interface {
	Model(ctx context.Context, template string, share map[string]string, selfMap map[string]string, arr, oneToOneArr, manyToOneArr, oneToManyArr []map[string]string) string
}
