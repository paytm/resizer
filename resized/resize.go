package resized

import (
	"gopkg.in/gographics/imagick.v2/imagick"
)

func Resize(width uint, height uint, quality uint, body []byte) ([]byte, error) {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	err := mw.ReadImageBlob(body)
	if err != nil {
		return nil, err
	}

	if width == 0 || height == 0 {
		owidth := mw.GetImageWidth()
		oheight := mw.GetImageHeight()

		if width == 0 && height == 0 {
			width = owidth
			height = oheight
		} else {
			var aspectRatio float64
			aspectRatio = float64(owidth) / float64(oheight)
			if height == 0 {
				height = uint(float64(width) / aspectRatio)
			} else {
				width = uint(float64(height) * aspectRatio)
			}
		}
	}

	err = mw.ThumbnailImage(width, height)
	if err != nil {
		return nil, err
	}

	err = mw.SetImageCompressionQuality(quality)
	if err != nil {
		return nil, err
	}

	bytes := mw.GetImageBlob()

	if len(bytes) == 0 {
		err = mw.GetLastError()
	}

	return bytes, err
}
