package web

import "testing"

func Test_validEmail(t *testing.T) {
	valid := []string{
		"testEmail2@email.com",
		"tester@test.no",
		"tes@test.no", // The least amount of characters in the name is 3
	}

	notValid := []string{
		"",
		"@test.com",   // Need name part
		"t@test.com",  // Name must be at least 3 characters
		"te@test.com", // As above
		"test@t.com",  // Domain must have at least 2 characters before and after the dot
		"test@test.n", // Only one character after the dot
		"test@t.c",    // Only one character before and after
	}

	for _, email := range valid {
		if !validEmail(email) {
			t.Errorf("'%s' should be valid", email)
		}
	}

	for _, email := range notValid {
		if validEmail(email) {
			t.Errorf("'%s' should NOT be valid", email)
		}
	}
}

func Test_validPassword(t *testing.T) {
	valid := []string{ // So far all passwords with >= 5 length are valid, so this
		"abcdef", // is a template if more complex validation is added later
		"ABCDef",
		"av2cde",
	}

	notValid := []string{
		"abcd",
	}

	for _, password := range valid {
		if !validPassword(password) {
			t.Errorf("'%s' should be valid", password)
		}
	}

	for _, password := range notValid {
		if validPassword(password) {
			t.Errorf("'%s' should NOT be valid", password)
		}
	}
}
