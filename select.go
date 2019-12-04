package promptui

import (
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/chzyer/readline"
	"github.com/tanema/promptui/term"
)

type selectMode int

const selectTemplate = `{{.Prefix}}{{- if .Done}}{{iconQ}} {{.Label}} (You chose: {{.Selected}}){{- else -}}
{{iconQ}} {{.Label}} {{.HelpText | yellow}}
{{- if eq .Mode 1}}
{{.Prefix}}{{.SelectTerm | green}} {{if eq .SelectTerm "Select: "}}{{.SelectHelp|blue}}{{end}}{{end}}
{{- if eq .Mode 2}}
{{.Prefix}}{{.SearchTerm | green}} {{if eq .SearchTerm "Filter: "}}{{.FilterHelp|blue}}{{end}}{{end}}
{{range $index, $item := .Items -}}
{{$.Prefix}}{{if eq $.Cursor $index}}{{iconSel|blue}} {{$item.Index|blue}} {{$item.Label|blue}}{{else}}{{$item.Index}} {{$item.Label}}{{end}}
{{else -}}
{{.Prefix}}no results
{{- end -}}
{{- end -}}`

const selectMultiTemplate = `{{.Prefix}}{{- if .Done}}{{iconQ}} {{.Label}} (You chose: {{.Selected}}){{- else -}}
{{iconQ}} {{.Label}} {{.HelpText | yellow}}
{{- if eq .Mode 1}}
{{.Prefix}}{{.SelectTerm | green}} {{if eq .SelectTerm "Select: "}}{{.SelectHelp|blue}}{{end}}{{end}}
{{- if eq .Mode 2}}
{{.Prefix}}{{.SearchTerm | green}} {{if eq .SearchTerm "Filter: "}}{{.FilterHelp|blue}}{{end}}{{end}}
{{.Prefix}}  0 Done
{{range $index, $item := .Items -}}
{{$.Prefix}}{{if eq $.Cursor $index}}{{iconSel|blue}} {{$item.Index|blue}} {{if $item.Chosen}}{{iconChk|blue}}{{else}}{{iconBox|blue}}{{end}} {{$item.Label|blue}}{{else}}  {{$item.Index}} {{if $item.Chosen}}{{iconChk}}{{else}}{{iconBox}}{{end}} {{if $item.Chosen}}{{$item.Label|cyan}}{{else}}{{$item.Label}}{{end}}{{end}}
{{else -}}
{{.Prefix}}no results
{{- end -}}
{{- end -}}`

const (
	normal selectMode = iota
	selecting
	filtering
)

type selectTemplateData struct {
	Prefix     string
	Label      string
	Items      []*selectItem
	Selected   string
	SearchTerm string
	SelectTerm string
	HelpText   string
	FilterHelp string
	SelectHelp string
	Mode       selectMode
	Done       bool
	Cursor     int
}

// Select represents a list of items used to enable selections, they can be used as search engines, menus
// or as a list of items in a cli based prompt.
type Select struct {
	ctx        *Ctx
	label      string
	items      []*selectItem
	searchTerm string
	selectTerm string
	mode       selectMode
	done       bool
	cursor     int
	multiple   bool
	scope      []*selectItem
	size       int
	start      int
}

type selectItem struct {
	Label  string
	Chosen bool
	Index  int
}

func convertSelectItems(in []string) []*selectItem {
	out := make([]*selectItem, len(in))
	for i, label := range in {
		out[i] = &selectItem{Label: label, Index: i + 1, Chosen: false}
	}
	return out
}

func newSelect(ctx *Ctx, label string, items []string) *Select {
	_, rows, _ := term.Size()
	sel := &Select{
		ctx:   ctx,
		label: label,
		items: convertSelectItems(items),
		size:  rows,
	}
	sel.cancelSearch()
	return sel
}

func newMultipleSelect(ctx *Ctx, label string, items []string) *Select {
	_, rows, _ := term.Size()
	sel := &Select{
		ctx:      ctx,
		label:    label,
		items:    convertSelectItems(items),
		multiple: true,
		size:     rows,
	}
	sel.cancelSearch()
	return sel
}

// Run executes the select list. It displays the label and the list of items, asking the user to chose any
// value within to list. Run will keep the prompt alive until it has been canceled from
// the command prompt or it has received a valid value. It will return the value and an error if any
// occurred during the select's execution.
func (s *Select) Run() (int, string, error) {
	s.done = false
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
	for !s.done && err == nil {
		_, err = rl.Readline()
		if err != nil {
			switch {
			case err == readline.ErrInterrupt, err.Error() == "Interrupt":
				err = nil
			case err == io.EOF:
				err = nil
			}
		}
	}
	rl.Write([]byte(term.ShowCursor()))
	rl.Clean()
	rl.Close()
	time.Sleep(10 * time.Millisecond)

	item := s.scope[s.cursor]
	return item.Index, item.Label, err
}

