package main

//golang.org package download cmd command : go get golang.org/..
import (
	"fmt"
	"os"
	"io/ioutil"
	"log"
	"image"
	"image/draw"
	// "reflect"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/key"
	// "github.com/nfnt/resize"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

const (
	CodeEscape = 41
	CodeRightArrow = 79
	CodeLeftArrow  = 80
	CodeDeleteForward = 76
	DirRelease = 2
)


func decode(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("could not decode %s: %v", filename, err)
	}
	return m, nil
}

func changeImg(ws *screen.Window, index int, 
	source *image.RGBA , img []image.Image, point image.Point, buffer *screen.Buffer) int {
	if index >= len(img) {
		index = 0
	} else if index < 0 {
		index = len(img)-1
	}
	draw.Draw(source, img[index].Bounds(), img[index], point, 0)
	(*ws).Upload((*buffer).Size(), *buffer, img[0].Bounds())
	(*ws).Publish() 

	return index
}

func DeleteImg(){
	
}

func main() {
	// var path string
	// fmt.Println("Input File Path : ")
	// fmt.Scanln(&path)

	driver.Main(func(s screen.Screen) {
		path := "./demo-image/" //Test path directory
		files, err := ioutil.ReadDir(path)
		if err != nil {
			log.Fatal(err)
		}	

		imgNames := []string{}
		resizeImg := []image.Image{}
		for i, file := range files {
			imgNames = append(imgNames,path + file.Name())
			src, err := decode(imgNames[i])
			if err != nil {
				log.Fatal(err)
			}
			// resizeImg = append(resizeImg, resize.Resize(300, 0, src, resize.Lanczos3))	
			resizeImg = append(resizeImg, src)
		}
		
		ws, err := s.NewWindow(nil)
		if err != nil {
			panic(err)
			return
		}
		defer ws.Release()
		point := image.Point{400, 200}
		buffer, err := s.NewBuffer(point)
		if err != nil {
			panic(err)
			return
		}
		defer buffer.Release()
		source := buffer.RGBA()
		count := 0
		count = changeImg(&ws, count, source, resizeImg, point, &buffer)

		for {
			switch e := ws.NextEvent().(type){
			case lifecycle.Event:
				if e.To == lifecycle.StageDead{
					return
				}
			case key.Event:
				if e.Direction ==  DirRelease {
					switch e.Code {
					case CodeEscape :
						fmt.Println("CLICK ESC")
						return
					case CodeRightArrow :
						fmt.Println("CLICK RIGHT")
						count = changeImg(&ws, count+1, source, resizeImg, point, &buffer)
						fmt.Println("index : ",count)
					case CodeLeftArrow :
						fmt.Println("CLICK LEFT")
						count = changeImg(&ws, count-1, source, resizeImg, point, &buffer)
						fmt.Println("index : ",count)
					case CodeDeleteForward :
						fmt.Println("CLCK DELETE")
					}
				}
			}
		}
		
	})
}
