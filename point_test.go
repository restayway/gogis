package gogis_test

import (
	"database/sql/driver"
	"encoding/hex"
	"testing"

	"github.com/restayway/gogis"
)

func TestPointString(t *testing.T) {
	tests := []struct {
		name     string
		point    gogis.Point
		expected string
	}{
		{
			name:     "positive coordinates",
			point:    gogis.Point{Lng: 11.292383687705296, Lat: 43.76857094631136},
			expected: "SRID=4326;POINT(11.292383687705296 43.76857094631136)",
		},
		{
			name:     "negative coordinates",
			point:    gogis.Point{Lng: -122.4194, Lat: -37.7749},
			expected: "SRID=4326;POINT(-122.4194 -37.7749)",
		},
		{
			name:     "zero coordinates",
			point:    gogis.Point{Lng: 0, Lat: 0},
			expected: "SRID=4326;POINT(0 0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.point.String()
			if result != tt.expected {
				t.Errorf("Point.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPointValue(t *testing.T) {
	tests := []struct {
		name     string
		point    gogis.Point
		expected driver.Value
	}{
		{
			name:     "standard point",
			point:    gogis.Point{Lng: 11.292383687705296, Lat: 43.76857094631136},
			expected: "SRID=4326;POINT(11.292383687705296 43.76857094631136)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.point.Value()
			if err != nil {
				t.Errorf("Point.Value() unexpected error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("Point.Value() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPointScan(t *testing.T) {
	// Well-Known Binary (WKB) for POINT(11.292383687705296 43.76857094631136)
	// Format: byte order (1) + geometry type (8) + X (8) + Y (8) = 25 bytes
	// Little endian, Point type (1), X coordinate, Y coordinate
	wkbHex := "01010000000000000000289150b39526401288638860e24540"

	tests := []struct {
		name        string
		input       any
		expectedLng float64
		expectedLat float64
		expectError bool
	}{
		{
			name:        "valid WKB hex string",
			input:       wkbHex,
			expectedLng: 11.292383687705296,
			expectedLat: 43.76857094631136,
			expectError: false,
		},
		{
			name:        "valid WKB as []uint8",
			input:       []uint8(wkbHex),
			expectedLng: 11.292383687705296,
			expectedLat: 43.76857094631136,
			expectError: false,
		},
		{
			name:        "nil input",
			input:       nil,
			expectedLng: 0,
			expectedLat: 0,
			expectError: false,
		},
		{
			name:        "invalid hex string",
			input:       "invalid_hex",
			expectedLng: 0,
			expectedLat: 0,
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expectedLng: 0,
			expectedLat: 0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &gogis.Point{}
			err := p.Scan(tt.input)

			if tt.expectError && err == nil {
				t.Errorf("Point.Scan() expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("Point.Scan() unexpected error = %v", err)
				return
			}

			if !tt.expectError {
				// Allow for small floating point differences
				const epsilon = 1e-9
				if absFloat(p.Lng-tt.expectedLng) > epsilon {
					t.Errorf("Point.Scan() Lng = %v, want %v", p.Lng, tt.expectedLng)
				}
				if absFloat(p.Lat-tt.expectedLat) > epsilon {
					t.Errorf("Point.Scan() Lat = %v, want %v", p.Lat, tt.expectedLat)
				}
			}
		})
	}
}

func TestPointScanBigEndian(t *testing.T) {
	// Big endian WKB for POINT(11.292383687705296 43.76857094631136)
	wkbHex := "000000000000000001402695b3509128004045e26088638812"

	p := &gogis.Point{}
	err := p.Scan(wkbHex)

	if err != nil {
		t.Errorf("Point.Scan() with big endian unexpected error = %v", err)
		return
	}

	const epsilon = 1e-9
	expectedLng := 11.292383687705296
	expectedLat := 43.76857094631136

	if absFloat(p.Lng-expectedLng) > epsilon {
		t.Errorf("Point.Scan() big endian Lng = %v, want %v", p.Lng, expectedLng)
	}
	if absFloat(p.Lat-expectedLat) > epsilon {
		t.Errorf("Point.Scan() big endian Lat = %v, want %v", p.Lat, expectedLat)
	}
}

func TestPointScanInvalidByteOrder(t *testing.T) {
	// Invalid byte order (2)
	invalidWKB, _ := hex.DecodeString("02")
	// Pad with zeros to avoid read errors
	invalidWKB = append(invalidWKB, make([]byte, 20)...)
	hexStr := hex.EncodeToString(invalidWKB)

	p := &gogis.Point{}
	err := p.Scan(hexStr)

	if err == nil {
		t.Errorf("Point.Scan() expected error for invalid byte order but got none")
	}
}

// Helper function for float comparison
func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
