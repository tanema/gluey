package promptui

import (
	"os"
	"strconv"
	"unicode"

	"github.com/chzyer/readline"
	"github.com/tanema/promptui/list"
	"github.com/tanema/promptui/term"
)

type selectMode int

const selectTemplate = `{{- if .Done}}{{iconQ}} {{.Label}}: (You chose: {{.List.Selected}}){{- else -}}
{{iconQ}} {{.Label}}: {{if ne .SearchTerm ""}}Filter: {{.SearchTerm}}{{end}}
{{range $index, $item := .List.Items -}}
{{$index}}. {{if $.Multiple}}{{if (index $.Chosen $index)}}☑{{else}}☐{{end}} {{end}}{{if eq $.List.Index $index}}{{$item | blue}}{{else}}{{$item}}{{end}}
{{else -}}
no results
{{- end -}}
{{- end -}}`

const (
	normal selectMode = iota
	selecting
	filtering
)

// Select represents a list of items used to enable selections, they can be used as search engines, menus
// or as a list of items in a cli based prompt.
type Select struct {
	Label      string
	Items      []string
	Chosen     []bool
	List       *list.List
	SearchTerm string
	mode       selectMode
	Done       bool
	Multiple   bool
}

// Run executes the select list. It displays the label and the list of items, asking the user to chose any
// value within to list. Run will keep the prompt alive until it has been canceled from
// the command prompt or it has received a valid value. It will return the value and an error if any
// occurred during the select's execution.
func (s *Select) Run() (int, string, error) {
	l, err := list.New(s.Items, 20)
	if err != nil {
		return 0, "", err
	}
	s.Done = false

	s.List = l

	stdin := readline.NewCancelableStdin(os.Stdin)
	c := &readline.Config{
		Stdin:          stdin,
		HistoryLimit:   -1,
		UniqueEditLine: true,
	}
	if err := c.Init(); err != nil {
		return 0, "", err
	}

	rl, err := readline.NewEx(c)
	if err != nil {
		return 0, "", err
	}

	rl.Write([]byte(term.HideCursor()))

	sb := term.NewScreenBuf(rl)
	c.SetListener(func(line []rune, pos int, key rune) ([]rune, int, bool) {
		if s.listen(line, key) {
			stdin.Close()
		}
		s.render(sb)
		return nil, 0, true
	})
	_, err = rl.Readline()
	rl.Write([]byte(term.ShowCursor()))
	rl.Clean()
	rl.Close()
	return s.List.Index(), s.List.Selected(), err
}

func (s *Select) listen(line []rune, key rune) bool {
	switch s.mode {
	case normal:
		switch {
		case key == readline.CharNext || key == 'j':
			s.List.Next()
		case key == readline.CharPrev || key == 'k':
			s.List.Prev()
		case key == readline.CharBackward || key == 'h':
			s.List.PageUp()
		case key == readline.CharForward || key == 'l':
			s.List.PageDown()
		case key == 'f' || key == '/':
			s.mode = filtering
		case unicode.IsNumber(key):
			cur, err := strconv.Atoi(string(key))
			if err == nil {
				s.List.SetCursor(cur - 1)
				s.Done = true
				return true
			}
		case key == readline.CharEnter && s.Multiple:
			s.Chosen[s.List.Index()] = !s.Chosen[s.List.Index()]
		case key == readline.CharEnter && !s.Multiple:
			s.Done = true
			return true
		}
	case selecting:
	case filtering:
		switch {
		case key == readline.CharEsc:
			s.mode = normal
			s.List.CancelSearch()
		case key == readline.CharBackspace:
			if len(s.SearchTerm) > 0 {
				s.SearchTerm = s.SearchTerm[:len(s.SearchTerm)-1]
				s.List.Search(s.SearchTerm)
			} else {
				s.mode = normal
				s.List.CancelSearch()
			}
		case key == readline.CharEnter:
			s.Done = true
			return true
		default:
			s.SearchTerm += string(line)
			s.List.Search(s.SearchTerm)
		}
	}
	return false
}

func (s *Select) render(sb *term.ScreenBuf) {
	sb.WriteTmpl(selectTemplate, s)
}
