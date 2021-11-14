package manual

import (
	"github.com/go-generator/core"
	"github.com/go-generator/core/generator"
	"strings"
)

func BuildFields(v metadata.Field) map[string]string {
	name := v.Name
	raw := generator.BuildSnakeName(name)
	names := map[string]string{
		"raw":      raw,
		"name":     strings.ToLower(string(name[0])) + name[1:],
		"Name":     strings.Title(name),
		"NAME":     strings.ToUpper(name),
		"constant": strings.ToUpper(raw),
		"lower":    strings.ToLower(name),
		"type":     v.Type,
		"default":  "",
	}
	if v.Key {
		names["id"] = "true"
	} else {
		names["default"] = "true"
	}
	return names
}
func JavaBuildFields(v metadata.Field) map[string]string {
	if strings.Contains(v.Type, "*") {
		return nil
	}
	name := v.Name
	raw := generator.BuildSnakeName(name)
	names := map[string]string{
		"raw":      raw,
		"name":     strings.ToLower(string(name[0])) + name[1:],
		"Name":     strings.Title(name),
		"NAME":     strings.ToUpper(name),
		"constant": strings.ToUpper(raw),
		"lower":    strings.ToLower(name),
		"type":     v.Type,
		"default":  "",
	}

	if v.Key {
		names["id"] = "true"
	} else {
		names["default"] = "true"
	}
	return names
}
func BuildNames(name string) map[string]string {
	names := make(map[string]string)
	if !strings.Contains(name, "_") {
		raw := generator.BuildSnakeName(name)
		names = map[string]string{
			"raw":      raw,
			"name":     strings.ToLower(string(name[0])) + name[1:],
			"Name":     strings.Title(name),
			"NAME":     strings.ToUpper(name),
			"constant": strings.ToUpper(raw),
			"lower":    strings.ToLower(name),
		}
	} else {
		raw := strings.ToLower(name)
		name = generator.UnBuildSnakeName(raw)
		names = map[string]string{
			"raw":       raw,
			"name":      strings.ToLower(string(name[0])) + name[1:],
			"Name":      strings.Title(name),
			"NAME":      strings.ToUpper(name),
			"constants": strings.ToUpper(raw),
			"lower":     strings.ToLower(name),
		}
	}
	return names
}
