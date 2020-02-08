package vips

// #cgo pkg-config: vips
// #include "vips.h"
import "C"
import (
	"errors"
	"io"
	"io/ioutil"
	"runtime"
	"unsafe"
)

func Init(name string) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := C.vips_init(C.CString(name)); err != 0 {
		C.vips_shutdown()
		panic("failed to initialize vips")
	}

	C.vips_concurrency_set(1)
}

func Shutdown() {
	C.vips_shutdown()
}

type Image struct {
	vi *C.VipsImage
}

// Width returns image width, in pixels
func (img *Image) Width() int {
	return int(img.vi.Xsize)
}

// Height returns image height, in pixels
func (img *Image) Height() int {
	return int(img.vi.Ysize)
}

// Bands returns number of image bands
func (img *Image) Bands() int {
	return int(img.vi.Bands)
}

// Format returns pixel format
func (img *Image) Format() int {
	return int(img.vi.BandFmt)
}

// Coding returns pixel coding
func (img *Image) Coding() int {
	return int(img.vi.Coding)
}

// Interpretation returns pixel interpretation
func (img *Image) Interpretation() int {
	return int(img.vi.Type)
}

// XRes returns horizontal pixels per millimetre
func (img *Image) XRes() int {
	return int(img.vi.Xres)
}

// YRes returns vertical pixels per millimetre
func (img *Image) YRes() int {
	return int(img.vi.Yres)
}

// XOffset returns image origin x coordinate, in pixels
func (img *Image) XOffset() int {
	return int(img.vi.Xoffset)
}

// YOffset returns image origin y coordinate, in pixels
func (img *Image) YOffset() int {
	return int(img.vi.Yoffset)
}

// Filename returns original image filename
func (img *Image) Filename() string {
	return C.GoString(img.vi.filename)
}

func (img *Image) MetaString(name string) string {
	var out *C.char
	C.vips_image_get_as_string(
		img.vi,
		C.CString(name),
		&out,
	)

	return C.GoString(out)
}

func (img *Image) ICC() string {
	return img.MetaString("icc-profile-data")
}

func Decode(r io.Reader) (*Image, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	vi := C.image_new_from_buffer(
		unsafe.Pointer(&buf[0]),
		C.size_t(len(buf)),
		C.CString(""),
	)
	if vi == nil {
		return nil, errors.New("vips function image_new_from_buffer returned nil")
	}

	return &Image{vi: vi}, nil
}

func (img *Image) EncodeJPEG(w io.Writer, quality int) error {
	var b unsafe.Pointer
	var s C.size_t

	status := C.jpegsave_buffer(
		img.vi,
		&b,
		&s,
		C.int(quality),
	)

	if status != 0 {
		return errors.New("vips function jpegsave_buffer returned non-zero code")
	}
	defer C.g_free(C.gpointer(b))

	_, err := w.Write((*[1 << 31]byte)(unsafe.Pointer(b))[:s:s])
	if err != nil {
		return err
	}

	return nil
}

func (img *Image) EncodePNG(w io.Writer, compression int) error {
	var b unsafe.Pointer
	var s C.size_t

	status := C.pngsave_buffer(
		img.vi,
		&b,
		&s,
		C.int(compression),
	)

	if status != 0 {
		return errors.New("vips function pngsave_buffer returned non-zero code")
	}
	defer C.g_free(C.gpointer(b))

	_, err := w.Write((*[1 << 31]byte)(unsafe.Pointer(b))[:s:s])
	if err != nil {
		return err
	}

	return nil
}

func (img *Image) EncodeTIFF(w io.Writer) error {
	var b unsafe.Pointer
	var s C.size_t

	status := C.tiffsave_buffer(
		img.vi,
		&b,
		&s,
	)

	if status != 0 {
		return errors.New("vips function tiffsave_buffer returned non-zero code")
	}
	defer C.g_free(C.gpointer(b))

	_, err := w.Write((*[1 << 31]byte)(unsafe.Pointer(b))[:s:s])
	if err != nil {
		return err
	}

	return nil
}

func (img *Image) EncodeWEBP(w io.Writer, quality int, loseless bool) error {
	var b unsafe.Pointer
	var s C.size_t

	status := C.webpsave_buffer(
		img.vi,
		&b,
		&s,
		C.int(quality),
		C.int(btoi(loseless)),
	)

	if status != 0 {
		return errors.New("vips function webpsave_buffer returned non-zero code")
	}
	defer C.g_free(C.gpointer(b))

	_, err := w.Write((*[1 << 31]byte)(unsafe.Pointer(b))[:s:s])
	if err != nil {
		return err
	}

	return nil
}

func btoi(b bool) int {
	if b {
		return 1
	}

	return 0
}
