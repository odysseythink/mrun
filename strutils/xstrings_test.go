package strutils

import (
	"fmt"
	"testing"
	"unicode/utf8"
)

func TestNextValidRune(t *testing.T) {
	r, sz := nextValidRune("-aaa您这边可以帮我重置一下那个ed的那个密码吗", utf8.RuneError)
	fmt.Println("-------", r, "---------", sz)

	for _, v := range []string{"235aaa", "- _-aaa", ",.，。/*?aaa", "HaaTTPServer", "FirstName", "NoHTTPS", "GO_PATH", "GO PATH", "GO-PATH", "http2xx", "HTTP20xOK", "Duration2m3s", "Bld4Floor3rd"} {
		wt, word, remaining := nextWord(v)
		fmt.Println("-------", wt, "---------", word, "-----", remaining)
	}
}
