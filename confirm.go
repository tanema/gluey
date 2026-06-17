package gluey

import (
	"fmt"
	"slices"
	"strings"
)

var validConfirmAnswers = []string{"yes", "no", "y", "n"}

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
