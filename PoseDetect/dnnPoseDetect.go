//menjalankan; go run dnnPoseDetect.go 0 pose_iter_440000.caffemodel openpose_pose_coco.prototxt openvino fp16

package main

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"gocv.io/x/gocv"
)

var net *gocv.Net
var images chan *gocv.Mat
var poses chan [][]image.Point
var pose [][]image.Point

func main() {
	if len(os.Args) < 4 {
		fmt.Println("How Run")
		return
	}

	deviceID := os.Args[1]
	proto := os.Args[2]
	model := os.Args[3]
	backend := gocv.NetBackendDefault

	if len(os.Args) > 4 {
		backend = gocv.ParseNetBackend(os.Args[4])
	}

	target := gocv.NetTargetCPU
	if len(os.Args) > 5 {
		target = gocv.ParseNetTarget(os.Args[5])
	}

	//Buka Device
	webcam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		fmt.Printf("Tidak bisa membuka device video")
		return
	}
	defer webcam.Close()

	window := gocv.NewWindow("Deep Neural Pose Detection")
	defer window.Close()

	img := gocv.NewMat()
	defer img.Close()

	//Open OpenPose
	n := gocv.ReadNet(model, proto)
	net = &n
	if net.Empty() {
		fmt.Println("Error reading Open Pose")
		return
	}
	defer net.Close()
	net.SetPreferableBackend(gocv.NetBackendType(backend))
	net.SetPreferableTarget(gocv.NetTargetType(target))
	fmt.Printf("Start Reading Device")

	images = make(chan *gocv.Mat, 1)
	poses = make(chan [][]image.Point)

	if ok := webcam.Read(&img); !ok {
		fmt.Printf("Error Read Device")
		return
	}

	processFrame(&img)

	go performDetection()

	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device Closed")
			return
		}
		if img.Empty() {
			continue
		}

		select {
		case pose = <-poses:
			processFrame(&img)

		default:

		}

		drawPose(&img)

		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}

func processFrame(i *gocv.Mat) {
	frame := gocv.NewMat()
	i.CopyTo(&frame)
	images <- &frame

}

func performDetection() {
	for {
		frame := <-images

		blob := gocv.BlobFromImage(*frame, 1.0/255.0, image.Pt(368, 368), gocv.NewScalar(0, 0, 0, 0), false, false)
		net.SetInput(blob, "")
		prob := net.Forward("")

		var midx int
		s := prob.Size()
		nparts, h, w := s[1], s[2], s[3]

		switch nparts {
		case 19:
			// COCO body
			midx = 0
			nparts = 18
		case 16:
			// MPI body
			midx = 1
			nparts = 15
		case 22:
			// hand
			midx = 2
		default:
			fmt.Println("")
			return
		}

		pts := make([]image.Point, 22)
		for i := 0; i < nparts; i++ {
			pts[i] = image.Pt(-1, -1)
			heatmap, _ := prob.FromPtr(h, w, gocv.MatTypeCV32F, 0, i)

			_, maxVal, _, maxLoc := gocv.MinMaxLoc(heatmap)
			if maxVal > 0.1 {
				pts[i] = maxLoc
			}
			heatmap.Close()
		}

		sX := int(float32(frame.Cols()) / float32(w))
		sY := int(float32(frame.Rows()) / float32(h))

		results := [][]image.Point{}
		for _, p := range PosePairs[midx] {
			a := pts[p[0]]
			b := pts[p[1]]

			if a.X <= 0 || a.Y <= 0 || b.X <= 0 || b.Y <= 0 {
				continue
			}

			a.X *= sX
			a.Y *= sY
			b.X *= sX
			b.Y *= sY

			results = append(results, []image.Point{a, b})
		}
		prob.Close()
		blob.Close()
		frame.Close()

		poses <- results
	}
}

func drawPose(frame *gocv.Mat) {
	for _, pts := range pose {
		gocv.Line(frame, pts[0], pts[1], color.RGBA{0, 255, 0, 0}, 2)
		gocv.Circle(frame, pts[0], 3, color.RGBA{0, 0, 200, 0}, -1)
		gocv.Circle(frame, pts[1], 3, color.RGBA{0, 0, 200, 0}, -1)
	}
}

// Selengkapnya https://github.com/CMU-Perceptual-Computing-Lab/openpose/blob/master/doc/output.md
var PosePairs = [3][20][2]int{
	{ // COCO
		{1, 2}, {1, 5}, {2, 3},
		{3, 4}, {5, 6}, {6, 7},
		{1, 8}, {8, 9}, {9, 10},
		{1, 11}, {11, 12}, {12, 13},
		{1, 0}, {0, 14},
		{14, 16}, {0, 15}, {15, 17},
	},
	{ // MPI
		{0, 1}, {1, 2}, {2, 3},
		{3, 4}, {1, 5}, {5, 6},
		{6, 7}, {1, 14}, {14, 8}, {8, 9},
		{9, 10}, {14, 11}, {11, 12}, {12, 13},
	},
	{ // hand
		{0, 1}, {1, 2}, {2, 3}, {3, 4}, // thumb
		{0, 5}, {5, 6}, {6, 7}, {7, 8}, // pinkie
		{0, 9}, {9, 10}, {10, 11}, {11, 12}, // middle
		{0, 13}, {13, 14}, {14, 15}, {15, 16}, // ring
		{0, 17}, {17, 18}, {18, 19}, {19, 20}, // small
	}}
