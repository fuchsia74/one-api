package helper

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"

	"github.com/Laisky/errors/v2"
)

var (
	ffprobePath string
	ffprobeOnce sync.Once
	ffprobeErr  error
)

func lookupFFProbe() (string, error) {
	ffprobeOnce.Do(func() {
		path, err := exec.LookPath("ffprobe")
		if err != nil {
			if alt, altErr := exec.LookPath("avprobe"); altErr == nil {
				ffprobePath = alt
				ffprobeErr = nil
				return
			}
			ffprobeErr = errors.Wrap(err, "ffprobe not found in PATH")
			return
		}
		ffprobePath = path
	})
	return ffprobePath, ffprobeErr
}

// SaveTmpFile saves data to a temporary file. The filename would be apppended with a random string.
func SaveTmpFile(filename string, data io.Reader) (string, error) {
	if data == nil {
		return "", errors.New("data is nil")
	}

	f, err := os.CreateTemp("", "*-"+filename)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create temporary file %s", filename)
	}
	defer f.Close()

	_, err = io.Copy(f, data)
	if err != nil {
		return "", errors.Wrapf(err, "failed to copy data to temporary file %s", filename)
	}

	return f.Name(), nil
}

// GetAudioTokens returns the number of tokens in an audio file.
func GetAudioTokens(ctx context.Context, audio io.Reader, tokensPerSecond float64) (float64, error) {
	filename, err := SaveTmpFile("audio", audio)
	if err != nil {
		return 0, errors.Wrap(err, "failed to save audio to temporary file")
	}
	defer os.Remove(filename)

	duration, err := GetAudioDuration(ctx, filename)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get audio tokens")
	}

	return duration * tokensPerSecond, nil
}

// GetAudioDuration returns the duration of an audio file in seconds.
func GetAudioDuration(ctx context.Context, filename string) (float64, error) {
	path, err := lookupFFProbe()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get audio duration")
	}
	// ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 {{input}}
	c := exec.CommandContext(ctx, path, "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filename)
	output, err := c.Output()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get audio duration")
	}

	// Actually gpt-4-audio calculates tokens with 0.1s precision,
	// while whisper calculates tokens with 1s precision
	return strconv.ParseFloat(string(bytes.TrimSpace(output)), 64)
}
