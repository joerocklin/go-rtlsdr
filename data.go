package rtlsdr

import "math"

// DataBuffer is the raw data coming off of the device
type DataBuffer struct {
	Size   int
	Length int
	Buffer []byte
}

func newDataBuffer(size int) *DataBuffer {
	return &DataBuffer{
		Size:   size,
		Length: 0,
		Buffer: make([]byte, size),
	}
}

// IqToMag converts I/Q data to magnitudes
func (d *DataBuffer) IqToMag() *MagBuffer {
	mags := NewMagBuffer(d.Length / 2)

	for j := 0; j < d.Length; j += 2 {
		i := d.Buffer[j] - 127
		q := d.Buffer[j+1] - 127
		if i < 0 {
			i = -i
		}
		if q < 0 {
			q = -q
		}
		sumOfSquares := (i * i) + (q * q)
		mag, _ := math.Modf(math.Sqrt(float64(sumOfSquares)))
		mag16 := (Mag)(int16(mag))
		mags.append(mag16)
	}

	return mags
}
