package strutils

import (
	"math/rand"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Shuffle randomizes runes in a string and returns the result.
// It uses default random source in `math/rand`.
func Shuffle(str string) string {
	if str == "" {
		return str
	}

	runes := []rune(str)
	index := 0

	for i := len(runes) - 1; i > 0; i-- {
		index = rand.Intn(i + 1)

		if i != index {
			runes[i], runes[index] = runes[index], runes[i]
		}
	}

	return string(runes)
}

// ToCamelCase is to convert words separated by space, underscore and hyphen to camel case.
//
// Some samples.
//
//	"some_words"      => "SomeWords"
//	"http_server"     => "HttpServer"
//	"no_https"        => "NoHttps"
//	"_complex__case_" => "_Complex_Case_"
//	"some words"      => "SomeWords"
func ToCamelCase(str string) string {
	if len(str) == 0 {
		return ""
	}

	buf := &strings.Builder{}
	var r0, r1 rune
	var size int

	// leading connector will appear in output.
	for len(str) > 0 {
		r0, size = utf8.DecodeRuneInString(str)
		str = str[size:]

		if !isConnector(r0) {
			r0 = unicode.ToUpper(r0)
			break
		}

		buf.WriteRune(r0)
	}

	if len(str) == 0 {
		// A special case for a string contains only 1 rune.
		if size != 0 {
			buf.WriteRune(r0)
		}

		return buf.String()
	}

	for len(str) > 0 {
		r1 = r0
		r0, size = utf8.DecodeRuneInString(str)
		str = str[size:]

		if isConnector(r0) && isConnector(r1) {
			buf.WriteRune(r1)
			continue
		}

		if isConnector(r1) {
			r0 = unicode.ToUpper(r0)
		} else {
			buf.WriteRune(r1)
		}
	}

	buf.WriteRune(r0)
	return buf.String()
}

// ToSnakeCase can convert all upper case characters in a string to
// snake case format.
//
// Some samples.
//
//	"FirstName"    => "first_name"
//	"HTTPServer"   => "http_server"
//	"NoHTTPS"      => "no_https"
//	"GO_PATH"      => "go_path"
//	"GO PATH"      => "go_path"  // space is converted to underscore.
//	"GO-PATH"      => "go_path"  // hyphen is converted to underscore.
//	"http2xx"      => "http_2xx" // insert an underscore before a number and after an alphabet.
//	"HTTP20xOK"    => "http_20x_ok"
//	"Duration2m3s" => "duration_2m3s"
//	"Bld4Floor3rd" => "bld4_floor_3rd"
func ToSnakeCase(str string) string {
	return camelCaseToLowerCase(str, '_')
}

// ToKebabCase can convert all upper case characters in a string to
// kebab case format.
//
// Some samples.
//
//	"FirstName"    => "first-name"
//	"HTTPServer"   => "http-server"
//	"NoHTTPS"      => "no-https"
//	"GO_PATH"      => "go-path"
//	"GO PATH"      => "go-path"  // space is converted to '-'.
//	"GO-PATH"      => "go-path"  // hyphen is converted to '-'.
//	"http2xx"      => "http-2xx" // insert an underscore before a number and after an alphabet.
//	"HTTP20xOK"    => "http-20x-ok"
//	"Duration2m3s" => "duration-2m3s"
//	"Bld4Floor3rd" => "bld4-floor-3rd"
func ToKebabCase(str string) string {
	return camelCaseToLowerCase(str, '-')
}
