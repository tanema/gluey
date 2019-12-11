package term

var (
	iconInitial       = icon{fGBlue, "?"}
	iconGood          = icon{fGGreen, "v"}
	iconWarn          = icon{fGYellow, "!"}
	iconBad           = icon{fGRed, "x"}
	iconSelect        = icon{fGBold, ">"}
	iconCheckboxCheck = icon{fGBold, "☑"}
	iconCheckbox      = icon{fGBold, "☐"}
)

// SpinGlyphs are used to display a spinner
var SpinGlyphs = []rune("▖▌▘▀▝▐▗▂")
