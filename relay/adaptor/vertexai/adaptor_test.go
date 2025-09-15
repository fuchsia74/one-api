package vertexai

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/meta"
)

func TestAdaptor_GetRequestURL(t *testing.T) {
	Convey("GetRequestURL", t, func() {
		adaptor := &Adaptor{}

		Convey("imagen models should use predict endpoint", func() {
			imagenTests := []struct {
				model  string
				region string
			}{
				{"imagen-4.0-fast-generate-001", "us-central1"},
				{"imagen-4.0-generate-001", "us-central1"},
				{"imagen-3.0-fast-generate-001", "europe-west4"},
				{"imagegeneration@006", "asia-southeast1"},
			}

			for _, tc := range imagenTests {
				Convey(tc.model+" predict", func() {
					meta := &meta.Meta{
						ActualModelName: tc.model,
						IsStream:        false,
						Config: model.ChannelConfig{
							Region:            tc.region,
							VertexAIProjectID: "test-project",
						},
					}

					url, err := adaptor.GetRequestURL(meta)
					So(err, ShouldBeNil)
					expectedURL := "https://" + tc.region + "-aiplatform.googleapis.com/v1/projects/test-project/locations/" + tc.region + "/publishers/google/models/" + tc.model + ":predict"
					So(url, ShouldEqual, expectedURL)
				})
			}
		})

		Convey("gemini-2.5-pro-preview models should use global endpoint", func() {
			testCases := []struct {
				name           string
				modelName      string
				isStream       bool
				expectedHost   string
				expectedLoc    string
				expectedSuffix string
			}{
				{
					name:           "gemini-2.5-pro-preview non-stream",
					modelName:      "gemini-2.5-pro-preview-05-06",
					isStream:       false,
					expectedHost:   "aiplatform.googleapis.com",
					expectedLoc:    "global",
					expectedSuffix: "generateContent",
				},
				{
					name:           "gemini-2.5-pro-preview stream",
					modelName:      "gemini-2.5-pro-preview-12-24",
					isStream:       true,
					expectedHost:   "aiplatform.googleapis.com",
					expectedLoc:    "global",
					expectedSuffix: "streamGenerateContent?alt=sse",
				},
				{
					name:           "gemini-2.5-pro-preview with additional suffix",
					modelName:      "gemini-2.5-pro-preview-latest",
					isStream:       false,
					expectedHost:   "aiplatform.googleapis.com",
					expectedLoc:    "global",
					expectedSuffix: "generateContent",
				},
			}

			for _, tc := range testCases {
				Convey(tc.name, func() {
					meta := &meta.Meta{
						ActualModelName: tc.modelName,
						IsStream:        tc.isStream,
						Config: model.ChannelConfig{
							Region:            "us-central1",
							VertexAIProjectID: "test-project",
						},
					}

					url, err := adaptor.GetRequestURL(meta)
					So(err, ShouldBeNil)

					expectedURL := "https://" + tc.expectedHost + "/v1/projects/test-project/locations/" + tc.expectedLoc + "/publishers/google/models/" + tc.modelName + ":" + tc.expectedSuffix
					So(url, ShouldEqual, expectedURL)
				})
			}
		})

		Convey("regular gemini models should use regional endpoint", func() {
			testCases := []struct {
				name           string
				modelName      string
				isStream       bool
				region         string
				expectedSuffix string
			}{
				{
					name:           "gemini-pro non-stream",
					modelName:      "gemini-pro",
					isStream:       false,
					region:         "us-central1",
					expectedSuffix: "generateContent",
				},
				{
					name:           "gemini-1.5-pro stream",
					modelName:      "gemini-1.5-pro",
					isStream:       true,
					region:         "europe-west4",
					expectedSuffix: "streamGenerateContent?alt=sse",
				},
				{
					name:           "gemini-flash",
					modelName:      "gemini-1.5-flash",
					isStream:       false,
					region:         "asia-southeast1",
					expectedSuffix: "generateContent",
				},
			}

			for _, tc := range testCases {
				Convey(tc.name, func() {
					meta := &meta.Meta{
						ActualModelName: tc.modelName,
						IsStream:        tc.isStream,
						Config: model.ChannelConfig{
							Region:            tc.region,
							VertexAIProjectID: "test-project",
						},
					}

					url, err := adaptor.GetRequestURL(meta)
					So(err, ShouldBeNil)

					expectedURL := "https://" + tc.region + "-aiplatform.googleapis.com/v1/projects/test-project/locations/" + tc.region + "/publishers/google/models/" + tc.modelName + ":" + tc.expectedSuffix
					So(url, ShouldEqual, expectedURL)
				})
			}
		})

		Convey("non-gemini models should use regional endpoint", func() {
			meta := &meta.Meta{
				ActualModelName: "claude-3-sonnet",
				IsStream:        false,
				Config: model.ChannelConfig{
					Region:            "us-central1",
					VertexAIProjectID: "test-project",
				},
			}

			url, err := adaptor.GetRequestURL(meta)
			So(err, ShouldBeNil)

			expectedURL := "https://us-central1-aiplatform.googleapis.com/v1/projects/test-project/locations/us-central1/publishers/google/models/claude-3-sonnet:rawPredict"
			So(url, ShouldEqual, expectedURL)
		})

		Convey("custom BaseURL should work for all models", func() {
			testCases := []struct {
				name        string
				modelName   string
				isStream    bool
				expectedLoc string
				suffix      string
			}{
				{
					name:        "gemini-2.5-pro-preview with custom BaseURL",
					modelName:   "gemini-2.5-pro-preview-05-06",
					isStream:    false,
					expectedLoc: "global",
					suffix:      "generateContent",
				},
				{
					name:        "regular gemini with custom BaseURL",
					modelName:   "gemini-pro",
					isStream:    false,
					expectedLoc: "us-central1",
					suffix:      "generateContent",
				},
			}

			for _, tc := range testCases {
				Convey(tc.name, func() {
					customBaseURL := "https://custom-vertex-proxy.example.com"
					meta := &meta.Meta{
						ActualModelName: tc.modelName,
						IsStream:        tc.isStream,
						BaseURL:         customBaseURL,
						Config: model.ChannelConfig{
							Region:            "us-central1",
							VertexAIProjectID: "test-project",
						},
					}

					url, err := adaptor.GetRequestURL(meta)
					So(err, ShouldBeNil)

					expectedURL := customBaseURL + "/v1/projects/test-project/locations/" + tc.expectedLoc + "/publishers/google/models/" + tc.modelName + ":" + tc.suffix
					So(url, ShouldEqual, expectedURL)
				})
			}
		})
	})
}

