package gluey

import (
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/chzyer/readline"
	"github.com/tanema/gluey/term"
)

type selectMode int

const selectTemplate = `{{.Prefix}}
{{- if .Done -}}
	{{ iconQ }} {{ .Label }} (You chose: {{ .Selected | italic }})
{{- else -}}
{{ iconQ }} {{ .Label }} {{ .HelpText | yellow }}
{{- if eq .Mode 1 }}
{{ .Prefix }}{{ .SelectTerm | green }} {{ .SelectHelp | blue }}
{{- end }}
{{- if eq .Mode 2 }}
{{ .Prefix }}{{ .SearchTerm | green }} {{ .FilterHelp | blue }}
{{- end}}
{{- if .Multiple }}
{{ .Prefix }}  0 {{ if gt .SelectCount 1 -}}
	{{ "Done" | bold }}
{{- else -}}
	Done
{{- end }}
{{- end }}
{{ range $index, $item := .Items -}}
	{{ $.Prefix }}
	{{- if eq $.Cursor $index -}}
		{{ iconSel | blue }} {{ $item.Index | blue }} {{ if $.Multiple }}
			{{- if .Chosen -}}
				{{ iconChk | blue }}
			{{- else -}}
				{{ iconBox | blue }}
			{{- end -}}
		{{- end }} {{$item.Label | blue }}
	{{- else }}  {{ $item.Index }} {{ if $.Multiple }}
			{{- if .Chosen -}}
				{{ iconChk }}
			{{- else -}}
				{{ iconBox }}
			{{- end -}}
		{{- end }} {{ if .Chosen -}}
			{{ $item.Label | bold }}
		{{- else -}}
			{{ $item.Label }}
		{{- end -}}
	{{- end }}
{{ else -}}
	{{ .Prefix }}no results
{{- end -}}
{{- end -}}
`

const (
	normal selectMode = iota
	selecting
	filtering
)

type selectTemplateData struct {
	Prefix      string
	Label       string
	Items       []*selectItem
	Selected    string
	SelectCount int
	SearchTerm  string
	SelectTerm  string
	HelpText    string
	FilterHelp  string
	SelectHelp  string
	Mode        selectMode
	Done        bool
	Multiple    bool
	Cursor      int
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
	sel := &Select{
		ctx:   ctx,
		label: label,
		items: convertSelectItems(items),
		size:  term.Height() - (1 + ctx.Indent),
	}
	sel.cancelSearch()
	return sel
}

func newMultipleSelect(ctx *Ctx, label string, items []string) *Select {
	sel := &Select{
		ctx:      ctx,
		label:    label,
		items:    convertSelectItems(items),
		multiple: true,
		size:     term.Height() - (2 + ctx.Indent),
	}
	sel.cancelSearch()
	return sel
}

