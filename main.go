package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// initial count 2532
func dbConnect() *gorm.DB {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}

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

var adminUserMap = make(map[int]int)

func getAdminUser() {
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

		adminID, err := strconv.Atoi(record[0])
		if err != nil {
			continue
		}

		orgID, err := strconv.ParseInt(record[1], 10, 64)
		if err != nil {
			continue
		}

		adminUserMap[int(orgID)] = int(adminID)
	}
}

func migrateData(db *gorm.DB, batchSize int) {
	fmt.Println("\n=== Starting Migration Process ===")

	var offset int
	processedCount := 0

	if batchSize <= 0 {
		batchSize = 200
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic in transaction: %v\n", r)
		}
	}()

	// Get total count for progress tracking
	var totalCount int64
	if err := db.Raw("SELECT COUNT(*) FROM `test_environments` WHERE configuration_id IS NULL").Scan(&totalCount).Error; err != nil {
		fmt.Printf("Error getting total count: %v\n", err)
		panic("failed to get total count")
	}
	fmt.Printf("Total records to migrate: %d\n", totalCount)

	for {

		tx := db.Begin()
		if tx.Error != nil {
			fmt.Printf("Error starting transaction: %v\n", tx.Error)
			panic("failed to start transaction")
		}

		fmt.Printf("\nProcessing batch starting at offset %d...\n", offset)

		var testEnvironments []TestEnvironment
		if err := tx.Raw("SELECT * FROM `test_environments` WHERE configuration_id IS NULL ORDER BY `id` LIMIT ?  FOR UPDATE", batchSize).Scan(&testEnvironments).Error; err != nil {
			fmt.Printf("Error fetching test environments: %v\n", err)
			tx.Rollback()
			panic("failed to fetch test environments")
		}

		if len(testEnvironments) == 0 {
			fmt.Println("No more records to process")
			break
		}

		fmt.Printf("Found %d records in current batch\n", len(testEnvironments))

		// Create configurations and store mappings
		configsToCreate := make([]Configurations, 0)
		envIDToUpdate := make([]int64, 0)
		for i, env := range testEnvironments {
			fmt.Printf("Processing record %d/%d in batch (ID: %d)\n", i+1, len(testEnvironments), env.ID)

			// Safely handle Platform field
			platform := "custom"
			validPlatform := map[string]bool{
				"real-device-mobile": true,
				"Desktop":            true,
				"Mobile":             true,
				"custom":             true,
			}
			if env.Platform != nil && validPlatform[*env.Platform] {
				platform = *env.Platform
			}

			// Check if admin user exists for this organization
			adminID, exists := adminUserMap[env.OrganizationID]
			if !exists {
				// Use a default admin ID or handle the error as needed
				adminID = 1 // Default admin ID
				fmt.Printf("Warning: No admin user found for organization ID %d, using default ID %d\n", env.OrganizationID, adminID)
			}

			config := Configurations{
				CommonModelPrimaryKey: CommonModelPrimaryKey{
					ID:        env.ID,
					CreatedAt: env.CreatedAt,
					UpdatedAt: env.UpdatedAt,
				},
				OrganizationID:    env.OrganizationID,
				Name:              env.Name,
				Platform:          platform,
				IsKaneSupported:   env.IsKaneSupported,
				IsManualSupported: !env.IsKaneSupported || (env.IsKaneSupported && platform == "real-device-mobile" && !env.PrivateCloud),
				IsDefault:         env.IsDefault,
				IsCustom:          env.IsCustom,
				DeletedAt:         env.DeletedAt,
				IsComplete:        env.IsComplete,
				CreatedBy:         adminID,
				UpdatedBy:         adminID,
			}
			configsToCreate = append(configsToCreate, config)
			envIDToUpdate = append(envIDToUpdate, env.ID)

		}

		if err := tx.CreateInBatches(configsToCreate, batchSize).Error; err != nil {
			fmt.Printf("Error creating configuration: %v\n", err)
			tx.Rollback()
			panic("failed to create configuration")
		}

		// Update the test_environments table with the configuration_id
		updateSQL := fmt.Sprintf("UPDATE `test_environments` SET `configuration_id` = `id` WHERE `id` IN ?")
		fmt.Printf("Updating test_environments: %s\n", updateSQL)
		if err := tx.Exec(updateSQL, envIDToUpdate).Error; err != nil {
			fmt.Printf("Error updating configuration_id: %v\n", err)
			tx.Rollback()
			panic("failed to update configuration_id")
		}

		if err := tx.Commit().Error; err != nil {
			fmt.Printf("Error committing transaction: %v\n", err)
			tx.Rollback()
			panic("failed to commit transaction")
		}

		processedCount += len(testEnvironments)
		fmt.Printf("Batch completed. Progress: %d/%d records Batch Number: %d (%.2f%%)\n",
			processedCount, totalCount, offset/batchSize, float64(processedCount)/float64(totalCount)*100)

		offset += batchSize

		// Short delay between batches to reduce database load
		time.Sleep(1 * time.Second)
	}

	fmt.Printf("\n=== Migration Phase Completed ===\n")
	fmt.Printf("Total records processed: %d\n", processedCount)
}

func main() {
	// Define CLI flag
	batchSize := flag.Int("batchSize", 200, "Number of records to process per batch")
	flag.Parse()

	fmt.Println("Migrating data from test_environments table to configurations table")
	db := dbConnect()

	// Check database connection
	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println("Error getting database instance:", err)
		return
	}
	// Ping the database to verify connection
	if err := sqlDB.Ping(); err != nil {
		fmt.Println("Database connection failed:", err)
		return
	}
	fmt.Println("DB connection successful")

	// populate org to admin user map
	getAdminUser()

	// migrate data
	migrateData(db, *batchSize)
	fmt.Println("Migration completed successfully")
}
