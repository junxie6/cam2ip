// +build !cv2,!cv3

// Package camera.
package camera

import (
	"fmt"
	"image"

	"github.com/disintegration/imaging"
	"github.com/korandiz/v4l"
	"github.com/korandiz/v4l/fmt/mjpeg"

	im "github.com/gen2brain/cam2ip/image"
)

// Options.
type Options struct {
	Index  int
	Rotate int
	Width  float64
	Height float64
}

// Camera represents camera.
type Camera struct {
	opts   Options
	camera *v4l.Device
}

// New returns new Camera for given camera index.
func New(opts Options) (camera *Camera, err error) {
	camera = &Camera{}
	camera.opts = opts

	devices := v4l.FindDevices()
	if len(devices) < opts.Index+1 {
		err = fmt.Errorf("camera: no camera at index %d", opts.Index)
		return
	}

	camera.camera, err = v4l.Open(devices[opts.Index].Path)
	if err != nil {
		err = fmt.Errorf("camera: %s", err.Error())
		return
	}

	if camera.camera == nil {
		err = fmt.Errorf("camera: can not open camera %d", opts.Index)
		return
	}

	config, err := camera.camera.GetConfig()
	if err != nil {
		err = fmt.Errorf("camera: %s", err.Error())
		return
	}

	config.Format = mjpeg.FourCC
	config.Width = int(opts.Width)
	config.Height = int(opts.Height)

	err = camera.camera.SetConfig(config)
	if err != nil {
		err = fmt.Errorf("camera: %s", err.Error())
		return
	}

	err = camera.camera.TurnOn()
	if err != nil {
		err = fmt.Errorf("camera: %s", err.Error())
		return
	}

	return
}

// Read reads next frame from camera and returns image.
func (c *Camera) Read() (img image.Image, err error) {

	buffer, err := c.camera.Capture()
	if err != nil {
		err = fmt.Errorf("camera: can not grab frame: %s", err.Error())
		return
	}

	img, err = im.NewDecoder(buffer).Decode()
	if err != nil {
		err = fmt.Errorf("camera: %s", err.Error())
		return
	}

	if c.opts.Rotate == 0 {
		return
	}

	switch c.opts.Rotate {
	case 90:
		img = imaging.Rotate90(img)
	case 180:
		img = imaging.Rotate180(img)
	case 270:
		img = imaging.Rotate270(img)
	}

	return
}

// GetProperty returns the specified camera property.
func (c *Camera) GetProperty(id int) float64 {
	ret, _ := c.camera.GetControl(uint32(id))
	return float64(ret)
}

// SetProperty sets a camera property.
func (c *Camera) SetProperty(id int, value float64) {
	c.camera.SetControl(uint32(id), int32(value))
}

// Close closes camera.
func (c *Camera) Close() (err error) {
	if c.camera == nil {
		err = fmt.Errorf("camera: camera is not opened")
		return
	}

	c.camera.TurnOff()

	c.camera.Close()
	c.camera = nil
	return
}
