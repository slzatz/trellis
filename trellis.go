package trellis

// HT16K33 Command Contstants
const (
	HT16K33_OSCILATOR_ON    = 0x21
	HT16K33_BLINK_CMD       = 0x80
	HT16K33_BLINK_DISPLAYON = 0x01
	HT16K33_CMD_BRIGHTNESS  = 0xE0
	HT16K33_KEY_READ_CMD    = 0x40
)

// LED Lookup Table

var  (
	ledLUT = [16]int{
     0x3A, 0x37, 0x35, 0x34,
      0x28, 0x29, 0x23, 0x24,
      0x16, 0x1B, 0x11, 0x10,
      0x0E, 0x0D, 0x0C, 0x02 }

  buttonLUT = [16]int{
    0x07, 0x04, 0x02, 0x22,
      0x05, 0x06, 0x00, 0x01,
      0x03, 0x10, 0x30, 0x21,
      0x13, 0x12, 0x11, 0x31 }
		)
var (
	ledLUT = [...]int{
		0x3A,
		0x37,
		0x35,
		0x34,
		0x28,
		0x29,
		0x23,
		0x24,
		0x16,
		0x1B,
		0x11,
		0x10,
		0x0E,
		0x0D,
		0x0C,
		0x02,
	}

	// Button Loookup Table
	buttonLUT = [...]int{
		0x07,
		0x04,
		0x02,
		0x22,
		0x05,
		0x06,
		0x00,
		0x01,
		0x03,
		0x10,
		0x30,
		0x21,
		0x13,
		0x12,
		0x11,
		0x31,
	}
)

type Trellis struct{}

func New(bus drivers.I2C) Trellis {
	return Trellis{}
}

func (t *Trellis) begin(addr uint8) {
	i2c.Tx(addr, []byte{0x21})


func (tl *TrellisLEDs) get(x int) (bool, error) {
	if 0 < x >= numLEDs {
		return false, Error
	}

	led := ledLUT[x %16] >> 4
	mask := 1 << (letLUT[x %16] & 0x0f)
	return bool(
		( 
			(
				ledBuffer[x / 16][led * 2) + 1 | ledBuffer[x / 16][(led *2) + 2] << 8
			)
			& mask
		)
		> 0
	), nil
}

