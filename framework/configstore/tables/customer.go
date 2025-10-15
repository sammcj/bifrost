package tables

import "time"

// TableCustomer represents a customer entity with budget
type TableCustomer struct {
	ID       string  `gorm:"primaryKey;type:varchar(255)" json:"id"`
	Name     string  `gorm:"type:varchar(255);not null" json:"name"`
	BudgetID *string `gorm:"type:varchar(255);index" json:"budget_id,omitempty"`

	// Relationships
	Budget      *TableBudget      `gorm:"foreignKey:BudgetID" json:"budget,omitempty"`
	Teams       []TableTeam       `gorm:"foreignKey:CustomerID" json:"teams"`
	VirtualKeys []TableVirtualKey `gorm:"foreignKey:CustomerID" json:"virtual_keys"`

	CreatedAt time.Time `gorm:"index;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"index;not null" json:"updated_at"`
}

// TableName sets the table name for each model
func (TableCustomer) TableName() string { return "governance_customers" }
