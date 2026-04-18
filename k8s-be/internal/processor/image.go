package processor

import (
	"github.com/disintegration/imaging"
)

func CompressImage(inputPath, outputPath string) error {
	src, err := imaging.Open(inputPath)

	if err != nil {
		return err
	}

	dst := imaging.Resize(src, 800, 0, imaging.Lanczos)

	return imaging.Save(dst, outputPath, imaging.JPEGQuality(60))
}
