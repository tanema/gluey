package term

import (
	"fmt"
	"strconv"
	"text/template"
)

type attribute int

type icon struct {
	color attribute
	char  string
}

var (
	iconInitial       = icon{fGBlue, "?"}
	iconGood          = icon{fGGreen, "*"}
	iconWarn          = icon{fGYellow, "⚠"}
	iconBad           = icon{fGRed, "X"}
	iconSelect        = icon{fGBold, ">"}
	iconCheckboxCheck = icon{fGBold, "☑"}
	iconCheckbox      = icon{fGBold, "☐"}
)

const (
	fGBold      attribute = 1
	fGFaint     attribute = 2
	fGItalic    attribute = 3
	fGUnderline attribute = 4

	fGBlack   attribute = 30
	fGRed     attribute = 31
	fGGreen   attribute = 32
	fGYellow  attribute = 33
	fGBlue    attribute = 94
	fGMagenta attribute = 35
	fGCyan    attribute = 36
	fGWhite   attribute = 97
)

const (
	bGBlack attribute = iota + 40
	bGRed
	bGGreen
	bGYellow
	bGBlue
	bGMagenta
	bGCyan
	bGWhite
)

var funcMap = template.FuncMap{
	"black":   styler(fGBlack),
	"red":     styler(fGRed),
	"green":   styler(fGGreen),
	"yellow":  styler(fGYellow),
	"blue":    styler(fGBlue),
	"magenta": styler(fGMagenta),
	"cyan":    styler(fGCyan),
	"white":   styler(fGWhite),

	"bgBlack":   styler(bGBlack),
	"bgRed":     styler(bGRed),
	"bgGreen":   styler(bGGreen),
	"bgYellow":  styler(bGYellow),
	"bgBlue":    styler(bGBlue),
	"bgMagenta": styler(bGMagenta),
	"bgCyan":    styler(bGCyan),
	"bgWhite":   styler(bGWhite),

	"bold":      styler(fGBold),
	"faint":     styler(fGFaint),
	"italic":    styler(fGItalic),
	"underline": styler(fGUnderline),

	"iconQ":    iconer(iconInitial),
	"iconGood": iconer(iconGood),
	"iconWarn": iconer(iconWarn),
	"iconBad":  iconer(iconBad),
	"iconSel":  iconer(iconSelect),
	"iconChk":  iconer(iconCheckboxCheck),
	"iconBox":  iconer(iconCheckbox),
}

func styler(attr attribute) func(interface{}) string {
	return func(v interface{}) string {
		s, ok := v.(string)
		if ok && s == ">>" {
			return fmt.Sprintf("\033[%sm", strconv.Itoa(int(attr)))
		}
		return fmt.Sprintf("\033[%sm%v%s", strconv.Itoa(int(attr)), v, "\033[0m")
	}
}

func iconer(ic icon) func() string {
	return func() string { return styler(ic.color)(ic.char) }
}
