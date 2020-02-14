package main

import (
	"strings"

	"github.com/meownoid/sharpei/vips"
)

var profileMapping = map[string]string{
	"gray":    "data/gray.icc",
	"srgb":    "data/sRGB2014.icc",
	"srgb-v2": "data/sRGB2014.icc",
	"srgb-v4": "data/sRGB_v4_ICC_preference.icc",
}

var profileCache = map[string][]byte{}

func getProfile(name string) ([]byte, error) {
	if profile, ok := profileCache[name]; ok {
		return profile, nil
	}

	nameLower := strings.ToLower(name)

	if profileName, ok := profileMapping[nameLower]; ok {
		profile, err := Asset(profileName)
		if err != nil {
			return nil, err
		}

		return profile, nil
	}

	profile, err := vips.LoadProfile(name)
	if err != nil {
		return nil, err
	}

	profileCache[name] = profile
	return profile, nil
}

type TransfromConfig struct {
	Width         int
	InputProfile  string
	OutputProfile string
}

func TransformImage(img *vips.Image, cfg TransfromConfig) (*vips.Image, error) {
	if cfg.InputProfile == "" {
		switch img.Interpretation() {
		case vips.INTERPRETATION_B_W, vips.INTERPRETATION_GREY16:
			cfg.InputProfile = "gray"
		default:
			cfg.InputProfile = "srgb"
		}
	}

	if cfg.OutputProfile == "" {
		cfg.OutputProfile = cfg.InputProfile
	}

	// If image has no attached profile, load default one and attach it to the image
	if !img.IsPropertySet("icc-profile-data") {
		inputProfile, err := getProfile(cfg.InputProfile)
		if err != nil {
			return nil, err
		}

		img.SetPropertyBlob("icc-profile-data", inputProfile)
	}

	// Import image to LAB PCS space using embedded profile with fallback to default profile
	imgImported, err := img.ICCImport(vips.INTENT_RELATIVE)
	if err != nil {
		return nil, err
	}
	defer imgImported.Destroy()

	scale := float64(cfg.Width) / float64(img.Width())

	// Resize image in the LAB PCS space
	imgResized, err := imgImported.Resize(scale, scale)
	if err != nil {
		return nil, err
	}
	defer imgResized.Destroy()

	imgResizedCopy, err := imgResized.Copy()
	if err != nil {
		return nil, err
	}
	defer imgResizedCopy.Destroy()

	// Load output profile and attach it to the image
	outputProfile, err := getProfile(cfg.OutputProfile)
	if err != nil {
		return nil, err
	}

	imgResizedCopy.SetPropertyBlob("icc-profile-data", outputProfile)

	// Export image to WEB-optimized ICC profile
	imgExported, err := imgResizedCopy.ICCExport(vips.INTENT_RELATIVE, 8)
	if err != nil {
		return nil, err
	}

	return imgExported, nil
}
