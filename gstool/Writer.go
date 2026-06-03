package gstool

import "strings"

type WriterString struct {
	Str *strings.Builder
}

func (h *WriterString) Write(p []byte) (n int, err error) {
	h.Str.Write(p)
	return len(p), nil
}

func NewWriterString() *WriterString {
	return &WriterString{
		Str: new(strings.Builder),
	}
}
