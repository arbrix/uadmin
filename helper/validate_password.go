package helper

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/muesli/crunchy"
)

const (
	minPasswordLengths = 14
)

var (
	ErrLength = errors.New("password is too short")
	ErrUpper  = errors.New("password should contains at least one uppercase letter")
	ErrLow    = errors.New("password should contains at least one lowercase letter")
	ErrNum    = errors.New("password should contains at least one digit")
	ErrSymbol = errors.New("password should contains at least one special symbol: @$!%*?&")
)

type PasswordValidator interface {
	Validate(string) error
}

var validator *passwordValidator

type passwordValidator struct {
	validator *crunchy.Validator
}

func NewPasswordValidator() PasswordValidator {
	if validator != nil {
		return validator
	}
	dictionaryPath := os.Getenv("PASS_DICTIONARY_PATH") // path from user
	validator := &passwordValidator{
		validator: crunchy.NewValidatorWithOpts(crunchy.Options{
			DictionaryPath: dictionaryPath, // is path is empty, crunchy will use default value  "/usr/share/dict"
		}),
	}
	return validator
}

func (p *passwordValidator) Validate(pass string) error {

	if len(pass) < minPasswordLengths {
		return ErrLength
	}

	uppercase := regexp.MustCompile(`[A-Z]`)
	if !uppercase.MatchString(pass) {
		return ErrUpper
	}

	lowercase := regexp.MustCompile(`[a-z]`)
	if !lowercase.MatchString(pass) {
		return ErrLow
	}

	number := regexp.MustCompile(`\d`)
	if !number.MatchString(pass) {
		return ErrNum
	}

	specialChar := regexp.MustCompile(`[@$!%*?&]`)
	if !specialChar.MatchString(pass) {
		return ErrSymbol
	}

	err := p.validator.Check(pass)
	if err != nil {
		return fmt.Errorf("weak pass: %s", err)
	}
	return nil
}
