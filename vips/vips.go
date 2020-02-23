package vips

// #cgo pkg-config: vips
// #include "vips.h"
import "C"
import (
	"errors"
	"fmt"
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

func getError(name string) string {
	defer C.vips_error_clear()

	maybeError := C.GoString(C.vips_error_buffer())

	if len(maybeError) > 0 {
		return maybeError
	}

	return fmt.Sprintf("unknown error in vips function %s", name)
}

type Image struct {
	vi *C.VipsImage
}

const (
	INTERPRETATION_ERROR     = int(C.VIPS_INTERPRETATION_ERROR)
	INTERPRETATION_MULTIBAND = int(C.VIPS_INTERPRETATION_MULTIBAND)
	INTERPRETATION_B_W       = int(C.VIPS_INTERPRETATION_B_W)
	INTERPRETATION_HISTOGRAM = int(C.VIPS_INTERPRETATION_HISTOGRAM)
	INTERPRETATION_XYZ       = int(C.VIPS_INTERPRETATION_XYZ)
	INTERPRETATION_LAB       = int(C.VIPS_INTERPRETATION_LAB)
	INTERPRETATION_CMYK      = int(C.VIPS_INTERPRETATION_CMYK)
	INTERPRETATION_LABQ      = int(C.VIPS_INTERPRETATION_LABQ)
	INTERPRETATION_RGB       = int(C.VIPS_INTERPRETATION_RGB)
	INTERPRETATION_CMC       = int(C.VIPS_INTERPRETATION_CMC)
	INTERPRETATION_LCH       = int(C.VIPS_INTERPRETATION_LCH)
	INTERPRETATION_LABS      = int(C.VIPS_INTERPRETATION_LABS)
	INTERPRETATION_sRGB      = int(C.VIPS_INTERPRETATION_sRGB)
	INTERPRETATION_YXY       = int(C.VIPS_INTERPRETATION_YXY)
	INTERPRETATION_FOURIER   = int(C.VIPS_INTERPRETATION_FOURIER)
	INTERPRETATION_RGB16     = int(C.VIPS_INTERPRETATION_RGB16)
	INTERPRETATION_GREY16    = int(C.VIPS_INTERPRETATION_GREY16)
	INTERPRETATION_MATRIX    = int(C.VIPS_INTERPRETATION_MATRIX)
	INTERPRETATION_scRGB     = int(C.VIPS_INTERPRETATION_scRGB)
	INTERPRETATION_HSV       = int(C.VIPS_INTERPRETATION_HSV)
	INTERPRETATION_LAST      = int(C.VIPS_INTERPRETATION_LAST)
)

const (
	INTENT_PERCEPTUAL = int(C.VIPS_INTENT_PERCEPTUAL)
	INTENT_RELATIVE   = int(C.VIPS_INTENT_RELATIVE)
	INTENT_SATURATION = int(C.VIPS_INTENT_SATURATION)
	INTENT_ABSOLUTE   = int(C.VIPS_INTENT_ABSOLUTE)
	INTENT_LAST       = int(C.VIPS_INTENT_LAST)
)

func (img *Image) Copy() (*Image, error) {
	var out *C.VipsImage

	if s := C.copy(img.vi, &out); s != 0 {
		return nil, errors.New(getError("copy"))
	}

	return &Image{vi: out}, nil
}

func (img *Image) Destroy() {
	C.g_object_unref(C.gpointer(img.vi))
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

// IsPropertySet returns true if property with that name is set on the image, false otherwise
func (img *Image) IsPropertySet(name string) bool {
	return C.vips_image_get_typeof(img.vi, C.CString(name)) != 0
}

// Properties returns list of names of all properties of an image
func (img *Image) Properties() []string {
	fields := C.image_get_fields(img.vi)
	defer C.g_strfreev(fields)

	fieldsSlice := (*[1 << 31]*C.char)(unsafe.Pointer(fields))

	result := make([]string, 0, 16)

	for _, field := range fieldsSlice {
		if field == nil {
			break
		}

		result = append(result, C.GoString(field))
	}

	return result
}

func (img *Image) RemoveProperty(name string) error {
	status := C.image_remove(img.vi, C.CString(name))

	if status == 0 {
		return fmt.Errorf("no metadata with name %s", name)
	}

	return nil
}

// PropertyString returns string value of the property with given name
func (img *Image) PropertyString(name string) string {
	var out *C.char
	C.vips_image_get_as_string(
		img.vi,
		C.CString(name),
		&out,
	)

	return C.GoString(out)
}

func (img *Image) SetPropertyBlob(name string, data []byte) {
	C.vips_image_set_blob_copy(
		img.vi,
		C.CString(name),
		unsafe.Pointer(&data[0]),
		C.size_t(len(data)),
	)
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
		return nil, errors.New(getError("image_new_from_buffer"))
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
		return errors.New(getError("jpegsave_buffer"))
	}
	defer C.g_free(C.gpointer(b))

	_, err := w.Write(view(b, int(s)))
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
		return errors.New(getError("pngsave_buffer"))
	}
	defer C.g_free(C.gpointer(b))

	_, err := w.Write(view(b, int(s)))
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
		return errors.New(getError("tiffsave_buffer"))
	}
	defer C.g_free(C.gpointer(b))

	_, err := w.Write(view(b, int(s)))
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
		return errors.New(getError("webpsave_buffer"))
	}
	defer C.g_free(C.gpointer(b))

	_, err := w.Write(view(b, int(s)))
	if err != nil {
		return err
	}

	return nil
}

func (img *Image) Resize(xscale float64, yscale float64) (*Image, error) {
	var out *C.VipsImage

	status := C.resize(
		img.vi,
		&out,
		C.double(xscale),
		C.double(yscale),
	)

	if status != 0 {
		return nil, errors.New(getError("resize"))
	}

	return &Image{vi: out}, nil
}

func (img *Image) ICCImport(intent int) (*Image, error) {
	var out *C.VipsImage

	status := C.icc_import(
		img.vi,
		&out,
		C.int(intent),
	)

	if status != 0 {
		return nil, errors.New(getError("icc_import"))
	}

	return &Image{vi: out}, nil
}

func (img *Image) ICCExport(intent int, depth int) (*Image, error) {
	var out *C.VipsImage

	status := C.icc_export(
		img.vi,
		&out,
		C.int(intent),
		C.int(depth),
	)

	if status != 0 {
		return nil, errors.New(getError("icc_export"))
	}

	return &Image{vi: out}, nil
}

func (img *Image) Autorot() (*Image, error) {
	var out *C.VipsImage

	status := C.autorot(
		img.vi,
		&out,
	)

	if status != 0 {
		return nil, errors.New(getError("autorot"))
	}

	return &Image{vi: out}, nil
}

func LoadProfile(name string) ([]byte, error) {
	var profileBlob *C.VipsBlob
	status := C.profile_load(
		C.CString(name),
		&profileBlob,
	)

	if status != 0 {
		return nil, errors.New(getError("profile_load"))
	}

	var length C.size_t
	ptr := C.vips_blob_get(
		profileBlob,
		&length,
	)

	return view(ptr, int(length)), nil
}

func btoi(b bool) int {
	if b {
		return 1
	}

	return 0
}

func view(ptr unsafe.Pointer, length int) []byte {
	return (*[1 << 31]byte)(ptr)[:length:length]
}
