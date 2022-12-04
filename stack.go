package main

import (
	"context"
	"fmt"
	"math"
	"strings"
)

// TODO represent input constraints; square number of inputs; as composite registration; also non-recursive
// composite interface?
// composite config?

func stackMatrix(n int) string {
	r := int(math.Sqrt(float64(n)))
	e := make([]string, 0)

	for h := 0; h < r; h++ {
		for w := 0; w < r; w++ {
			s := ""
			if w == 0 {
				s += "0"
			} else {
				for wi := 0; wi < w; wi++ {
					if wi != 0 {
						s += "+"
					}
					s += fmt.Sprintf("w%d", wi)
				}
			}
			s += "_"
			if h == 0 {
				s += "0"
			} else {
				for hi := 0; hi < h; hi++ {
					if hi != 0 {
						s += "+"
					}
					s += fmt.Sprintf("h%d", hi)
				}
			}
			e = append(e, s)
		}
	}

	return strings.Join(e, "|")
}

// assumes format and output have already been validated
// assumes input count is a perfect square greater than one
func renderStack(ctx context.Context, f Format, o Output, d Duration, input []Input) error {
	l := len(input)
	args := make([]string, 4*l+9)
	for i, p := range input {
		args[4*i], args[4*i+1] = "-stream_loop", "-1"
		args[4*i+2], args[4*i+3] = "-i", string(p)
	}
	rest := 4 * l
	args[rest], args[rest+1] = "-filter_complex", fmt.Sprintf(
		"xstack=inputs=%d:layout=%s,scale=%dx%d",
		l, stackMatrix(l), f.Width, f.Height)
	args[rest+2], args[rest+3] = "-filter_complex", fmt.Sprintf(
		"amix=inputs=%d,loudnorm", l)
	args[rest+4], args[rest+5] = "-ac", "2"
	args[rest+6], args[rest+7] = "-to", fmt.Sprintf("%dus", d.Microseconds())
	args[rest+8] = string(o)
	return ffmpeg(ctx, args...)
}
