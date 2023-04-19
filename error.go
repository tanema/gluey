package gluey

import (
	"fmt"
	"strings"
)

// GroupError will combine errors for a group of concurrent jobs and make them
// act as a single error but with the ability to still access the errors
type GroupError struct {
	Errors map[string]error
}

func (err *GroupError) Error() string {
	points := make([]string, 0, len(err.Errors))
	for name, e := range err.Errors {
		points = append(points, fmt.Sprintf("* %s: %s", name, e))
	}
	return fmt.Sprintf("%d errors occurred:\n\t%s\n\n", len(err.Errors), strings.Join(points, "\n\t"))
}
