package gogis_test

import (
	"bytes"
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/restayway/gogis"
)

func TestGeometryCollectionString(t *testing.T) {
	tests := []struct {
		name       string
		collection gogis.GeometryCollection
		expected   string
	}{
		{
			name: "point and linestring",
			collection: gogis.GeometryCollection{
				Geometries: []gogis.Geometry{
					&gogis.Point{Lng: 2, Lat: 0},
					&gogis.LineString{
						Points: []gogis.Point{
							{Lng: 0, Lat: 0},
							{Lng: 1, Lat: 1},
						},
					},
				},
			},
			expected: "SRID=4326;GEOMETRYCOLLECTION(POINT(2 0),LINESTRING(0 0,1 1))",
		},
		{
			name: "point and polygon",
			collection: gogis.GeometryCollection{
				Geometries: []gogis.Geometry{
					&gogis.Point{Lng: 2, Lat: 0},
					&gogis.Polygon{
						Rings: [][]gogis.Point{
							{
								{Lng: 0, Lat: 0},
								{Lng: 1, Lat: 0},
								{Lng: 1, Lat: 1},
								{Lng: 0, Lat: 1},
								{Lng: 0, Lat: 0},
							},
						},
					},
				},
			},
			expected: "SRID=4326;GEOMETRYCOLLECTION(POINT(2 0),POLYGON((0 0,1 0,1 1,0 1,0 0)))",
		},
		{
			name: "multiple points",
			collection: gogis.GeometryCollection{
				Geometries: []gogis.Geometry{
					&gogis.Point{Lng: 1, Lat: 1},
					&gogis.Point{Lng: 2, Lat: 2},
					&gogis.Point{Lng: 3, Lat: 3},
				},
			},
			expected: "SRID=4326;GEOMETRYCOLLECTION(POINT(1 1),POINT(2 2),POINT(3 3))",
		},
		{
			name:       "empty collection",
			collection: gogis.GeometryCollection{Geometries: []gogis.Geometry{}},
			expected:   "SRID=4326;GEOMETRYCOLLECTION EMPTY",
		},
		{
			name:       "nil geometries",
			collection: gogis.GeometryCollection{Geometries: nil},
			expected:   "SRID=4326;GEOMETRYCOLLECTION EMPTY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.collection.String()
			if result != tt.expected {
				t.Errorf("GeometryCollection.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGeometryCollectionValue(t *testing.T) {
	gc := gogis.GeometryCollection{
		Geometries: []gogis.Geometry{
			&gogis.Point{Lng: 2, Lat: 0},
			&gogis.Point{Lng: 3, Lat: 1},
		},
	}

	expected := "SRID=4326;GEOMETRYCOLLECTION(POINT(2 0),POINT(3 1))"

	result, err := gc.Value()
	if err != nil {
		t.Errorf("GeometryCollection.Value() unexpected error = %v", err)
		return
	}

	if result != driver.Value(expected) {
		t.Errorf("GeometryCollection.Value() = %v, want %v", result, expected)
	}
}

func TestGeometryCollectionScan(t *testing.T) {
	// Helper function to create WKB for GeometryCollection
	createGeometryCollectionWKB := func(byteOrder binary.ByteOrder, geometries []gogis.Geometry) string {
		var buf bytes.Buffer

		// Byte order
		if byteOrder == binary.LittleEndian {
			binary.Write(&buf, binary.LittleEndian, uint8(1))
		} else {
			binary.Write(&buf, binary.LittleEndian, uint8(0))
		}

		// Geometry type (7 for GeometryCollection)
		binary.Write(&buf, byteOrder, uint64(7))

		// Number of geometries
		binary.Write(&buf, byteOrder, uint32(len(geometries)))

		// Geometries
		for _, geom := range geometries {
			switch g := geom.(type) {
			case *gogis.Point:
				// Point WKB
				if byteOrder == binary.LittleEndian {
					binary.Write(&buf, binary.LittleEndian, uint8(1)) // Little endian for geometry
				} else {
					binary.Write(&buf, binary.LittleEndian, uint8(0)) // Big endian for geometry
				}
				binary.Write(&buf, byteOrder, uint64(1)) // Point type
				binary.Write(&buf, byteOrder, g.Lng)
				binary.Write(&buf, byteOrder, g.Lat)

			case *gogis.LineString:
				// LineString WKB
				if byteOrder == binary.LittleEndian {
					binary.Write(&buf, binary.LittleEndian, uint8(1)) // Little endian for geometry
				} else {
					binary.Write(&buf, binary.LittleEndian, uint8(0)) // Big endian for geometry
				}
				binary.Write(&buf, byteOrder, uint64(2)) // LineString type
				binary.Write(&buf, byteOrder, uint32(len(g.Points)))
				for _, p := range g.Points {
					binary.Write(&buf, byteOrder, p.Lng)
					binary.Write(&buf, byteOrder, p.Lat)
				}

			case *gogis.Polygon:
				// Polygon WKB
				if byteOrder == binary.LittleEndian {
					binary.Write(&buf, binary.LittleEndian, uint8(1)) // Little endian for geometry
				} else {
					binary.Write(&buf, binary.LittleEndian, uint8(0)) // Big endian for geometry
				}
				binary.Write(&buf, byteOrder, uint64(3)) // Polygon type
				binary.Write(&buf, byteOrder, uint32(len(g.Rings)))
				for _, ring := range g.Rings {
					binary.Write(&buf, byteOrder, uint32(len(ring)))
					for _, p := range ring {
						binary.Write(&buf, byteOrder, p.Lng)
						binary.Write(&buf, byteOrder, p.Lat)
					}
				}
			}
		}

		return hex.EncodeToString(buf.Bytes())
	}

	tests := []struct {
		name          string
		input         any
		expectedCount int
		expectedTypes []string
		expectError   bool
	}{
		{
			name: "valid WKB hex string - point and linestring",
			input: createGeometryCollectionWKB(binary.LittleEndian, []gogis.Geometry{
				&gogis.Point{Lng: 2, Lat: 0},
				&gogis.LineString{
					Points: []gogis.Point{
						{Lng: 0, Lat: 0},
						{Lng: 1, Lat: 1},
					},
				},
			}),
			expectedCount: 2,
			expectedTypes: []string{"Point", "LineString"},
			expectError:   false,
		},
		{
			name: "valid WKB as []uint8 - multiple points",
			input: []uint8(createGeometryCollectionWKB(binary.LittleEndian, []gogis.Geometry{
				&gogis.Point{Lng: 1, Lat: 1},
				&gogis.Point{Lng: 2, Lat: 2},
				&gogis.Point{Lng: 3, Lat: 3},
			})),
			expectedCount: 3,
			expectedTypes: []string{"Point", "Point", "Point"},
			expectError:   false,
		},
		{
			name: "big endian WKB - point and polygon",
			input: createGeometryCollectionWKB(binary.BigEndian, []gogis.Geometry{
				&gogis.Point{Lng: 5, Lat: 5},
				&gogis.Polygon{
					Rings: [][]gogis.Point{
						{
							{Lng: 0, Lat: 0},
							{Lng: 1, Lat: 0},
							{Lng: 1, Lat: 1},
							{Lng: 0, Lat: 1},
							{Lng: 0, Lat: 0},
						},
					},
				},
			}),
			expectedCount: 2,
			expectedTypes: []string{"Point", "Polygon"},
			expectError:   false,
		},
		{
			name:          "nil input",
			input:         nil,
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "invalid hex string",
			input:         "invalid_hex",
			expectedCount: 0,
			expectError:   true,
		},
		{
			name:          "empty string",
			input:         "",
			expectedCount: 0,
			expectError:   true,
		},
		{
			name:          "wrong geometry type",
			input:         "01010000000000000000000000000000000000000000000000",
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gc := &gogis.GeometryCollection{}
			err := gc.Scan(tt.input)

			if tt.expectError && err == nil {
				t.Errorf("GeometryCollection.Scan() expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("GeometryCollection.Scan() unexpected error = %v", err)
				return
			}

			if !tt.expectError {
				if len(gc.Geometries) != tt.expectedCount {
					t.Errorf("GeometryCollection.Scan() got %d geometries, want %d", len(gc.Geometries), tt.expectedCount)
				}

				for i, expectedType := range tt.expectedTypes {
					if i < len(gc.Geometries) {
						var actualType string
						switch gc.Geometries[i].(type) {
						case *gogis.Point:
							actualType = "Point"
						case *gogis.LineString:
							actualType = "LineString"
						case *gogis.Polygon:
							actualType = "Polygon"
						default:
							actualType = "Unknown"
						}

						if actualType != expectedType {
							t.Errorf("Geometry[%d] type = %s, want %s", i, actualType, expectedType)
						}
					}
				}

				// Test specific geometry values for the first test case
				if tt.name == "valid WKB hex string - point and linestring" && len(gc.Geometries) >= 2 {
					if point, ok := gc.Geometries[0].(*gogis.Point); ok {
						const epsilon = 1e-9
						if absFloat(point.Lng-2) > epsilon || absFloat(point.Lat-0) > epsilon {
							t.Errorf("First geometry Point = (%v, %v), want (2, 0)", point.Lng, point.Lat)
						}
					}

					if lineString, ok := gc.Geometries[1].(*gogis.LineString); ok {
						if len(lineString.Points) != 2 {
							t.Errorf("Second geometry LineString has %d points, want 2", len(lineString.Points))
						}
					}
				}
			}
		})
	}
}

func TestGeometryCollectionScanInvalidByteOrder(t *testing.T) {
	// Invalid byte order (2)
	invalidWKB, _ := hex.DecodeString("02")
	// Pad with zeros to avoid read errors
	invalidWKB = append(invalidWKB, make([]byte, 20)...)
	hexStr := hex.EncodeToString(invalidWKB)

	gc := &gogis.GeometryCollection{}
	err := gc.Scan(hexStr)

	if err == nil {
		t.Errorf("GeometryCollection.Scan() expected error for invalid byte order but got none")
	}
}

func TestGeometryCollectionScanInvalidType(t *testing.T) {
	// Test scanning with an unsupported type
	gc := &gogis.GeometryCollection{}
	err := gc.Scan(12345) // int is not supported

	if err == nil {
		t.Errorf("GeometryCollection.Scan() expected error for unsupported type but got none")
	}
}

func TestGeometryCollectionScanUnsupportedGeometry(t *testing.T) {
	// Create a GeometryCollection WKB with an unsupported geometry type (e.g., type 4)
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint8(1))  // byte order
	binary.Write(&buf, binary.LittleEndian, uint64(7)) // GeometryCollection type
	binary.Write(&buf, binary.LittleEndian, uint32(1)) // number of geometries
	binary.Write(&buf, binary.LittleEndian, uint8(1))  // geometry byte order
	binary.Write(&buf, binary.LittleEndian, uint64(4)) // unsupported geometry type
	// Add some dummy data
	binary.Write(&buf, binary.LittleEndian, uint64(0))
	binary.Write(&buf, binary.LittleEndian, uint64(0))

	hexStr := hex.EncodeToString(buf.Bytes())

	gc := &gogis.GeometryCollection{}
	err := gc.Scan(hexStr)

	if err == nil {
		t.Errorf("GeometryCollection.Scan() expected error for unsupported geometry type but got none")
	}
}
