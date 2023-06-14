package bmp

import (
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"io"
)

type BMPHeader struct {
	Magic         uint16
	FileSize      uint32
	Reserved1     uint16
	Reserved2     uint16
	Offset        uint32
	DibHeaderSize uint32
	Width         int32
	Height        int32
	Planes        uint16
	Bpp           uint16
	Compression   uint32
	ImageSize     uint32
	Xppm          int32
	Yppm          int32
	Colors        uint32
	ImportantClr  uint32
}

type Pixel struct {
	ColorIndex byte
}

type Bitmap struct {
	Width  int32
	Height int32
	Pixels [][]Pixel
}

type Color struct {
	B uint8
	G uint8
	R uint8
	A uint8
}

type ColorTable struct {
	Rows []Color
}

var top2bottom bool
var imageHeight int32

func Decode(r io.Reader) (image.Image, error) {
	header := BMPHeader{}
	err := binary.Read(r, binary.LittleEndian, &header)
	if err != nil {
		return nil, err
	}

	if header.Magic != 0x4D42 {
		err = errors.New("not a valid BMP file")
		return nil, err
	}

	if header.Compression > 2 {
		err = errors.New("BMP compression not supported by this decoder")
		return nil, err
	}

	if header.Width > 32767 || header.Height > 32767 { // > 32KB?
		err = errors.New("image too large")
		return nil, err
	}

	var imgBytes []byte
	chunk := make([]byte, 32768) // 32KB
	for {
		n, err := r.Read(chunk) // read all

		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		imgBytes = append(imgBytes, chunk[:n]...)
	}

	var img image.Image

	imageHeight = header.Height
	top2bottom = false
	if int(header.Height) > 0 && header.Compression == 0 {
		top2bottom = false
	} else {
		top2bottom = true
		if int(header.Height) < 0 {
			imageHeight = 0 - header.Height
		}
	}

	if header.Bpp == 24 || header.Bpp == 32 {
		img = decodeImg(imgBytes, &header)
	} else if header.Bpp == 16 && header.Compression == 0 {
		img = decode16(imgBytes, &header)
	} else if header.Bpp == 8 {
		if int(header.Compression) == 1 {
			bmp, err := decodeRle(imgBytes, &header, 8)
			if err != nil {
				return nil, err
			}
			clrTable := readColorTable(imgBytes, &header)
			img = decodeBmp(bmp, &clrTable)
		} else if int(header.Compression) == 0 {
			clrTable := readColorTable(imgBytes, &header)
			img = decode8(imgBytes, &header, &clrTable)
		} else {
			err = errors.New("bad compression value")
			return nil, err
		}
	} else if header.Bpp == 4 {
		if int(header.Compression) == 2 {
			bmp, err := decodeRle(imgBytes, &header, 4)
			if err != nil {
				return nil, err
			}
			clrTable := readColorTable(imgBytes, &header)
			img = decodeBmp(bmp, &clrTable)
		} else if int(header.Compression) == 0 {
			clrTable := readColorTable(imgBytes, &header)
			img = decode4(imgBytes, &header, &clrTable)
		} else {
			err = errors.New("bad compression value")
			return nil, err
		}
	} else {
		err = errors.New("unsupported BMP format")
		return nil, err
	}

	return img, nil
}

func decodeImg(in []byte, header *BMPHeader) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, int(header.Width), int(imageHeight)))
	byteSize := int32(header.Bpp) / 8
	offset := int32(header.Offset) - 54
	y1 := 0
	var z0, z1, zd int32 = imageHeight - 1, -1, -1
	if top2bottom {
		z0, z1, zd = 0, imageHeight, +1
	}

	for y := z0; y != z1; y += zd {
		for x := 0; x < int(header.Width); x++ {
			i := (y*header.Width+int32(x))*byteSize + offset // pointer to data

			b := in[i]
			g := in[i+1]
			r := in[i+2]

			c := color.RGBA{r, g, b, 0xFF}
			img.Set(x, y1, c)
		}
		y1++
	}

	return img
}

func decode16(in []byte, header *BMPHeader) image.Image {
	var R, G, B byte
	img := image.NewRGBA(image.Rect(0, 0, int(header.Width), int(imageHeight)))
	offset := int32(header.Offset) - 54
	y1 := 0
	var z0, z1, zd int32 = imageHeight - 1, -1, -1
	if top2bottom {
		z0, z1, zd = 0, imageHeight, +1
	}

	for y := z0; y != z1; y += zd {
		x1 := 0
		for x := 0; x < int(header.Width)*2; x++ {
			i := (y*header.Width*2 + int32(x)) + offset // pointer to data
			val := uint16(in[i]) | uint16(in[i+1])<<8
			R = byte(val & 0b0111110000000000 >> 7) // mask and shift into 8 bits
			G = byte(val & 0b1111100000 >> 2)
			B = byte(val & 0b00011111 << 3)

			c := color.RGBA{R, G, B, 0xFF}
			img.Set(x1, y1, c)
			x++
			x1++
		}
		y1++
	}

	return img
}

