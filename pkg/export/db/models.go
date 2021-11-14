package db

import (
	"github.com/go-generator/core"
	"strings"
)

type FilesDetails struct {
	Env    []string         `json:"env"`
	Entity []string         `json:"entity"`
	Model  string           `json:"model"`
	Files  []metadata.Model `json:"files"`
}

type Folders struct {
	ModelFile []FilesDetails `json:"folders"`
}

type Connection struct {
	TableName       string
	ReferencedTable string
	Fields          []metadata.Link
}

type JavaComPK struct {
	Package    string
	PKName     string
	Array      []string
	Capitalize []string
	AllGet     string
}

func NewJavaComPK(pkg, pkName string, arr []string) *JavaComPK {
	jPk := JavaComPK{
		Package:    pkg,
		PKName:     pkName,
		Array:      arr,
		Capitalize: make([]string, len(arr)),
		AllGet:     "",
	}
	copy(jPk.Capitalize, jPk.Array)
	for i := range jPk.Capitalize {
		jPk.Capitalize[i] = strings.Title(jPk.Capitalize[i])
		if i == len(jPk.Capitalize)-1 {
			jPk.AllGet += "get" + strings.Title(jPk.Capitalize[i]) + "()"
		} else {
			jPk.AllGet += "get" + strings.Title(jPk.Capitalize[i]) + "(),"
		}
	}
	return &jPk
}

type JavaTemplate struct {
	Package   string
	TableName string
	Array     []string
	IDFields  []string
	IDClass   string
	TableRef  []TableRef
}

type TableRef struct {
	Name        string
	JoinColumns []ColumnRef
}

type ColumnRef struct {
	Col           string
	ReferencedCol string
}

func NewJavaTemplate(pkg, table string, idField, arr []string) *JavaTemplate {
	jTem := JavaTemplate{
		Package:   pkg,
		TableName: table,
		Array:     arr,
		IDFields:  idField,
	}
	if len(idField) > 1 {
		jTem.IDClass = "@IdClass(" + strings.Title(jTem.TableName) + "PK.class)"
	}
	return &jTem
}
