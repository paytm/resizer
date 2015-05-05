package resized

import (
  "errors"
)

func EncodeWebp(ibuf []byte,wquality float32) ( []byte, error) {
  err := errors.New("webp resize unsupported on darwin")
  return nil,err
}
