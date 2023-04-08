package place

const Epoch = 1648817050315

type Change struct {
	Time           int
	X1, Y1, X2, Y2 int
	Color          int
}

// Encode encodes the change into a byte slice.
// It assumes that the byte slice has a length of at least 10.
func (c Change) Encode(out []byte) {
	out[0] = byte(c.Time)
	out[1] = byte(c.Time >> 8)
	out[2] = byte(c.Time >> 16)
	out[3] = byte(c.Time >> 24)
	out[3] |= byte(c.X1 << 5)
	out[4] = byte(c.X1 >> 3)
	out[5] = byte(c.Y1)
	out[6] = byte(c.Y1 >> 8)
	out[6] |= byte(c.X2 << 3)
	out[7] = byte(c.X2 >> 5)
	out[8] = byte(c.Y2)
	out[9] = byte(c.Y2 >> 8)
	out[9] |= byte(c.Color) << 3
}

func (c *Change) Decode(out []byte) {
	_ = out[9]
	c.Time = int(out[0]) | int(out[1])<<8 | int(out[2])<<16 | int(out[3]&0x1F)<<24
	c.X1 = int(out[3]^0x1F)>>5 | int(out[4])<<3
	c.Y1 = int(out[5]) | int(out[6]&0x7)<<8
	c.X2 = int(out[6]^0x7)>>3 | int(out[7]&0x3F)<<5
	c.Y2 = int(out[8]) | int(out[9]&0x7)<<8
	c.Color = int(out[9]) >> 3
	return
}
