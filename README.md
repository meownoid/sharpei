# Sharpei

Sharpei is a tool for creating thumbnails and easy photo publishing built on top of the [vips](https://github.com/libvips/libvips) library.

- Focus on image quality
- Fast
- Easy and convinient command line interface
- Aware of ICC profiles
- Aware of gray images

## Installation

Vips library must be installed. On Mac OS you can install it via homebrew:

```bash
brew install vips
```

For instructions for other platforms pleasse visit [vips homepage](https://github.com/libvips/libvips).

Now, if you have go toolchain installed:

```bash
go install github.com/meownoid/sharpei
```

## Usage

Simple usage:

```bash
sharpei -width 1024 image.jpg
```

This command will create `image_thumbnail.jpg` in the same directory.

For vertical images you can also specify height:

```bash
sharpei -width 1024 -height 512 image.jpg
```

And ICC profiles:

```bash
sharpei \
    -width 1024 \
    -height 512 \
    -input_profile srgb \
    -output_profile /System/Library/ColorSync/Profiles/AdobeRGB1998.icc \
    image.jpeg
```

Sharpei includes 3 icc profiles appropriate for free distribution: `srgb-v2` (`srgb`), `srgb-v4`, `gray`.

**But wait, there is more!**

You can also create a config in the working directory (`sharpei.yml` or `.sharpei.yml`) or in the home directory (`~/.sharpei.yml`). Config makes more features available than cli.

### Example config
```yaml
output: 'images/'
format: '{name}_{profile}'
rewrite: true

profiles:
    small:
        width: 512
        height: 256
        type: 'png'
        compression: 5
    
    medium:
        width: 1024
        height: 512
        type: 'jpeg'
        quality: 95
    
    large:
        width: 2048
```
