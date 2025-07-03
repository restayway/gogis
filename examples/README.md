# GoGIS Examples

This directory contains comprehensive examples demonstrating how to use the GoGIS library with GORM for PostGIS geometry operations.

## Prerequisites

Before running these examples, ensure you have:

1. **PostgreSQL with PostGIS extension**:
   ```sql
   CREATE EXTENSION IF NOT EXISTS postgis;
   ```

2. **Go dependencies**:
   ```bash
   go mod init your-project
   go get github.com/restayway/gogis
   go get gorm.io/gorm
   go get gorm.io/driver/postgres
   ```

3. **Database connection**: Update the DSN in each example file:
   ```go
   dsn := "host=localhost user=postgres password=yourpassword dbname=testdb port=5432 sslmode=disable"
   ```

## Examples Overview

### 1. Basic Usage (`basic_usage/`)
**What it demonstrates:**
- Setting up GORM with PostGIS
- Creating and querying Point geometries
- Basic spatial operations (distance, nearest points)
- Spatial indexing recommendations

**Key concepts:**
- Point creation and storage
- `ST_DWithin` for proximity queries
- `ST_Distance` for distance calculations
- Distance-based ordering

**Run with:**
```bash
cd examples/basic_usage
go run main.go
```

### 2. LineString Example (`linestring_example/`)
**What it demonstrates:**
- Working with LineString geometries (paths, routes)
- Line-based spatial queries
- Length calculations and route analysis
- Intersection detection

**Key concepts:**
- Multi-point line creation
- `ST_Length` for distance measurements
- `ST_Intersects` for line crossings
- `ST_ClosestPoint` for proximity analysis
- Buffer operations with lines

**Run with:**
```bash
cd examples/linestring_example
go run main.go
```

### 3. Polygon Example (`polygon_example/`)
**What it demonstrates:**
- Creating polygons with holes (complex areas)
- Point-in-polygon containment
- Area calculations
- Polygon relationships (overlaps, touches, contains)

**Key concepts:**
- Polygon rings (outer boundary + holes)
- `ST_Contains` for containment testing
- `ST_Area` for area calculations
- `ST_Overlaps`, `ST_Touches` for relationships
- `ST_Buffer` for zone analysis
- `ST_Centroid` for center points

**Run with:**
```bash
cd examples/polygon_example
go run main.go
```

### 4. GeometryCollection Example (`geometry_collection_example/`)
**What it demonstrates:**
- Complex geometries with mixed types
- Heterogeneous spatial data modeling
- Component analysis and extraction
- Advanced spatial relationships

**Key concepts:**
- Mixed geometry types in single column
- `ST_CollectionExtract` for type filtering
- `ST_NumGeometries` for component counting
- `ST_Envelope` for bounding boxes
- Complex intersection analysis

**Run with:**
```bash
cd examples/geometry_collection_example
go run main.go
```

## Common Spatial Functions

The examples demonstrate these essential PostGIS functions:

### Distance and Proximity
- `ST_Distance(geom1, geom2)` - Calculate distance between geometries
- `ST_DWithin(geom1, geom2, distance)` - Check if within distance
- `ST_ClosestPoint(geom1, geom2)` - Find closest point on geometry

### Containment and Relationships
- `ST_Contains(geom1, geom2)` - Check if geom1 contains geom2
- `ST_Within(geom1, geom2)` - Check if geom1 is within geom2
- `ST_Intersects(geom1, geom2)` - Check if geometries intersect
- `ST_Overlaps(geom1, geom2)` - Check if geometries overlap
- `ST_Touches(geom1, geom2)` - Check if geometries touch

### Measurements
- `ST_Area(geom)` - Calculate area (use with ST_Transform for meters)
- `ST_Length(geom)` - Calculate length (use with ST_Transform for meters)
- `ST_Perimeter(geom)` - Calculate perimeter

### Transformations
- `ST_Buffer(geom, distance)` - Create buffer around geometry
- `ST_Transform(geom, srid)` - Transform to different coordinate system
- `ST_Centroid(geom)` - Calculate center point

### Utility Functions
- `ST_AsText(geom)` - Convert to Well-Known Text
- `ST_GeomFromText(wkt)` - Parse Well-Known Text
- `ST_Point(x, y)` - Create point from coordinates

## Performance Tips

### 1. Spatial Indexes
Always create spatial indexes for better performance:
```sql
CREATE INDEX idx_table_geom_column ON table_name USING GIST (geometry_column);
```

### 2. Coordinate System Considerations
- Use SRID 4326 (WGS84) for global lat/lng data
- Transform to projected coordinate systems (like 3857) for accurate distance/area calculations:
  ```sql
  ST_Area(ST_Transform(geom, 3857))  -- Area in square meters
  ```

### 3. Query Optimization
- Use `ST_DWithin` instead of `ST_Distance` with WHERE clauses
- Combine spatial indexes with other indexes for complex queries
- Consider using `&&` operator for bounding box intersection as a pre-filter

## Error Handling

All examples include proper error handling for:
- Database connection failures
- Migration errors
- Spatial query failures
- PostGIS extension availability

## Database Schema

The examples automatically create and clean up their tables, but you can also create them manually:

```sql
-- For basic usage
CREATE TABLE locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    description TEXT,
    point GEOMETRY(Point, 4326) NOT NULL,
    created_at BIGINT
);

-- For routes
CREATE TABLE routes (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    description TEXT,
    path GEOMETRY(LineString, 4326) NOT NULL,
    length DOUBLE PRECISION,
    created_at BIGINT
);

-- For regions
CREATE TABLE regions (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    type VARCHAR,
    description TEXT,
    area GEOMETRY(Polygon, 4326) NOT NULL,
    area_size DOUBLE PRECISION,
    created_at BIGINT
);

-- For complex places
CREATE TABLE places (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    type VARCHAR,
    description TEXT,
    geometries GEOMETRY(GeometryCollection, 4326) NOT NULL,
    created_at BIGINT
);

-- Add spatial indexes
CREATE INDEX idx_locations_point ON locations USING GIST (point);
CREATE INDEX idx_routes_path ON routes USING GIST (path);
CREATE INDEX idx_regions_area ON regions USING GIST (area);
CREATE INDEX idx_places_geometries ON places USING GIST (geometries);
```

## Next Steps

After running these examples, you can:
1. Modify the sample data to match your use case
2. Experiment with different spatial functions
3. Combine multiple geometry types in your applications
4. Implement real-world spatial applications like mapping services, location-based apps, or GIS systems