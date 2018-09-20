package main

//golang.org package download cmd command : go get golang.org/..
import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	// "reflect"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/nfnt/resize"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
)

const (
	// CodeEscape is the numeric code for the escape key
	CodeEscape = 41
	// CodeRightArrow is the numeric code for the right arrow key
	CodeRightArrow = 79
	// CodeLeftArrow is the numeric code for the left arrow key
	CodeLeftArrow = 80
	// CodeDeleteForward is the numeric code for the delete key
	CodeDeleteForward = 76
	// DirRelease needs explanation...
	DirRelease = 2
)

// DecodeImage decodes a single image by its name
func DecodeImage(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m, _, err := image.Decode(f)
	if err != nil {
		fmt.Printf("Unable to decode %s", filename)
		return nil, err
	}
	return m, nil
}

// ChangeImage progresses the slide show
func ChangeImage(ws *screen.Window, index int, maxRect image.Rectangle,
	source *image.RGBA, img []image.Image, buffer *screen.Buffer) int {
	if index >= len(img) {
		index = 0
	} else if index < 0 {
		index = len(img) - 1
	}
	black := color.RGBA{0, 0, 0, 0}
	draw.Draw(source, maxRect.Bounds(), &image.Uniform{black}, image.ZP, 1)
	draw.Draw(source, maxRect.Bounds(), img[index], image.ZP, 1)
	(*ws).Upload(image.ZP, *buffer, maxRect.Bounds())
	(*ws).Publish()

	return index
}

// DeleteImage deletes a single image
func DeleteImage(img []image.Image, index int, path string) ([]image.Image, error) {
	result := []image.Image{}
	switch index {
	case len(img) - 1:
		result = img[:index]
	case 0:
		result = img[1:]
	default:
		result = append(img[:index], img[index+1:]...)
	}
	err := os.Remove(path)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ReadFiles recursively searches an entire directory for all the files in that directory
func ReadFiles(path string) []string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	re := regexp.MustCompile("[.]")
	imgNames := []string{}
	childPath := []string{}
	for _, file := range files {
		if re.MatchString(file.Name()) {
			imgNames = append(imgNames, path+"/"+file.Name())
		} else {
			childPath = append(childPath, file.Name())
			imgNames = append(imgNames, ReadFiles(path+"/"+file.Name())...)
		}
	}
	return imgNames
}

func main() {
	var path string
	fmt.Println("Input path directory : ")
	fmt.Scanln(&path)

	driver.Main(func(s screen.Screen) {

		resizeImg := []image.Image{}
		var w, h int
		imgNames := ReadFiles(path)
		// path := "./demo-image" //Test path directory

		for i, imageName := range imgNames {
			src, err := DecodeImage(imageName)
			if err != nil {
				log.Fatal(err)
			}
			resizeImg = append(resizeImg, resize.Resize(500, 0, src, resize.Lanczos3))
			if w < resizeImg[i].Bounds().Max.X {
				w = resizeImg[i].Bounds().Max.X
			}
			if h < resizeImg[i].Bounds().Max.Y {
				h = resizeImg[i].Bounds().Max.Y
			}
		}

		ws, err := s.NewWindow(nil)
		if err != nil {
			log.Fatal(fmt.Sprintf("Error creating a new window: %v", err))
		}
		defer ws.Release()

		buffer, err := s.NewBuffer(image.Pt(w, h))
		if err != nil {
			log.Fatal(fmt.Sprintf("Error creating a new buffer: %v", err))
		}
		defer buffer.Release()
		maxRect := image.Rect(0, 0, w, h)
		source := buffer.RGBA()
		count := 0
		count = ChangeImage(&ws, count, maxRect, source, resizeImg, &buffer)

		for {
			switch e := ws.NextEvent().(type) {
			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}
			case key.Event:
				if e.Direction == DirRelease {
					switch e.Code {
					case CodeEscape:
						buffer.Release()
						fmt.Println("BYE!")
						return
					case CodeRightArrow:
						count = ChangeImage(&ws, count+1, maxRect, source, resizeImg, &buffer)
					case CodeLeftArrow:
						count = ChangeImage(&ws, count-1, maxRect, source, resizeImg, &buffer)
					case CodeDeleteForward:
						pathName := imgNames[count]
						resizeImg, err = DeleteImage(resizeImg, count, pathName)
						if err != nil {
							log.Fatal(err)
						}
						count = ChangeImage(&ws, count, maxRect, source, resizeImg, &buffer)
						fmt.Println("SUCCESS DELETE")
					}
				}
			}
		}

	})
}
