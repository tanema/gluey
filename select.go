package promptui

import (
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/chzyer/readline"
	"github.com/tanema/promptui/term"
)

type selectMode int

const selectTemplate = `{{- if .Done}}{{iconQ}} {{.Label}}: (You chose: {{.Selected}}){{- else -}}
{{iconQ}} {{.Label}}: {{if ne .SearchTerm ""}}Filter: {{.SearchTerm}}{{end}}
{{range $index, $item := .FilteredItems -}}
{{(inc $index)}}. {{if $.Multiple}}{{if (index $.Chosen $index)}}☑{{else}}☐{{end}} {{end}}{{if eq $.Index $index}}{{iconSel | blue}}{{$item | blue}}{{else}}{{$item}}{{end}}
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
	ctx        *Ctx
	Label      string
	Items      []string
	Chosen     []bool
	SearchTerm string
	mode       selectMode
	Done       bool
	Multiple   bool

	scope  []string
	cursor int
	size   int
	start  int
}

func newSelect(ctx *Ctx, label string, items []string) *Select {
	_, rows, _ := term.Size()
	sel := &Select{
		ctx:   ctx,
		Label: label,
		Items: items,
		scope: items,
		size:  rows,
	}
	return sel
}

func newMultipleSelect(ctx *Ctx, label string, items []string) *Select {
	_, rows, _ := term.Size()
	sel := &Select{
		ctx:    ctx,
		Label:  label,
		Items:  items,
		Chosen: make([]bool, len(items)),
		size:   rows,
	}
	return sel
}

// Run executes the select list. It displays the label and the list of items, asking the user to chose any
// value within to list. Run will keep the prompt alive until it has been canceled from
// the command prompt or it has received a valid value. It will return the value and an error if any
// occurred during the select's execution.
func (s *Select) Run() (int, string, error) {
	s.Done = false
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
	time.Sleep(10 * time.Millisecond)
	return s.Index(), s.Selected(), err
}

func (s *Select) listen(line []rune, key rune) bool {
	switch s.mode {
	case normal:
		switch {
		case key == readline.CharNext || key == 'j':
			s.Next()
		case key == readline.CharPrev || key == 'k':
			s.Prev()
		case key == 'f' || key == '/':
			s.mode = filtering
		case unicode.IsNumber(key):
			cur, err := strconv.Atoi(string(key))
			if err == nil {
				s.SetCursor(cur - 1)
				s.Done = true
				return true
			}
		case key == readline.CharEnter && s.Multiple:
			s.Chosen[s.Index()] = !s.Chosen[s.Index()]
		case key == readline.CharEnter && !s.Multiple:
			s.Done = true
			return true
		}
	case selecting:
	case filtering:
		switch {
		case key == readline.CharEsc:
			s.mode = normal
			s.CancelSearch()
		case key == readline.CharBackspace:
			if len(s.SearchTerm) > 0 {
				s.SearchTerm = s.SearchTerm[:len(s.SearchTerm)-1]
				s.Search(s.SearchTerm)
			} else {
				s.mode = normal
				s.CancelSearch()
			}
		case key == readline.CharEnter:
			s.Done = true
			return true
		default:
			s.SearchTerm += string(line)
			s.Search(s.SearchTerm)
		}
	}
	return false
}

func (s *Select) Prev() {
	if s.cursor > 0 {
		s.cursor--
	}

	if s.start > s.cursor {
		s.start = s.cursor
	}
}

func (s *Select) Search(term string) {
	term = strings.Trim(term, " ")
	s.cursor = 0
	s.start = 0
	s.search(term)
}

func (s *Select) CancelSearch() {
	s.cursor = 0
	s.start = 0
	s.scope = s.Items
}

func (s *Select) search(term string) {
	scope := []string{}
	for _, item := range s.Items {
		if strings.Contains(strings.ToLower(item), strings.ToLower(term)) {
			scope = append(scope, item)
		}
	}
	s.scope = scope
}

func (s *Select) SetCursor(i int) {
	max := len(s.scope) - 1
	if i >= max {
		i = max
	}
	if i < 0 {
		i = 0
	}
	s.cursor = i

	if s.start > s.cursor {
		s.start = s.cursor
	} else if s.start+s.size <= s.cursor {
		s.start = s.cursor - s.size + 1
	}
}

func (s *Select) Next() {
	max := len(s.scope) - 1

	if s.cursor < max {
		s.cursor++
	}

	if s.start+s.size <= s.cursor {
		s.start = s.cursor - s.size + 1
	}
}

func (s *Select) Cursor() (int, string) {
	selected := s.scope[s.cursor]
	for i, item := range s.Items {
		if item == selected {
			return i, item
		}
	}
	return -1, ""
}

func (s *Select) Index() int {
	i, _ := s.Cursor()
	return i
}

func (s *Select) Selected() string {
	_, item := s.Cursor()
	return item
}

func (s *Select) FilteredItems() []string {
	var result []string
	max := len(s.scope)
	end := s.start + s.size

	if end > max {
		end = max
	}

	for i, j := s.start, 0; i < end; i, j = i+1, j+1 {
		result = append(result, s.scope[i])
	}

	return result
}

func (s *Select) render(sb *term.ScreenBuf) {
	sb.WriteTmpl(selectTemplate, s)
}
