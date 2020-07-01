package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	col "github.com/fatih/color"
	"github.com/meownoid/sharpei/vips"
	"github.com/meownoid/stempl"
	"github.com/pkg/errors"
)

func usage() {
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [OPTIONS] PATH [PATH] ...\n", os.Args[0])
	flag.PrintDefaults()
}

type outputFile struct {
	buf *bytes.Buffer
	ext string
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
		inputProfile  = flag.String("input-profile", "", "input icc profile")
		outputProfile = flag.String("output-profile", "", "output icc profile")

		noColor = flag.Bool("no-color", false, "disable colorized output")
	)

	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if *noColor {
		col.NoColor = true
	}

	var cfg *Config

	if *width != 0 || *height != 0 || *inputProfile != "" || *outputProfile != "" {
		cfg = &Config{
			Output:  *output,
			Format:  *format,
			Rewrite: *rewrite,
			Profiles: map[string]ProfileConfig{
				"thumbnail": {
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

	pathsToProcess := make([]string, 0, flag.NArg())

	for _, path := range flag.Args() {
		stat, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("%s: %s\n", path, col.RedString("does not exists, skipping"))
				continue
			}
			log.Fatal(err)
		}

		if stat.IsDir() {
			err := filepath.Walk(path,
				func(walkPath string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}

					if !info.IsDir() && isImage(info.Name()) {
						pathsToProcess = append(pathsToProcess, walkPath)
					}

					return nil
				})

			if err != nil {
				log.Fatal(err)
			}

			continue
		}

		if !isImage(path) {
			fmt.Printf("%s: %s\n", path, col.RedString("not an image, skipping"))
			continue
		}

		pathsToProcess = append(pathsToProcess, path)
	}

	if len(pathsToProcess) == 0 {
		col.Green("No images to process")
		os.Exit(0)
	}

	vips.Init(os.Args[0])
	defer vips.Shutdown()

	for _, inputPath := range pathsToProcess {
		func(inputPath string) {
			reader, err := os.Open(inputPath)
			if err != nil {
				fmt.Printf("%s: %s\n", inputPath, col.RedString(err.Error()))
				return
			}
			defer func() { _ = reader.Close() }()

			img, err := vips.Decode(reader)
			if err != nil {
				fmt.Printf("%s: %s\n", inputPath, col.RedString(err.Error()))
				return
			}
			defer img.Destroy()

			// Autorotate
			imgRotated, err := img.Autorot()
			if err != nil {
				imgRotated = img
			} else {
				defer imgRotated.Destroy()
			}

			imgRotatedCopy, err := imgRotated.Copy()
			if err != nil {
				fmt.Printf("%s: %s\n", inputPath, col.RedString(err.Error()))
				return
			}
			defer imgRotatedCopy.Destroy()

			// Remove EXIF metadata
			for _, p := range imgRotatedCopy.Properties() {
				if strings.HasPrefix(p, "exif") || strings.HasPrefix(p, "iptc") || strings.HasPrefix(p, "xmp") || p == "orientation" {
					_ = imgRotatedCopy.RemoveProperty(p)
				}
			}

			basename := filepath.Base(inputPath)
			name := strings.TrimSuffix(basename, filepath.Ext(basename))

			for profileName, profile := range cfg.Profiles {
				func(profileName string, profile ProfileConfig) {
					out, err := processProfile(profile, imgRotatedCopy)
					if err != nil {
						fmt.Printf("%s: error while processing profile %s: %s\n", inputPath, profileName, col.RedString(err.Error()))
						return
					}

					filename, err := stempl.Format(
						cfg.Format,
						map[string]string{
							"profile": profileName,
							"name":    name,
						},
					)
					if err != nil {
						fmt.Printf("%s: error in format string for profile %s: %s\n", inputPath, profileName, col.RedString(err.Error()))
						return
					}

					filename = fmt.Sprintf("%s.%s", filename, out.ext)

					inputDir := filepath.Dir(inputPath)
					outputDir := filepath.Join(cfg.Output, inputDir)

					stat, err := os.Stat(outputDir)

					if err != nil {
						if os.IsNotExist(err) {
							err = os.MkdirAll(outputDir, 0755)
						}
						if err != nil {
							fmt.Printf("%s: %s\n", outputDir, col.RedString(err.Error()))
							return
						}
					} else if !stat.IsDir() {
						fmt.Printf("%s: %s\n", outputDir, col.RedString("exists and not a directory, skipping"))
						return
					}

					outputPath := filepath.Join(outputDir, filename)

					if _, err := os.Stat(outputPath); err == nil && !cfg.Rewrite {
						fmt.Printf("%s: %s\n", outputPath, col.RedString("already exists, skipping"))
						return
					}

					f, err := os.Create(outputPath)
					if err != nil {
						fmt.Printf("%s: %s\n", outputPath, col.RedString(err.Error()))
						return
					}
					defer func() { _ = f.Close() }()

					_, err = out.buf.WriteTo(f)
					if err != nil {
						fmt.Printf("%s: %s\n", outputPath, col.RedString(err.Error()))
						return
					}

					fmt.Printf("%s: %s\n", outputPath, col.GreenString("OK"))
				}(profileName, profile)
			}
		}(inputPath)
	}
}

func isImage(filename string) bool {
	switch strings.ToLower(filepath.Ext(filepath.Base(filename))) {
	case "", ".jpeg", ".jpg", ".jpe", ".jif", ".jfif", ".jfi":
		return true
	case ".png":
		return true
	case ".tiff", ".tif":
		return true
	case ".webp":
		return true
	}

	return false
}

func processProfile(profile ProfileConfig, img *vips.Image) (*outputFile, error) {
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
		return nil, err
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
		return nil, errors.Errorf("unsupported file type %s, use jpg, png, webp or tiff", fileType)
	}

	if err != nil {
		return nil, err
	}

	return &outputFile{
		buf: buf,
		ext: fileType,
	}, nil
}
