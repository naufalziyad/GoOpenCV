//menjalankan go run tracking.go 0
package main

import (
	"fmt"
	"image/color"
	"os"

	"gocv.io/x/gocv"
	"gocv.io/x/gocv/contrib"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("How to run:\n\ttracking [camera ID]")
		return
	}

	deviceID := os.Args[1]

	// open webcam
	webcam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		fmt.Printf("Error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	window := gocv.NewWindow("Tracking")
	defer window.Close()

	tracker := contrib.NewTrackerMOSSE()
	defer tracker.Close()

	img := gocv.NewMat()
	defer img.Close()

	if ok := webcam.Read(&img); !ok {
		fmt.Printf("cannot read device %v\n", deviceID)
		return
	}

	rect := gocv.SelectROI("Tracking", img)
	if rect.Max.X == 0 {
		fmt.Printf("user cancelled roi selection\n")
		return
	}

	init := tracker.Init(img, rect)
	if !init {
		fmt.Printf("Tidak Bisa Menginisiasi")
		return
	}

	// color for the rect to draw
	blue := color.RGBA{0, 0, 255, 0}
	fmt.Printf("Start reading device: %v\n", deviceID)
	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device closed: %v\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		rect, _ := tracker.Update(img)

		gocv.Rectangle(&img, rect, blue, 3)

		window.IMShow(img)
		if window.WaitKey(10) >= 0 {
			break
		}
	}
}
