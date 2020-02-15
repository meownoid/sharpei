package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/meownoid/sharpei/vips"
	"github.com/meownoid/stempl"
	"github.com/pkg/errors"
)

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [OPTIONS] PATH [PATH] ...\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage

	var (
		config = flag.String("config", "", "path to config")

		output  = flag.String("output", ".", "output directory")
		format  = flag.String("format", "{name}_{profile}", "format of output filenames")
		rewrite = flag.Bool("rewrite", false, "to rewrite existing files")

		width         = flag.Int("width", 0, "width of the output image")
		height        = flag.Int("height", 0, "height of the output image")
		inputProfile  = flag.String("input_profile", "", "input icc profile")
		outputProfile = flag.String("output_profile", "", "output icc profile")
	)

	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	var cfg *Config

	if *width != 0 || *height != 0 || *inputProfile != "" || *outputProfile != "" {
		cfg = &Config{
			Output:  *output,
			Format:  *format,
			Rewrite: *rewrite,
			Profiles: map[string]ProfileConfig{
				"thumbnail": ProfileConfig{
					Width:         *width,
					Height:        *height,
					InputProfile:  *inputProfile,
					OutputProfile: *outputProfile,
					Type:          "jpg",
				},
			},
		}
	}

	if *config != "" {
		if cfg != nil {
			log.Fatal("either external or cli config should be present, not both")
		}

		var err error
		cfg, err = loadConfig(*config)
		if err != nil {
			log.Fatal(err)
		}
	}

	if cfg == nil {
		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		defaultPaths := []string{
			"sharpei.yaml",
			"sharpei.yml",
			".sharpei.yaml",
			".sharpei.yml",
			filepath.Join(usr.HomeDir, ".sharpei.yaml"),
			filepath.Join(usr.HomeDir, ".sharpei.yml"),
		}
		for _, path := range defaultPaths {
			if _, err := os.Stat(path); err == nil {
				cfg, err = loadConfig(path)
				if err != nil {
					log.Fatal(errors.Wrap(err, path))
				}
				break
			}
		}
	}

	if cfg == nil {
		log.Fatal("no config found, use cli arguments, config flag, sharpei.yml or ~/.sharpei.yml")
	}

	vips.Init(os.Args[0])
	defer vips.Shutdown()

	for _, arg := range flag.Args() {
		err := process(cfg, arg)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func process(cfg *Config, file string) error {
	reader, err := os.Open(file)
	if err != nil {
		return err
	}
	defer reader.Close()

	img, err := vips.Decode(reader)
	if err != nil {
		return err
	}
	defer img.Destroy()

	for profileName, profile := range cfg.Profiles {
		transformedImg, err := TransformImage(
			img,
			TransfromConfig{
				Width:         profile.Width,
				Height:        profile.Height,
				InputProfile:  profile.InputProfile,
				OutputProfile: profile.OutputProfile,
			},
		)
		if err != nil {
			return err
		}
		defer transformedImg.Destroy()

		quality := profile.Quality
		if quality == 0 {
			quality = 95
		}
		if quality < 1 {
			quality = 1
		}
		if quality > 100 {
			quality = 100
		}

		compression := profile.Compression
		if compression == 0 {
			compression = 7
		}
		if compression < 1 {
			compression = 1
		}
		if compression > 9 {
			compression = 9
		}

		fileType := strings.ToLower(profile.Type)

		buf := bytes.NewBuffer([]byte{})

		switch fileType {
		case "jpeg", "jpg", "jpe", "jif", "jfif", "jfi":
			err = transformedImg.EncodeJPEG(buf, quality)
		case "png":
			err = transformedImg.EncodePNG(buf, compression)
		case "tiff", "tif":
			err = transformedImg.EncodeTIFF(buf)
		case "webp":
			err = transformedImg.EncodeWEBP(buf, quality, false)
		default:
			return errors.Errorf("unsupported file type %s, use jpg, png, webp or tiff", fileType)
		}

		if err != nil {
			return err
		}

		random := randString(32)

		basename := filepath.Base(file)
		name := strings.TrimSuffix(basename, filepath.Ext(basename))

		filename, err := stempl.Format(
			cfg.Format,
			map[string]string{
				"profile":  profileName,
				"name":     name,
				"random32": random,
				"random16": random[:16],
				"random8":  random[:8],
				"random4":  random[:4],
			},
		)
		if err != nil {
			return errors.Wrap(err, "format error")
		}

		filename = fmt.Sprintf("%s.%s", filename, fileType)

		if _, err := os.Stat(filename); err == nil && !cfg.Rewrite {
			return errors.Errorf("file %s already exists and rewrite=false", filename)
		}

		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		buf.WriteTo(f)
	}

	return nil
}

const randStringBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = randStringBytes[rand.Int63()%int64(len(randStringBytes))]
	}
	return string(b)
}