func decode8(in []byte, header *BMPHeader, table *ColorTable) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, int(header.Width), int(imageHeight)))
	offset := int32(header.Offset) - 54
	y1 := 0
	var z0, z1, zd int32 = imageHeight - 1, -1, -1
	if top2bottom {
		z0, z1, zd = 0, imageHeight, +1
	}

	for y := z0; y != z1; y += zd {
		for x := 0; x < int(header.Width); x++ {
			i := (y*header.Width + int32(x)) + offset // pointer to data
			val := in[i]                              // value of color index

			b := table.Rows[val].B
			g := table.Rows[val].G
			r := table.Rows[val].R

			c := color.RGBA{r, g, b, 0xFF}
			img.Set(x, y1, c)
		}
		y1++
	}

	return img
}

func decode4(in []byte, header *BMPHeader, table *ColorTable) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, int(header.Width), int(imageHeight)))
	offset := int32(header.Offset) - 54
	y1 := 0
	var z0, z1, zd int32 = imageHeight - 1, -1, -1
	if top2bottom {
		z0, z1, zd = 0, imageHeight, +1
	}

	for y := z0; y != z1; y += zd {
		x1 := 0
		for x := 0; x < int(header.Width)/2; x++ {
			i := (y*header.Width/2 + int32(x)) + offset // pointer to data
			val := in[i]                                // value of color index
			v1 := val >> 4
			v2 := val & 0x0F

			b := table.Rows[v1].B
			g := table.Rows[v1].G
			r := table.Rows[v1].R
			c := color.RGBA{r, g, b, 0xFF}
			img.Set(x1, y1, c)
			x1++

			b = table.Rows[v2].B
			g = table.Rows[v2].G
			r = table.Rows[v2].R
			c = color.RGBA{r, g, b, 0xFF}
			img.Set(x1, y1, c)
			x1++
		}
		y1++
	}

	return img
}

func decodeRle(in []byte, header *BMPHeader, bits int) (*Bitmap, error) {
	if (bits != 4) && (bits != 8) {
		err := errors.New("bad RLE bits value")
		return nil, err
	}

	width := header.Width
	height := header.Height

	bmp := &Bitmap{
		Width:  width,
		Height: height,
		Pixels: make([][]Pixel, height),
	}

	for i := range bmp.Pixels {
		bmp.Pixels[i] = make([]Pixel, width)
	}

	offset := int32(header.Offset) - 54 // full BMP header
	ptr := offset
	var x int32
	var eol bool
	var b, b1, b2, dx, dy byte

	var z0, z1, zd int32 = imageHeight - 1, -1, -1
	if top2bottom {
		z0, z1, zd = 0, imageHeight, +1
	}

	for y := z0; y != z1; y += zd {
		x = 0
		eol = false

		for !eol {
			b1 = in[ptr]
			b2 = in[ptr+1]
			ptr += 2

			if b1 == 0 {
				switch b2 {
				case 0:
					eol = true
				case 1:
					return bmp, nil
				case 2:
					dx = in[ptr]
					dy = in[ptr+1]
					ptr += 2

					x += int32(dx)
					y -= int32(dy)
				default:
					// Absolute mode
					n := int(b2)

					for i := 0; i < n; i++ {
						b = in[ptr]
						ptr++

						if bits == 4 {
							if i%2 == 0 {
								bmp.Pixels[y][x] = Pixel{ColorIndex: b >> 4}
							} else {
								bmp.Pixels[y][x] = Pixel{ColorIndex: b & 0x0F}
							}
						} else {
							bmp.Pixels[y][x] = Pixel{ColorIndex: b}
						}

						x++
						if x >= width {
							break
						}
					}

					if n%2 == 1 {
						ptr++ // skip odd padding
					}
				}
			} else {
				// Encoded mode
				n := int(b1)
				c := b2
				c1 := b2 >> 4
				c2 := b2 & 0x0F

				for i := 0; i < n; i++ {
					if bits == 4 {
						if i%2 == 0 {
							bmp.Pixels[y][x] = Pixel{ColorIndex: c1}
						} else {
							bmp.Pixels[y][x] = Pixel{ColorIndex: c2}
						}
					} else {
						bmp.Pixels[y][x] = Pixel{ColorIndex: c}
					}

					x++
					if x >= width {
						break
					}
				}
			}
		}
	}

	return bmp, nil
}

func readColorTable(in []byte, header *BMPHeader) ColorTable {
	clrTable := &ColorTable{
		Rows: make([]Color, int(header.Colors)),
	}

	byteSize := 4
	offset := header.DibHeaderSize - 40 // DIB header

	for i := int(header.Colors) - 1; i >= 0; i-- {
		ptr := i*byteSize + int(offset)
		clrTable.Rows[i].R = in[ptr+2]
		clrTable.Rows[i].G = in[ptr+1]
		clrTable.Rows[i].B = in[ptr]
		clrTable.Rows[i].A = in[ptr+3]
	}

	return *clrTable
}

func decodeBmp(bmp *Bitmap, table *ColorTable) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, int(bmp.Width), int(bmp.Height)))
	for j := 0; j < int(bmp.Height); j++ {
		for i := 0; i < int(bmp.Width); i++ {
			k := bmp.Pixels[j][i].ColorIndex
			c := color.RGBA{table.Rows[int(k)].R, table.Rows[int(k)].G, table.Rows[int(k)].B, 0xFF}
			img.Set(i, j, c)
		}
	}

	return img
}