func (s *Select) listen(line []rune, key rune) bool {
	switch s.mode {
	case normal:
		switch {
		case key == readline.CharNext || key == 'j':
			s.next()
		case key == readline.CharPrev || key == 'k':
			s.prev()
		case key == 'f' || key == '/':
			s.mode = filtering
		case key == 'e':
			s.mode = selecting
		case unicode.IsNumber(key) && s.multiple:
			cur, _ := strconv.Atoi(string(key))
			if cur == 0 {
				s.done = true
				return true
			} else if cur < len(s.items) {
				s.items[cur-1].Chosen = !s.items[cur-1].Chosen
			}
		case unicode.IsNumber(key) && !s.multiple:
			cur, _ := strconv.Atoi(string(key))
			s.SetCursor(cur - 1)
			s.done = true
			return true
		case key == readline.CharEnter && s.multiple:
			s.scope[s.cursor].Chosen = !s.scope[s.cursor].Chosen
		case key == readline.CharEnter && !s.multiple:
			s.done = true
			return true
		}
	case selecting:
		switch {
		case key == readline.CharEsc:
			s.mode = normal
			s.selectTerm = ""
		case key == readline.CharBackspace:
			if len(s.selectTerm) > 0 {
				s.selectTerm = s.selectTerm[:len(s.selectTerm)-1]
				cur, _ := strconv.Atoi(s.selectTerm)
				s.SetCursor(cur + 1)
			} else {
				s.mode = normal
			}
		case key == readline.CharEnter && s.multiple:
			s.scope[s.cursor].Chosen = !s.scope[s.cursor].Chosen
		case key == readline.CharEnter && !s.multiple:
			s.done = true
			return true
		default:
			s.selectTerm += string(line)
			cur, _ := strconv.Atoi(s.selectTerm)
			s.SetCursor(cur + 1)
		}
	case filtering:
		switch {
		case key == readline.CharEsc:
			s.mode = normal
			s.cancelSearch()
		case key == readline.CharBackspace:
			if len(s.searchTerm) > 0 {
				s.searchTerm = s.searchTerm[:len(s.searchTerm)-1]
				s.search(s.searchTerm)
			} else {
				s.mode = normal
				s.cancelSearch()
			}
		case key == readline.CharEnter && s.multiple:
			s.scope[s.cursor].Chosen = !s.scope[s.cursor].Chosen
		case key == readline.CharEnter && !s.multiple:
			s.done = true
			return true
		default:
			s.searchTerm += string(line)
			s.search(s.searchTerm)
		}
	}
	return false
}

func (s *Select) prev() {
	if s.cursor > 0 {
		s.cursor--
	}
	if s.start > s.cursor {
		s.start = s.cursor
	}
}

func (s *Select) search(term string) {
	term = strings.Trim(term, " ")
	s.cursor = 0
	s.start = 0
	scope := []*selectItem{}
	for _, item := range s.items {
		if strings.Contains(strings.ToLower(item.Label), strings.ToLower(term)) {
			scope = append(scope, item)
		}
	}
	s.scope = scope
}

func (s *Select) cancelSearch() {
	s.cursor = 0
	s.start = 0
	s.scope = s.items
}

// SetCursor will set the list cursor to a single item in the list
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

func (s *Select) next() {
	max := len(s.scope) - 1
	if s.cursor < max {
		s.cursor++
	}
	if s.start+s.size <= s.cursor {
		s.start = s.cursor - s.size + 1
	}
}

func (s *Select) render(sb *term.ScreenBuf) {
	var items []*selectItem
	end := s.start + s.size
	if end > len(s.scope) {
		end = len(s.scope)
	}
	for i, j := s.start, 0; i < end; i, j = i+1, j+1 {
		items = append(items, s.scope[i])
	}

	template := selectTemplate
	templateData := selectTemplateData{
		Prefix:     s.ctx.Prefix(),
		Label:      s.label,
		Items:      items,
		HelpText:   "(Choose with ↑ ↓ ⏎, filter with 'f')",
		FilterHelp: "Ctrl-D anytime or Backspace now to exit",
		SelectHelp: "e, q, or up/down anytime to exit",
		SelectTerm: "Select: " + s.selectTerm,
		SearchTerm: "Filter: " + s.searchTerm,
		Mode:       s.mode,
		Done:       s.done,
		Cursor:     s.cursor,
	}

	if len(s.items) > 9 {
		templateData.HelpText = "(Choose with ↑ ↓ ⏎, filter with 'f', enter option with 'e')"
	}

	if s.multiple {
		selected := []string{}
		for _, item := range s.items {
			if item.Chosen {
				selected = append(selected, item.Label)
			}
		}

		template = selectMultiTemplate
		if len(selected) == 1 {
			templateData.Selected = selected[0]
		} else if len(selected) == 2 {
			templateData.Selected = selected[0] + " and " + selected[1]
		} else if len(selected) > 2 {
			templateData.Selected = strconv.Itoa(len(selected)) + " Items"
		}
		templateData.HelpText = strings.Replace(templateData.HelpText, "Choose", "Toggle", 1)
	} else if s.cursor < len(s.scope) && s.scope[s.cursor] != nil {
		templateData.Selected = s.scope[s.cursor].Label
	}

	sb.WriteTmpl(template, templateData)
}
