package main

import (
	"context"
	"fmt"
)

// assumes inputs are validated and input files are the same format
func renderConcat(ctx context.Context, o Output, i []Input) error {
	l := len(i)
	args := make([]string, 2*l+3)
	for n, p := range i {
		args[2*n] = "-i"
		args[2*n+1] = string(p)
	}
	rest := 2 * l
	args[rest], args[rest+1], args[rest+2] =
		"-filter_complex",
		fmt.Sprintf("concat=n=%d:v=1:a=1", l),
		string(o)
	return ffmpeg(ctx, args...)
}
