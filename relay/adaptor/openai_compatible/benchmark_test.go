package openai_compatible

import (
	"strings"
	"testing"
)

// BenchmarkStringConcatenation benchmarks the old string concatenation approach
func BenchmarkStringConcatenation(b *testing.B) {
	chunks := []string{
		"Hello world, this is chunk 1",
		"This is chunk 2 with some content",
		"Another chunk with more text content",
		"Even more content in this chunk",
		"Final chunk with additional text",
	}

	for b.Loop() {
		responseText := ""
		for _, chunk := range chunks {
			responseText += chunk
		}
		_ = responseText
	}
}

// BenchmarkStringBuilder benchmarks the new StringBuilder approach
func BenchmarkStringBuilder(b *testing.B) {
	chunks := []string{
		"Hello world, this is chunk 1",
		"This is chunk 2 with some content",
		"Another chunk with more text content",
		"Even more content in this chunk",
		"Final chunk with additional text",
	}

	for b.Loop() {
		var responseTextBuilder strings.Builder
		responseTextBuilder.Grow(DefaultBuilderCapacity)
		for _, chunk := range chunks {
			responseTextBuilder.WriteString(chunk)
		}
		_ = responseTextBuilder.String()
	}
}

// BenchmarkStringConcatenationLarge benchmarks string concatenation with many chunks
func BenchmarkStringConcatenationLarge(b *testing.B) {
	chunks := make([]string, 100)
	for i := range chunks {
		chunks[i] = "This is a repetitive chunk of text that simulates streaming data "
	}

	for b.Loop() {
		responseText := ""
		for _, chunk := range chunks {
			responseText += chunk
		}
		_ = responseText
	}
}

// BenchmarkStringBuilderLarge benchmarks StringBuilder with many chunks
func BenchmarkStringBuilderLarge(b *testing.B) {
	chunks := make([]string, 100)
	for i := range chunks {
		chunks[i] = "This is a repetitive chunk of text that simulates streaming data "
	}

	for b.Loop() {
		var responseTextBuilder strings.Builder
		responseTextBuilder.Grow(LargeBuilderCapacity)
		for _, chunk := range chunks {
			responseTextBuilder.WriteString(chunk)
		}
		_ = responseTextBuilder.String()
	}
}

// BenchmarkMemoryAllocations tests memory allocation patterns
func BenchmarkMemoryAllocations(b *testing.B) {
	b.Run("StringConcatenation", func(b *testing.B) {
		b.ReportAllocs()
		chunk := "test chunk data"

		for b.Loop() {
			responseText := ""
			for range 50 {
				responseText += chunk
			}
			_ = responseText
		}
	})

	b.Run("StringBuilder", func(b *testing.B) {
		b.ReportAllocs()
		chunk := "test chunk data"

		for b.Loop() {
			var responseTextBuilder strings.Builder
			responseTextBuilder.Grow(DefaultBuilderCapacity)
			for range 50 {
				responseTextBuilder.WriteString(chunk)
			}
			_ = responseTextBuilder.String()
		}
	})
}

// BenchmarkEnhancedUnifiedStreamHandler tests the enhanced unified implementation
func BenchmarkEnhancedUnifiedStreamHandler(b *testing.B) {
	b.Run("RegularContent", func(b *testing.B) {
		b.ReportAllocs()
		chunks := []string{
			"Hello world, this is chunk 1",
			"This is chunk 2 with some content",
			"Another chunk with more text content",
			"Even more content in this chunk",
			"Final chunk with additional text",
		}

		for b.Loop() {
			var responseTextBuilder strings.Builder
			responseTextBuilder.Grow(DefaultBuilderCapacity)

			// Simulate enhanced unified approach
			for _, chunk := range chunks {
				responseTextBuilder.WriteString(chunk)
			}
			_ = responseTextBuilder.String()
		}
	})

	b.Run("ThinkingContent", func(b *testing.B) {
		b.ReportAllocs()
		chunks := []string{
			"Hello <think>first thought</think> world",
			"More content here",
			"<think>another thought</think> but only first processed",
			"Final content chunk",
		}

		for b.Loop() {
			var responseTextBuilder strings.Builder
			responseTextBuilder.Grow(DefaultBuilderCapacity)

			// Simulate enhanced unified thinking processing
			hasProcessedThinkTag := false
			for _, chunk := range chunks {
				responseTextBuilder.WriteString(chunk)

				// Simulate thinking detection logic
				if !hasProcessedThinkTag && strings.Contains(chunk, "<think>") {
					hasProcessedThinkTag = true
				}
			}
			_ = responseTextBuilder.String()
		}
	})
}

// BenchmarkEnhancedVsOriginal compares all approaches
func BenchmarkEnhancedVsOriginal(b *testing.B) {
	chunks := make([]string, 50)
	for i := range chunks {
		if i%10 == 0 {
			chunks[i] = "Content with <think>thinking block</think> included"
		} else {
			chunks[i] = "Regular streaming content chunk"
		}
	}

	b.Run("OriginalStringConcatenation", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			responseText := ""
			for _, chunk := range chunks {
				responseText += chunk
			}
			_ = responseText
		}
	})

	b.Run("EnhancedUnifiedStringBuilder", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var responseTextBuilder strings.Builder
			responseTextBuilder.Grow(LargeBuilderCapacity)

			hasProcessedThinkTag := false
			for _, chunk := range chunks {
				responseTextBuilder.WriteString(chunk)

				// Enhanced thinking detection
				if !hasProcessedThinkTag && strings.Contains(chunk, "<think>") {
					hasProcessedThinkTag = true
				}
			}
			_ = responseTextBuilder.String()
		}
	})
}
