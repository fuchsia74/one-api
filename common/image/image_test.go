package image_test

import (
	"bytes"
	"encoding/base64"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "golang.org/x/image/webp"

	"github.com/songquanpeng/one-api/common/client"
	img "github.com/songquanpeng/one-api/common/image"
)

type CountingReader struct {
	reader    io.Reader
	BytesRead int
}

func (r *CountingReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.BytesRead += n
	return n, err
}

// retryHTTPGet retries HTTP GET requests with exponential backoff to handle network issues in CI
func retryHTTPGet(url string, maxRetries int) (*http.Response, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			return resp, nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		lastErr = err
		if i < maxRetries-1 {
			time.Sleep(time.Duration(1<<uint(i)) * time.Second) // exponential backoff
		}
	}
	return nil, lastErr
}

var (
	cases = []struct {
		url    string
		format string
		width  int
		height int
	}{
		{"https://s3.laisky.com/uploads/2025/05/Gfp-wisconsin-madison-the-nature-boardwalk.jpg", "jpeg", 2560, 1669},
		{"https://s3.laisky.com/uploads/2025/05/Basshunter_live_performances.png", "png", 4500, 2592},
		{"https://s3.laisky.com/uploads/2025/05/TO_THE_ONE_SOMETHINGNESS.webp", "webp", 984, 985},
		{"https://s3.laisky.com/uploads/2025/05/01_Das_Sandberg-Modell.gif", "gif", 1917, 1533},
		{"https://s3.laisky.com/uploads/2025/05/102Cervus.jpg", "jpeg", 270, 230},
	}
)

func TestMain(m *testing.M) {
	client.Init()
	m.Run()
}

func TestDecode(t *testing.T) {
	t.Parallel()

	// Bytes read: varies sometimes
	// jpeg: 1063892
	// png: 294462
	// webp: 99529
	// gif: 956153
	// jpeg#01: 32805
	for _, c := range cases {
		t.Run("Decode:"+c.format, func(t *testing.T) {
			t.Logf("testing %s", c.url)
			resp, err := retryHTTPGet(c.url, 3)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equalf(t, http.StatusOK, resp.StatusCode, "status code from %s", c.url)

			reader := &CountingReader{reader: resp.Body}
			img, format, err := image.Decode(reader)
			require.NoErrorf(t, err, "decode image from %s", c.url)
			size := img.Bounds().Size()
			require.Equal(t, c.format, format)
			require.Equal(t, c.width, size.X)
			require.Equal(t, c.height, size.Y)
			t.Logf("Bytes read: %d", reader.BytesRead)
		})
	}

	// Bytes read:
	// jpeg: 4096
	// png: 4096
	// webp: 4096
	// gif: 4096
	// jpeg#01: 4096
	for _, c := range cases {
		t.Run("DecodeConfig:"+c.format, func(t *testing.T) {
			resp, err := retryHTTPGet(c.url, 3)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equalf(t, http.StatusOK, resp.StatusCode, "status code from %s", c.url)

			reader := &CountingReader{reader: resp.Body}
			config, format, err := image.DecodeConfig(reader)
			require.NoError(t, err)
			require.Equal(t, c.format, format)
			require.Equal(t, c.width, config.Width)
			require.Equal(t, c.height, config.Height)
			t.Logf("Bytes read: %d", reader.BytesRead)
		})
	}
}

