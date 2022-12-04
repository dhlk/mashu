package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func mpv(path string) {
	exec.Command("mpv",
		"--player-operation-mode=pseudo-gui",
		"--loop=inf",
		"--loop-playlist=inf",
		path).Start()
}

func vim(path string) error {
	cmd := exec.Command("vim", path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type probeResult struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

func (r probeResult) tracksOf(codecType string) (c uint) {
	for _, t := range r.Streams {
		if t.CodecType == codecType {
			c++
		}
	}
	return
}

func (r probeResult) VideoTracks() uint {
	return r.tracksOf("video")
}

func (r probeResult) AudioTracks() uint {
	return r.tracksOf("audio")
}

func (r probeResult) SubtitleTracks() uint {
	return r.tracksOf("subtitle")
}

func (r probeResult) Duration() (d Duration, err error) {
	err = (&d).UnmarshalJSON([]byte(fmt.Sprintf("\"%ss\"", r.Format.Duration)))
	return
}

func ffprobe(path string) (r probeResult, err error) {
	var buffer bytes.Buffer
	cmd := exec.CommandContext(context.TODO(),
		"ffprobe",
		"-v", "error",
		"-show_entries", "stream=codec_type:format=duration",
		"-of", "json",
		path)
	cmd.Stdout = &buffer
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return
	}

	err = json.Unmarshal(buffer.Bytes(), &r)
	return
}

func addTracksFromVideo(path string, s *Source) (err error) {
	var r probeResult
	if r, err = ffprobe(path); err != nil {
		return err
	}

	var d Duration
	if d, err = r.Duration(); err != nil {
		return
	}
	s.Regions = append(s.Regions, TaggedRegion{Region: Region{End: d}})

	v, a, t := r.VideoTracks(), r.AudioTracks(), r.SubtitleTracks()
	if v > 0 {
		s.Video = &Track{Path: Input(path), Track: v - 1}
	}
	if a > 0 {
		s.Audio = &Track{Path: Input(path), Track: a - 1}
	}
	if t > 0 {
		s.Subtitle = &Track{Path: Input(path), Track: t - 1}
	}

	return
}

func buildSource(path string) (s Source, err error) {
	mpv(path)

	s.Key = path

	var m3u []string
	if m3u, err = getM3uEntries(path); err == nil {
		if len(m3u) == 0 {
			err = fmt.Errorf("empty m3u")
			return
		}
		if err = addTracksFromVideo(m3u[0], &s); err != nil {
			return
		}
	}
	if errors.Is(err, ErrNoM3UHeader) {
		err = nil
		s.Video = &Track{}
		s.Audio = &Track{}
		s.Subtitle = &Track{}
	} else if err != nil {
		return
	}

	var f *os.File
	if f, err = os.CreateTemp("", "mashu-build-source-*.json"); err != nil {
		return
	}
	defer f.Close()
	defer os.Remove(f.Name())

	fmt.Fprintln(f, path)
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "\t")
	if err = encoder.Encode(s); err != nil {
		return
	}
	if err = f.Close(); err != nil {
		return
	}

	if err = vim(f.Name()); err != nil {
		return
	}

	if err = decodeJsonFromFile(f.Name(), &s); err != nil {
		return
	}

	if err = s.Valid(); err != nil {
		return
	}

	return
}
