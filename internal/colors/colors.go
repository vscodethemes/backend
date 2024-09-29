package colors

import (
	"fmt"
	"math"
	"strconv"
)

func HexToLabString(hex string) (string, error) {
	l, a, b, err := HexToLab(hex)
	if err != nil {
		return "", fmt.Errorf("failed to convert hex to lab: %w", err)
	}

	return fmt.Sprintf("(%.3f, %.3f, %.3f)", l, a, b), nil
}

func LabStringToHex(lab string) (string, error) {
	// Split into L, A, and B.
	var l, a, b float64
	_, err := fmt.Sscanf(lab, "(%f,%f,%f)", &l, &a, &b)

	if err != nil {
		return "", fmt.Errorf("failed to parse LAB color %s: %w", lab, err)
	}

	// Convert to RGB.
	return LabToHex(l, a, b), nil
}

func HexToLab(hex string) (float64, float64, float64, error) {
	sR, sG, sB, err := HexToRgb(hex)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to convert hex to rgb: %w", err)
	}

	x, y, z := RgbToXyz(sR, sG, sB)

	l, a, b := XyzToLab(x, y, z)

	return l, a, b, nil
}

func LabToHex(l, a, b float64) string {
	x, y, z := LabToXyz(l, a, b)

	sR, sG, sB := XyzToRgb(x, y, z)

	hex := RgbToHex(sR, sG, sB)

	return hex

}

func HexToRgb(hex string) (int, int, int, error) {
	r, err := strconv.ParseInt(hex[1:3], 16, 0)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse red: %w", err)
	}

	g, err := strconv.ParseInt(hex[3:5], 16, 0)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse green: %w", err)
	}

	b, err := strconv.ParseInt(hex[5:7], 16, 0)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse blue: %w", err)
	}

	return int(r), int(g), int(b), nil
}

func RgbToHex(r, g, b int) string {
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// http://www.easyrgb.com/en/math.php (Standard-RGB → XYZ)
func RgbToXyz(sR, sG, sB int) (float64, float64, float64) {
	r := float64(sR) / 255
	g := float64(sG) / 255
	b := float64(sB) / 255

	if r > 0.04045 {
		r = math.Pow((r+0.055)/1.055, 2.4)
	} else {
		r = r / 12.92
	}

	if g > 0.04045 {
		g = math.Pow((g+0.055)/1.055, 2.4)
	} else {
		g = g / 12.92
	}

	if b > 0.04045 {
		b = math.Pow((b+0.055)/1.055, 2.4)
	} else {
		b = b / 12.92
	}

	r = r * 100
	g = g * 100
	b = b * 100

	x := r*0.4124 + g*0.3576 + b*0.1805
	y := r*0.2126 + g*0.7152 + b*0.0722
	z := r*0.0193 + g*0.1192 + b*0.9505

	return x, y, z
}

// http://www.easyrgb.com/en/math.php (XYZ → Standard-RGB)
func XyzToRgb(x, y, z float64) (int, int, int) {
	x = x / 100
	y = y / 100
	z = z / 100

	r := x*3.2406 + y*-1.5372 + z*-0.4986
	g := x*-0.9689 + y*1.8758 + z*0.0415
	b := x*0.0557 + y*-0.2040 + z*1.0570

	if r > 0.0031308 {
		r = 1.055*math.Pow(r, 1/2.4) - 0.055
	} else {
		r = 12.92 * r
	}

	if g > 0.0031308 {
		g = 1.055*math.Pow(g, 1/2.4) - 0.055
	} else {
		g = 12.92 * g
	}

	if b > 0.0031308 {
		b = 1.055*math.Pow(b, 1/2.4) - 0.055
	} else {
		b = 12.92 * b
	}

	sR := r * 255
	sG := g * 255
	sB := b * 255

	return int(sR), int(sG), int(sB)
}

const (
	// Reference values for D65-1931 (Daylight, sRGB, Adobe-RGB).
	referenceX = 95.047
	referenceY = 100
	referenceZ = 108.883
)

// http://www.easyrgb.com/en/math.php (XYZ → CIE-L*ab)
func XyzToLab(x, y, z float64) (float64, float64, float64) {
	x = x / referenceX
	y = y / referenceY
	z = z / referenceZ

	if x > 0.008856 {
		x = math.Pow(x, float64(1)/float64(3))
	} else {
		x = (7.787 * x) + (float64(16) / float64(116))
	}

	if y > 0.008856 {
		y = math.Pow(y, float64(1)/float64(3))
	} else {
		y = (7.787 * y) + (float64(16) / float64(116))
	}

	if z > 0.008856 {
		z = math.Pow(z, float64(1)/float64(3))
	} else {
		z = (7.787 * z) + (float64(16) / float64(116))
	}

	l := 116*y - 16
	a := 500 * (x - y)
	b := 200 * (y - z)

	return l, a, b
}

// http://www.easyrgb.com/en/math.php (CIE-L*ab → XYZ)
func LabToXyz(l, a, b float64) (float64, float64, float64) {
	y := (l + 16) / 116
	x := a/500 + y
	z := y - b/200

	if math.Pow(y, 3) > 0.008856 {
		y = math.Pow(y, 3)
	} else {
		y = (y - 16.0/116.0) / 7.787
	}

	if math.Pow(x, 3) > 0.008856 {
		x = math.Pow(x, 3)
	} else {
		x = (x - 16.0/116.0) / 7.787
	}

	if math.Pow(z, 3) > 0.008856 {
		z = math.Pow(z, 3)
	} else {
		z = (z - 16.0/116.0) / 7.787
	}

	x = x * referenceX
	y = y * referenceY
	z = z * referenceZ

	return x, y, z
}
