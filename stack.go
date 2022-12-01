package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
)

// TODO represent input constraints; square number of inputs; as composite registration; also non-recursive
// composite interface?
// composite config?

type Stack struct {
	Format Format
	Output Output
	Input  []string
}

func (s Stack) Valid() error {
	if err := s.Format.Valid(); err != nil {
		return fmt.Errorf("mashu.Stack.Valid: bad format: %w", err)
	}
	if err := s.Output.Valid(s.Format); err != nil {
		return fmt.Errorf("mashu.Stack.Valid: bad output: %w", err)
	}
	if len(s.Input) < 2 || len(s.Input) != int(math.Sqrt(float64(len(s.Input))))*int(math.Sqrt(float64(len(s.Input)))) {
		return fmt.Errorf("mashu.Stack.Valid: input count must be a perfect square greater than one (%d)", len(s.Input))
	}
	for _, input := range s.Input {
		if fi, err := os.Stat(input); err != nil || fi.IsDir() {
			return fmt.Errorf("mashu.Stack.Valid: cannot stat input or is a directory ('%s'): %w", input, err)
		}
	}

	return nil
}

func renderStacksFromFile(path string) (err error) {
	var f *os.File
	if f, err = os.Open(path); err != nil {
		return
	}
	defer f.Close()

	d := json.NewDecoder(f)
	for d.More() {
		var stack Stack
		if err = d.Decode(&stack); err != nil {
			return
		}
		if err = stack.Valid(); err != nil {
			log.Printf("mashu.renderStacksFromFile: error validating stack (output: %#v): %v", stack.Output, err)
			err = nil
			continue
		}
		if err = renderStack(stack.Format, stack.Output, stack.Input); err != nil {
			log.Printf("mashu.renderStacksFromFile: error rendering stack (output: %#v): %v", stack.Output, err)
			err = nil
			continue
		}
	}

	return
}

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
func renderStack(f Format, o Output, input []string) error {
	l := len(input)
	args := make([]string, 2*l+7)
	for i, p := range input {
		args[2*i] = "-i"
		args[2*i+1] = p
	}
	rest := 2 * l
	args[rest], args[rest+1] = "-filter_complex", fmt.Sprintf(
		"xstack=inputs=%d:layout=%s,scale=%dx%d",
		l, stackMatrix(l), f.Width, f.Height)
	args[rest+2], args[rest+3] = "-filter_complex", fmt.Sprintf(
		"amix=inputs=%d,loudnorm", l)
	args[rest+4], args[rest+5], args[rest+6] = "-ac", "2", string(o)
	return ffmpeg(context.TODO(), args...)
}
