package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func mpv(path ...string) {
	exec.Command("mpv", append([]string{
		"--player-operation-mode=pseudo-gui",
		"--loop=inf",
		"--loop-playlist=inf"},
		path...)...).Start()
}

func locateMacros() (macros string, err error) {
	if err = executableError; err != nil {
		return
	}

	var target string
	if target, err = filepath.EvalSymlinks(executable); err != nil {
		return
	}

	macros = filepath.Join(filepath.Dir(target), "macros.vim")

	return
}

func vim(path ...string) (err error) {
	var macroFile string
	if macroFile, err = locateMacros(); err != nil {
		return
	}

	cmd := exec.Command("vim", append([]string{
		"-S", macroFile, "--"}, path...)...)
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

func buildSources(paths ...string) (sources []Source, err error) {
	names := make([]string, 0)
	mpvPaths := make([]string, 0)

	for _, path := range paths {
		var s Source
		s.Key = path

		var m3u []string
		if m3u, err = getM3uEntries(path); err == nil {
			if len(m3u) == 0 {
				log.Printf("mashu.buildSources: skipping %s: empty m3u", path)
				continue
			}
			if err = addTracksFromVideo(m3u[0], &s); err != nil {
				log.Printf("mashu.buildSources: skipping %s: %v", path, err)
				err = nil
				continue
			}
		}
		if errors.Is(err, ErrNoM3UHeader) {
			err = nil
			s.Video = &Track{}
			s.Audio = &Track{}
			s.Subtitle = &Track{}
			s.Regions = []TaggedRegion{TaggedRegion{}}
		} else if err != nil {
			log.Printf("mashu.buildSources: skipping %s: %v", path, err)
			err = nil
			continue
		}

		var f *os.File
		if f, err = os.CreateTemp("", "mashu-build-source-*.json"); err != nil {
			return
		}
		defer f.Close()
		defer os.Remove(f.Name())

		names = append(names, f.Name())
		mpvPaths = append(mpvPaths, path)
		fmt.Fprintln(f, path)
		encoder := json.NewEncoder(f)
		encoder.SetIndent("", "\t")
		if err = encoder.Encode(s); err != nil {
			return
		}
		if err = f.Close(); err != nil {
			return
		}
	}

	if len(names) == 0 {
		return
	}

	mpv(mpvPaths...)
	if err = vim(names...); err != nil {
		return
	}

	for _, name := range names {
		var s Source
		if err = decodeJsonFromFile(name, &s); err != nil {
			log.Printf("mashu.buildSources: unable to import %s: %v", name, err)
			continue
		}
		sources = append(sources, s)
	}

	for _, s := range sources {
		if err = s.Valid(); err != nil {
			return
		}
	}

	return
}
