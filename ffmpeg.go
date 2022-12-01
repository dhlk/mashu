package main

import (
	"context"
	"os"
	"os/exec"
)

var loglevel = func() string {
	l, d := os.LookupEnv("LOGLEVEL")
	if !d {
		return "error"
	}
	return l
}()

func ffmpeg(ctx context.Context, arg ...string) error {
	args := []string{"-loglevel", loglevel,
		"-analyzeduration", "2147483647", "-probesize", "2147483647"}
	args = append(args, arg...)
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
