package main

import (
	"fmt"
	"log"

	"github.com/restayway/gogis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Location represents a place with geographic coordinates
type Location struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Description string
	Point       gogis.Point `gorm:"type:geometry(Point,4326);not null"`
	CreatedAt   int64
}

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=yourpassword dbname=testdb port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Enable PostGIS extension (requires superuser privileges)
	// You may need to run this manually: CREATE EXTENSION IF NOT EXISTS postgis;
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis").Error; err != nil {
		log.Printf("Warning: Failed to create PostGIS extension: %v", err)
		log.Printf("Please run manually: CREATE EXTENSION IF NOT EXISTS postgis;")
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&Location{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Create sample locations
	locations := []Location{
		{
			Name:        "Statue of Liberty",
			Description: "Iconic symbol of freedom and democracy",
			Point: gogis.Point{
				Lng: -74.0445, // Longitude (X coordinate)
				Lat: 40.6892,  // Latitude (Y coordinate)
			},
		},
		{
			Name:        "Central Park",
			Description: "Large public park in Manhattan",
			Point: gogis.Point{
				Lng: -73.9665,
				Lat: 40.7812,
			},
		},
		{
			Name:        "Brooklyn Bridge",
			Description: "Historic suspension bridge",
			Point: gogis.Point{
				Lng: -73.9969,
				Lat: 40.7061,
			},
		},
		{
			Name:        "Times Square",
			Description: "Commercial intersection and tourist destination",
			Point: gogis.Point{
				Lng: -73.9857,
				Lat: 40.7580,
			},
		},
	}

	// Insert locations
	for _, location := range locations {
		if err := db.Create(&location).Error; err != nil {
			log.Printf("Failed to create location %s: %v", location.Name, err)
		} else {
			fmt.Printf("Created location: %s at (%f, %f)\n",
				location.Name, location.Point.Lng, location.Point.Lat)
		}
	}

	// Basic queries
	fmt.Println("\n=== Basic Queries ===")

	// Find all locations
	var allLocations []Location
	db.Find(&allLocations)
	fmt.Printf("Total locations: %d\n", len(allLocations))

	// Find location by name
	var centralPark Location
	db.Where("name = ?", "Central Park").First(&centralPark)
	fmt.Printf("Found: %s at (%f, %f)\n",
		centralPark.Name, centralPark.Point.Lng, centralPark.Point.Lat)

	// Spatial queries
	fmt.Println("\n=== Spatial Queries ===")

	// Find locations within 2km of Central Park (approximately 0.018 degrees)
	var nearbyLocations []Location
	centralParkLng := -73.9665
	centralParkLat := 40.7812

	err = db.Where("ST_DWithin(point, ST_Point(?, ?), ?)",
		centralParkLng, centralParkLat, 0.018).Find(&nearbyLocations).Error
	if err != nil {
		log.Printf("Spatial query failed: %v", err)
	} else {
		fmt.Printf("Locations within 2km of Central Park: %d\n", len(nearbyLocations))
		for _, loc := range nearbyLocations {
			fmt.Printf("- %s\n", loc.Name)
		}
	}

	// Find nearest locations to a point (using distance ordering)
	var nearestLocations []Location
	err = db.Order("ST_Distance(point, ST_Point(?, ?))").
		Limit(3).Find(&nearestLocations, centralParkLng, centralParkLat).Error
	if err != nil {
		log.Printf("Nearest query failed: %v", err)
	} else {
		fmt.Println("\nNearest locations to Central Park:")
		for i, loc := range nearestLocations {
			fmt.Printf("%d. %s\n", i+1, loc.Name)
		}
	}

	// Calculate distances
	fmt.Println("\n=== Distance Calculations ===")

	type LocationWithDistance struct {
		Location
		Distance float64 `gorm:"column:distance"`
	}

	var locationsWithDistance []LocationWithDistance
	err = db.Table("locations").
		Select("*, ST_Distance(point, ST_Point(?, ?)) as distance",
			centralParkLng, centralParkLat).
		Order("distance").
		Find(&locationsWithDistance).Error

	if err != nil {
		log.Printf("Distance calculation failed: %v", err)
	} else {
		fmt.Println("Locations with distances from Central Park:")
		for _, loc := range locationsWithDistance {
			fmt.Printf("- %s: %.6f degrees\n", loc.Name, loc.Distance)
		}
	}

	// Cleanup (optional)
	fmt.Println("\n=== Cleanup ===")
	if err := db.Where("1 = 1").Delete(&Location{}).Error; err != nil {
		log.Printf("Failed to cleanup: %v", err)
	} else {
		fmt.Println("Cleaned up all test locations")
	}
}
