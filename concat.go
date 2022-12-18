package main

import (
	"context"
	"fmt"
	"strings"
)

// assumes inputs are validated and input files are the same format
func renderConcat(ctx context.Context, f Format, o Output, i []Input) error {
	l := len(i)
	args := make([]string, 2*l)
	for n, p := range i {
		args[2*n] = "-i"
		args[2*n+1] = string(p)
	}

	args = append(args,
		"-filter_complex", fmt.Sprintf(
			"concat=n=%d:v=1:a=1", l))

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
		"-map_metadata", "-1",
		"-map_chapters", "-1",
		string(o))

	return ffmpeg(ctx, args...)
}
