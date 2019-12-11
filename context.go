package gluey

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
	"github.com/tanema/gluey/term"
)

type fnCompleter struct{}

func (fc *fnCompleter) Do(line []rune, pos int) ([][]rune, int) {
	sug, _ := filepath.Glob(string(line) + "*")
	items := [][]rune{}
	for _, item := range sug {
		suggestion := strings.TrimPrefix(item, string(line))
		info, err := os.Stat(item)
		if err == nil && info.Mode().IsDir() {
			suggestion += string(os.PathSeparator)
		}
		items = append(items, []rune(suggestion))
	}
	return items, len(line)
}

// Ctx allows use to keep a root object for all elements
type Ctx struct {
	*log.Logger
	Indent int
}

// New builds a new UI context that every element will be based on
func New() *Ctx {
	return &Ctx{
		Logger: log.New(os.Stdout, "", 0),
	}
}

// Fmt will format a string template with color and icons
func Fmt(template string, data interface{}) string {
	return term.Sprintf(template, data)
}

// Ask will prompt the user for a string input and will not return until a value
// is passed. If the value is an empty string, the user will be re-prompted.
func (ctx *Ctx) Ask(label string) (input string, err error) {
	ctx.Println(Fmt(`{{iconQ}} {{.}}`, label))
	for input = ""; input == "" && err == nil; input, err = ctx.ask("") {
	}
	return input, err
}

// AskDefault will prompt the user for a string input. If the input is an empty
// string then the defalt value will be returned
func (ctx *Ctx) AskDefault(label, what string) (string, error) {
	ctx.Println(Fmt(`{{iconQ}} {{.Lab}} {{.Def | faint}}`, struct{ Lab, Def string }{label, "[default = " + what + "]"}))
	return ctx.ask(what)
}

// AskFile will prompt the user for a filepath with autocomplete
func (ctx *Ctx) AskFile(label string) (string, error) {
	ctx.Println(Fmt(`{{iconQ}} {{.}}`, label))
	c := &readline.Config{
		AutoComplete: &fnCompleter{},
		Stdin:        os.Stdin,
	}
	if err := c.Init(); err != nil {
		return "", err
	}
	rl, err := readline.NewEx(c)
	if err != nil {
		return "", err
	}
	defer rl.Close()
	return rl.Readline()
}

func (ctx *Ctx) ask(what string) (string, error) {
	prompt := Fmt(`{{.}}{{">" | blue}}`, ctx.Prefix())
	rdl, err := readline.New(prompt + Fmt(`{{.}} `, term.Sgr("93")))
	if err != nil {
		return "", err
	}

	result, err := rdl.Readline()
	if err != nil {
		return "", err
	}

	if what != "" && result == "" {
		ctx.Println(term.PreviousLine() + term.ClearToEndOfLine() + prompt + Fmt(` {{.|yellow}} `, what))
		result = what
	}
	return result, nil
}

// Select will propt the user with a list and will allow them to select a single option
func (ctx *Ctx) Select(label string, items []string) (int, string, error) {
	return newSelect(ctx, label, items).Run()
}

// SelectMultiple will propt the user with a list and will allow them to select multiple options
func (ctx *Ctx) SelectMultiple(label string, items []string) (int, string, error) {
	return newMultipleSelect(ctx, label, items).Run()
}

// Confirm will prompt the user with a yes/no option. The dflt setting will decide
// if the first option is yes or no so that the user can just press enter
func (ctx *Ctx) Confirm(label string, dflt bool) (bool, error) {
	var err error
	var result string
	if dflt {
		_, result, err = ctx.Select(label, []string{"yes", "no"})
	} else {
		_, result, err = ctx.Select(label, []string{"no", "yes"})
	}
	return result == "yes", err
}

// Password prompts the user for a password input. Characters are not echoed
func (ctx *Ctx) Password(label string) (string, error) {
	rdl, err := readline.New("")
	if err != nil {
		return "", err
	}
	result, err := rdl.ReadPassword(Fmt(`{{iconQ}} {{.}}`, label))
	return string(result), err
}

// InFrame will format output to be inside a frame
func (ctx *Ctx) InFrame(title string, fn func(*Ctx) error) error {
	width, _ := term.Size()
	ctx.Println(Fmt("{{ . | cyan }}", "┏ "+title+" "+strings.Repeat("━", width-len(title)-ctx.Indent-3)))
	nestedCtx := &Ctx{
		Logger: log.New(ctx.Writer(), Fmt(`{{.}}{{ "┃" | cyan }} `, ctx.Prefix()), 0),
		Indent: ctx.Indent + 2,
	}
	if err := fn(nestedCtx); err != nil {
		ctx.Println(Fmt("{{ . | red }}", "┣"+strings.Repeat("━", width-ctx.Indent-1)))
		ctx.Println(Fmt("{{ . | red }}", "┃"+err.Error()))
		ctx.Println(Fmt("{{ . | red }}", "┗"+strings.Repeat("━", width-ctx.Indent-1)))
		return err
	}
	ctx.Println(Fmt("{{ . | cyan }}", "┗"+strings.Repeat("━", width-ctx.Indent-1)))
	return nil
}
