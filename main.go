package main

import (
	"encoding/csv"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

var (
	backgroundWidth  = 650
	backgroundHeight = 150
	utf8FontFile     = "arial.ttf"
	utf8FontSize     = float64(24.0)
	spacing          = float64(1.5)
	dpi              = float64(72)
	ctx              = new(freetype.Context)
	utf8Font         = new(truetype.Font)
	black            = color.RGBA{0, 0, 0, 255}
	// more color at https://github.com/golang/image/blob/master/colornames/table.go
)

var data = make(map[string]string)

func Chunks(s string, chunkSize int) []string {
	if len(s) == 0 {
		return nil
	}
	if chunkSize >= len(s) {
		return []string{s}
	}
	var chunks []string = make([]string, 0, (len(s)-1)/chunkSize+1)
	currentLen := 0
	currentStart := 0
	for i := range s {
		if currentLen == chunkSize {
			chunks = append(chunks, s[currentStart:i])
			currentLen = 0
			currentStart = i
		}
		currentLen++
	}
	chunks = append(chunks, s[currentStart:])
	return chunks
}

func main() {
	var filename string
	fmt.Println("Укажите имя файла .csv")
	fmt.Scanf("%s\n", &filename)

	csv := readCsvFile(filename)
	if err := os.Mkdir(fmt.Sprintf("Result-%s", filename), os.ModePerm); err != nil {
		log.Fatal(err)
	}

	for j := 0; j < len(csv); j++ {
		data[csv[j][0]] = csv[j][1]
	}
	for key, val := range data {
		path := fmt.Sprintf("./%s/%s.png", fmt.Sprintf("Result-%s", filename), val)
		barCode, _ := code128.EncodeWithoutChecksum(val)
		barCode, _ = barcode.Scale(barCode, 600, 400)
		img := createImageFromBarCode(barCode)
		addLabel(img, 0, 400, Chunks(key, 50))

		file, _ := os.Create(path)

		defer file.Close()

		png.Encode(file, img)
	}
}

func readCsvFile(filePath string) [][]string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}
	return records
}

func createImageFromBarCode(code barcode.Barcode) *image.RGBA {
	b := code.Bounds()
	img := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()+100))
	draw.Draw(img, img.Bounds(), code, b.Min, draw.Src)
	return img
}

func addLabel(img *image.RGBA, x, y int, label []string) {
	fontBytes, err := ioutil.ReadFile(utf8FontFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	utf8Font, err = freetype.ParseFont(fontBytes)
	if err != nil {
		fmt.Println(err)
		return
	}

	fontForeGroundColor := image.NewUniform(black)

	//draw.Draw(img, img.Bounds(), nil, image.ZP, draw.Src)

	ctx = freetype.NewContext()
	ctx.SetDPI(dpi) //screen resolution in Dots Per Inch
	ctx.SetFont(utf8Font)
	ctx.SetFontSize(utf8FontSize) //font size in points
	ctx.SetClip(img.Bounds())
	ctx.SetDst(img)
	ctx.SetSrc(fontForeGroundColor)

	// Draw the text to the background
	pt := freetype.Pt(x, y+int(ctx.PointToFixed(utf8FontSize)>>6))

	// not all utf8 fonts are supported by wqy-zenhei.ttf
	// use your own language true type font file if your language cannot be printed

	for _, str := range label {
		_, err := ctx.DrawString(str, pt)
		if err != nil {
			fmt.Println(err)
			return
		}
		pt.Y += ctx.PointToFixed(utf8FontSize * spacing)
	}
}
