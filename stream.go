package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"math"
	"os"
)

type Clip struct {
	Format Format
	Source Source
	Region Region
	Output Output
}

func (c Clip) Valid() error {
	if err := c.Format.Valid(); err != nil {
		return fmt.Errorf("mashu.Clip.Valid: bad format: %w", err)
	}
	if err := c.Source.Valid(); err != nil {
		return fmt.Errorf("mashu.Clip.Valid: bad source: %w", err)
	}
	if err := c.Region.Valid(); err != nil {
		return fmt.Errorf("mashu.Clip.Valid: bad region: %w", err)
	}
	if err := c.Output.ValidVideo(c.Format); err != nil {
		return fmt.Errorf("mashu.Clip.Valid: bad output: %w", err)
	}

	return nil
}

type Stack struct {
	Format Format
	Output Output
	Input  []Input
}

func (s Stack) Valid() error {
	if err := s.Format.Valid(); err != nil {
		return fmt.Errorf("mashu.Stack.Valid: bad format: %w", err)
	}
	if err := s.Output.ValidVideo(s.Format); err != nil {
		return fmt.Errorf("mashu.Stack.Valid: bad output: %w", err)
	}
	if len(s.Input) < 2 || len(s.Input) != int(math.Sqrt(float64(len(s.Input))))*int(math.Sqrt(float64(len(s.Input)))) {
		return fmt.Errorf("mashu.Stack.Valid: input count must be a perfect square greater than one (%d)", len(s.Input))
	}
	for _, input := range s.Input {
		if err := input.Valid(); err != nil {
			return fmt.Errorf("mashu.Stack.Valid: invalid input ('%s'): %w", input, err)
		}
	}

	return nil
}

type Blend struct {
	Name        string
	Format      Format
	Output      Output
	Attachments Attachments
}

func (b Blend) Valid() error {
	if err := b.Format.Valid(); err != nil {
		return fmt.Errorf("mashu.Blend.Valid: bad format: %w", err)
	}
	if err := b.Output.ValidVideo(b.Format); err != nil {
		return fmt.Errorf("mashu.Blend.Valid: bad output: %w", err)
	}
	if err := b.Attachments.Valid(); err != nil {
		return fmt.Errorf("mashu.Blend.Valid: bad attachments: %w", err)
	}

	return nil
}

type Concat struct {
	Format Format
	Output Output
	Input  []Input
}

func (c Concat) Valid() error {
	if err := c.Format.Valid(); err != nil {
		return fmt.Errorf("mashu.Concat.Valid: bad format: %w", err)
	}
	if err := c.Output.ValidVideo(c.Format); err != nil {
		return fmt.Errorf("mashu.Concat.Valid: bad output: %w", err)
	}
	if len(c.Input) < 2 {
		return fmt.Errorf("mashu.Concat.Valid: input count must be greater than one (%d)", len(c.Input))
	}
	for _, input := range c.Input {
		if err := input.Valid(); err != nil {
			return fmt.Errorf("mashu.Concat.Valid: invalid input ('%s'): %w", input, err)
		}
	}

	return nil
}

func renderStreamFromFile(ctx context.Context, path string) (err error) {
	var f *os.File
	if f, err = os.Open(path); err != nil {
		return
	}
	defer f.Close()

	d := json.NewDecoder(f)
	for d.More() {
		var s string
		if err = d.Decode(&s); err != nil {
			return
		}

		switch s {
		case "clip":
			var clip Clip
			if err = d.Decode(&clip); err != nil {
				return
			}
			if err = clip.Valid(); err != nil {
				if !errors.Is(err, fs.ErrExist) {
					log.Printf("mashu.renderStreamFromFile: error validating clip (key: %#v): %v", clip.Source.Key, err)
				}
				err = nil
				continue
			}
			if err = renderClip(ctx, clip.Format, clip.Source, clip.Region, clip.Output); err != nil {
				log.Printf("mashu.renderStreamFromFile: error rendering clip (key: %#v): %v", clip.Source.Key, err)
				err = nil
				continue
			}
		case "stack":
			var stack Stack
			if err = d.Decode(&stack); err != nil {
				return
			}
			if err = stack.Valid(); err != nil {
				if !errors.Is(err, fs.ErrExist) {
					log.Printf("mashu.renderStreamFromFile: error validating stack (output: %#v): %v", stack.Output, err)
				}
				err = nil
				continue
			}
			if err = renderStack(ctx, stack.Format, stack.Output, stack.Input); err != nil {
				log.Printf("mashu.renderSreamFromFile: error rendering stack (output: %#v): %v", stack.Output, err)
				err = nil
				continue
			}
		case "blend":
			var blend Blend
			if err = d.Decode(&blend); err != nil {
				return
			}
			if err = blend.Valid(); err != nil {
				if !errors.Is(err, fs.ErrExist) {
					log.Printf("mashu.renderStreamFromFile: error validating blend (output: %#v): %v", blend.Output, err)
				}
				err = nil
				continue
			}
			if err = renderBlend(ctx, blend.Name, blend.Format, blend.Output, blend.Attachments); err != nil {
				log.Printf("mashu.renderSreamFromFile: error rendering blend (output: %#v): %v", blend.Output, err)
				err = nil
				continue
			}
		case "concat":
			var concat Concat
			if err = d.Decode(&concat); err != nil {
				return
			}
			if err = concat.Valid(); err != nil {
				if !errors.Is(err, fs.ErrExist) {
					log.Printf("mashu.renderStreamFromFile: error validating concat (output: %#v): %v", concat.Output, err)
				}
				err = nil
				continue
			}
			if err = renderConcat(ctx, concat.Output, concat.Input); err != nil {
				log.Printf("mashu.renderSreamFromFile: error rendering concat (output: %#v): %v", concat.Output, err)
				err = nil
				continue
			}
		default:
			return fmt.Errorf("mashu.renderStreamFromFile: unknown identifier '%s'", s)
		}

	}

	return
}
