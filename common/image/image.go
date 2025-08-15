package image

import (
	"bytes"
	"encoding/base64"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/Laisky/errors/v2"
	_ "golang.org/x/image/webp"

	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/common/config"
)

// Regex to match data URL pattern
var dataURLPattern = regexp.MustCompile(`data:image/([^;]+);base64,(.*)`)

func IsImageUrl(url string) (bool, error) {
	resp, err := client.UserContentRequestHTTPClient.Head(url)
	if err != nil {
		return false, errors.Wrapf(err, "failed to fetch image URL: %s", url)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// this file may not support HEAD method
		resp, err = client.UserContentRequestHTTPClient.Get(url)
		if err != nil {
			return false, errors.Wrapf(err, "failed to fetch image URL: %s", url)
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return false, errors.Errorf("failed to fetch image URL: %s, status code: %d", url, resp.StatusCode)
	}

	maxSize := int64(config.MaxInlineImageSizeMB) * 1024 * 1024
	if resp.ContentLength > maxSize {
		return false, errors.Errorf("image size should not exceed %dMB: %s, size: %d", config.MaxInlineImageSizeMB, url, resp.ContentLength)
	}

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if !strings.HasPrefix(contentType, "image/") &&
		!strings.Contains(contentType, "application/octet-stream") {
		return false,
			errors.Errorf("invalid content type: %s, expected image type", contentType)
	}

	return true, nil
}

func GetImageSizeFromUrl(url string) (width int, height int, err error) {
	isImage, err := IsImageUrl(url)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to fetch image URL")
	}
	if !isImage {
		return 0, 0, errors.New("not an image URL")
	}
	resp, err := client.UserContentRequestHTTPClient.Get(url)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to get image from URL")
	}
	defer resp.Body.Close()

	img, _, err := image.DecodeConfig(resp.Body)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to decode image")
	}
	return img.Width, img.Height, nil
}

func GetImageFromUrl(url string) (mimeType string, data string, err error) {
	// Check if the URL is a data URL
	matches := dataURLPattern.FindStringSubmatch(url)
	if len(matches) == 3 {
		// URL is a data URL
		mimeType = "image/" + matches[1]
		data = matches[2]
		return
	}

	isImage, err := IsImageUrl(url)
	if err != nil {
		return mimeType, data, errors.Wrap(err, "failed to fetch image URL")
	}
	if !isImage {
		return mimeType, data, errors.New("not an image URL")
	}

	resp, err := client.UserContentRequestHTTPClient.Get(url)
	if err != nil {
		return mimeType, data, errors.Wrap(err, "failed to get image from URL")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return mimeType, data, errors.Errorf("failed to fetch image URL: %s, status code: %d", url, resp.StatusCode)
	}
	maxSize := int64(config.MaxInlineImageSizeMB) * 1024 * 1024
	if resp.ContentLength > maxSize {
		return mimeType, data, errors.Errorf("image size should not exceed %dMB: %s, size: %d", config.MaxInlineImageSizeMB, url, resp.ContentLength)
	}

	buffer := bytes.NewBuffer(nil)
	_, err = buffer.ReadFrom(resp.Body)
	if err != nil {
		return mimeType, data, errors.Wrap(err, "failed to read image data from response")
	}

	mimeType = resp.Header.Get("Content-Type")
	data = base64.StdEncoding.EncodeToString(buffer.Bytes())
	return mimeType, data, nil
}

var (
	reg = regexp.MustCompile(`data:image/([^;]+);base64,`)
)

var readerPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Reader{}
	},
}

func GetImageSizeFromBase64(encoded string) (width int, height int, err error) {
	decoded, err := base64.StdEncoding.DecodeString(reg.ReplaceAllString(encoded, ""))
	if err != nil {
		return 0, 0, err
	}

	reader := readerPool.Get().(*bytes.Reader)
	defer readerPool.Put(reader)
	reader.Reset(decoded)

	img, _, err := image.DecodeConfig(reader)
	if err != nil {
		return 0, 0, err
	}

	return img.Width, img.Height, nil
}

func GetImageSize(image string) (width int, height int, err error) {
	if strings.HasPrefix(image, "data:image/") {
		return GetImageSizeFromBase64(image)
	}
	return GetImageSizeFromUrl(image)
}
