package resized

import (
  "github.com/chai2010/webp"
  "bytes"
  "image"
  _ "image/jpeg"
  _ "image/png"
)

func EncodeWebp(ibuf []byte,wquality float32) ( []byte, error) {
  var data bytes.Buffer
  var obuf [] byte
  src := bytes.NewBuffer(ibuf)

  img,_,err := image.Decode(src)

  if err == nil {
    if webp.Encode(&data, img, &webp.Options{ false, wquality }); err == nil {
      obuf = data.Bytes()
    }
  }

  return obuf,err
}
