package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

// TODO random generation should microsecond align; start and duration

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
	if err := c.Output.Valid(c.Format); err != nil {
		return fmt.Errorf("mashu.Clip.Valid: bad output: %w", err)
	}

	return nil
}

func renderClipsFromFile(path string) (err error) {
	var f *os.File
	if f, err = os.Open(path); err != nil {
		return
	}
	defer f.Close()

	d := json.NewDecoder(f)
	for d.More() {
		var clip Clip
		if err = d.Decode(&clip); err != nil {
			return
		}
		if err = clip.Valid(); err != nil {
			log.Printf("mashu.renderClipsFromFile: error validating clip (key: %#v): %v", clip.Source.Key, err)
			err = nil
			continue
		}
		if err = renderClip(clip.Format, clip.Source, clip.Region, clip.Output); err != nil {
			log.Printf("mashu.renderClipsFromFile: error rendering clip (key: %#v): %v", clip.Source.Key, err)
			err = nil
			continue
		}
	}

	return
}

func renderClip(f Format, s Source, r Region, o Output) (err error) {
	args := make([]string, 0)
	filters := make([]string, 0)

	inputLink, videoLink, audioLink := 0, 0, 0

	// add inputs
	if s.Video != nil {
		args = append(args,
			"-ss", fmt.Sprintf("%dus", r.Start.Microseconds()),
			"-to", fmt.Sprintf("%dus", r.End.Microseconds()),
			"-i", s.Video.Path)
		filters = append(filters, fmt.Sprintf(
			"[%d:v:%d]null[v%d]", inputLink, s.Video.Track, videoLink))
		inputLink += 1
	}

	if s.Audio != nil {
		args = append(args,
			"-ss", fmt.Sprintf("%dus", r.Start.Microseconds()),
			"-to", fmt.Sprintf("%dus", r.End.Microseconds()),
			"-i", s.Audio.Path)
		filters = append(filters, fmt.Sprintf(
			"[%d:a:%d]loudnorm,aresample=%d[a%d]", inputLink, s.Audio.Track, f.SampleRate, audioLink))
		if s.Video == nil {
			filters = append(filters, fmt.Sprintf(
				"[%d:a:%d]avectorscope=size=%dx%d:rate=%d[v%d]", inputLink, s.Audio.Track, f.Width, f.Height, f.FrameRate, videoLink))
		}
		inputLink += 1
	} else {
		filters = append(filters, fmt.Sprintf(
			"anullsrc=sample_rate=%d:duration=%dus[a%d]", f.SampleRate, r.Duration().Microseconds(), audioLink))
	}

	if s.Video == nil && s.Audio == nil {
		filters = append(filters, fmt.Sprintf(
			"nullsrc=size=%dx%d:rate=%d:duration=%dus,geq=random(1)*255:128:128[v%d]",
			f.Width, f.Height, f.FrameRate, r.Duration().Microseconds(), videoLink))
	}

	if s.Video != nil && s.Video.Filter != nil {
		filters = append(filters, fmt.Sprintf(
			"[v%d]%s[v%d]", videoLink, *s.Video.Filter, videoLink+1))
		videoLink += 1
	}
	if s.Audio != nil && s.Audio.Filter != nil {
		filters = append(filters, fmt.Sprintf(
			"[a%d]%s[a%d]", audioLink, *s.Audio.Filter, audioLink+1))
		audioLink += 1
	}

	// extract subs
	var subtitleFile *os.File
	if s.Subtitle != nil {
		if subtitleFile, err = os.CreateTemp(os.TempDir(), "mashu-clip-*.ass"); err != nil {
			return
		}
		subtitleFile.Close()
		defer os.Remove(subtitleFile.Name())

		if err = ffmpeg(context.TODO(), "-y",
			"-itsoffset", fmt.Sprintf("-%dus", r.Start.Microseconds()),
			"-i", s.Subtitle.Path,
			"-map", fmt.Sprintf("0:s:%d", s.Subtitle.Track),
			subtitleFile.Name()); err != nil {
			log.Printf("mashu.clip: unable to extract subtitles: %w", err)
			err = nil
			os.Remove(subtitleFile.Name())
			subtitleFile = nil
		} else {
			filters = append(filters, fmt.Sprintf(
				"[v%d]subtitles=filename='%s'[v%d]", videoLink, subtitleFile.Name(), videoLink+1))
			videoLink += 1
		}
	}

	// scale/transcribe
	if s.Video != nil {
		filters = append(filters, fmt.Sprintf(
			"[v%d]scale=width=%d:height=%d:force_original_aspect_ratio=decrease,pad=width=%d:height=%d:x=(ow-iw)/2:y=(oh-ih)/2,setsar=1:1[v%d]",
			videoLink, f.Width, f.Height, f.Width, f.Height, videoLink+1))
		videoLink += 1
	}

	// add stamp
	stamp := f.Stamp
	if s.Stamp != nil {
		stamp = *s.Stamp
	}
	filters = append(filters, fmt.Sprintf(
		"[v%d]drawtext=borderw=2:fontcolor=%s:fontfile=%s:fontsize=%d:text='%s':x=w-tw-8:y=h-th-8[v%d]",
		videoLink, stamp.Color, stamp.Font, stamp.Size, s.Key, videoLink+1))
	videoLink += 1

	// apply filters
	args = append(args, "-filter_complex", strings.Join(filters, ";"))

	// configure output
	switch f.VideoCodec {
	case "H264":
		args = append(args, "-codec:v", "libx264", "-x264-params", fmt.Sprintf("log-level=%s", loglevel))
	case "H265":
		args = append(args, "-codec:v", "libx265", "-x265-params", fmt.Sprintf("log-level=%s", loglevel))
	case "VP9":
		args = append(args, "-codec:v", "vp9")
	}

	args = append(args,
		"-r", fmt.Sprintf("%d", f.FrameRate),
		"-codec:a", strings.ToLower(f.AudioCodec),
		"-ac", "2",
		"-map", fmt.Sprintf("[v%d]", videoLink),
		"-map", fmt.Sprintf("[a%d]", audioLink),
		"-map_metadata", "-1",
		"-map_chapters", "-1",
		string(o))

	err = ffmpeg(context.TODO(), args...)
	return
}
