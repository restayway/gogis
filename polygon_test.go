package gogis_test

import (
	"bytes"
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/restayway/gogis"
)

func TestPolygonString(t *testing.T) {
	tests := []struct {
		name     string
		polygon  gogis.Polygon
		expected string
	}{
		{
			name: "simple polygon",
			polygon: gogis.Polygon{
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
			expected: "SRID=4326;POLYGON((0 0,1 0,1 1,0 1,0 0))",
		},
		{
			name: "polygon with hole",
			polygon: gogis.Polygon{
				Rings: [][]gogis.Point{
					{
						{Lng: 0, Lat: 0},
						{Lng: 10, Lat: 0},
						{Lng: 10, Lat: 10},
						{Lng: 0, Lat: 10},
						{Lng: 0, Lat: 0},
					},
					{
						{Lng: 1, Lat: 1},
						{Lng: 1, Lat: 2},
						{Lng: 2, Lat: 2},
						{Lng: 2, Lat: 1},
						{Lng: 1, Lat: 1},
					},
				},
			},
			expected: "SRID=4326;POLYGON((0 0,10 0,10 10,0 10,0 0),(1 1,1 2,2 2,2 1,1 1))",
		},
		{
			name:     "empty polygon",
			polygon:  gogis.Polygon{Rings: [][]gogis.Point{}},
			expected: "SRID=4326;POLYGON EMPTY",
		},
		{
			name:     "nil rings",
			polygon:  gogis.Polygon{Rings: nil},
			expected: "SRID=4326;POLYGON EMPTY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.polygon.String()
			if result != tt.expected {
				t.Errorf("Polygon.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPolygonValue(t *testing.T) {
	poly := gogis.Polygon{
		Rings: [][]gogis.Point{
			{
				{Lng: 0, Lat: 0},
				{Lng: 1, Lat: 0},
				{Lng: 1, Lat: 1},
				{Lng: 0, Lat: 1},
				{Lng: 0, Lat: 0},
			},
		},
	}

	expected := "SRID=4326;POLYGON((0 0,1 0,1 1,0 1,0 0))"

	result, err := poly.Value()
	if err != nil {
		t.Errorf("Polygon.Value() unexpected error = %v", err)
		return
	}

	if result != driver.Value(expected) {
		t.Errorf("Polygon.Value() = %v, want %v", result, expected)
	}
}

func TestPolygonScan(t *testing.T) {
	// Helper function to create WKB for Polygon
	createPolygonWKB := func(byteOrder binary.ByteOrder, rings [][]gogis.Point) string {
		var buf bytes.Buffer

		// Byte order
		if byteOrder == binary.LittleEndian {
			binary.Write(&buf, binary.LittleEndian, uint8(1))
		} else {
			binary.Write(&buf, binary.LittleEndian, uint8(0))
		}

		// Geometry type (3 for Polygon)
		binary.Write(&buf, byteOrder, uint64(3))

		// Number of rings
		binary.Write(&buf, byteOrder, uint32(len(rings)))

		// Rings
		for _, ring := range rings {
			// Number of points in ring
			binary.Write(&buf, byteOrder, uint32(len(ring)))

			// Points
			for _, p := range ring {
				binary.Write(&buf, byteOrder, p.Lng)
				binary.Write(&buf, byteOrder, p.Lat)
			}
		}

		return hex.EncodeToString(buf.Bytes())
	}

	tests := []struct {
		name          string
		input         any
		expectedRings int
		expectedData  [][]gogis.Point
		expectError   bool
	}{
		{
			name: "valid WKB hex string - simple polygon",
			input: createPolygonWKB(binary.LittleEndian, [][]gogis.Point{
				{
					{Lng: 0, Lat: 0},
					{Lng: 1, Lat: 0},
					{Lng: 1, Lat: 1},
					{Lng: 0, Lat: 1},
					{Lng: 0, Lat: 0},
				},
			}),
			expectedRings: 1,
			expectedData: [][]gogis.Point{
				{
					{Lng: 0, Lat: 0},
					{Lng: 1, Lat: 0},
					{Lng: 1, Lat: 1},
					{Lng: 0, Lat: 1},
					{Lng: 0, Lat: 0},
				},
			},
			expectError: false,
		},
		{
			name: "valid WKB as []uint8 - polygon with hole",
			input: []uint8(createPolygonWKB(binary.LittleEndian, [][]gogis.Point{
				{
					{Lng: 0, Lat: 0},
					{Lng: 10, Lat: 0},
					{Lng: 10, Lat: 10},
					{Lng: 0, Lat: 10},
					{Lng: 0, Lat: 0},
				},
				{
					{Lng: 2, Lat: 2},
					{Lng: 2, Lat: 8},
					{Lng: 8, Lat: 8},
					{Lng: 8, Lat: 2},
					{Lng: 2, Lat: 2},
				},
			})),
			expectedRings: 2,
			expectedData: [][]gogis.Point{
				{
					{Lng: 0, Lat: 0},
					{Lng: 10, Lat: 0},
					{Lng: 10, Lat: 10},
					{Lng: 0, Lat: 10},
					{Lng: 0, Lat: 0},
				},
				{
					{Lng: 2, Lat: 2},
					{Lng: 2, Lat: 8},
					{Lng: 8, Lat: 8},
					{Lng: 8, Lat: 2},
					{Lng: 2, Lat: 2},
				},
			},
			expectError: false,
		},
		{
			name: "big endian WKB",
			input: createPolygonWKB(binary.BigEndian, [][]gogis.Point{
				{
					{Lng: -122.4, Lat: 37.8},
					{Lng: -122.4, Lat: 37.7},
					{Lng: -122.3, Lat: 37.7},
					{Lng: -122.3, Lat: 37.8},
					{Lng: -122.4, Lat: 37.8},
				},
			}),
			expectedRings: 1,
			expectedData: [][]gogis.Point{
				{
					{Lng: -122.4, Lat: 37.8},
					{Lng: -122.4, Lat: 37.7},
					{Lng: -122.3, Lat: 37.7},
					{Lng: -122.3, Lat: 37.8},
					{Lng: -122.4, Lat: 37.8},
				},
			},
			expectError: false,
		},
		{
			name:          "nil input",
			input:         nil,
			expectedRings: 0,
			expectError:   false,
		},
		{
			name:          "invalid hex string",
			input:         "invalid_hex",
			expectedRings: 0,
			expectError:   true,
		},
		{
			name:          "empty string",
			input:         "",
			expectedRings: 0,
			expectError:   true,
		},
		{
			name:          "wrong geometry type",
			input:         "01010000000000000000000000000000000000000000000000",
			expectedRings: 0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &gogis.Polygon{}
			err := p.Scan(tt.input)

			if tt.expectError && err == nil {
				t.Errorf("Polygon.Scan() expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("Polygon.Scan() unexpected error = %v", err)
				return
			}

			if !tt.expectError {
				if len(p.Rings) != tt.expectedRings {
					t.Errorf("Polygon.Scan() got %d rings, want %d", len(p.Rings), tt.expectedRings)
				}

				const epsilon = 1e-9
				for i, expectedRing := range tt.expectedData {
					if i < len(p.Rings) {
						if len(p.Rings[i]) != len(expectedRing) {
							t.Errorf("Ring[%d] has %d points, want %d", i, len(p.Rings[i]), len(expectedRing))
							continue
						}

						for j, expectedPt := range expectedRing {
							if j < len(p.Rings[i]) {
								if absFloat(p.Rings[i][j].Lng-expectedPt.Lng) > epsilon {
									t.Errorf("Ring[%d] Point[%d].Lng = %v, want %v", i, j, p.Rings[i][j].Lng, expectedPt.Lng)
								}
								if absFloat(p.Rings[i][j].Lat-expectedPt.Lat) > epsilon {
									t.Errorf("Ring[%d] Point[%d].Lat = %v, want %v", i, j, p.Rings[i][j].Lat, expectedPt.Lat)
								}
							}
						}
					}
				}
			}
		})
	}
}

func TestPolygonScanInvalidByteOrder(t *testing.T) {
	// Invalid byte order (2)
	invalidWKB, _ := hex.DecodeString("02")
	// Pad with zeros to avoid read errors
	invalidWKB = append(invalidWKB, make([]byte, 20)...)
	hexStr := hex.EncodeToString(invalidWKB)

	p := &gogis.Polygon{}
	err := p.Scan(hexStr)

	if err == nil {
		t.Errorf("Polygon.Scan() expected error for invalid byte order but got none")
	}
}

func TestPolygonScanInvalidType(t *testing.T) {
	// Test scanning with an unsupported type
	p := &gogis.Polygon{}
	err := p.Scan(12345) // int is not supported

	if err == nil {
		t.Errorf("Polygon.Scan() expected error for unsupported type but got none")
	}
}
