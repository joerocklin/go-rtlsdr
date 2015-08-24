package rtlsdr

/*
	#cgo LDFLAGS: -lrtlsdr -lusb
	#include <stdint.h>
  #include <stdlib.h>
	#include <rtl-sdr.h>
*/
import "C"
import (
	"fmt"
	"log"
	"unsafe"
)

// GetDeviceCount returns a uint of the count of rtlsdr devices
func GetDeviceCount() int {
	return int(C.rtlsdr_get_device_count())
}

// GetDeviceName returns the name of the device
func GetDeviceName(index int) string {
	cidx := C.uint32_t(index)
	devname := C.rtlsdr_get_device_name(cidx)
	return C.GoString(devname)
}

// GetDeviceUsbStrings Gets USB device strings for manufacturer name, product
// name, and serial number if they are available.
func GetDeviceUsbStrings(index int) (string, string, string, error) {
	var manufacturer *C.char
	manufacturer = (*C.char)(C.calloc(1, 256))
	defer C.free(unsafe.Pointer(manufacturer))

	var product *C.char
	product = (*C.char)(C.calloc(1, 256))
	defer C.free(unsafe.Pointer(product))

	var serial *C.char
	serial = (*C.char)(C.calloc(1, 256))
	defer C.free(unsafe.Pointer(serial))

	// Returns 0 on success
	retval := C.rtlsdr_get_device_usb_strings(
		C.uint32_t(index),
		manufacturer,
		product,
		serial)

	if retval != 0 {
		return "", "", "", fmt.Errorf("GetDeviceUsbStrings returned error value: %d", retval)
	}

	return C.GoString(manufacturer), C.GoString(product), C.GoString(serial), nil
}

// Open a handle to the RTLSDR device
func Open(index int) (*Radio, error) {
	dev := NewRadio()

	// Possible return values:
	//  * 0 on success
	//  * -ENOMEM
	// RTLSDR_API int rtlsdr_open(rtlsdr_dev_t **dev, uint32_t index);
	retval := C.rtlsdr_open(&dev.devptr, C.uint32_t(index))

	if retval != 0 {
		defer dev.Cleanup()
		return nil, fmt.Errorf("rtlsdr_open returned error value: %d", retval)
	}

	return dev, nil
}

// Close a handle to the RTLSDR device
func (dev *Radio) Close() error {
	// RTLSDR_API int rtlsdr_close(rtlsdr_dev_t *dev);
	retval := C.rtlsdr_close(dev.devptr)

	if retval != 0 {
		defer dev.Cleanup()
		return fmt.Errorf("rtlsdr_close returned error value: %d", retval)
	}

	return nil
}

// GetTunerGains retrieves and stores the list of available tuner gain options
func (dev *Radio) GetTunerGains() ([]int, error) {
	dev.lock.Lock()
	defer dev.lock.Unlock()

	// The underlying rtlsdr call returns a pointer to an array of ints. The first
	// call figures out how big that array will be
	retval := C.rtlsdr_get_tuner_gains(dev.devptr, nil)
	if retval <= 0 {
		return nil, fmt.Errorf("rtlsdr_get_tuner_gains returned error value: %d", retval)
	}

	cGains := make([]C.int, 100)
	retval = C.rtlsdr_get_tuner_gains(dev.devptr, (*C.int)(unsafe.Pointer(&cGains[0])))
	if retval <= 0 {
		return nil, fmt.Errorf("rtlsdr_get_tuner_gains returned error value: %d", retval)
	}

	for i := 0; i < int(retval); i++ {
		dev.GainOptions = append(dev.GainOptions, int(cGains[i]))
	}

	return dev.GainOptions, nil
}

// SetTunerGainMode Sets the tuner gain mode to manual or automatic
func (dev *Radio) SetTunerGainMode(mode GainMode) error {
	retval := C.rtlsdr_set_tuner_gain_mode(dev.devptr, C.int(mode))

	if retval != 0 {
		return fmt.Errorf("rtlsdr_set_tuner_gain_mode returned error value: %d", retval)
	}
	return nil
}

// SetTunerGain Sets the tuner gain.
// Gain is specified as an int in tenths of dB, 115 means 11.5dB
// See: GetTunerGains to get a list of valid options
func (dev *Radio) SetTunerGain(gain int) error {
	retval := C.rtlsdr_set_tuner_gain(dev.devptr, C.int(gain))

	if retval != 0 {
		return fmt.Errorf("rtlsdr_set_tuner_gain returned error value: %d", retval)
	}
	return nil
}

