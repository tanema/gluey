package term

var (
	iconInitial       = icon{fGBlue, "?"}
	iconGood          = icon{fGGreen, "*"}
	iconWarn          = icon{fGYellow, "!"}
	iconBad           = icon{fGRed, "X"}
	iconSelect        = icon{fGBold, ">"}
	iconCheckboxCheck = icon{fGBold, "█"}
	iconCheckbox      = icon{fGBold, "░"}
)

// SpinGlyphs are the glyphs that the spinner uses for animation
var SpinGlyphs = []rune(`|/-\|/-\`)

// ReturnLabel allows platform dependent icon for return
var ReturnLabel = "[Ret]"
