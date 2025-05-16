package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func dbConnect() *gorm.DB {
	dbAddress := os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") +
		"@(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" +
		os.Getenv("DB_NAME") + "?parseTime=true"
	dialector := &mysql.Dialector{}
	dialector.Config = &mysql.Config{
		DriverName: "mysql",
		DSN:        dbAddress,
	}
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

func getAdminUser(organizationID uint) int {
	// get from csv
	csvFile, err := os.Open("admin_users.csv")
	if err != nil {
		panic("failed to open admin_users.csv")
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	records, err := reader.ReadAll()
	if err != nil {
		panic("failed to read admin_users.csv")
	}
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 2 {
			continue
		}

		orgID, err := strconv.ParseUint(record[1], 10, 32)
		if err != nil {
			continue
		}

		if uint(orgID) == organizationID {
			adminID, err := strconv.Atoi(record[0])
			if err != nil {
				continue
			}
			return adminID
		}
	}

	// Return -1 if no matching organization ID is found
	return -1
}

func migrateData(db *gorm.DB) {
	// Map to store env.ID to config.ID
	envToConfigMap := make(map[uint]uint)

	var offset uint
	batchSize := 200

	for {
		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		var testEnvironments []TestEnvironment
		if err := tx.Limit(batchSize).Offset(int(offset)).Find(&testEnvironments).Error; err != nil {
			tx.Rollback()
			panic("failed to fetch test environments")
		}

		if len(testEnvironments) == 0 {
			tx.Commit()
			break
		}

		// Create configurations and store mappings
		for _, env := range testEnvironments {
			config := Configurations{
				CommonModelPrimaryKey: CommonModelPrimaryKey{
					ID:        env.ID,
					CreatedAt: env.CreatedAt,
					UpdatedAt: env.UpdatedAt,
				},
				OrganizationID:    env.OrganizationID,
				Name:              env.Name,
				Platform:          *env.Platform,
				IsKaneSupported:   env.IsKaneSupported,
				IsManualSupported: !env.IsKaneSupported || (env.IsKaneSupported && *env.Platform == "real-device-mobile"),
				IsDefault:         env.IsDefault,
				IsCustom:          env.IsCustom,
				DeletedAt:         env.DeletedAt,
				IsComplete:        env.IsComplete,
				CreatedBy:         getAdminUser(uint(env.OrganizationID)),
				UpdatedBy:         getAdminUser(uint(env.OrganizationID)),
			}

			if err := tx.Create(&config).Error; err != nil {
				tx.Rollback()
				panic("failed to create configuration")
			}
			envToConfigMap[uint(env.ID)] = uint(config.ID)
		}

		if err := tx.Commit().Error; err != nil {
			panic("failed to commit transaction")
		}

		offset += uint(batchSize)
	}

	// Backfill configuration_id in TestEnvironment
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for envID, configID := range envToConfigMap {
		if err := tx.Model(&TestEnvironment{}).Where("id = ?", envID).
			Update("configuration_id", configID).Error; err != nil {
			tx.Rollback()
			panic("failed to backfill configuration_id")
		}
	}

	if err := tx.Commit().Error; err != nil {
		panic("failed to commit backfill transaction")
	}
}

func main() {
	fmt.Println("Migrating data from test_environments table to configurations table")
	db := dbConnect()
	migrateData(db)
	fmt.Println("Migration completed successfully")
}
