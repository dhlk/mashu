package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var executable, executableError = os.Executable()

func locateBlend(blend string) (script, path string, err error) {
	if err = executableError; err != nil {
		return
	}

	var target string
	if target, err = filepath.EvalSymlinks(executable); err != nil {
		return
	}

	script = filepath.Join(filepath.Dir(target), "blend.py")
	if _, err = os.Stat(script); err != nil {
		err = fmt.Errorf("mashu.locateBlend: unable to locate python module ('%s'): %w", script, err)
		return
	}

	path = filepath.Join(filepath.Dir(target), blend+".blend")
	if _, err = os.Stat(path); err != nil {
		err = fmt.Errorf("mashu.locateBlend: unable to locate target blend ('%s'): %w", path, err)
		return
	}

	return
}

// assumes inputs have been validated
func renderBlend(ctx context.Context, blendName string, f Format, o Output, a Attachments) (err error) {
	var script, blend string
	if script, blend, err = locateBlend(blendName); err != nil {
		return
	}

	args := []string{"--background",
		"--factory-startup", blend,
		"--python", script,
		"--threads", "0",
		"--render-anim", "--",
		"-output", string(o),
		"-format", f.Format,
		"-samples", fmt.Sprintf("%d", f.Samples),
		"-quality", f.Quality,
		"-speed", f.Speed,
		"-vcodec", f.VideoCodec,
		"-acodec", f.AudioCodec,
		"-fps", fmt.Sprintf("%d", f.FrameRate),
		"-samplerate", fmt.Sprintf("%d", f.SampleRate),
		"-bitrate", fmt.Sprintf("%d", f.BitRate),
		"-gopsize", fmt.Sprintf("%d", f.Gopsize),
		"-width", fmt.Sprintf("%d", f.Width),
		"-height", fmt.Sprintf("%d", f.Height),
	}
	for k, v := range a {
		args = append(args, "-attach", k, v)
	}

	cmd := exec.CommandContext(ctx, "blender", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// TODO dropped -frame variant
//	blender --background --factory-startup "$blend" --python "$py" --threads 0 --render-format PNG --render-frame 1 -- "$@"