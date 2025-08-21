package ratio

var ImageSizeRatios = map[string]map[string]float64{
	"dall-e-2": {
		"256x256":   1,
		"512x512":   1.125,
		"1024x1024": 1.25,
	},
	"dall-e-3": {
		"1024x1024": 1,
		"1024x1792": 2,
		"1792x1024": 2,
	},
	"gpt-image-1": {
		"1024x1024": 1,
		"1024x1536": 16.0 / 11,
		"1536x1024": 16.0 / 11,
	},
	"ali-stable-diffusion-xl": {
		"512x1024":  1,
		"1024x768":  1,
		"1024x1024": 1,
		"576x1024":  1,
		"1024x576":  1,
	},
	"ali-stable-diffusion-v1.5": {
		"512x1024":  1,
		"1024x768":  1,
		"1024x1024": 1,
		"576x1024":  1,
		"1024x576":  1,
	},
	"wanx-v1": {
		"1024x1024": 1,
		"720x1280":  1,
		"1280x720":  1,
	},
	"step-1x-medium": {
		"256x256":   1,
		"512x512":   1,
		"768x768":   1,
		"1024x1024": 1,
		"1280x800":  1,
		"800x1280":  1,
	},
}

var ImageGenerationAmounts = map[string][2]int{
	"dall-e-2":                  {1, 10},
	"dall-e-3":                  {1, 1}, // OpenAI allows n=1 currently.
	"gpt-image-1":               {1, 1},
	"ali-stable-diffusion-xl":   {1, 4}, // Ali
	"ali-stable-diffusion-v1.5": {1, 4}, // Ali
	"wanx-v1":                   {1, 4}, // Ali
	"cogview-3":                 {1, 1},
	"step-1x-medium":            {1, 1},
}

var ImagePromptLengthLimitations = map[string]int{
	"dall-e-2":                  1000,
	"dall-e-3":                  4000,
	"gpt-image-1":               4000,
	"ali-stable-diffusion-xl":   4000,
	"ali-stable-diffusion-v1.5": 4000,
	"wanx-v1":                   4000,
	"cogview-3":                 833,
	"step-1x-medium":            4000,
}

var ImageOriginModelName = map[string]string{
	"ali-stable-diffusion-xl":   "stable-diffusion-xl",
	"ali-stable-diffusion-v1.5": "stable-diffusion-v1.5",
}

// ImageTierTables defines structured multipliers per model by quality and size.
// Keys:
//
//	model -> quality -> size -> multiplier
//
// Quality "default" applies when no specific quality tier is set.
var ImageTierTables = map[string]map[string]map[string]float64{
	// DALL·E 2: default tiers from ImageSizeRatios
	"dall-e-2": {
		"default": {
			"256x256":   1,
			"512x512":   1.125,
			"1024x1024": 1.25,
		},
	},
	// DALL·E 3: default + hd tiers
	"dall-e-3": {
		"default": {
			"1024x1024": 1,
			"1024x1792": 2,
			"1792x1024": 2,
		},
		// hd multiplies default by 2 (square) or 1.5 (non-square)
		"hd": {
			"1024x1024": 2, // 1 * 2
			"1024x1792": 3, // 2 * 1.5
			"1792x1024": 3, // 2 * 1.5
		},
	},
	// GPT Image: explicit per-quality tiers
	"gpt-image-1": {
		"default": {
			"1024x1024": 1,
			"1024x1536": 16.0 / 11,
			"1536x1024": 16.0 / 11,
		},
		"low": {
			"1024x1024": 1,
			"1024x1536": 16.0 / 11,
			"1536x1024": 16.0 / 11,
		},
		"medium": {
			"1024x1024": 42.0 / 11,
			"1024x1536": 63.0 / 11,
			"1536x1024": 63.0 / 11,
		},
		"high": {
			"1024x1024": 167.0 / 11,
			"1024x1536": 250.0 / 11,
			"1536x1024": 250.0 / 11,
		},
		"auto": { // same as high
			"1024x1024": 167.0 / 11,
			"1024x1536": 250.0 / 11,
			"1536x1024": 250.0 / 11,
		},
	},
	// Ali SD variants and others: copy defaults from ImageSizeRatios
	"ali-stable-diffusion-xl": {
		"default": {
			"512x1024":  1,
			"1024x768":  1,
			"1024x1024": 1,
			"576x1024":  1,
			"1024x576":  1,
		},
	},
	"ali-stable-diffusion-v1.5": {
		"default": {
			"512x1024":  1,
			"1024x768":  1,
			"1024x1024": 1,
			"576x1024":  1,
			"1024x576":  1,
		},
	},
	"wanx-v1": {
		"default": {
			"1024x1024": 1,
			"720x1280":  1,
			"1280x720":  1,
		},
	},
	"step-1x-medium": {
		"default": {
			"256x256":   1,
			"512x512":   1,
			"768x768":   1,
			"1024x1024": 1,
			"1280x800":  1,
			"800x1280":  1,
		},
	},
}
