//menjalankan go run handGesture.go 0

package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"

	"gocv.io/x/gocv"
)

const MinimunArea = 3000

func main() {
	if len(os.Args) < 2 {
		fmt.Println("How to run:\n\thand-gestures [camera ID]")
		return
	}

	deviceID := os.Args[1]

	webcam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		fmt.Printf("Error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	window := gocv.NewWindow("Gestur Tangan")
	defer window.Close()

	img := gocv.NewMat()
	defer img.Close()

	imgGrey := gocv.NewMat()
	defer imgGrey.Close()

	imgBlur := gocv.NewMat()
	defer imgBlur.Close()

	imgThresh := gocv.NewMat()
	defer imgThresh.Close()

	hull := gocv.NewMat()
	defer hull.Close()

	defects := gocv.NewMat()
	defer defects.Close()

	// rgba Warna Merah untuk penulisan status
	red := color.RGBA{255, 0, 0, 0}

	fmt.Printf("Star reading device: %v\n", deviceID)
	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device Closed")
			return
		}
		if img.Empty() {
			continue
		}

		gocv.CvtColor(img, &imgGrey, gocv.ColorBGRToGray)
		gocv.GaussianBlur(imgGrey, &imgBlur, image.Pt(35, 35), 0, 0, gocv.BorderDefault)
		gocv.Threshold(imgBlur, &imgThresh, 0, 255, gocv.ThresholdBinaryInv+gocv.ThresholdOtsu)

		contours := gocv.FindContours(imgThresh, gocv.RetrievalExternal, gocv.ChainApproxSimple)
		c := getBiggestContour(contours)

		gocv.ConvexHull(c, &hull, true, false)
		gocv.ConvexityDefects(c, hull, &defects)

		var angle float64
		defectCount := 0
		for i := 0; i < defects.Rows(); i++ {
			start := c[defects.GetIntAt(i, 0)]
			end := c[defects.GetIntAt(i, 1)]
			far := c[defects.GetIntAt(i, 2)]

			a := math.Sqrt(math.Pow(float64(end.X-start.X), 2) + math.Pow(float64(end.Y-start.Y), 2))
			b := math.Sqrt(math.Pow(float64(far.X-start.X), 2) + math.Pow(float64(far.Y-start.Y), 2))
			c := math.Sqrt(math.Pow(float64(end.X-far.X), 2) + math.Pow(float64(end.Y-far.Y), 2))

			angle = math.Acos((math.Pow(b, 2)+math.Pow(c, 2)-math.Pow(a, 2))/(2*b*c)) * 57
			if angle <= 90 {
				defectCount++
				gocv.Circle(&img, far, 1, red, 2)
			}
		}

		hasil := fmt.Sprintf("TERDETEKSI: %d", defectCount)

		rect := gocv.BoundingRect(c)
		gocv.Rectangle(&img, rect, color.RGBA{255, 255, 255, 0}, 2)

		//setting untuk tulisan hasil
		gocv.PutText(&img, hasil, image.Pt(100, 100), gocv.FontHersheyPlain, 3.2, red, 2)

		window.IMShow(img)
		if window.WaitKey(1) == 27 {
			break
		}
	}
}

func getBiggestContour(contours [][]image.Point) []image.Point {
	var area float64
	index := 0
	for i, c := range contours {
		newArea := gocv.ContourArea(c)
		if newArea > area {
			area = newArea
			index = i
		}
	}
	return contours[index]
}
