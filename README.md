# BMP Decoding and Encoding in Go

bmp-go is a Go package aiming to provide abilities to decode BMP image files having 4, 8, 16, 24 and 32 bits color depths. It supports RLE decoding which might be applied as a compression method to 4 bits and 8 bits BMP images. The decoder outputs an image.Image with an *image.RGBA format.

This could also be used to directly encode an *image.RGBA format and writes it into an io.Writer as a 32 bits RGBA color BMP file. See sample usage.

Sample BMP images with various formats can be found in `/images` directory. The specifications for encoding and decoding are from [https://en.wikipedia.org/wiki/BMP_file_format](https://en.wikipedia.org/wiki/BMP_file_format). The code within this repo is easy to read and modify to suit your needs.

## Installation

Install with `go get -u github.com/dvertx/bmp-go` or by manually cloning this repository into `$GOPATH/src/github.com/dvertx/`

## Sample Usage

Test both, decoding and encoding, with:

```go
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image/png"
	"os"

	"github.com/dvertx/bmp-go"
)

func main() {
	f, err := os.Open("anyold.bmp")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer f.Close()

	img, err := bmp.Decode(f)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	pngbuffer := new(bytes.Buffer)
	pngwriter := bufio.NewWriter(pngbuffer)
	png.Encode(pngwriter, img)
	pngwriter.Flush()
	err = os.WriteFile("test.png", pngbuffer.Bytes(), 0644)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	out, err := os.Create("newfile.bmp")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer out.Close()

	err = bmp.Encode(out, img)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}
```
