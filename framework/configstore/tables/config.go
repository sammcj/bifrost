package tables

// TableConfig represents generic configuration key-value pairs
type TableConfig struct {
	Key   string `gorm:"primaryKey;type:varchar(255)" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}

// TableName sets the table name for each model
func (TableConfig) TableName() string { return "governance_config" }
