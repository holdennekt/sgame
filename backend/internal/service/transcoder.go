package service

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

const (
	IMAGE_MIME_TYPE = "image/webp"
	AUDIO_MIME_TYPE = "audio/mp4"
	VIDEO_MIME_TYPE = "video/mp4"

	maxImageSide    = 1920
	webpQuality     = 80
	maxAudioBitRate = 128_000
	videoCRF        = 23
)

// only downscales, never upscales; -2 keeps the height even for libx264
var downscaleFilter = fmt.Sprintf("scale='min(%d,iw)':-2", maxImageSide)

// ContentError marks deterministic decode failures: retrying the same bytes fails the same way.
type ContentError struct {
	Err error
}

func (e ContentError) Error() string { return e.Err.Error() }
func (e ContentError) Unwrap() error { return e.Err }

func runFFmpeg(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg", append([]string{"-y", "-hide_banner", "-loglevel", "error"}, args...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("ffmpeg canceled: %w", ctx.Err())
		}
		msg := strings.TrimSpace(stderr.String())
		if len(msg) > 500 {
			msg = msg[:500]
		}
		return ContentError{Err: fmt.Errorf("ffmpeg failed: %v: %s", err, msg)}
	}
	return nil
}

// -loop 0 keeps real animations animated; others get one frame, so e.g. motion-photo JPEGs don't become garbled WebPs.
func transcodeImage(ctx context.Context, inPath, outPath string, animated bool) error {
	args := []string{"-i", inPath, "-vf", downscaleFilter}
	if !animated {
		args = append(args, "-frames:v", "1")
	}
	args = append(args,
		"-c:v", "libwebp", "-quality", fmt.Sprint(webpQuality), "-loop", "0",
		"-f", "webp", outPath,
	)
	return runFFmpeg(ctx, args...)
}

func audioBitRate(sourceBitRate int) int {
	if sourceBitRate > 0 && sourceBitRate < maxAudioBitRate {
		return sourceBitRate
	}
	return maxAudioBitRate
}

func transcodeAudio(ctx context.Context, inPath, outPath string, sourceBitRate int) error {
	return runFFmpeg(ctx,
		"-i", inPath,
		"-vn",
		"-c:a", "aac", "-b:a", fmt.Sprint(audioBitRate(sourceBitRate)),
		"-movflags", "+faststart",
		"-f", "mp4", outPath,
	)
}

func transcodeVideo(ctx context.Context, inPath, outPath string, sourceAudioBitRate int) error {
	return runFFmpeg(ctx,
		"-i", inPath,
		"-vf", downscaleFilter,
		"-c:v", "libx264", "-crf", fmt.Sprint(videoCRF), "-preset", "veryfast", "-pix_fmt", "yuv420p",
		"-c:a", "aac", "-b:a", fmt.Sprint(audioBitRate(sourceAudioBitRate)),
		"-movflags", "+faststart",
		"-f", "mp4", outPath,
	)
}

func transcodeAvatar(ctx context.Context, inPath, outPath string, size int) error {
	return runFFmpeg(ctx,
		"-i", inPath,
		"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d", size, size, size, size),
		"-frames:v", "1",
		"-c:v", "libwebp", "-quality", fmt.Sprint(webpQuality),
		"-f", "webp", outPath,
	)
}
