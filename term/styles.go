// +build !windows

package term

var (
	iconInitial       = icon{fGBlue, "?"}
	iconGood          = icon{fGGreen, "✔"}
	iconWarn          = icon{fGYellow, "⚠"}
	iconBad           = icon{fGRed, "✗"}
	iconSelect        = icon{fGBold, "▸"}
	iconCheckboxCheck = icon{fGBold, "☑"}
	iconCheckbox      = icon{fGBold, "☐"}
)

// SpinGlyphs are used to display a spinner
//var SpinGlyphs = []rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")
var SpinGlyphs = []rune("◴◷◶◵◐◓◑◒")
