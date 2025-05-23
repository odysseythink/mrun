package mrun

import (
	"fmt"
	"strconv"
	"testing"
)

func F1(val string) string {
	fmt.Println("------hello world=", val)
	return ""
}

func BenchmarkSignalSend(b *testing.B) {
	msgSig, err := NewSignal("message_was_created", func(string) string { return "" }, NewSignalConcurrencyOption(100))
	if err != nil {
		fmt.Println("new signal failed:", err)
		return
	}
	err = Connect(msgSig, F1)
	if err != nil {
		fmt.Println("connect signal failed:", err)
		return
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msgSig.Emit("message", strconv.Itoa(i))
	}
}
