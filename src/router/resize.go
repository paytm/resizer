package router

import (
  "github.com/gographics/imagick/imagick"
)

func Resize(width uint, height uint, quality uint, body []byte) ([]byte, error) {
  mw := imagick.NewMagickWand()
  defer mw.Destroy()

  err := mw.ReadImageBlob(body)
  if err != nil {
    panic(err)
  }

  if (width == 0 || height == 0) {
    owidth := mw.GetImageWidth()
    oheight := mw.GetImageHeight()

    if (width == 0 && height == 0) {
      width = owidth
      height = oheight
    } else {
      aspectRatio := (width * 1.0)/height
      if (height == 0) {
        height =  width* (1.0/aspectRatio)
      } else {
        width = height * aspectRatio
      }
    }
  }

  err = mw.ThumbnailImage(width,height)
  if err != nil {
    return nil,err
  }

  err = mw.SetImageCompressionQuality(quality)
  if err != nil {
    return nil,err
  }

  bytes := mw.GetImageBlob()

  if len(bytes) == 0 {
    err = mw.GetLastError()
  }

  return bytes,err
}
