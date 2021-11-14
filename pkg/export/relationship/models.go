package relationship

const (
	OneToOne    = "one to one"
	OneToMany   = "one to many"
	ManyToOne   = "many to one"
	ManyToMany  = "many to many"
	Unknown     = "unknown"
	Unsupported = "unsupported"
)

type RelTables struct {
	Table            string `gorm:"column:table"`
	Column           string `gorm:"column:column"`
	ReferencedTable  string `gorm:"column:referenced_table"`
	ReferencedColumn string `gorm:"column:referenced_column"`
	Relationship     string
}

type MySqlUnique struct {
	Column    string `gorm:"column:Column_name"`
	NonUnique bool   `gorm:"column:Non_unique"` // False mean it's unique, True means it can contain duplicate
	Key       string `gorm:"column:Key_name"`
}

type PgUnique struct {
	TableName string `gorm:"column:tablename"`
	IndexName string `gorm:"column:indexname"`
	IndexDef  string `gorm:"column:indexdef"`
}

type CompositeKey struct {
	Column string `gorm:"column:column"`
}
