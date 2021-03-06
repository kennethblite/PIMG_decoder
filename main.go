package main

import (
	"bytes"
	"compress/zlib"
	"io/ioutil"
	"fmt"
	"regexp"
	"image"
	"image/color"
	"image/png"
	"flag"
	"log"
	"os"
	"io"
)

func main() {
	input := flag.String("i", "", "Input file")
	output := flag.String("o", "output.png", "Output file")
	isdata := flag.Bool("z", false, "Input file is just a data file instead of zlib")
	help := flag.Bool("h", false, "Help flag")
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}
	if *input == "" {
		fmt.Println("No input file specified (-i=<filename>)")
		return
	}
	buff, err := ioutil.ReadFile(*input)
	if err != nil{
		fmt.Println(err)
		return
	}
	var b io.Reader
	if *isdata{
		b = bytes.NewReader(buff)
	}else{
		b, err = zlib.NewReader(bytes.NewReader(buff))
		if err != nil{
		fmt.Println(err)
		return
		}
	}
	enflated, err := ioutil.ReadAll(b)
	if err != nil {
		fmt.Println(err)
	}
	width, height := int(enflated[0x11])<<8+int(enflated[0x10]), int(enflated[0x13])<<8+int(enflated[0x12])
	encoding :=  enflated[0x14]

	// Create a colored image of the given width and height.
	re, err := regexp.Compile("IDAT")
	if err != nil{
	    fmt.Println(err)
	    return
	}
	re1, err := regexp.Compile("IEND")
	if err != nil{
	    fmt.Println(err)
	    return
	}
	re2, err := regexp.Compile("PIMG")
	if err != nil{
	    fmt.Println(err)
	    return
	}
	idat := re.FindStringIndex(string(enflated))
	iend := re1.FindStringIndex(string(enflated))
	header := re2.FindStringIndex(string(enflated))
	fmt.Println("Height: ",width," Width: ", height)
	increment := 0
	if encoding == 7 {
		fmt.Println("Its RGBA")
		increment = 4
	}else if encoding == 6 {
		fmt.Println("Its RGB")
		increment = 3
	}else{
		fmt.Println("UNKNOWN ENCODING", encoding, "ABORTING")
		return
	}
	var pixel_range []byte
	if len(header) == 0 || header[0] != 0 {
		fmt.Println("NOT A PIMG FILE: ABORTING")
		return
	}
	if len(idat) == 0 {
		fmt.Println("CANNOT FIND IDAT SECTION, DOUBLE CHECK FILE: ABORTING")
		return
	}
	if len(iend) == 0 {
		pixel_range = enflated[idat[1]:]
		fmt.Println("There is no IEND, Image is possibly incomplete: CONTINUING")
	}else{
		pixel_range = enflated[idat[1]:iend[0]]
	}
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for i := idat[1]; i+increment < len(pixel_range); i=i+increment{
		x := ((i-idat[1])/increment)%width
		y := ((i-idat[1])/increment)/height
		var a byte
		if encoding == 7 {
			a = enflated[i+3]
		}else if encoding == 6{
			a = 255
		}
		img.Set(x, y, color.NRGBA{
			R: enflated[i],
			G: enflated[i+1],
			B: enflated[i+2],
			A: a,
		})
	}

	fmt.Println("PARSING COMPLETE, Writing ", *output)
	f, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, img); err != nil {
		f.Close()
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
