//Untuk menjalankan : go run faceDetection.go 0 haarcascade_frontalface_default.xml

package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"strconv"

	"gocv.io/x/gocv"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("How to run:\n\tfacedetect [camera ID] [classifier XML file]")
		return
	}

	deviceID, _ := strconv.Atoi(os.Args[1])
	xmlFile := os.Args[2]

	//Buka Webcam
	webcam, err := gocv.VideoCaptureDevice(int(deviceID))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer webcam.Close()

	//Buka display windows
	window := gocv.NewWindow("Face Detection Using Go Language")
	defer window.Close()

	img := gocv.NewMat()
	defer img.Close()

	blue := color.RGBA{0, 0, 255, 0} //warna untuk tulisan & bingkai

	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	if !classifier.Load(xmlFile) {
		fmt.Printf("Error reading cascade file:%v\n", xmlFile)
		return
	}

	fmt.Printf("start reading camera device : %v\n", deviceID)
	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("cannot read device %d\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		//Mendeteksi wajah
		rects := classifier.DetectMultiScale(img)
		fmt.Printf("found %d faces\n", len(rects))

		for _, r := range rects {
			gocv.Rectangle(&img, r, blue, 3)

			size := gocv.GetTextSize("This is Me", gocv.FontHersheyPlain, 1.2, 2)
			pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
			gocv.PutText(&img, "This me", pt, gocv.FontHersheyPlain, 1.2, blue, 2)
		}

		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}

	}

}
