package processor

import (
	"bytes"
	"image/jpeg"

	"github.com/disintegration/imaging"
)

func CompressImage(file []byte) ([]byte, error) {

	src, err := imaging.Decode(bytes.NewReader(file))

	if err != nil {
		return nil, err
	}

	dst := imaging.Resize(src, 800, 0, imaging.Lanczos)

	buf := new(bytes.Buffer)
	if err = jpeg.Encode(buf, dst, &jpeg.Options{Quality: 75}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
