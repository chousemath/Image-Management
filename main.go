package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
)

const(
	maxWidth = 1920
	maxHeight = 1080
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

// DrawImage draw a single image on window
func DrawImage(
	ws *screen.Window,
	buffer *screen.Buffer,
	imgNames []string, 
	index int){
	src, err := DecodeImage(imgNames[index])
	if err != nil {
		log.Fatal(err)
	}
	source := (*buffer).RGBA()
	// draw background
	black := color.RGBA{0, 0, 0, 0}
	draw.Draw(source, (*buffer).Bounds(), &image.Uniform{black}, image.ZP, 1)
	// draw data image
	draw.Draw(source, src.Bounds(), src , image.ZP, 1)
	// upload image on screen
	(*ws).Upload(image.ZP, *buffer, (*buffer).Bounds())
	(*ws).Publish()

}

// DeleteFile deletes a single file path
func DeleteFile(imgNames []string, index int) ([]string, error){
	err := os.Remove(imgNames[index])
	if err != nil {
		return nil, err
	}
	switch index {
	case len(imgNames) - 1:
		// you have reached the end of the list
		imgNames = imgNames[:index]
	case 0:
		// you are at the start of the list
		imgNames = imgNames[1:]
	default:
		// you are somewhere between the end and the start of the list
		imgNames = append(imgNames[:index], imgNames[index+1:]...)
	}
	return imgNames, nil
}

func CheckOutOfIndex(slice []string, index int) int{
	switch index {
	case len(slice):
		return 0
	case -1:
		return len(slice) - 1
	default:
		return index
	}
}

func main() {
	var path string
	fmt.Println("Input path directory : ")
	fmt.Scanln(&path)

	driver.Main(func(s screen.Screen) {
		ws, err := s.NewWindow(nil)
		if err != nil {
			log.Fatal(fmt.Sprintf("Error creating a new window: %v", err))
		}
		defer ws.Release()

		buffer, err := s.NewBuffer(image.Pt(maxWidth, maxHeight))
		if err != nil {
			log.Fatal(fmt.Sprintf("Error creating a new buffer: %v", err))
		}
		defer buffer.Release()

		imgNames := ReadFiles(path)
		curIndex := 0
		DrawImage(&ws, &buffer, imgNames, curIndex)
		
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
						curIndex = CheckOutOfIndex(imgNames, curIndex+1)
						DrawImage(&ws, &buffer, imgNames, curIndex)
					case key.CodeLeftArrow:
						curIndex = CheckOutOfIndex(imgNames, curIndex-1)
						DrawImage(&ws, &buffer, imgNames, curIndex)
					case key.CodeDeleteForward, key.CodeDeleteBackspace:
						// TODO : change delete method 
						imgNames, err = DeleteFile(imgNames, curIndex)
						if err != nil{
							log.Fatal(fmt.Sprintf("Error deleteing a file : %v", err))
						}
						DrawImage(&ws, &buffer, imgNames, curIndex)
					}
				}
			}
		}
	})
}
