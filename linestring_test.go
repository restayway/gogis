package gogis_test

import (
	"bytes"
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/restayway/gogis"
)

func TestLineStringString(t *testing.T) {
	tests := []struct {
		name       string
		lineString gogis.LineString
		expected   string
	}{
		{
			name: "simple line",
			lineString: gogis.LineString{
				Points: []gogis.Point{
					{Lng: 0, Lat: 0},
					{Lng: 1, Lat: 1},
					{Lng: 2, Lat: 1},
					{Lng: 2, Lat: 2},
				},
			},
			expected: "SRID=4326;LINESTRING(0 0,1 1,2 1,2 2)",
		},
		{
			name: "two point line",
			lineString: gogis.LineString{
				Points: []gogis.Point{
					{Lng: -122.4194, Lat: 37.7749},
					{Lng: -118.2437, Lat: 34.0522},
				},
			},
			expected: "SRID=4326;LINESTRING(-122.4194 37.7749,-118.2437 34.0522)",
		},
		{
			name:       "empty linestring",
			lineString: gogis.LineString{Points: []gogis.Point{}},
			expected:   "SRID=4326;LINESTRING EMPTY",
		},
		{
			name:       "nil points",
			lineString: gogis.LineString{Points: nil},
			expected:   "SRID=4326;LINESTRING EMPTY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.lineString.String()
			if result != tt.expected {
				t.Errorf("LineString.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLineStringValue(t *testing.T) {
	ls := gogis.LineString{
		Points: []gogis.Point{
			{Lng: 0, Lat: 0},
			{Lng: 1, Lat: 1},
		},
	}

	expected := "SRID=4326;LINESTRING(0 0,1 1)"

	result, err := ls.Value()
	if err != nil {
		t.Errorf("LineString.Value() unexpected error = %v", err)
		return
	}

	if result != driver.Value(expected) {
		t.Errorf("LineString.Value() = %v, want %v", result, expected)
	}
}

func TestLineStringScan(t *testing.T) {
	// Helper function to create WKB for LineString
	createLineStringWKB := func(byteOrder binary.ByteOrder, points []gogis.Point) string {
		var buf bytes.Buffer

		// Byte order
		if byteOrder == binary.LittleEndian {
			binary.Write(&buf, binary.LittleEndian, uint8(1))
		} else {
			binary.Write(&buf, binary.LittleEndian, uint8(0))
		}

		// Geometry type (2 for LineString)
		binary.Write(&buf, byteOrder, uint64(2))

		// Number of points
		binary.Write(&buf, byteOrder, uint32(len(points)))

		// Points
		for _, p := range points {
			binary.Write(&buf, byteOrder, p.Lng)
			binary.Write(&buf, byteOrder, p.Lat)
		}

		return hex.EncodeToString(buf.Bytes())
	}

	tests := []struct {
		name        string
		input       any
		expectedLen int
		expectedPts []gogis.Point
		expectError bool
	}{
		{
			name: "valid WKB hex string",
			input: createLineStringWKB(binary.LittleEndian, []gogis.Point{
				{Lng: 0, Lat: 0},
				{Lng: 1, Lat: 1},
				{Lng: 2, Lat: 1},
				{Lng: 2, Lat: 2},
			}),
			expectedLen: 4,
			expectedPts: []gogis.Point{
				{Lng: 0, Lat: 0},
				{Lng: 1, Lat: 1},
				{Lng: 2, Lat: 1},
				{Lng: 2, Lat: 2},
			},
			expectError: false,
		},
		{
			name: "valid WKB as []uint8",
			input: []uint8(createLineStringWKB(binary.LittleEndian, []gogis.Point{
				{Lng: -122.4194, Lat: 37.7749},
				{Lng: -118.2437, Lat: 34.0522},
			})),
			expectedLen: 2,
			expectedPts: []gogis.Point{
				{Lng: -122.4194, Lat: 37.7749},
				{Lng: -118.2437, Lat: 34.0522},
			},
			expectError: false,
		},
		{
			name: "big endian WKB",
			input: createLineStringWKB(binary.BigEndian, []gogis.Point{
				{Lng: 10.5, Lat: 20.5},
				{Lng: 30.5, Lat: 40.5},
				{Lng: 50.5, Lat: 60.5},
			}),
			expectedLen: 3,
			expectedPts: []gogis.Point{
				{Lng: 10.5, Lat: 20.5},
				{Lng: 30.5, Lat: 40.5},
				{Lng: 50.5, Lat: 60.5},
			},
			expectError: false,
		},
		{
			name:        "nil input",
			input:       nil,
			expectedLen: 0,
			expectError: false,
		},
		{
			name:        "invalid hex string",
			input:       "invalid_hex",
			expectedLen: 0,
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expectedLen: 0,
			expectError: true,
		},
		{
			name:        "wrong geometry type",
			input:       "01010000000000000000000000000000000000000000000000",
			expectedLen: 0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := &gogis.LineString{}
			err := ls.Scan(tt.input)

			if tt.expectError && err == nil {
				t.Errorf("LineString.Scan() expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("LineString.Scan() unexpected error = %v", err)
				return
			}

			if !tt.expectError {
				if len(ls.Points) != tt.expectedLen {
					t.Errorf("LineString.Scan() got %d points, want %d", len(ls.Points), tt.expectedLen)
				}

				const epsilon = 1e-9
				for i, expectedPt := range tt.expectedPts {
					if i < len(ls.Points) {
						if absFloat(ls.Points[i].Lng-expectedPt.Lng) > epsilon {
							t.Errorf("Point[%d].Lng = %v, want %v", i, ls.Points[i].Lng, expectedPt.Lng)
						}
						if absFloat(ls.Points[i].Lat-expectedPt.Lat) > epsilon {
							t.Errorf("Point[%d].Lat = %v, want %v", i, ls.Points[i].Lat, expectedPt.Lat)
						}
					}
				}
			}
		})
	}
}

func TestLineStringScanInvalidByteOrder(t *testing.T) {
	// Invalid byte order (2)
	invalidWKB, _ := hex.DecodeString("02")
	// Pad with zeros to avoid read errors
	invalidWKB = append(invalidWKB, make([]byte, 20)...)
	hexStr := hex.EncodeToString(invalidWKB)

	ls := &gogis.LineString{}
	err := ls.Scan(hexStr)

	if err == nil {
		t.Errorf("LineString.Scan() expected error for invalid byte order but got none")
	}
}

func TestLineStringScanInvalidType(t *testing.T) {
	// Test scanning with an unsupported type
	ls := &gogis.LineString{}
	err := ls.Scan(12345) // int is not supported

	if err == nil {
		t.Errorf("LineString.Scan() expected error for unsupported type but got none")
	}
}
