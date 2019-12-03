package term

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"
)

type attribute int

type icon struct {
	color attribute
	char  string
}

const (
	reset       attribute = 0
	fGBold      attribute = 1
	fGFaint     attribute = 2
	fGItalic    attribute = 3
	fGUnderline attribute = 4
)

const (
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

var resetCode = fmt.Sprintf("\033[%dm", reset)

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
}

func styler(attrs ...attribute) func(interface{}) string {
	attrstrs := make([]string, len(attrs))
	for i, v := range attrs {
		attrstrs[i] = strconv.Itoa(int(v))
	}
	seq := strings.Join(attrstrs, ";")
	return func(v interface{}) string {
		end := ""
		s, ok := v.(string)
		if !ok || !strings.HasSuffix(s, resetCode) {
			end = resetCode
		}
		return fmt.Sprintf("\033[%sm%v%s", seq, v, end)
	}
}

func iconer(ic icon) func() string {
	return func() string {
		return styler(ic.color)(ic.char)
	}
}

// Sprintf formats a string template and outputs console ready text
func Sprintf(in string, data interface{}) string {
	return string(renderStringTemplate(in, data))
}

func renderStringTemplate(in string, data interface{}) []byte {
	tpl, err := template.New("").Funcs(funcMap).Parse(in)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}
