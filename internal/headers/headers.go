package headers

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"unicode"
)

type Headers map[string]string

const crlf = "\r\n"

func NewHeaders() Headers {
	headers := make(map[string]string)

	return headers
}

func (h Headers) Get(key string) (string, bool) {
	val, ok := h[strings.ToLower(key)]

	return val, ok
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}

	if idx == 0 {
		return 2, true, nil
	}

	headerText := string(data[:idx])
	parts := strings.SplitN(headerText, ":", 2)

	readBytes := len(headerText) + 2

	// fmt.Printf("header text: %v\n", headerText)
	if len(parts) < 2 {
		return 0, true, errors.New("invalid header parts")
	}

	key := parts[0]
	// key cannot have any space before semicolom
	lastChar := key[len(key)-1]
	if unicode.IsSpace(rune(lastChar)) {
		return 0, false, errors.New("invalid key")
	}

	finalKey := strings.ToLower(strings.TrimSpace(key))
	finalVal := strings.TrimSpace(parts[1])

	if !isValidKey(finalKey) {
		return 0, false, errors.New("invalid key")
	}

	val, ok := h[finalKey]

	if ok {
		h[finalKey] = val + ", " + finalVal
	} else {
		h[finalKey] = finalVal
	}

	return readBytes, false, nil
}

func isValidKey(key string) bool {
	validSpecialChars := []string{"!", "#", "$", "%", "&", "'", "*", "+", "-", "^", "_", "`", "|", "~"}

	for _, char := range key {
		isDigit := unicode.IsDigit(char)
		isAlpha := unicode.IsLetter(char)
		isInSpecialChars := slices.Contains(validSpecialChars, string(char))

		if !isDigit && !isAlpha && !isInSpecialChars {
			return false
		}
	}

	return true
}
