package visual

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

/* Pine Script v5 color constants matching TradingView hex values */
const (
	Aqua    = "#00BCD4"
	Black   = "#363A45"
	Blue    = "#2962FF"
	Fuchsia = "#E040FB"
	Gray    = "#787B86"
	Green   = "#4CAF50"
	Lime    = "#00E676"
	Maroon  = "#880E4F"
	Navy    = "#311B92"
	Olive   = "#808000"
	Orange  = "#FF9800"
	Purple  = "#9C27B0"
	Red     = "#FF5252"
	Silver  = "#B2B5BE"
	Teal    = "#00897B"
	White   = "#FFFFFF"
	Yellow  = "#FFEB3B"
)

/* parseHexColor extracts RGBA components from #RRGGBB or #RRGGBBAA hex strings.
 * Returns (r, g, b, a) where a defaults to 255 (opaque) for 6-digit hex.
 */
func parseHexColor(hex string) (r, g, b, a uint8) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) < 6 {
		return 0, 0, 0, 255
	}
	rv, _ := strconv.ParseUint(hex[0:2], 16, 8)
	gv, _ := strconv.ParseUint(hex[2:4], 16, 8)
	bv, _ := strconv.ParseUint(hex[4:6], 16, 8)
	av := uint64(255)
	if len(hex) >= 8 {
		av, _ = strconv.ParseUint(hex[6:8], 16, 8)
	}
	return uint8(rv), uint8(gv), uint8(bv), uint8(av)
}

/* transpToAlpha converts Pine transparency (0=opaque, 100=invisible) to hex alpha (FF=opaque, 00=invisible) */
func transpToAlpha(transp float64) uint8 {
	clamped := math.Max(0, math.Min(100, transp))
	return uint8(math.Round(255 * (100 - clamped) / 100))
}

/* alphaToTransp converts hex alpha (FF=opaque, 00=invisible) to Pine transparency (0=opaque, 100=invisible) */
func alphaToTransp(alpha uint8) float64 {
	return math.Round(100 - float64(alpha)/255*100)
}

/* PineColorNew applies transparency to a base color. Pine: color.new(color, transp) */
func PineColorNew(baseHex string, transp float64) string {
	r, g, b, _ := parseHexColor(baseHex)
	alpha := transpToAlpha(transp)
	return fmt.Sprintf("#%02X%02X%02X%02X", r, g, b, alpha)
}

/* PineColorRGB creates a color from RGBA components. Pine: color.rgb(red, green, blue, transp) */
func PineColorRGB(red, green, blue, transp float64) string {
	r := uint8(math.Max(0, math.Min(255, math.Round(red))))
	g := uint8(math.Max(0, math.Min(255, math.Round(green))))
	b := uint8(math.Max(0, math.Min(255, math.Round(blue))))
	alpha := transpToAlpha(transp)
	return fmt.Sprintf("#%02X%02X%02X%02X", r, g, b, alpha)
}

/* PineColorR extracts the red component (0-255) from a hex color. Pine: color.r(color) */
func PineColorR(hex string) float64 {
	r, _, _, _ := parseHexColor(hex)
	return float64(r)
}

/* PineColorG extracts the green component (0-255) from a hex color. Pine: color.g(color) */
func PineColorG(hex string) float64 {
	_, g, _, _ := parseHexColor(hex)
	return float64(g)
}

/* PineColorB extracts the blue component (0-255) from a hex color. Pine: color.b(color) */
func PineColorB(hex string) float64 {
	_, _, b, _ := parseHexColor(hex)
	return float64(b)
}

/* PineColorT extracts the transparency (0-100) from a hex color. Pine: color.t(color) */
func PineColorT(hex string) float64 {
	_, _, _, a := parseHexColor(hex)
	return alphaToTransp(a)
}

/* PineColorFromGradient interpolates between two colors based on a value's position within a range.
 * Pine: color.from_gradient(value, bottom_value, top_value, bottom_color, top_color)
 */
func PineColorFromGradient(value, bottomValue, topValue float64, bottomColor, topColor string) string {
	if topValue == bottomValue {
		r, g, b, a := parseHexColor(bottomColor)
		return fmt.Sprintf("#%02X%02X%02X%02X", r, g, b, a)
	}
	ratio := (value - bottomValue) / (topValue - bottomValue)
	ratio = math.Max(0, math.Min(1, ratio))

	r1, g1, b1, a1 := parseHexColor(bottomColor)
	r2, g2, b2, a2 := parseHexColor(topColor)

	r := uint8(math.Round(float64(r1) + ratio*(float64(r2)-float64(r1))))
	g := uint8(math.Round(float64(g1) + ratio*(float64(g2)-float64(g1))))
	b := uint8(math.Round(float64(b1) + ratio*(float64(b2)-float64(b1))))
	a := uint8(math.Round(float64(a1) + ratio*(float64(a2)-float64(a1))))

	return fmt.Sprintf("#%02X%02X%02X%02X", r, g, b, a)
}
