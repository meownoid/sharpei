package main

import (
	"log"
	"os"

	"github.com/meownoid/s3-publisher/vips"
)

func main() {
	vips.Init(os.Args[0])
	defer vips.Shutdown()

	reader, err := os.Open("test.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	img, err := vips.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(img.Filename())
	log.Printf("(%v, %v) - (%v, %v)", img.XOffset(), img.YOffset(), img.XOffset()+img.Width(), img.YOffset()+img.Height())
	log.Printf("%v band(s)", img.Bands())
	log.Printf("ICC: %v", img.ICC())

	f, err := os.Create("out.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	err = img.EncodeJPEG(f, 95)
	if err != nil {
		log.Fatal(err)
	}
}
