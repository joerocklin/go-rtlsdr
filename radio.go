package rtlsdr

/*
	#cgo LDFLAGS: -lrtlsdr -lusb
	#include <stdint.h>
	#include <stdio.h>
  #include <stdlib.h>
	#include <rtl-sdr.h>
*/
import "C"
import (
	"log"
	"sync"
	"unsafe"
)

// Radio is the representation of the RTLSDR device.
type Radio struct {
	PpmError        int
	TunerGainMode   GainMode
	GainOptions     []int
	Gain            int
	AutoGainControl AgcMode

	CenterFrequencyHz uint32
	SampleRate        uint32

	DataBufferSize int

	lock   *sync.RWMutex
	devptr *C.struct_rtlsdr_dev
}

// GainMode for determining what gain type is being used
type GainMode int

const (
	// AutoGain for automatic gain handling
	AutoGain GainMode = iota
	// ManualGain for manual gain handling
	ManualGain
)

// AgcMode for enabling or disabling the digital AGC
type AgcMode int

const (
	// AgcDisabled for disabling it
	AgcDisabled = iota
	// AgcEnabled for enabling it
	AgcEnabled
)

// NewRadio initializes the Radio struct appropriately
// Make sure to call the Cleanup method before exiting!
func NewRadio() *Radio {
	return &Radio{
		PpmError:        0,
		TunerGainMode:   ManualGain,
		AutoGainControl: AgcEnabled,
		DataBufferSize:  256 * 1024,
		lock:            new(sync.RWMutex),
		devptr:          (*C.struct_rtlsdr_dev)(C.calloc(1, 1000)),
	}
}

// Cleanup frees the memory from the C interface appropriately
func (r *Radio) Cleanup() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.CancelAsync()
	log.Print("Freeing Radio memory for C objects")
	C.free(unsafe.Pointer(r.devptr))
}