// Run executes the select list. It displays the label and the list of items, asking the user to chose any
// value within to list. Run will keep the prompt alive until it has been canceled from
// the command prompt or it has received a valid value. It will return the value and an error if any
// occurred during the select's execution.
func (s *Select) run() ([]int, []string, error) {
	s.done = false
	stdin := readline.NewCancelableStdin(os.Stdin)
	c := &readline.Config{
		Stdin:          stdin,
		Stdout:         s.ctx.Writer(),
		HistoryLimit:   -1,
		UniqueEditLine: true,
	}
	if err := c.Init(); err != nil {
		return []int{}, []string{}, err
	}

	rl, err := readline.NewEx(c)
	if err != nil {
		return []int{}, []string{}, err
	}

	sb := term.NewScreenBuf(rl)
	defer sb.Done()

	c.SetListener(func(line []rune, pos int, key rune) ([]rune, int, bool) {
		s.listen(line, key)
		s.render(sb)
		if s.done {
			stdin.Close()
		}
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
	rl.Clean()
	rl.Close()
	time.Sleep(10 * time.Millisecond)

	indexes, items := s.Selected()
	return indexes, items, err
}

func (s *Select) listen(line []rune, key rune) {
	switch s.mode {
	case normal:
		switch {
		case key == readline.CharNext || key == 'j':
			s.next()
		case key == readline.CharPrev || key == 'k':
			s.prev()
		case key == 'f' || key == '/':
			s.mode = filtering
		case unicode.IsNumber(key):
			s.keyedSelectItem(key)
		case key == readline.CharEnter || key == ' ':
			s.selectItem(s.cursor)
		}
	case selecting:
		switch {
		case key == readline.CharEsc || key == 'j' || key == 'k' || key == readline.CharNext || key == readline.CharPrev:
			s.mode = normal
			s.selectTerm = ""
		case key == readline.CharBackspace:
			if len(s.selectTerm) > 0 {
				s.selectTerm = s.selectTerm[:len(s.selectTerm)-1]
				cur, _ := strconv.Atoi(s.selectTerm)
				s.SetCursor(cur - 1)
			} else {
				s.mode = normal
			}
		case key == readline.CharEnter || key == ' ':
			s.selectItem(s.cursor)
		default:
			s.keyedSelectItem(key)
		}
	case filtering:
		switch {
		case key == readline.CharNext:
			s.next()
		case key == readline.CharPrev:
			s.prev()
		case key == readline.CharEsc || key == readline.CharDelete:
			s.cancelSearch()
		case key == readline.CharBackspace:
			if len(s.searchTerm) > 0 {
				s.searchTerm = s.searchTerm[:len(s.searchTerm)-1]
				s.search(s.searchTerm)
			} else {
				s.cancelSearch()
			}
		case key == readline.CharEnter || key == ' ':
			s.selectItem(s.cursor)
		default:
			s.searchTerm += string(line)
			s.search(s.searchTerm)
		}
	}
}

func (s *Select) keyedSelectItem(key rune) {
	if len(s.items) > 9 {
		cur, err := strconv.Atoi(s.selectTerm + string(key))
		if err != nil {
			return
		}
		s.mode = selecting
		s.selectTerm += string(key)
		s.SetCursor(cur - 1)
		return
	}
	cur, err := strconv.Atoi(string(key))
	if err != nil {
		return
	}
	s.selectItem(cur - 1)
}

func (s *Select) selectItem(cursor int) {
	if len(s.scope) == 0 && cursor < 0 && cursor >= len(s.scope) {
		return
	}
	s.scope[cursor].Chosen = !s.scope[cursor].Chosen
	s.SetCursor(cursor)
	if !s.multiple {
		s.done = true
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
	s.mode = normal
	s.cursor = 0
	s.start = 0
	s.scope = s.items
}

// SetCursor will set the list cursor to a single item in the list
func (s *Select) SetCursor(i int) {
	s.cursor = clamp(i, 0, len(s.scope)-1)
	if s.start > s.cursor {
		s.start = s.cursor
	} else if s.start+s.size <= s.cursor {
		s.start = s.cursor - s.size + 1
	}
	if s.multiple && i == -1 && len(s.scope) > 0 {
		s.done = true
	}
}

func (s *Select) next() {
	if s.cursor >= len(s.scope)-1 {
		s.SetCursor(0)
	} else {
		s.SetCursor(s.cursor + 1)
	}
}

func (s *Select) prev() {
	if s.cursor <= 0 {
		s.SetCursor(len(s.scope) - 1)
	} else {
		s.SetCursor(s.cursor - 1)
	}
}

func (s *Select) scopedItems() []*selectItem {
	var items []*selectItem
	for i := s.start; i < min(s.start+s.size, len(s.scope)); i++ {
		items = append(items, s.scope[i])
	}
	return items
}

func (s *Select) selectedItems() []*selectItem {
	selected := []*selectItem{}
	for _, item := range s.items {
		if item.Chosen {
			selected = append(selected, item)
		}
	}
	return selected
}

// Selected returns the options that have been chosen
func (s *Select) Selected() ([]int, []string) {
	indexes := []int{}
	selected := []string{}
	for _, item := range s.items {
		if item.Chosen {
			indexes = append(indexes, item.Index)
			selected = append(selected, item.Label)
		}
	}
	return indexes, selected
}

func (s *Select) selectedLabel() string {
	selected := s.selectedItems()
	if len(selected) == 1 {
		return selected[0].Label
	} else if len(selected) == 2 {
		return selected[0].Label + " and " + selected[1].Label
	} else if len(selected) > 2 {
		return strconv.Itoa(len(selected)) + " Items"
	}
	return "<nothing>"
}

func (s *Select) render(sb *term.ScreenBuf) {
	template := selectTemplate
	templateData := selectTemplateData{
		Prefix:      s.ctx.Prefix(),
		Label:       s.label,
		Items:       s.scopedItems(),
		HelpText:    "(Choose with ↑ ↓ [Return], filter with 'f')",
		FilterHelp:  "Ctrl-D, Esc anytime or Backspace to exit",
		SelectHelp:  "Esc or up/down anytime to exit",
		SelectTerm:  "Select: " + s.selectTerm,
		SearchTerm:  "Filter: " + s.searchTerm,
		Selected:    s.selectedLabel(),
		SelectCount: len(s.selectedItems()),
		Mode:        s.mode,
		Done:        s.done,
		Multiple:    s.multiple,
		Cursor:      s.cursor - s.start,
	}

	if s.multiple {
		templateData.HelpText = strings.Replace(templateData.HelpText, "Choose", "Toggle", 1)
	}

	sb.WriteTmpl(template, templateData)
}

func max(x, y int) int {
	if x >= y {
		return x
	}
	return y
}

func min(x, y int) int {
	if x <= y {
		return x
	}
	return y
}

func clamp(a, minVal, maxVal int) int {
	return max(min(a, maxVal), minVal)
}