func TestBase64(t *testing.T) {
	t.Parallel()

	// Bytes read:
	// jpeg: 1063892
	// png: 294462
	// webp: 99072
	// gif: 953856
	// jpeg#01: 32805
	for _, c := range cases {
		t.Run("Decode:"+c.format, func(t *testing.T) {
			resp, err := retryHTTPGet(c.url, 3)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equalf(t, http.StatusOK, resp.StatusCode, "status code from %s", c.url)

			require.Equalf(t, http.StatusOK, resp.StatusCode, "status code from %s", c.url)

			data, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			encoded := base64.StdEncoding.EncodeToString(data)
			body := base64.NewDecoder(base64.StdEncoding, strings.NewReader(encoded))
			reader := &CountingReader{reader: body}
			img, format, err := image.Decode(reader)
			require.NoError(t, err)
			size := img.Bounds().Size()
			require.Equal(t, c.format, format)
			require.Equal(t, c.width, size.X)
			require.Equal(t, c.height, size.Y)
			t.Logf("Bytes read: %d", reader.BytesRead)
		})
	}

	// Bytes read:
	// jpeg: 1536
	// png: 768
	// webp: 768
	// gif: 1536
	// jpeg#01: 3840
	for _, c := range cases {
		t.Run("DecodeConfig:"+c.format, func(t *testing.T) {
			resp, err := retryHTTPGet(c.url, 3)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equalf(t, http.StatusOK, resp.StatusCode, "status code from %s", c.url)

			data, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			encoded := base64.StdEncoding.EncodeToString(data)
			body := base64.NewDecoder(base64.StdEncoding, strings.NewReader(encoded))
			reader := &CountingReader{reader: body}
			config, format, err := image.DecodeConfig(reader)
			require.NoError(t, err)
			require.Equal(t, c.format, format)
			require.Equal(t, c.width, config.Width)
			require.Equal(t, c.height, config.Height)
			t.Logf("Bytes read: %d", reader.BytesRead)
		})
	}
}

func TestGetImageSize(t *testing.T) {
	t.Parallel()

	for i, c := range cases {
		t.Run("Decode:"+strconv.Itoa(i), func(t *testing.T) {
			width, height, err := img.GetImageSize(c.url)
			require.NoError(t, err)
			require.Equal(t, c.width, width)
			require.Equal(t, c.height, height)
		})
	}
}

func TestGetImageSizeFromBase64(t *testing.T) {
	t.Parallel()

	// Use embedded minimal images to avoid network flakiness
	// 1x1 transparent PNG
	b64PNG1x1 := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9YwVf0sAAAAASUVORK5CYII="
	// 1x1 GIF
	b64GIF1x1 := "R0lGODlhAQABAPAAAP///wAAACH5BAAAAAAALAAAAAABAAEAAAICRAEAOw=="

	tests := []struct {
		name   string
		b64    string
		width  int
		height int
	}{
		{name: "PNG_1x1", b64: b64PNG1x1, width: 1, height: 1},
		{name: "GIF_1x1", b64: b64GIF1x1, width: 1, height: 1},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			width, height, err := img.GetImageSizeFromBase64(tt.b64)
			require.NoError(t, err)
			require.Equal(t, tt.width, width)
			require.Equal(t, tt.height, height)
		})
	}
}

func TestGetImageFromUrl(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		wantMime   string
		wantErr    bool
		errMessage string
	}{
		{
			name:     "Valid JPEG URL",
			input:    cases[0].url, // Using the existing JPEG test case
			wantMime: "image/jpeg",
			wantErr:  false,
		},
		{
			name:     "Valid PNG URL",
			input:    cases[1].url, // Using the existing PNG test case
			wantMime: "image/png",
			wantErr:  false,
		},
		{
			name:     "Valid Data URL",
			input:    "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==",
			wantMime: "image/png",
			wantErr:  false,
		},
		{
			name:       "Invalid URL",
			input:      "https://invalid.example.com/nonexistent.jpg",
			wantErr:    true,
			errMessage: "failed to fetch image URL",
		},
		{
			name:       "Non-image URL",
			input:      "https://ario.laisky.com/alias/doc",
			wantErr:    true,
			errMessage: "invalid content type",
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mimeType, data, err := img.GetImageFromUrl(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMessage != "" {
					require.Contains(t, err.Error(), tt.errMessage)
				}
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, data)

			// For data URLs, we should verify the mime type matches the input
			if strings.HasPrefix(tt.input, "data:image/") {
				require.Equal(t, tt.wantMime, mimeType)
				return
			}

			// For regular URLs, verify the base64 data is valid and can be decoded
			decoded, err := base64.StdEncoding.DecodeString(data)
			require.NoError(t, err)
			require.NotEmpty(t, decoded)

			// Verify the decoded data is a valid image
			reader := bytes.NewReader(decoded)
			_, format, err := image.DecodeConfig(reader)
			require.NoError(t, err)
			require.Equal(t, strings.TrimPrefix(tt.wantMime, "image/"), format)
		})
	}
}
