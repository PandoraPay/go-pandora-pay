//based on https://pkg.go.dev/github.com/StellarCN/stellar-identicon-go

package identicon

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
)

const (
	columns = 7
	rows    = 7
)

func GenerateToBytes(key []byte, width, height int) ([]byte, error) {
	img, err := Generate(key, width, height)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err = png.Encode(buf, img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Generates an identicon for the given stellar address
func Generate(keyData []byte, width, height int) (*image.RGBA, error) {

	var matrix [columns][rows]bool

	columnsForCalculation := int(math.Ceil(columns / 2.0))
	for column := 0; column < columnsForCalculation; column++ {
		for row := 0; row < rows; row++ {
			if getBit(column+row*columnsForCalculation, keyData[1:]) {
				matrix[row][column] = true
				matrix[row][columns-column-1] = true
			}
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	foreground := getColor(keyData)
	blockWidth := width / columns
	blockHeight := height / rows

	for row, rowColumns := range matrix {
		for column, cell := range rowColumns {
			if cell {
				x0 := column * blockWidth
				y0 := row * blockHeight
				x1 := (column+1)*blockWidth - 1
				y1 := (row+1)*blockHeight - 1
				for x := x0; x <= x1; x++ {
					for y := y0; y <= y1; y++ {
						img.Set(x, y, foreground)
					}
				}
			}
		}
	}
	return img, nil
}

func getColor(data []byte) color.RGBA {
	c := byte(0)
	for _, it := range data {
		c += it
	}

	hsv := HSV{
		H: float64(c) / 255,
		S: 0.7,
		V: 0.8,
	}
	rgb := hsv.RGB()
	return color.RGBA{
		R: uint8(math.Round(rgb.R * 255)), G: uint8(math.Round(rgb.G * 255)), B: uint8(math.Round(rgb.B * 255)), A: 255,
	}
}

func getBit(n int, keyData []byte) bool {
	if keyData[n/8]>>(8-((n%8)+1))&1 == 1 {
		return true
	}
	return false
}
