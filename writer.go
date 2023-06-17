package bmp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"io"
)

func Encode(w io.Writer, img image.Image) error {
	header := &BMPHeader{
		Magic:         0x4D42,
		Reserved1:     0,
		Reserved2:     0,
		Offset:        54, // full BMP header
		DibHeaderSize: 40,
		Planes:        1,
		Bpp:           32,
		Compression:   0,
		Colors:        0,
		ImportantClr:  0,
	}

	bounds := img.Bounds().Size()
	if (bounds.X < 0) || (bounds.Y < 0) {
		err := errors.New("image has negative boundaries")
		return err
	}
	header.Width = int32(bounds.X)
	header.Height = int32(bounds.Y)
	header.ImageSize = uint32(bounds.X) * uint32(bounds.Y) * 4
	header.FileSize = header.ImageSize + 54

	img, ok := img.(*image.RGBA)
	if !ok {
		err := errors.New("image is not an RGBA type")
		return err
	}

	var buf bytes.Buffer

	pix := img.(*image.RGBA).Pix
	width := img.Bounds().Dx() * 4

	for i := len(pix) - width; i >= 0; i -= width {
		for j := i; j < i+width; j += 4 {
			pix[j], pix[j+2] = pix[j+2], pix[j]
		}
		buf.Write(pix[i : i+width])
	}

	data := buf.Bytes()

	err := binary.Write(w, binary.LittleEndian, header)
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, data)
	if err != nil {
		return err
	}

	return nil
}
