package rtlsdr

import "math"

// MagnitudeLUT is a lookup table of I/Q -> Magnitude values
var MagnitudeLUT = buildMagLut()

// Mag is a magnitude value
type Mag int16

// MagBuffer is a buffer of the I/Q data converted to magnitudes
type MagBuffer struct {
	Buffer []Mag
}

// NewMagBuffer Consturcts a new magnitude buffer
func NewMagBuffer(size int) *MagBuffer {
	return &MagBuffer{
		Buffer: make([]Mag, size),
	}
}

func buildMagLut() [][]Mag {
	lut := make([][]Mag, 129)
	for i := 0; i <= 128; i++ {
		lut[i] = make([]Mag, 129)
		for q := 0; q <= 128; q++ {
			sumOfSquares := (i * i) + (q * q)
			mag, _ := math.Modf(math.Sqrt(float64(sumOfSquares)))
			lut[i][q] = (Mag)(int16(mag))
		}
	}
	return lut
}

func (m *MagBuffer) append(mag Mag) {
	m.Buffer = append(m.Buffer, mag)
}
