package main

import (
	"fmt"
	"io"
	"bytes"
)

const debug = true

func main() {
	var buf *bytes.Buffer
	if debug {
		buf = new(bytes.Buffer)
	}
	f(buf)
	if debug {
		fmt.Println("debug")
	}
	fmt.Errorf("HTML aarse is hogehoge.")
}
func f(out io.Writer) {
	if out != nil {
		out.Write([]byte("done!\n"))
	}
}
