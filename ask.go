package gluey

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
	"github.com/tanema/gluey/term"
)

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

func (ctx *Ctx) AnyKey(label string) error {
	ctx.Println(label)
	_, _, err := bufio.NewReader(os.Stdin).ReadRune()
	return err
}

func AnyKey(label string) error {
	return New().AnyKey(label)
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

// AskPassword prompts the user for a password input. Characters are not echoed
func (ctx *Ctx) AskPassword(label string) (string, error) {
	rdl, err := readline.New("")
	if err != nil {
		return "", err
	}
	result, err := rdl.ReadPassword(Fmt(`{{iconQ}} {{.}}`, label))
	return string(result), err
}

func AskPassword(label string) (string, error) {
	return New().AskPassword(label)
}

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
