package generator

import (
	"context"
	"github.com/go-generator/core"
)

type Generator interface {
	Generate(ctx context.Context, project metadata.Project, templates map[string]string) []metadata.File
}
