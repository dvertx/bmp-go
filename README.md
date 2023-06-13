# BMP Decoding and Encoding in Go

bmp-go is a Go package aiming to provide abilities to decode BMP image files having 4, 8, 16, 24 and 32 bits color depths. It supports RLE decoding which might be applied as a compression method to 4 bits and 8 bits images.

The code is easy to read and modify to suit your needs. Currently only includes the decoding part. Will update with the encoding part later.  

## Install

Install with `go get -u github.com/dvertx/bmp-go` or by manually cloning this repository into `$GOPATH/src/github.com/dvertx/`

## Sample Usage

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
}
```
