package main

import (
	"fmt"
	"log"

	"github.com/restayway/gogis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Place represents a complex geographical entity that may contain multiple geometry types
type Place struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Type        string // e.g., "campus", "complex", "infrastructure"
	Description string
	Geometries  gogis.GeometryCollection `gorm:"type:geometry(GeometryCollection,4326);not null"`
	CreatedAt   int64
}

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=yourpassword dbname=testdb port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&Place{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Create sample places with geometry collections
	places := []Place{
		{
			Name:        "University Campus",
			Type:        "campus",
			Description: "Campus with buildings (points), walkways (lines), and grounds (polygon)",
			Geometries: gogis.GeometryCollection{
				Geometries: []gogis.Geometry{
					// Main campus area (polygon)
					&gogis.Polygon{
						Rings: [][]gogis.Point{
							{
								{Lng: -73.9600, Lat: 40.8000},
								{Lng: -73.9500, Lat: 40.8000},
								{Lng: -73.9500, Lat: 40.8100},
								{Lng: -73.9600, Lat: 40.8100},
								{Lng: -73.9600, Lat: 40.8000},
							},
						},
					},
					// Library building (point)
					&gogis.Point{Lng: -73.9550, Lat: 40.8050},
					// Student center (point)
					&gogis.Point{Lng: -73.9520, Lat: 40.8030},
					// Main walkway (linestring)
					&gogis.LineString{
						Points: []gogis.Point{
							{Lng: -73.9600, Lat: 40.8010},
							{Lng: -73.9550, Lat: 40.8050},
							{Lng: -73.9500, Lat: 40.8090},
						},
					},
				},
			},
		},
		{
			Name:        "Transportation Hub",
			Type:        "infrastructure",
			Description: "Complex with station areas, platforms, and access routes",
			Geometries: gogis.GeometryCollection{
				Geometries: []gogis.Geometry{
					// Main terminal building (polygon)
					&gogis.Polygon{
						Rings: [][]gogis.Point{
							{
								{Lng: -74.0100, Lat: 40.7200},
								{Lng: -74.0050, Lat: 40.7200},
								{Lng: -74.0050, Lat: 40.7250},
								{Lng: -74.0100, Lat: 40.7250},
								{Lng: -74.0100, Lat: 40.7200},
							},
						},
					},
					// Information kiosk (point)
					&gogis.Point{Lng: -74.0075, Lat: 40.7225},
					// Platform 1 (linestring)
					&gogis.LineString{
						Points: []gogis.Point{
							{Lng: -74.0080, Lat: 40.7210},
							{Lng: -74.0080, Lat: 40.7240},
						},
					},
					// Platform 2 (linestring)
					&gogis.LineString{
						Points: []gogis.Point{
							{Lng: -74.0070, Lat: 40.7210},
							{Lng: -74.0070, Lat: 40.7240},
						},
					},
					// Access road (linestring)
					&gogis.LineString{
						Points: []gogis.Point{
							{Lng: -74.0100, Lat: 40.7225},
							{Lng: -74.0120, Lat: 40.7225},
							{Lng: -74.0140, Lat: 40.7230},
						},
					},
				},
			},
		},
		{
			Name:        "Mixed Development",
			Type:        "complex",
			Description: "Development with various geometric features",
			Geometries: gogis.GeometryCollection{
				Geometries: []gogis.Geometry{
					// Residential area (polygon)
					&gogis.Polygon{
						Rings: [][]gogis.Point{
							{
								{Lng: -73.9700, Lat: 40.7500},
								{Lng: -73.9650, Lat: 40.7500},
								{Lng: -73.9650, Lat: 40.7550},
								{Lng: -73.9700, Lat: 40.7550},
								{Lng: -73.9700, Lat: 40.7500},
							},
						},
					},
					// Commercial area (polygon)
					&gogis.Polygon{
						Rings: [][]gogis.Point{
							{
								{Lng: -73.9650, Lat: 40.7500},
								{Lng: -73.9600, Lat: 40.7500},
								{Lng: -73.9600, Lat: 40.7530},
								{Lng: -73.9650, Lat: 40.7530},
								{Lng: -73.9650, Lat: 40.7500},
							},
						},
					},
					// Shopping center entrance (point)
					&gogis.Point{Lng: -73.9625, Lat: 40.7515},
					// Main street (linestring)
					&gogis.LineString{
						Points: []gogis.Point{
							{Lng: -73.9700, Lat: 40.7525},
							{Lng: -73.9600, Lat: 40.7525},
						},
					},
				},
			},
		},
	}

	// Insert places
	for _, place := range places {
		if err := db.Create(&place).Error; err != nil {
			log.Printf("Failed to create place %s: %v", place.Name, err)
		} else {
			fmt.Printf("Created place: %s (%s) with %d geometries\n",
				place.Name, place.Type, len(place.Geometries.Geometries))
		}
	}

	// Query examples
	fmt.Println("\n=== GeometryCollection Queries ===")

	// Find all places
	var allPlaces []Place
	db.Find(&allPlaces)
	fmt.Printf("Total places: %d\n", len(allPlaces))

	// Display place details
	for _, place := range allPlaces {
		fmt.Printf("\nPlace: %s (%s)\n", place.Name, place.Type)
		fmt.Printf("Description: %s\n", place.Description)
		fmt.Printf("Geometry count: %d\n", len(place.Geometries.Geometries))
		fmt.Printf("WKT: %s\n", place.Geometries.String())

		// Analyze geometry types in collection
		geometryTypes := make(map[string]int)
		for _, geom := range place.Geometries.Geometries {
			switch geom.(type) {
			case *gogis.Point:
				geometryTypes["Point"]++
			case *gogis.LineString:
				geometryTypes["LineString"]++
			case *gogis.Polygon:
				geometryTypes["Polygon"]++
			}
		}

		fmt.Printf("Geometry breakdown:")
		for geomType, count := range geometryTypes {
			fmt.Printf(" %s: %d", geomType, count)
		}
		fmt.Println()
	}

	// Spatial queries with GeometryCollection
	fmt.Println("\n=== Spatial Analysis ===")

	// Find places that contain a specific point
	testPoint := gogis.Point{Lng: -73.9550, Lat: 40.8050} // Library location

	var containingPlaces []Place
	err = db.Where("ST_Contains(geometries, ST_Point(?, ?))",
		testPoint.Lng, testPoint.Lat).Find(&containingPlaces).Error

	if err != nil {
		log.Printf("Containment query failed: %v", err)
	} else {
		fmt.Printf("Places containing point (%.4f, %.4f): %d\n",
			testPoint.Lng, testPoint.Lat, len(containingPlaces))
		for _, place := range containingPlaces {
			fmt.Printf("- %s\n", place.Name)
		}
	}

	// Find places within distance of a point
	var nearbyPlaces []Place
	err = db.Where("ST_DWithin(geometries, ST_Point(?, ?), ?)",
		testPoint.Lng, testPoint.Lat, 0.02).Find(&nearbyPlaces).Error

	if err != nil {
		log.Printf("Distance query failed: %v", err)
	} else {
		fmt.Printf("\nPlaces within 2km of test point: %d\n", len(nearbyPlaces))
		for _, place := range nearbyPlaces {
			fmt.Printf("- %s\n", place.Name)
		}
	}

	// Calculate areas of polygon components
	fmt.Println("\n=== Component Analysis ===")

	type PlaceWithArea struct {
		Place
		TotalArea float64 `gorm:"column:total_area"`
	}

	var placesWithArea []PlaceWithArea
	err = db.Table("places").
		Select("*, ST_Area(ST_Transform(geometries, 3857)) as total_area").
		Find(&placesWithArea).Error

	if err != nil {
		log.Printf("Area calculation failed: %v", err)
	} else {
		fmt.Println("Total areas (polygons only):")
		for _, place := range placesWithArea {
			if place.TotalArea > 0 {
				fmt.Printf("- %s: %.2f square meters\n", place.Name, place.TotalArea)
			}
		}
	}

	// Find intersecting places
	fmt.Println("\n=== Intersection Analysis ===")

	var universityPlace Place
	db.Where("name = ?", "University Campus").First(&universityPlace)

	var intersectingPlaces []Place
	err = db.Where("id != ? AND ST_Intersects(geometries, ?)",
		universityPlace.ID, universityPlace.Geometries.String()).Find(&intersectingPlaces).Error

	if err != nil {
		log.Printf("Intersection query failed: %v", err)
	} else {
		fmt.Printf("Places intersecting with University Campus: %d\n", len(intersectingPlaces))
		for _, place := range intersectingPlaces {
			fmt.Printf("- %s\n", place.Name)
		}
	}

	// Extract specific geometry types from collections
	fmt.Println("\n=== Geometry Type Extraction ===")

	// Note: PostGIS provides functions like ST_CollectionExtract to extract specific geometry types
	type PlaceWithPoints struct {
		Place
		PointCount int `gorm:"column:point_count"`
	}

	var placesWithPoints []PlaceWithPoints
	err = db.Table("places").
		Select("*, ST_NumGeometries(ST_CollectionExtract(geometries, 1)) as point_count").
		Find(&placesWithPoints).Error

	if err != nil {
		log.Printf("Point extraction failed: %v", err)
	} else {
		fmt.Println("Point counts in geometry collections:")
		for _, place := range placesWithPoints {
			fmt.Printf("- %s: %d points\n", place.Name, place.PointCount)
		}
	}

	// Bounding box analysis
	fmt.Println("\n=== Bounding Box Analysis ===")

	type PlaceWithBounds struct {
		Place
		BoundingBox string `gorm:"column:bounding_box"`
	}

	var placesWithBounds []PlaceWithBounds
	err = db.Table("places").
		Select("*, ST_AsText(ST_Envelope(geometries)) as bounding_box").
		Find(&placesWithBounds).Error

	if err != nil {
		log.Printf("Bounding box calculation failed: %v", err)
	} else {
		fmt.Println("Bounding boxes:")
		for _, place := range placesWithBounds {
			fmt.Printf("- %s: %s\n", place.Name, place.BoundingBox)
		}
	}

	// Complex spatial relationships
	fmt.Println("\n=== Complex Relationships ===")

	// Find places where any geometry component intersects with our test point's buffer
	var placesInBuffer []Place
	err = db.Where("ST_Intersects(geometries, ST_Buffer(ST_Point(?, ?), ?))",
		testPoint.Lng, testPoint.Lat, 0.01).Find(&placesInBuffer).Error

	if err != nil {
		log.Printf("Buffer intersection failed: %v", err)
	} else {
		fmt.Printf("Places intersecting 1km buffer around test point: %d\n", len(placesInBuffer))
		for _, place := range placesInBuffer {
			fmt.Printf("- %s\n", place.Name)
		}
	}

	// Cleanup
	fmt.Println("\n=== Cleanup ===")
	if err := db.Where("1 = 1").Delete(&Place{}).Error; err != nil {
		log.Printf("Failed to cleanup: %v", err)
	} else {
		fmt.Println("Cleaned up all test places")
	}
}
