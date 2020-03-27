package errors

import (
	"testing"
)

// GenerateError Syntactic sugar to display errors
func GenerateError(t *testing.T, text string) {
	t.Errorf("An error occurred : %s", text)
}
