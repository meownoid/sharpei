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
