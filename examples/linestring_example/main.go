package main

import (
	"fmt"
	"log"

	"github.com/restayway/gogis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Route represents a path or route with multiple waypoints
type Route struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Description string
	Path        gogis.LineString `gorm:"type:geometry(LineString,4326);not null"`
	Length      float64          // Length in meters (calculated)
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
	if err := db.AutoMigrate(&Route{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Create sample routes
	routes := []Route{
		{
			Name:        "Broadway Walk",
			Description: "Walking tour along Broadway in Manhattan",
			Path: gogis.LineString{
				Points: []gogis.Point{
					{Lng: -73.9857, Lat: 40.7580}, // Times Square
					{Lng: -73.9857, Lat: 40.7614}, // North on Broadway
					{Lng: -73.9856, Lat: 40.7648}, // Continue north
					{Lng: -73.9855, Lat: 40.7682}, // Further north
					{Lng: -73.9854, Lat: 40.7716}, // End point
				},
			},
		},
		{
			Name:        "Central Park Loop",
			Description: "Scenic loop around Central Park",
			Path: gogis.LineString{
				Points: []gogis.Point{
					{Lng: -73.9734, Lat: 40.7644}, // Southeast corner
					{Lng: -73.9734, Lat: 40.7947}, // Northeast corner
					{Lng: -73.9812, Lat: 40.7947}, // Northwest corner
					{Lng: -73.9812, Lat: 40.7644}, // Southwest corner
					{Lng: -73.9734, Lat: 40.7644}, // Back to start (closed loop)
				},
			},
		},
		{
			Name:        "Brooklyn Bridge Approach",
			Description: "Route from Manhattan to Brooklyn Bridge",
			Path: gogis.LineString{
				Points: []gogis.Point{
					{Lng: -74.0059, Lat: 40.7127}, // City Hall area
					{Lng: -73.9997, Lat: 40.7081}, // Approach
					{Lng: -73.9969, Lat: 40.7061}, // Brooklyn Bridge center
					{Lng: -73.9941, Lat: 40.7041}, // Brooklyn side
				},
			},
		},
	}

	// Insert routes
	for _, route := range routes {
		if err := db.Create(&route).Error; err != nil {
			log.Printf("Failed to create route %s: %v", route.Name, err)
		} else {
			fmt.Printf("Created route: %s with %d waypoints\n",
				route.Name, len(route.Path.Points))
		}
	}

	// Query examples
	fmt.Println("\n=== LineString Queries ===")

	// Find all routes
	var allRoutes []Route
	db.Find(&allRoutes)
	fmt.Printf("Total routes: %d\n", len(allRoutes))

	// Display route details
	for _, route := range allRoutes {
		fmt.Printf("\nRoute: %s\n", route.Name)
		fmt.Printf("Description: %s\n", route.Description)
		fmt.Printf("Waypoints: %d\n", len(route.Path.Points))
		fmt.Printf("WKT: %s\n", route.Path.String())
	}

	// Spatial queries with LineStrings
	fmt.Println("\n=== Spatial Analysis ===")

	// Calculate route lengths in meters
	type RouteWithLength struct {
		Route
		LengthMeters float64 `gorm:"column:length_meters"`
	}

	var routesWithLength []RouteWithLength
	err = db.Table("routes").
		Select("*, ST_Length(ST_Transform(path, 3857)) as length_meters").
		Find(&routesWithLength).Error

	if err != nil {
		log.Printf("Length calculation failed: %v", err)
	} else {
		fmt.Println("Route lengths:")
		for _, route := range routesWithLength {
			fmt.Printf("- %s: %.2f meters\n", route.Name, route.LengthMeters)
		}
	}

	// Find routes that pass within distance of a point
	targetPoint := gogis.Point{Lng: -73.9665, Lat: 40.7812} // Central Park

	var nearbyRoutes []Route
	err = db.Where("ST_DWithin(path, ST_Point(?, ?), ?)",
		targetPoint.Lng, targetPoint.Lat, 0.01).Find(&nearbyRoutes).Error

	if err != nil {
		log.Printf("Nearby routes query failed: %v", err)
	} else {
		fmt.Printf("\nRoutes passing within 1km of Central Park: %d\n", len(nearbyRoutes))
		for _, route := range nearbyRoutes {
			fmt.Printf("- %s\n", route.Name)
		}
	}

	// Find intersecting routes (routes that cross each other)
	fmt.Println("\n=== Route Intersections ===")

	var route1 Route
	db.Where("name = ?", "Broadway Walk").First(&route1)

	var intersectingRoutes []Route
	err = db.Where("id != ? AND ST_Intersects(path, ?)",
		route1.ID, route1.Path.String()).Find(&intersectingRoutes).Error

	if err != nil {
		log.Printf("Intersection query failed: %v", err)
	} else {
		fmt.Printf("Routes that intersect with '%s': %d\n", route1.Name, len(intersectingRoutes))
		for _, route := range intersectingRoutes {
			fmt.Printf("- %s\n", route.Name)
		}
	}

	// Find closest point on route to a location
	fmt.Println("\n=== Closest Points ===")

	type RouteWithClosestPoint struct {
		Route
		ClosestPointWKT string  `gorm:"column:closest_point"`
		Distance        float64 `gorm:"column:distance"`
	}

	var routesWithClosestPoints []RouteWithClosestPoint
	err = db.Table("routes").
		Select(`*, 
			ST_AsText(ST_ClosestPoint(path, ST_Point(?, ?))) as closest_point,
			ST_Distance(path, ST_Point(?, ?)) as distance`,
			targetPoint.Lng, targetPoint.Lat,
			targetPoint.Lng, targetPoint.Lat).
		Order("distance").
		Find(&routesWithClosestPoints).Error

	if err != nil {
		log.Printf("Closest point query failed: %v", err)
	} else {
		fmt.Printf("Closest points on routes to Central Park:\n")
		for _, route := range routesWithClosestPoints {
			fmt.Printf("- %s: %s (distance: %.6f)\n",
				route.Name, route.ClosestPointWKT, route.Distance)
		}
	}

	// Advanced: Buffer analysis
	fmt.Println("\n=== Buffer Analysis ===")

	var routesInBuffer []Route
	bufferDistance := 0.005 // approximately 500m

	err = db.Where("ST_Within(path, ST_Buffer(ST_Point(?, ?), ?))",
		targetPoint.Lng, targetPoint.Lat, bufferDistance).Find(&routesInBuffer).Error

	if err != nil {
		log.Printf("Buffer analysis failed: %v", err)
	} else {
		fmt.Printf("Routes completely within 500m buffer of Central Park: %d\n", len(routesInBuffer))
		for _, route := range routesInBuffer {
			fmt.Printf("- %s\n", route.Name)
		}
	}

	// Cleanup
	fmt.Println("\n=== Cleanup ===")
	if err := db.Where("1 = 1").Delete(&Route{}).Error; err != nil {
		log.Printf("Failed to cleanup: %v", err)
	} else {
		fmt.Println("Cleaned up all test routes")
	}
}
