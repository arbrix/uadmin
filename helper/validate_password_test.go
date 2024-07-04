package helper

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Validate(t *testing.T) {

	testCases := []struct {
		name     string
		arg      string
		expError error
	}{
		{
			name:     "success",
			arg:      "1q?btAjhpztqnln",
			expError: nil,
		},
		{
			name:     "failed, password is too short",
			arg:      "",
			expError: ErrLength,
		},
		{
			name:     "failed, password should contains at least one uppercase letter",
			arg:      "1q?btajhpztqnln",
			expError: ErrUpper,
		},
		{
			name:     "failed, password should contains at least one lowercase letter",
			arg:      "1Q?BTAJHPZTQNLTN",
			expError: ErrLow,
		},
		{
			name:     "failed, password should contains at least one digit",
			arg:      "aq?btAjhpztqnln",
			expError: ErrNum,
		},
		{
			name:     "failed, password should contains at least one special symbol",
			arg:      "aq1Btajhpztqnln",
			expError: ErrSymbol,
		},
		{
			name:     "failed, password should contains don't repeat parts",
			arg:      "1234567812345678?Aa",
			expError: fmt.Errorf("weak pass: Password is too systematic"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := NewPasswordValidator()

			err := v.Validate(tc.arg)
			if tc.expError != nil {
				assert.Equal(t, tc.expError.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}

		})
	}
}

func Test_Validate_search_in_dictionary(t *testing.T) {
	// create temp folder with dictionary
	tmp := t.TempDir()

	err := os.WriteFile(fmt.Sprintf("%s/dictionary", tmp), []byte("123PasswordFromDictionary!"), os.FileMode(0644))
	require.Nil(t, err)

	err = os.Setenv("PASS_DICTIONARY_PATH", tmp)
	require.Nil(t, err)
	defer os.Unsetenv("PASS_DICTIONARY_PATH")

	v := NewPasswordValidator()

	err = v.Validate("123PasswordFromDictionary!")
	require.NotNil(t, err)
	assert.Equal(t, "weak pass: Password is too common / from a dictionary", err.Error())
}
