package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/jpeg"
	"mygoml"
	"mygoml/kmeans"
	"os"
)

type Uint32Color struct {
	r, g, b uint32
}

func (u Uint32Color) RGBA() (r, g, b, a uint32) {
	return u.r, u.g, u.b, 4294967295
}

type ImagePoint struct {
	x, y  int
	color color.Color
}

func (ip ImagePoint) Features() []float64 {
	r, g, b, _ := ip.color.RGBA()
	return []float64{float64(r), float64(g), float64(b)}
}

type Image struct {
	img image.Image
}

func (im Image) DataPoints() []mygoml.UnsupervisedDataPoint {
	var out []mygoml.UnsupervisedDataPoint
	rect := im.img.Bounds()
	for x := rect.Min.X; x < rect.Max.X; x++ {
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			c := im.img.At(x, y)
			out = append(out, ImagePoint{x, y, c})
		}
	}
	return out
}

func main() {
	imageFile, _ := os.Open("cmd/kmeans_app/image_compression/girl3.jpg")
	defer imageFile.Close()

	img, _, _ := image.Decode(imageFile)
	ds := Image{img}

	for _, count := range []int{3, 5, 10, 15, 20} {
		model := kmeans.Model{ClusterCount: count}
		clusters := model.Clustering(ds)

		rect := img.Bounds()
		newImg := image.NewRGBA(rect)
		for _, c := range clusters {
			rc, _ := c.(*kmeans.Cluster)
			center := rc.Center()
			centerColor := Uint32Color{uint32(center[0]), uint32(center[1]), uint32(center[2])}
			for _, m := range c.Members() {
				rm, _ := m.(ImagePoint)
				newImg.Set(rm.x, rm.y, centerColor)
			}
		}

		newImgFile, _ := os.Create(fmt.Sprintf("cmd/kmeans_app/image_compression/girl3_clustering_final_K%d.jpg", count))
		defer newImgFile.Close()
		jpeg.Encode(newImgFile, newImg, &jpeg.Options{Quality: 80})
	}
}
