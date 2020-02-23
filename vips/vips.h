#include <stdlib.h>
#include <vips/vips.h>

VipsImage* image_new_from_buffer(
	const void *buf,
	size_t len,
	const char *option_string
) {
	return vips_image_new_from_buffer(
		buf,
		len,
		option_string,
		NULL
	);
}

int jpegsave_buffer(
	VipsImage *in,
	void **buf,
	size_t *size,
	int quality
) {
	return vips_jpegsave_buffer(
		in,
		buf,
		size,
		"Q", quality,
		"optimize_coding", TRUE,
		NULL
	);
}

int pngsave_buffer(
	VipsImage *in,
	void **buf,
	size_t *size,
	int compression
) {
	return vips_pngsave_buffer(
		in,
		buf,
		size,
		"compression", compression,
		"filter", VIPS_FOREIGN_PNG_FILTER_NONE,
		NULL
	);
}

int webpsave_buffer(
	VipsImage *in,
	void **buf,
	size_t *size,
	int quality,
	int loseless
) {
	return vips_webpsave_buffer(
		in,
		buf,
		size,
		"Q", quality,
		"loseless", loseless,
		NULL
	);
}

int tiffsave_buffer(
	VipsImage *in,
	void **buf,
	size_t *size
) {
	return vips_tiffsave_buffer(
		in,
		buf,
		size,
		NULL
	);
}

int resize(
	VipsImage *in,
	VipsImage **out,
	double xscale,
	double yscale
) {
	return vips_resize(
		in,
		out,
		xscale,
		"vscale", yscale,
		NULL
	);
}

int icc_import(
	VipsImage *in,
    VipsImage **out,
	int intent
) {
	return vips_icc_import(
		in,
		out,
		"intent", intent,
		"embedded", TRUE,
		"pcs", VIPS_PCS_LAB,
		NULL
	);
}

int icc_export(
	VipsImage *in,
    VipsImage **out,
	int intent,
	int depth
) {
	return vips_icc_export(
		in,
		out,
		"intent", intent,
		"depth", depth,
		"pcs", VIPS_PCS_LAB,
		NULL
	);
}

int copy(
	VipsImage *in,
    VipsImage **out
) {
	return vips_copy(in, out, NULL);
}

int profile_load(
	const char *name,
    VipsBlob **profile
) {
	return vips_profile_load(
		name,
		profile,
		NULL
	);
}

int autorot(
	VipsImage *in,
	VipsImage **out
) {
	return vips_autorot(in, out, NULL);
}

gchar **image_get_fields(
	VipsImage *image
) {
	return vips_image_get_fields(image);
}

int image_remove(
	VipsImage *image,
    const char *name
) {
	return (int)vips_image_remove(image, name);
}
