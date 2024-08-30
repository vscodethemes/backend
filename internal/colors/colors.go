package colors

import (
	"fmt"
	"math"
	"strconv"
)

func HexToLab(hex string) (float64, float64, float64, error) {
	sR, sG, sB, err := HexToRGB(hex)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to convert hex to rgb: %w", err)
	}

	x, y, z := RgbToXyz(sR, sG, sB)

	l, a, b := XyzToLab(x, y, z)

	return l, a, b, nil
}

func HexToRGB(hex string) (int, int, int, error) {
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

// http://www.easyrgb.com/en/math.php (XYZ → CIE-L*ab)
func XyzToLab(x, y, z float64) (float64, float64, float64) {
	// Using reference values D65-1931 (Daylight, sRGB, Adobe-RGB).
	x = x / 95.047
	y = y / 100
	z = z / 108.883

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
