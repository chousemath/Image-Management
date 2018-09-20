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
func ChangeImage(
	ws *screen.Window,
	index int,
	maxRect image.Rectangle,
	source *image.RGBA,
	images []image.Image,
	buffer *screen.Buffer,
) int {
	if index >= len(images) {
		index = 0
	} else if index < 0 {
		index = len(images) - 1
	}
	black := color.RGBA{0, 0, 0, 0}
	draw.Draw(source, maxRect.Bounds(), &image.Uniform{black}, image.ZP, 1)
	draw.Draw(source, maxRect.Bounds(), images[index], image.ZP, 1)
	(*ws).Upload(image.ZP, *buffer, maxRect.Bounds())
	(*ws).Publish()
	return index
}

// DeleteImage deletes a single image
func DeleteImage(images []image.Image, index int, path string) ([]image.Image, error) {
	// this function modifies the images collection in place
	switch index {
	case len(images) - 1:
		// you have reached the end of the list
		images = images[:index]
	case 0:
		// you are at the start of the list
		images = images[1:]
	default:
		// you are somewhere between the end and the start of the list
		images = append(images[:index], images[index+1:]...)
	}
	err := os.Remove(path)
	if err != nil {
		return nil, err
	}
	return images, nil
}

// ReadFiles recursively searches an entire directory for all the files in that directory
func ReadFiles(path string) []string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	re := regexp.MustCompile("[.]")
	imgNames := []string{}
	for _, file := range files {
		fullPath := fmt.Sprintf("%s/%s", path, file.Name())
		if re.MatchString(file.Name()) {
			imgNames = append(imgNames, fullPath)
		} else {
			imgNames = append(imgNames, ReadFiles(fullPath)...)
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
				if e.Direction == key.DirRelease {
					switch e.Code {
					case key.CodeEscape:
						buffer.Release()
						return
					case key.CodeRightArrow:
						count = ChangeImage(&ws, count+1, maxRect, source, resizeImg, &buffer)
					case key.CodeLeftArrow:
						count = ChangeImage(&ws, count-1, maxRect, source, resizeImg, &buffer)
					case key.CodeDeleteForward, key.CodeDeleteBackspace:
						pathName := imgNames[count]
						resizeImg, err = DeleteImage(resizeImg, count, pathName)
						if err != nil {
							log.Fatal(err)
						}
						count = ChangeImage(&ws, count, maxRect, source, resizeImg, &buffer)
					}
				}
			}
		}
	})
}
