package test

import _ "embed"

// VisionImage is a bundled sample used by the end-to-end vision regression tests.
//
//go:embed vision.jpeg
var VisionImage []byte
