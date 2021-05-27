package identicon

import "math"

type HSV struct {
	H, S, V float64
}

type RGB struct {
	R, G, B float64
}

func (c *HSV) RGB() *RGB {
	var r, g, b float64
	if c.S == 0 {
		return &RGB{
			R: c.V,
			G: c.V,
			B: c.V,
		}
	}

	i := math.Floor(c.H * 6.0)
	f := c.H*6.0 - i
	p := c.V * (1 - c.S)
	q := c.V * (1 - c.S*f)
	t := c.V * (1 - c.S*(1-f))

	switch int(i) % 6 {
	case 0:
		r = c.V
		g = t
		b = p
	case 1:
		r = q
		g = c.V
		b = p
	case 2:
		r = p
		g = c.V
		b = t
	case 3:
		r = p
		g = q
		b = c.V
	case 4:
		r = t
		g = p
		b = c.V
	case 5:
		r = c.V
		g = p
		b = q
	}

	rgb := &RGB{r, g, b}
	return rgb
}
