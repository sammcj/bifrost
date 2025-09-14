package configstore

import (
	"fmt"

	"github.com/maximhq/bifrost/framework/configstore/internal/migration"
	"gorm.io/gorm"
)

// Migrate performs the necessary database migrations.
func triggerMigrations(db *gorm.DB) error {
	var err error
	err = migrationInit(db)
	if err != nil {
		return err
	}
	return nil
}

// migrationInit is the first migration
func migrationInit(db *gorm.DB) error {
	m := migration.New(db, migration.DefaultOptions, []*migration.Migration{{
		ID: "init",
		Migrate: func(tx *gorm.DB) error {
			migrator := tx.Migrator()
			if !migrator.HasTable(&TableConfigHash{}) {
				if err := migrator.CreateTable(&TableConfigHash{}); err != nil {
					return err
				}
			}
			if !migrator.HasTable(&TableProvider{}) {
				if err := migrator.CreateTable(&TableProvider{}); err != nil {
					return err
				}
			}
			if !migrator.HasTable(&TableKey{}) {
				if err := migrator.CreateTable(&TableKey{}); err != nil {
					return err
				}
			}
			if !migrator.HasTable(&TableModel{}) {
				if err := migrator.CreateTable(&TableModel{}); err != nil {
					return err
				}
			}
			if !migrator.HasTable(&TableMCPClient{}) {
				if err := migrator.CreateTable(&TableMCPClient{}); err != nil {
					return err
				}
			}
			if !migrator.HasTable(&TableClientConfig{}) {
				if err := migrator.CreateTable(&TableClientConfig{}); err != nil {
					return err
				}
			} else if !migrator.HasColumn(&TableClientConfig{}, "max_request_body_size_mb") {
				if err := migrator.AddColumn(&TableClientConfig{}, "max_request_body_size_mb"); err != nil {
					return err
				}
			}
			if !migrator.HasTable(&TableEnvKey{}) {
				if err := migrator.CreateTable(&TableEnvKey{}); err != nil {
					return err
				}
			}
			if !migrator.HasTable(&TableVectorStoreConfig{}) {
				if err := migrator.CreateTable(&TableVectorStoreConfig{}); err != nil {
					return err
				}
			}
			if !migrator.HasTable(&TableLogStoreConfig{}) {
				if err := migrator.CreateTable(&TableLogStoreConfig{}); err != nil {
					return err
				}
			}
			if !migrator.HasTable(&TableBudget{}) {
				if err := migrator.CreateTable(&TableBudget{}); err != nil {
					return err
				}
			}
			if !migrator.HasTable(&TableRateLimit{}) {
				if err := migrator.CreateTable(&TableRateLimit{}); err != nil {
					return err
				}
			}
			if !migrator.HasTable(&TableCustomer{}) {
				if err := migrator.CreateTable(&TableCustomer{}); err != nil {
					return err
				}
			}
			if !migrator.HasTable(&TableTeam{}) {
				if err := migrator.CreateTable(&TableTeam{}); err != nil {
					return err
				}
			}
			if !migrator.HasTable(&TableVirtualKey{}) {
				if err := migrator.CreateTable(&TableVirtualKey{}); err != nil {
					return err
				}
			}
			if !migrator.HasTable(&TableConfig{}) {
				if err := migrator.CreateTable(&TableConfig{}); err != nil {
					return err
				}
			}
			if !migrator.HasTable(&TableModelPricing{}) {
				if err := migrator.CreateTable(&TableModelPricing{}); err != nil {
					return err
				}
			}
			if !migrator.HasTable(&TablePlugin{}) {
				if err := migrator.CreateTable(&TablePlugin{}); err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			migrator := tx.Migrator()
			// Drop children first, then parents (adjust if your actual FKs differ)
			if err := migrator.DropTable(&TableVirtualKey{}); err != nil {
				return err
			}
			if err := migrator.DropTable(&TableKey{}); err != nil {
				return err
			}
			if err := migrator.DropTable(&TableTeam{}); err != nil {
				return err
			}
			if err := migrator.DropTable(&TableProvider{}); err != nil {
				return err
			}
			if err := migrator.DropTable(&TableCustomer{}); err != nil {
				return err
			}
			if err := migrator.DropTable(&TableBudget{}); err != nil {
				return err
			}
			if err := migrator.DropTable(&TableRateLimit{}); err != nil {
				return err
			}
			if err := migrator.DropTable(&TableModel{}); err != nil {
				return err
			}
			if err := migrator.DropTable(&TableMCPClient{}); err != nil {
				return err
			}
			if err := migrator.DropTable(&TableClientConfig{}); err != nil {
				return err
			}
			if err := migrator.DropTable(&TableEnvKey{}); err != nil {
				return err
			}
			if err := migrator.DropTable(&TableVectorStoreConfig{}); err != nil {
				return err
			}
			if err := migrator.DropTable(&TableLogStoreConfig{}); err != nil {
				return err
			}
			if err := migrator.DropTable(&TableConfig{}); err != nil {
				return err
			}
			if err := migrator.DropTable(&TableModelPricing{}); err != nil {
				return err
			}
			if err := migrator.DropTable(&TablePlugin{}); err != nil {
				return err
			}
			if err := migrator.DropTable(&TableConfigHash{}); err != nil {
				return err
			}
			return nil
		},
	}})
	err := m.Migrate()
	if err != nil {
		return fmt.Errorf("error while running db migration: %s", err.Error())
	}
	return nil
}
