package utilities

import "strings"

func InitShareMap(packageName, projectName string) map[string]string {
	return map[string]string{
		"package_name": packageName,
		"projectName":  projectName,
		"Directory":    strings.Title(projectName),
	}
}

func InitFullMap(env string, entity string, projectName string) map[string]string {
	return map[string]string{
		"static_package": env,
		"projectName":    projectName,
		"Directory":      strings.Title(projectName),
		"Entity":         entity,
		"entity":         strings.ToLower(string(entity[0])) + entity[1:],
	}
}
func MergeTemplate(template string, fullMap map[string]string) string {
	text := template
	for k, v := range fullMap {
		text = strings.ReplaceAll(text, k, v)
	}
	return text
}
func MergeArrayTemplate(template string, share map[string]string, arr []map[string]string) string {
	text := template
	for k, v := range share {
		text = strings.ReplaceAll(text, k, v)
	}
	for strings.Contains(text, "{begin}") {
		begin := strings.Index(text, "{begin}")
		end := strings.Index(text, "{end}")
		tmpText := text[:begin]
		for j := 0; j < len(arr); j++ {
			tmp := text[begin+len("{begin}") : end-1]
			for k, v := range arr[j] {
				tmp = strings.ReplaceAll(tmp, k, v)
			}
			tmpText += tmp
		}
		text = tmpText + text[end+len("{end}"):]
		//text = strings.Replace(text, "{begin}", "", 1)
		//text = strings.Replace(text, "{end}", "", 1)
	}
	return text
}

func BuildFileName(s string, path string) string {
	ext := ".go"
	if strings.Contains(path, ".") {
		ext = ""
	}
	path = strings.ReplaceAll(path, "/", "_")
	if s == "" {
		return path + ext
	}
	s2 := strings.ToLower(s)
	s3 := ""
	for i := range s {
		if s2[i] != s[i] {
			s3 += "_" + string(s2[i])
		} else {
			s3 += string(s2[i])
		}
	}

	return s3[1:] + "_" + path + ext
}
