package service

import (
	"image"
	"os"
	"strconv"
	"strings"
	"time"
	"errors"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"path/filepath"
	"image-processing/internal/models"
	"github.com/disintegration/imaging"
)

func TransformAndSaveImage(src image.Image, req *models.TransformationRequest, originalPath string) (string, error) {
	dst := src
	// Resize
	if req.Resize != nil {
		dst = imaging.Resize(dst, req.Resize.Width, req.Resize.Height, imaging.Lanczos)
	}

	// Crop
	if req.Crop != nil {
		rect := image.Rect(req.Crop.X, req.Crop.Y, req.Crop.X+req.Crop.Width, req.Crop.Y+req.Crop.Height)
		dst = imaging.Crop(dst, rect)
	}

	// Rotate
	if req.Rotate != nil {
		dst = imaging.Rotate(dst, *req.Rotate, color.Transparent)
	}

	// Flip
	if req.Flip {
		dst = imaging.FlipV(dst)
	}

	// Mirror
	if req.Mirror {
		dst = imaging.FlipH(dst)
	}

	// Filter
	switch strings.ToLower(req.Filter) {
	case "grayscale":
		dst = imaging.Grayscale(dst)
	case "sepia":
		dst = imaging.AdjustSaturation(dst, -100)
		dst = imaging.AdjustContrast(dst, 10)
		dst = imaging.AdjustGamma(dst, 0.9)
	}

	// Watermark
	if req.Watermark != nil {
		watermark := image.NewRGBA(image.Rect(0, 0, 400, 50))
		draw.Draw(watermark, watermark.Bounds(), &image.Uniform{color.RGBA{255, 255, 255, 100}}, image.Point{}, draw.Over)
		dst = imaging.Overlay(dst, watermark, image.Pt(req.Watermark.X, req.Watermark.Y), 1.0)
	}

	// Generate output path
	ext := filepath.Ext(originalPath) // get extension
	base := strings.TrimPrefix(filepath.Base(originalPath), ext) // get filename
	dir := filepath.Dir(originalPath) // 

	timestamp := time.Now().Unix()
	// Change format here
	format := strings.ToLower(req.Format)
	if format == ""{
		format = "jpeg"
	}
	newPath := filepath.Join(dir, base + "_transformed" + format + "_" + strconv.FormatInt(timestamp, 10) + "." + format)

	outFile, err := os.Create(newPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	switch format {
	case "jpeg", "jpg":
		// Default compression is quality 90 (on a scale of 1–100).
		// If user specified a compression value, it's used instead.
		opts := jpeg.Options{Quality: 90}
		if req.Compress != nil {
			opts.Quality = *req.Compress
		}
		err = jpeg.Encode(outFile, dst, &opts)
	case "png":
		err = png.Encode(outFile, dst)
	default:
		return "", errors.New("unsupported format: " + format)
	}

	if err != nil {
		return "", err
	}

	return newPath, nil
}
