package export

import (
	"context"
	"database/sql"
	"github.com/go-generator/core"
	edb "github.com/go-generator/core/export/db"
	"github.com/go-generator/core/export/relationship"
	"github.com/go-generator/core/generator"
	"strconv"
)

func ToModel(types map[string]string, table string, rt []relationship.RelTables, hasCompositeKey bool, sqlTable []edb.TableFields) (*metadata.Model, error) { //s *TableInfo, conn *gorm.DB, tables []string, packageName, output string) {
	var m metadata.Model
	tableNames := generator.BuildNames(table)
	m.Name = tableNames["Name"]
	m.Table = tableNames["name"]
	for _, v := range sqlTable {
		columns := generator.BuildNames(v.Column)
		var f metadata.Field
		if hasCompositeKey {
			f.Source = columns["name"]
		} else {
			if v.ColumnKey == "PRI" {
				f.Source = "_id"
			} else {
				f.Source = columns["name"]
			}
		}
		f.Name = columns["name"]
		f.Type = types[v.DataType]
		if v.Length.Valid {
			l, err := strconv.Atoi(v.Length.String)
			if err != nil {
				return nil, err
			}
			f.Length = l
		}
		if v.ColumnKey == "PRI" {
			f.Key = true
		}
		rl := getRelationship(v.Column, rt)
		if rl != nil {
			var rls metadata.Relationship
			var foreign metadata.Field
			tmpMap := generator.BuildNames(rl.Table)
			foreign.Name = tmpMap["name"]
			foreign.Source = tmpMap["name"]
			foreign.Type = "*[]" + tmpMap["Names"]                                        // for many to many relationship
			if rl.Relationship == relationship.ManyToOne && table == rl.ReferencedTable { // have Many to One relation, add a field to the current struct
				rls.Ref = rl.Table
				rls.Fields = append(rls.Fields, metadata.Link{
					Column: rl.Column,
					To:     rl.ReferencedColumn,
				})
				if m.Arrays == nil {
					m.Arrays = append(m.Arrays, rls)
				} else {
					for j := range m.Arrays {
						if m.Arrays[j].Ref == rls.Ref {
							m.Arrays[j].Fields = append(m.Arrays[j].Fields, rls.Fields...)
							break
						}
						if j == len(m.Arrays)-1 {
							m.Arrays = append(m.Arrays, rls)
						}
					}
				}
				for i := range m.Fields {
					if m.Fields[i] == foreign {
						break
					}
					if i == len(m.Fields)-1 {
						m.Fields = append(m.Fields, foreign)
					}
				}
			}
		}
		m.Fields = append(m.Fields, f)
	}
	return &m, nil
}

func getRelationship(column string, rt []relationship.RelTables) *relationship.RelTables {
	for _, v := range rt {
		if column == v.ReferencedColumn {
			return &v
		}
	}
	return nil
}

func ToModels(ctx context.Context, db *sql.DB, database string, tables []string, rt []relationship.RelTables, types map[string]string) ([]metadata.Model, error) {
	var projectModels []metadata.Model
	for _, t := range tables {
		var tablesData edb.TableInfo
		err := edb.InitTables(ctx, db, database, t, &tablesData)
		if err != nil {
			return nil, err
		}
		m, err := ToModel(types, t, rt, tablesData.HasCompositeKey, tablesData.Fields)
		if err != nil {
			return nil, err
		}
		projectModels = append(projectModels, *m)
	}
	return projectModels, nil
}