func TestIsRequireGlobalEndpoint(t *testing.T) {
	Convey("IsRequireGlobalEndpoint", t, func() {
		testCases := []struct {
			model    string
			expected bool
		}{
			{"gemini-2.5-pro-preview", true},
			{"gemini-2.5-pro-preview-05-06", true},
			{"gemini-2.5-pro-preview-12-24", true},
			{"gemini-2.5-pro-preview-latest", true},
			{"gemini-2.5-pro-preview-experimental", true},
			{"gemini-pro", false},
			{"gemini-1.5-pro", false},
			{"gemini-1.0-pro", false},
			{"gemini-1.5-flash", false},
			{"claude-3-sonnet", false},
			{"gpt-4", false},
			{"imagen-3.0", false},
			{"", false},
		}

		for _, tc := range testCases {
			Convey("model "+tc.model, func() {
				result := IsRequireGlobalEndpoint(tc.model)
				So(result, ShouldEqual, tc.expected)
			})
		}
	})
}

func TestIsImagenModel(t *testing.T) {
	Convey("isImagenModel", t, func() {
		cases := []struct {
			model    string
			expected bool
		}{
			{"imagen-4.0-fast-generate-001", true},
			{"imagen-3.0-generate-002", true},
			{"imagegeneration@006", true},
			{"gemini-1.5-pro", false},
			{"claude-3-sonnet", false},
			{"", false},
		}
		for _, c := range cases {
			Convey("model "+c.model, func() {
				So(isImagenModel(c.model), ShouldEqual, c.expected)
			})
		}
	})
}
