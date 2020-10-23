# Sharpei

Sharpei is a command line tool for creating thumbnails and easy photo 
publishing built on top of the [vips](https://github.com/libvips/libvips) library.

Key features:

- focuses on image quality;
- fast;
- provides easy and convenient command line interface;
- aware of ICC profiles;
- aware of grayscale images.

## Installation

Vips library must be installed first. On the macOS you can install it via the homebrew:

```shell script
brew install vips
```

On the Debian or Ubuntu you can install it via the apt:

```shell script
sudo apt install libvips libvips-dev
```

For instructions for other platforms please visit 
the [vips homepage](https://github.com/libvips/libvips).

After that you can install the sharpei:

```shell script
go install github.com/meownoid/sharpei
```

## Usage

For simple usage specify only one parameter `width`. 
Image will be scaled proportionally to the specified width.

```shell script
sharpei -width 1024 image.jpg
```

This command will create `image_thumbnail.jpg` in the same directory.

For vertical images you can also specify height. Maximal dimension will be used. 
Horizontal images will be scaled to the specified width and vertical images
will be scaled to the specified height.

```shell script
sharpei -width 1024 -height 512 image.jpg
```

You can also specify input and output ICC profiles.

```shell script
sharpei \
    -width 1024 \
    -height 512 \
    -input_profile srgb \
    -output_profile /System/Library/ColorSync/Profiles/AdobeRGB1998.icc \
    image.jpeg
```

Sharpei includes 3 ICC profiles appropriate for free distribution: `srgb-v2` (`srgb`), `srgb-v4`, `gray`.

## Command line arguments

* `-config` – path to the config
* `-output` – output directory
* `-format` – format of the output file name (for example: `{name}_transformed`)
* `-rewrite` – use it to rewrite existing files
* `-width` – image width
* `-height` – image height
* `-input-profile` – input ICC profile (name or path)
* `-output-profile` – output ICC profile (name or path)
* `-no-color` – disable colorized terminal output

**But wait, there is more!**

## Configuration file

You can also create a config in the working directory (`sharpei.yml` or `.sharpei.yml`) 
or in the home directory (`~/.sharpei.yml`). There are more options available with 
the configuration file.

For example, you can create different profiles and transform each image to 
multiple thumbnails with different sizes or ICC profiles.

### Example configuration

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
