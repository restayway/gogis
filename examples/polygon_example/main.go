package main

import (
	"fmt"
	"log"

	"github.com/restayway/gogis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Region represents a geographic area defined by a polygon
type Region struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Type        string // e.g., "park", "neighborhood", "district"
	Description string
	Area        gogis.Polygon `gorm:"type:geometry(Polygon,4326);not null"`
	AreaSize    float64       // Area in square meters (calculated)
	CreatedAt   int64
}

// Location represents a point location for testing containment
type Location struct {
	ID    uint        `gorm:"primaryKey"`
	Name  string      `gorm:"not null"`
	Point gogis.Point `gorm:"type:geometry(Point,4326);not null"`
}

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=yourpassword dbname=testdb port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&Region{}, &Location{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Create sample regions (polygons)
	regions := []Region{
		{
			Name:        "Central Park",
			Type:        "park",
			Description: "Large public park in Manhattan",
			Area: gogis.Polygon{
				Rings: [][]gogis.Point{
					{ // Outer boundary (simplified rectangle)
						{Lng: -73.9812, Lat: 40.7644}, // Southwest corner
						{Lng: -73.9734, Lat: 40.7644}, // Southeast corner
						{Lng: -73.9734, Lat: 40.7947}, // Northeast corner
						{Lng: -73.9812, Lat: 40.7947}, // Northwest corner
						{Lng: -73.9812, Lat: 40.7644}, // Close the ring
					},
				},
			},
		},
		{
			Name:        "Times Square District",
			Type:        "district",
			Description: "Commercial and entertainment district",
			Area: gogis.Polygon{
				Rings: [][]gogis.Point{
					{
						{Lng: -73.9885, Lat: 40.7540}, // Southwest
						{Lng: -73.9835, Lat: 40.7540}, // Southeast
						{Lng: -73.9835, Lat: 40.7610}, // Northeast
						{Lng: -73.9885, Lat: 40.7610}, // Northwest
						{Lng: -73.9885, Lat: 40.7540}, // Close
					},
				},
			},
		},
		{
			Name:        "Complex Polygon with Hole",
			Type:        "example",
			Description: "Polygon with internal hole (donut shape)",
			Area: gogis.Polygon{
				Rings: [][]gogis.Point{
					{ // Outer ring
						{Lng: -74.0200, Lat: 40.7000},
						{Lng: -74.0100, Lat: 40.7000},
						{Lng: -74.0100, Lat: 40.7100},
						{Lng: -74.0200, Lat: 40.7100},
						{Lng: -74.0200, Lat: 40.7000},
					},
					{ // Inner ring (hole)
						{Lng: -74.0180, Lat: 40.7020},
						{Lng: -74.0180, Lat: 40.7080},
						{Lng: -74.0120, Lat: 40.7080},
						{Lng: -74.0120, Lat: 40.7020},
						{Lng: -74.0180, Lat: 40.7020},
					},
				},
			},
		},
	}

	// Insert regions
	for _, region := range regions {
		if err := db.Create(&region).Error; err != nil {
			log.Printf("Failed to create region %s: %v", region.Name, err)
		} else {
			fmt.Printf("Created region: %s (%s) with %d rings\n",
				region.Name, region.Type, len(region.Area.Rings))
		}
	}

	// Create sample locations for testing containment
	locations := []Location{
		{Name: "Bethesda Fountain", Point: gogis.Point{Lng: -73.9712, Lat: 40.7739}}, // In Central Park
		{Name: "TKTS Red Steps", Point: gogis.Point{Lng: -73.9857, Lat: 40.7580}},    // In Times Square
		{Name: "Statue of Liberty", Point: gogis.Point{Lng: -74.0445, Lat: 40.6892}}, // Outside both
		{Name: "Hole Center", Point: gogis.Point{Lng: -74.0150, Lat: 40.7050}},       // In the hole
		{Name: "Outer Ring", Point: gogis.Point{Lng: -74.0110, Lat: 40.7010}},        // In outer ring but not hole
	}

	for _, location := range locations {
		if err := db.Create(&location).Error; err != nil {
			log.Printf("Failed to create location %s: %v", location.Name, err)
		}
	}

	// Query examples
	fmt.Println("\n=== Polygon Queries ===")

	// Find all regions
	var allRegions []Region
	db.Find(&allRegions)
	fmt.Printf("Total regions: %d\n", len(allRegions))

	// Display region details
	for _, region := range allRegions {
		fmt.Printf("\nRegion: %s (%s)\n", region.Name, region.Type)
		fmt.Printf("Rings: %d (outer + %d holes)\n", len(region.Area.Rings), len(region.Area.Rings)-1)
		fmt.Printf("WKT: %s\n", region.Area.String())
	}

	// Calculate polygon areas
	fmt.Println("\n=== Area Calculations ===")

	type RegionWithArea struct {
		Region
		AreaSquareMeters float64 `gorm:"column:area_square_meters"`
	}

	var regionsWithArea []RegionWithArea
	err = db.Table("regions").
		Select("*, ST_Area(ST_Transform(area, 3857)) as area_square_meters").
		Find(&regionsWithArea).Error

	if err != nil {
		log.Printf("Area calculation failed: %v", err)
	} else {
		fmt.Println("Region areas:")
		for _, region := range regionsWithArea {
			fmt.Printf("- %s: %.2f square meters\n", region.Name, region.AreaSquareMeters)
		}
	}

	// Point-in-polygon queries
	fmt.Println("\n=== Point Containment Analysis ===")

	var allLocations []Location
	db.Find(&allLocations)

	for _, location := range allLocations {
		fmt.Printf("\nLocation: %s (%.4f, %.4f)\n",
			location.Name, location.Point.Lng, location.Point.Lat)

		// Find which regions contain this point
		var containingRegions []Region
		err = db.Where("ST_Contains(area, ST_Point(?, ?))",
			location.Point.Lng, location.Point.Lat).Find(&containingRegions).Error

		if err != nil {
			log.Printf("Containment query failed: %v", err)
		} else if len(containingRegions) > 0 {
			fmt.Printf("  Contained in:")
			for _, region := range containingRegions {
				fmt.Printf(" %s (%s)", region.Name, region.Type)
			}
			fmt.Println()
		} else {
			fmt.Printf("  Not contained in any region\n")
		}

		// Check if point is within a certain distance of polygon boundaries
		var nearbyRegions []Region
		err = db.Where("ST_DWithin(area, ST_Point(?, ?), ?)",
			location.Point.Lng, location.Point.Lat, 0.01).Find(&nearbyRegions).Error

		if err == nil && len(nearbyRegions) > 0 {
			fmt.Printf("  Within 1km of:")
			for _, region := range nearbyRegions {
				fmt.Printf(" %s", region.Name)
			}
			fmt.Println()
		}
	}

	// Polygon relationship queries
	fmt.Println("\n=== Polygon Relationships ===")

	// Find overlapping regions
	var centralPark Region
	db.Where("name = ?", "Central Park").First(&centralPark)

	var overlappingRegions []Region
	err = db.Where("id != ? AND ST_Overlaps(area, ?)",
		centralPark.ID, centralPark.Area.String()).Find(&overlappingRegions).Error

	if err != nil {
		log.Printf("Overlap query failed: %v", err)
	} else {
		fmt.Printf("Regions overlapping with Central Park: %d\n", len(overlappingRegions))
		for _, region := range overlappingRegions {
			fmt.Printf("- %s\n", region.Name)
		}
	}

	// Find adjacent regions (touching but not overlapping)
	var adjacentRegions []Region
	err = db.Where("id != ? AND ST_Touches(area, ?)",
		centralPark.ID, centralPark.Area.String()).Find(&adjacentRegions).Error

	if err != nil {
		log.Printf("Adjacent query failed: %v", err)
	} else {
		fmt.Printf("Regions adjacent to Central Park: %d\n", len(adjacentRegions))
		for _, region := range adjacentRegions {
			fmt.Printf("- %s\n", region.Name)
		}
	}

	// Buffer operations
	fmt.Println("\n=== Buffer Operations ===")

	// Create a buffer around Central Park and find what it contains
	var locationsInBuffer []Location
	err = db.Where("ST_Within(point, ST_Buffer(?, ?))",
		centralPark.Area.String(), 0.005).Find(&locationsInBuffer).Error // ~500m buffer

	if err != nil {
		log.Printf("Buffer query failed: %v", err)
	} else {
		fmt.Printf("Locations within 500m buffer of Central Park: %d\n", len(locationsInBuffer))
		for _, location := range locationsInBuffer {
			fmt.Printf("- %s\n", location.Name)
		}
	}

	// Centroid calculation
	fmt.Println("\n=== Centroids ===")

	type RegionWithCentroid struct {
		Region
		CentroidWKT string `gorm:"column:centroid"`
	}

	var regionsWithCentroid []RegionWithCentroid
	err = db.Table("regions").
		Select("*, ST_AsText(ST_Centroid(area)) as centroid").
		Find(&regionsWithCentroid).Error

	if err != nil {
		log.Printf("Centroid calculation failed: %v", err)
	} else {
		fmt.Println("Region centroids:")
		for _, region := range regionsWithCentroid {
			fmt.Printf("- %s: %s\n", region.Name, region.CentroidWKT)
		}
	}

	// Cleanup
	fmt.Println("\n=== Cleanup ===")
	if err := db.Where("1 = 1").Delete(&Region{}).Error; err != nil {
		log.Printf("Failed to cleanup regions: %v", err)
	}
	if err := db.Where("1 = 1").Delete(&Location{}).Error; err != nil {
		log.Printf("Failed to cleanup locations: %v", err)
	}
	if err == nil {
		fmt.Println("Cleaned up all test data")
	}
}
