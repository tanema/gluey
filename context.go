package gluey

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
	"github.com/k0kubun/go-ansi"
	"github.com/tanema/gluey/term"
	"golang.org/x/exp/slices"
)

var validConfirmAnswers = []string{"yes", "no", "y", "n"}

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
	return &Ctx{Logger: log.New(ansi.NewAnsiStdout(), "", 0)}
}

// Fmt will format a string template with color and icons
func Fmt(template string, data any) string {
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

func Ask(label string) (input string, err error) {
	return New().Ask(label)
}

// AskDefault will prompt the user for a string input. If the input is an empty
// string then the defalt value will be returned
func (ctx *Ctx) AskDefault(label, what string) (string, error) {
	ctx.Println(Fmt(`{{iconQ}} {{.Lab}} {{.Def | faint}}`, struct{ Lab, Def string }{label, "[default = " + what + "]"}))
	return ctx.ask(what)
}

func AskDefault(label, what string) (string, error) {
	return New().AskDefault(label, what)
}

// Confirm will as a yes or no question and wait for an answer that is one of those.
func (ctx *Ctx) Confirm(question string) (bool, error) {
	for {
		if response, err := ctx.Ask(fmt.Sprintf("%s [y/n]: ", question)); err != nil {
			return false, err
		} else if res := strings.ToLower(strings.TrimSpace(response)); slices.Contains(validConfirmAnswers, res) {
			return res == "y" || res == "yes", nil
		}
	}
}

func Confirm(question string) (bool, error) {
	return New().Confirm(question)
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

func AskFile(label string) (string, error) {
	return New().AskFile(label)
}

func (ctx *Ctx) ask(what string) (string, error) {
	prompt := Fmt(`{{.}}{{blue ">"}} {{yellow ">>"}}`, ctx.Prefix())
	rdl, err := readline.New(prompt)
	if err != nil {
		return "", err
	}

	result, err := rdl.Readline()
	if err != nil {
		return "", err
	}

	if what != "" && result == "" {
		term.ClearLines(ctx.Writer(), 1)
		ctx.Println(prompt + Fmt(`{{.|yellow}} `, what))
		result = what
	}
	return result, nil
}

// ConfirmSelect will prompt the user with a yes/no option. The dflt setting will
// decide if the first option is yes or no so that the user can just press enter
func (ctx *Ctx) ConfirmSelect(label string, dflt bool) (bool, error) {
	var err error
	var result string
	if dflt {
		_, result, err = ctx.Select(label, []string{"yes", "no"})
	} else {
		_, result, err = ctx.Select(label, []string{"no", "yes"})
	}
	return result == "yes", err
}

func ConfirmSelect(label string, dflt bool) (bool, error) {
	return New().ConfirmSelect(label, dflt)
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

func Password(label string) (string, error) {
	return New().Password(label)
}

// InFrame will format output to be inside a frame
func (ctx *Ctx) InFrame(title string, fn FrameFunc) error {
	return newFrame(ctx).run(title, fn)
}

func InFrame(title string, fn FrameFunc) error {
	return newFrame(New()).run(title, fn)
}

// Debreif is a convienience method to format multi error return from
// spin groups and progress groups, it will also return the first error
// for returning errors inside frames
func (ctx *Ctx) Debreif(errors map[string]error) error {
	if len(errors) == 0 {
		return nil
	}
	var firstErrTitle string
	for title, err := range errors {
		if firstErrTitle == "" {
			firstErrTitle = title
		}
		frame := newFrame(ctx)
		frame.SetColor("red")
		frame.run("Task Failed: "+title, func(c *Ctx, f *Frame) error {
			c.Println(err)
			return nil
		})
	}
	return errors[firstErrTitle]
}
