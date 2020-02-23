package main

import (
	"errors"
	"math"
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
	Height        int
	InputProfile  string
	OutputProfile string
}

func TransformImage(img *vips.Image, cfg TransfromConfig) (*vips.Image, error) {
	if cfg.Width < 0 {
		cfg.Width = 0
	}

	if cfg.Height < 0 {
		cfg.Height = 0
	}

	if cfg.Width == 0 && cfg.Height == 0 {
		return nil, errors.New("either width or height should be greater than zero")
	}

	// Calculate scale
	scalex := float64(cfg.Width) / float64(img.Width())
	scaley := float64(cfg.Height) / float64(img.Height())
	scale := math.Max(scalex, scaley)

	isEmbeddedICC := img.IsPropertySet("icc-profile-data")

	if cfg.OutputProfile == "same" && isEmbeddedICC {
		// Resize image in the original color space
		imgResized, err := img.Resize(scale, scale)
		if err != nil {
			return nil, err
		}

		return imgResized, nil
	}

	if cfg.InputProfile == "" {
		switch img.Interpretation() {
		case vips.INTERPRETATION_B_W, vips.INTERPRETATION_GREY16:
			cfg.InputProfile = "gray"
		default:
			cfg.InputProfile = "srgb"
		}
	}

	imgWithICCProfile := img
	profileAttached := ""

	// If image has no attached profile, load default one and embedd it into the image
	if !isEmbeddedICC {
		profileAttached = cfg.InputProfile

		inputProfile, err := getProfile(cfg.InputProfile)
		if err != nil {
			return nil, err
		}

		imgWithICCProfile, err = img.Copy()
		if err != nil {
			return nil, err
		}
		defer imgWithICCProfile.Destroy()

		imgWithICCProfile.SetPropertyBlob("icc-profile-data", inputProfile)
	}

	// Import image to LAB PCS space using embedded profile
	imgImported, err := imgWithICCProfile.ICCImport(vips.INTENT_RELATIVE)
	if err != nil {
		return nil, err
	}
	defer imgImported.Destroy()

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

	if cfg.OutputProfile == "" || cfg.OutputProfile == "same" {
		switch img.Interpretation() {
		case vips.INTERPRETATION_B_W, vips.INTERPRETATION_GREY16:
			cfg.OutputProfile = "gray"
		default:
			cfg.OutputProfile = "srgb"
		}
	}

	// Load output profile and attach it to the image
	if cfg.OutputProfile != profileAttached {
		outputProfile, err := getProfile(cfg.OutputProfile)
		if err != nil {
			return nil, err
		}

		imgResizedCopy.SetPropertyBlob("icc-profile-data", outputProfile)
	}

	// Export image to output ICC profile
	imgExported, err := imgResizedCopy.ICCExport(vips.INTENT_RELATIVE, 8)
	if err != nil {
		return nil, err
	}

	return imgExported, nil
}
