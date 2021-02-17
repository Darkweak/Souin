package auth

import (
	"github.com/darkweak/souin/errors"
	"testing"
)

func TestSignatureError_Error(t *testing.T) {
	s := signatureError{}
	if s.Error() != "An error occurred, Impossible to sign the JWT" {
		errors.GenerateError(t, "Signature Error function not compliant")
	}
}

func TestTokenError_Error(t *testing.T) {
	s := tokenError{}
	if s.Error() != "An error occurred, Invalid request" {
		errors.GenerateError(t, "Token Error function not compliant")
	}
}
