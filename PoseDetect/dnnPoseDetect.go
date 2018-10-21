package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"gocv.io/x/gocv"

)
var net *gocv.Net
var images chan 