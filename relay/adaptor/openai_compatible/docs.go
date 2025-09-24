// Package openai_compatible provides high-performance streaming response processing
// with specialized thinking block support for OpenAI-compatible APIs.
//
// This package implements optimized algorithms for real-time processing of streaming
// chat completions, with particular focus on extracting and handling <think></think>
// blocks commonly used by reasoning models.
//
// Key features:
//   - Ultra-low latency streaming processing
//   - Memory-efficient buffer management
//   - Thread-safe thinking block extraction
//   - Comprehensive usage tracking
//   - Error resilience and validation
//
// Performance considerations:
//   - Uses strings.Builder for O(1) string concatenation
//   - Single-pass thinking block processing
//   - Pre-allocated buffer capacities
//   - Minimal memory allocations during streaming
package openai_compatible
