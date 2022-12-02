package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Track struct {
	Path   string
	Track  uint
	Filter *string
}

func (t Track) Valid() error {
	if _, err := os.Stat(t.Path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("mashu.Track.Valid: path must exist (%s): %w", t.Path, err)
		}
		return fmt.Errorf("mashu.Track.Valid: unable to stat track path (%s): %w", t.Path, err)
	}
	// TODO validate the track number is valid

	if t.Filter != nil {
		if len(*t.Filter) == 0 {
			return fmt.Errorf("mashu.Track.Valid: non-nil filter must not be empty")
		}
	}

	return nil
}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	if d.Duration, err = time.ParseDuration(s); err != nil {
		return err
	}

	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

type Region struct {
	Start, End Duration
}

func (r Region) Duration() time.Duration {
	return r.End.Duration - r.Start.Duration
}

func (r Region) Valid() error {
	if r.Start == r.End || r.Start.Duration > r.End.Duration {
		return fmt.Errorf("mashu.Region.Valid: start and end must be different and end must be after start (not %v to %v)", r.Start, r.End)
	}

	return nil
}

type TaggedRegion struct {
	Region
	Tags []string
}

func (r TaggedRegion) Duration() time.Duration {
	return r.Region.Duration()
}

func (r TaggedRegion) Valid() error {
	return r.Region.Valid()
}

type Stamp struct {
	Color string
	Font  string
	Size  uint
}

func (s Stamp) Valid() error {
	if len(s.Color) == 0 {
		return fmt.Errorf("mashu.Format.Valid: stamp color must not be empty")
	}
	if _, err := os.Stat(s.Font); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("mashu.Format.Valid: stamp font must exist (%s): %w", s.Font, err)
		}
		return fmt.Errorf("mashu.Format.Valid: unable to stat stamp font (%s): %w", s.Font, err)
	}

	return nil
}

type Source struct {
	Key      string
	Video    *Track
	Audio    *Track
	Subtitle *Track
	Regions  []TaggedRegion
	Stamp    *Stamp
}

func (s Source) Valid() error {
	if s.Video != nil {
		if err := s.Video.Valid(); err != nil {
			return fmt.Errorf("mashu.Source.Valid: invalid video specified: %w", err)
		}
	}
	if s.Audio != nil {
		if err := s.Audio.Valid(); err != nil {
			return fmt.Errorf("mashu.Source.Valid: invalid audio specified: %w", err)
		}
	}
	if s.Subtitle != nil {
		if err := s.Subtitle.Valid(); err != nil {
			return fmt.Errorf("mashu.Source.Valid: invalid subtitle specified: %w", err)
		}
	}
	if len(s.Regions) == 0 {
		return fmt.Errorf("mashu.Source.Valid: must include a region list")
	}
	for _, r := range s.Regions {
		if err := r.Valid(); err != nil {
			return fmt.Errorf("mashu.Source.Valid: bad region: %w", err)
		}
	}
	if s.Stamp != nil {
		if err := s.Stamp.Valid(); err != nil {
			return fmt.Errorf("mashu.Source.Valid: invalid stamp: %w", err)
		}
	}

	return nil
}

type Format struct {
	Format     string
	Samples    uint
	Quality    string
	Speed      string
	VideoCodec string
	AudioCodec string
	FrameRate  uint
	SampleRate uint
	BitRate    uint
	Gopsize    uint
	Width      uint
	Height     uint
	Stamp      Stamp
}

func (f Format) Valid() error {
	if f.Format != "MKV" {
		return fmt.Errorf("mashu.Format.Valid: format must be MKV (not '%s')", f.Format)
	}
	if f.VideoCodec != "H264" {
		return fmt.Errorf("mashu.Format.Valid: video codec must be H264 (not '%s')", f.VideoCodec)
	}
	if f.AudioCodec != "AAC" && f.AudioCodec != "FLAC" {
		return fmt.Errorf("mashu.Format.Valid: audio codec must be AAC or FLAC (not '%s')", f.AudioCodec)
	}
	if err := f.Stamp.Valid(); err != nil {
		return fmt.Errorf("mashu.Format.Valid: invalid stamp: %w", err)
	}
	if f.FrameRate == 0 {
		return fmt.Errorf("mashu.Format.Valid: framerate must be non-zero")
	}
	if f.SampleRate == 0 {
		return fmt.Errorf("mashu.Format.Valid: samplerate must be non-zero")
	}
	if f.BitRate == 0 {
		return fmt.Errorf("mashu.Format.Valid: bitrate must be non-zero")
	}
	if f.Gopsize == 0 {
		return fmt.Errorf("mashu.Format.Valid: gopsize must be non-zero")
	}
	if f.Width == 0 {
		return fmt.Errorf("mashu.Format.Valid: width must be non-zero")
	}
	if f.Height == 0 {
		return fmt.Errorf("mashu.Format.Valid: height must be non-zero")
	}
	// TODO blender properties (f.Samples, f.Quality, f.Speed)

	return nil
}

var DefaultFormat = Format{
	Format:     "MKV",
	Samples:    8,
	Quality:    "MEDIUM",
	Speed:      "GOOD",
	VideoCodec: "H264",
	AudioCodec: "AAC",
	FrameRate:  30,
	SampleRate: 48000,
	BitRate:    192,
	Gopsize:    18,
	Width:      1920,
	Height:     1080,
	Stamp: Stamp{
		Color: "'Snow'",
		Font:  "/usr/share/fonts/noto/NotoSansMono-Regular.ttf",
		Size:  64,
	},
}

type Output string

func (o Output) Valid(f Format) error {
	if len(o) == 0 {
		return fmt.Errorf("mashu.Output.Valid: output path must be non-empty")
	}
	if filepath.Ext(string(o)) != "."+strings.ToLower(f.Format) {
		return fmt.Errorf("mashu.Output.Valid: output extension ('%s') much mach format ('%s')", filepath.Ext(string(o)), f.Format)
	}
	if _, err := os.Stat(string(o)); err == nil || !os.IsNotExist(err) {
		if err == nil {
			return fmt.Errorf("mashu.Output.Valid: path must not exist (%s): %w", o, os.ErrExist)
		}
		return fmt.Errorf("mashu.Output.Valid: unable to stat track path (%s): %w", o, err)
	}

	return nil
}

type Attachments map[string]string

func (a Attachments) Valid() error {
	for _, v := range a {
		if _, err := os.Stat(v); err != nil {
			return fmt.Errorf("mashu.Attachments.Valid: unable to stat attachment path (%s): %w", v, err)
		}
	}

	return nil
}