// GetTunerGain gets the actual gian the device is configured to
func (dev *Radio) GetTunerGain() (int, error) {
	retval := C.rtlsdr_get_tuner_gain(dev.devptr)

	if retval == 0 {
		return 0, fmt.Errorf("rtlsdr_get_tuner_gain encountered an error")
	}
	return int(retval), nil
}

// SetAgcMode Enables or Disables the internal digital AGC of the RTL2832
func (dev *Radio) SetAgcMode(mode AgcMode) error {
	retval := C.rtlsdr_set_agc_mode(dev.devptr, C.int(mode))

	if retval != 0 {
		return fmt.Errorf("rtlsdr_set_agc_mode returned error value: %d", retval)
	}
	return nil
}

// GetCenterFrequency gets the actual frequency the device is tuned to in HZ
func (dev *Radio) GetCenterFrequency() (uint32, error) {
	retval := C.rtlsdr_get_center_freq(dev.devptr)

	if retval == 0 {
		return 0, fmt.Errorf("rtlsdr_get_center_freq encountered an error")
	}
	return uint32(retval), nil
}

// SetCenterFrequency tunes the device in HZ
func (dev *Radio) SetCenterFrequency(freq uint32) error {
	retval := C.rtlsdr_set_center_freq(dev.devptr, C.uint32_t(freq))

	if retval != 0 {
		return fmt.Errorf("rtlsdr_set_center_freq returned error value: %d", retval)
	}
	return nil
}

// GetSampleRate gets the actual sample rate the device is configured for
func (dev *Radio) GetSampleRate() (uint32, error) {
	retval := C.rtlsdr_get_sample_rate(dev.devptr)

	if retval == 0 {
		return 0, fmt.Errorf("rtlsdr_get_sample_rate encountered an error")
	}
	return uint32(retval), nil
}

// SetSampleRate sets the sample rate for the device, also selects the
// baseband filters according to the requested sample rate for tuners
// where this is possible.
// possible values are:
// 		    225001 - 300000 Hz
// 		    900001 - 3200000 Hz
//		    sample loss is to be expected for rates > 2400000
func (dev *Radio) SetSampleRate(rate uint32) error {
	retval := C.rtlsdr_set_sample_rate(dev.devptr, C.uint32_t(rate))

	if retval != 0 {
		return fmt.Errorf("rtlsdr_set_sample_rate returned error value: %d", retval)
	}
	return nil
}

// GetFrequencyCorrection gets the frequency correction value for the device in
// parts per million
func (dev *Radio) GetFrequencyCorrection() int {
	retval := C.rtlsdr_get_freq_correction(dev.devptr)
	return int(retval)
}

// SetFrequencyCorrection sets the frequency correction value for the device in
// parts per million
func (dev *Radio) SetFrequencyCorrection(ppm int) error {
	retval := C.rtlsdr_set_freq_correction(dev.devptr, C.int(ppm))

	if retval != 0 {
		return fmt.Errorf("rtlsdr_set_freq_correction returned error value: %d", retval)
	}
	return nil
}

// =============== Streaming Functions ===============

// ResetBuffer resets the internal buffer on the rtlsdr
func (dev *Radio) ResetBuffer() error {
	retval := C.rtlsdr_reset_buffer(dev.devptr)

	if retval != 0 {
		return fmt.Errorf("rtlsdr_set_freq_correction returned error value: %d", retval)
	}
	return nil
}

// ReadAsync wraps the blocked call to read data asyncrhonously from the rtlsdr
func (dev *Radio) ReadAsync(dataPipe chan *DataBuffer) error {
	go dev.readSyncInternal(dataPipe)
	return nil
}

func (dev *Radio) readSyncInternal(dataPipe chan *DataBuffer) {
	var data C.int

	for {
		buf := newDataBuffer(dev.DataBufferSize)

		retval := C.rtlsdr_read_sync(
			dev.devptr, unsafe.Pointer(&buf.Buffer[0]), C.int(buf.Size), &data,
		)

		if retval != 0 {
			fmt.Printf("rtlsdr_read_sync returned error value: %d", retval)
		} else {
			buf.Length = int(data)
			dataPipe <- buf
		}
	}
}

// CancelAsync cancels all pending asynchronous operations on the device
func (dev *Radio) CancelAsync() error {
	retval := C.rtlsdr_cancel_async(dev.devptr)

	if retval != 0 {
		return fmt.Errorf("rtlsdr_cancel_async returned error value: %d", retval)
	}

	log.Print("Async operations canceled")
	return nil
}
