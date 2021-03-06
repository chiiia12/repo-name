package main

import (
	"io"
)

type limitReader struct {
	r       io.Reader
	limit   int64
	hasRead bool
}

//EOFが呼ばれるまで呼び続けるっぽい
func (lr *limitReader) Read(p []byte) (n int, err error) {
	if lr.hasRead {
		//nが0で返さないとHello,      空白が入る
		return 0, io.EOF
	}
	n, err = lr.r.Read(p[:lr.limit])
	if err != nil {
		return n, nil
	}
	lr.hasRead = true
	return int(lr.limit), nil
}

func LimitReader(r io.Reader, n int64) io.Reader {
	return &limitReader{r, n, false}
}
