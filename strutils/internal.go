package strutils

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type wordType int

const (
	invalidWord wordType = iota
	numberWord
	upperCaseWord
	alphabetWord
	connectorWord
	punctWord
	otherWord
)

const minCJKCharacter = '\u3400'

// Checks r is a letter but not CJK character.
func isAlphabet(r rune) bool {
	if !unicode.IsLetter(r) {
		return false
	}

	switch {
	// Quick check for non-CJK character.
	case r < minCJKCharacter:
		return true

	// Common CJK characters.
	case r >= '\u4E00' && r <= '\u9FCC':
		return false

	// Rare CJK characters.
	case r >= '\u3400' && r <= '\u4D85':
		return false

	// Rare and historic CJK characters.
	case r >= '\U00020000' && r <= '\U0002B81D':
		return false
	}

	return true
}

func isConnector(r rune) bool {
	return r == '-' || r == '_' || unicode.IsSpace(r)
}

func toLower(buf *strings.Builder, wt wordType, str string, connector rune) {
	buf.Grow(buf.Len() + len(str))

	if wt != upperCaseWord && wt != connectorWord {
		buf.WriteString(str)
		return
	}

	for len(str) > 0 {
		r, size := utf8.DecodeRuneInString(str)
		str = str[size:]

		if isConnector(r) {
			buf.WriteRune(connector)
		} else if unicode.IsUpper(r) {
			buf.WriteRune(unicode.ToLower(r))
		} else {
			buf.WriteRune(r)
		}
	}
}
func nextWord(str string) (wt wordType, word, remaining string) {
	if len(str) == 0 {
		return
	}

	var offset int
	remaining = str
	r, size := nextValidRune(remaining, utf8.RuneError)
	offset += size

	if r == utf8.RuneError {
		wt = invalidWord
		word = str[:offset]
		remaining = str[offset:]
		return
	}

	switch {
	case isConnector(r):
		wt = connectorWord
		remaining = remaining[size:]

		for len(remaining) > 0 {
			r, size = nextValidRune(remaining, r)

			if !isConnector(r) {
				break
			}

			offset += size
			remaining = remaining[size:]
		}

	case unicode.IsPunct(r):
		wt = punctWord
		remaining = remaining[size:]

		for len(remaining) > 0 {
			r, size = nextValidRune(remaining, r)

			if !unicode.IsPunct(r) {
				break
			}

			offset += size
			remaining = remaining[size:]
		}

	case unicode.IsUpper(r):
		wt = upperCaseWord
		remaining = remaining[size:]

		if len(remaining) == 0 {
			break
		}

		r, size = nextValidRune(remaining, r)

		switch {
		case unicode.IsUpper(r):
			prevSize := size
			offset += size
			remaining = remaining[size:]

			for len(remaining) > 0 {
				r, size = nextValidRune(remaining, r)

				if !unicode.IsUpper(r) {
					break
				}

				prevSize = size
				offset += size
				remaining = remaining[size:]
			}

			// it's a bit complex when dealing with a case like "HTTPStatus".
			// it's expected to be splitted into "HTTP" and "Status".
			// Therefore "S" should be in remaining instead of word.
			if len(remaining) > 0 && isAlphabet(r) {
				offset -= prevSize
				remaining = str[offset:]
			}

		case isAlphabet(r):
			offset += size
			remaining = remaining[size:]

			for len(remaining) > 0 {
				r, size = nextValidRune(remaining, r)

				if !isAlphabet(r) || unicode.IsUpper(r) {
					break
				}

				offset += size
				remaining = remaining[size:]
			}
		}

	case isAlphabet(r):
		wt = alphabetWord
		remaining = remaining[size:]

		for len(remaining) > 0 {
			r, size = nextValidRune(remaining, r)
			if !isAlphabet(r) || unicode.IsUpper(r) {
				break
			}

			offset += size
			remaining = remaining[size:]
		}

	case unicode.IsNumber(r):
		wt = numberWord
		remaining = remaining[size:]

		for len(remaining) > 0 {
			r, size = nextValidRune(remaining, r)

			if !unicode.IsNumber(r) {
				break
			}

			offset += size
			remaining = remaining[size:]
		}

	default:
		wt = otherWord
		remaining = remaining[size:]

		for len(remaining) > 0 {
			r, size = nextValidRune(remaining, r)

			if size == 0 || isConnector(r) || isAlphabet(r) || unicode.IsNumber(r) || unicode.IsPunct(r) {
				break
			}

			offset += size
			remaining = remaining[size:]
		}
	}

	word = str[:offset]
	return
}

func nextValidRune(str string, prev rune) (r rune, size int) {
	var sz int

	for len(str) > 0 {
		r, sz = utf8.DecodeRuneInString(str)
		size += sz

		if r != utf8.RuneError {
			return
		}

		str = str[sz:]
	}

	r = prev
	return
}

func camelCaseToLowerCase(str string, connector rune) string {
	if len(str) == 0 {
		return ""
	}

	wt, word, remaining := nextWord(str)
	buf := &strings.Builder{}

	for len(remaining) > 0 {
		if wt != connectorWord {
			toLower(buf, wt, word, connector)
		}

		prev := wt
		last := word
		wt, word, remaining = nextWord(remaining)

		switch prev {
		case numberWord:
			for wt == alphabetWord || wt == numberWord {
				toLower(buf, wt, word, connector)
				wt, word, remaining = nextWord(remaining)
			}

			if wt != invalidWord && wt != punctWord && wt != connectorWord {
				buf.WriteRune(connector)
			}

		case connectorWord:
			toLower(buf, prev, last, connector)

		case punctWord:
			// nothing.

		default:
			if wt != numberWord {
				if wt != connectorWord && wt != punctWord {
					buf.WriteRune(connector)
				}

				break
			}

			if len(remaining) == 0 {
				break
			}

			last := word
			wt, word, remaining = nextWord(remaining)

			// consider number as a part of previous word.
			// e.g. "Bld4Floor" => "bld4_floor"
			if wt != alphabetWord {
				toLower(buf, numberWord, last, connector)

				if wt != connectorWord && wt != punctWord {
					buf.WriteRune(connector)
				}

				break
			}

			// if there are some lower case letters following a number,
			// add connector before the number.
			// e.g. "HTTP2xx" => "http_2xx"
			buf.WriteRune(connector)
			toLower(buf, numberWord, last, connector)

			for wt == alphabetWord || wt == numberWord {
				toLower(buf, wt, word, connector)
				wt, word, remaining = nextWord(remaining)
			}

			if wt != invalidWord && wt != connectorWord && wt != punctWord {
				buf.WriteRune(connector)
			}
		}
	}

	toLower(buf, wt, word, connector)
	return buf.String()
}
