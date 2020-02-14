package main

import (
	"log"
	"os"

	"github.com/meownoid/sharpei/vips"
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

	transformedImg, err := TransformImage(img, TransfromConfig{Width: 2075})
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create("out.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	err = transformedImg.EncodeJPEG(f, 95)
	if err != nil {
		log.Fatal(err)
	}
}
