package list

import (
	"fmt"
	"strings"
)

// NotFound is an index returned when no item was selected. This could
// happen due to a search without results.
const NotFound = -1

// List holds a collection of items that can be displayed with an N number of
// visible items. The list can be moved up, down by one item of time or an
// entire page (ie: visible size). It keeps track of the current selected item.
type List struct {
	items  []string
	scope  []string
	cursor int // cursor holds the index of the current selected item
	size   int // size is the number of visible options
	start  int
}

// New creates and initializes a list of searchable items. The items attribute must be a slice type with a
// size greater than 0. Error will be returned if those two conditions are not met.
func New(items []string, size int) (*List, error) {
	if size < 1 {
		return nil, fmt.Errorf("list size %d must be greater than 0", size)
	}
	return &List{size: size, items: items, scope: items}, nil
}

// Prev moves the visible list back one item. If the selected item is out of
// view, the new select item becomes the last visible item. If the list is
// already at the top, nothing happens.
func (l *List) Prev() {
	if l.cursor > 0 {
		l.cursor--
	}

	if l.start > l.cursor {
		l.start = l.cursor
	}
}

// Search allows the list to be filtered by a given term. The list must
// implement the searcher function signature for this functionality to work.
func (l *List) Search(term string) {
	term = strings.Trim(term, " ")
	l.cursor = 0
	l.start = 0
	l.search(term)
}

// CancelSearch stops the current search and returns the list to its
// original order.
func (l *List) CancelSearch() {
	l.cursor = 0
	l.start = 0
	l.scope = l.items
}

func (l *List) search(term string) {
	scope := []string{}
	for _, item := range l.items {
		if strings.Contains(strings.ToLower(item), strings.ToLower(term)) {
			scope = append(scope, item)
		}
	}
	l.scope = scope
}

// Start returns the current render start position of the list.
func (l *List) Start() int {
	return l.start
}

// SetStart sets the current scroll position. Values out of bounds will be
// clamped.
func (l *List) SetStart(i int) {
	if i < 0 {
		i = 0
	}
	if i > l.cursor {
		l.start = l.cursor
	} else {
		l.start = i
	}
}

// SetCursor sets the position of the cursor in the list. Values out of bounds
// will be clamped.
func (l *List) SetCursor(i int) {
	max := len(l.scope) - 1
	if i >= max {
		i = max
	}
	if i < 0 {
		i = 0
	}
	l.cursor = i

	if l.start > l.cursor {
		l.start = l.cursor
	} else if l.start+l.size <= l.cursor {
		l.start = l.cursor - l.size + 1
	}
}

// Next moves the visible list forward one item. If the selected item is out of
// view, the new select item becomes the first visible item. If the list is
// already at the bottom, nothing happens.
func (l *List) Next() {
	max := len(l.scope) - 1

	if l.cursor < max {
		l.cursor++
	}

	if l.start+l.size <= l.cursor {
		l.start = l.cursor - l.size + 1
	}
}

// PageUp moves the visible list backward by x items. Where x is the size of the
// visible items on the list. The selected item becomes the first visible item.
// If the list is already at the bottom, the selected item becomes the last
// visible item.
func (l *List) PageUp() {
	start := l.start - l.size
	if start < 0 {
		l.start = 0
	} else {
		l.start = start
	}

	cursor := l.start

	if cursor < l.cursor {
		l.cursor = cursor
	}
}

// PageDown moves the visible list forward by x items. Where x is the size of
// the visible items on the list. The selected item becomes the first visible
// item.
func (l *List) PageDown() {
	start := l.start + l.size
	max := len(l.scope) - l.size

	switch {
	case len(l.scope) < l.size:
		l.start = 0
	case start > max:
		l.start = max
	default:
		l.start = start
	}

	cursor := l.start

	if cursor == l.cursor {
		l.cursor = len(l.scope) - 1
	} else if cursor > l.cursor {
		l.cursor = cursor
	}
}

// CanPageDown returns whether a list can still PageDown().
func (l *List) CanPageDown() bool {
	max := len(l.scope)
	return l.start+l.size < max
}

// CanPageUp returns whether a list can still PageUp().
func (l *List) CanPageUp() bool {
	return l.start > 0
}

// Index returns the index of the item currently selected inside the searched list. If no item is selected,
// the NotFound (-1) index is returned.
func (l *List) Index() int {
	selected := l.scope[l.cursor]

	for i, item := range l.items {
		if item == selected {
			return i
		}
	}

	return NotFound
}

func (l *List) Selected() string {
	selected := l.scope[l.cursor]

	for _, item := range l.items {
		if item == selected {
			return item
		}
	}

	return ""
}

// Items returns a slice equal to the size of the list with the current visible
// items
func (l *List) Items() []string {
	var result []string
	max := len(l.scope)
	end := l.start + l.size

	if end > max {
		end = max
	}

	for i, j := l.start, 0; i < end; i, j = i+1, j+1 {
		result = append(result, l.scope[i])
	}

	return result
}
