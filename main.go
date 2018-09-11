package main

//golang.org package download cmd command : go get golang.org/..
import (
	"fmt"
	"os"
	"io/ioutil"
	"log"
	"image"
	"image/draw"
	// "image/color"
	// "reflect"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/key"
	"github.com/nfnt/resize"
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
	source *image.RGBA , img []image.Image, buffer *screen.Buffer) int {
	point := image.ZP
	if index >= len(img) {
		index = 0
	} else if index < 0 {
		index = len(img)-1
	}
	draw.Draw(source, img[index].Bounds(), img[index], point, 1)
	(*ws).Upload(image.ZP, *buffer, img[0].Bounds())
	(*ws).Publish() 

	return index
}

func deleteImg(img []image.Image, index int, path string) ([]image.Image, error){
	result := []image.Image{}
	if index == len(img)-1{
		result = img[:index]
	} else if index == 0 {
		result = img[1:]
	} else {
		result = append(img[:index], img[index+1:]...)
	}
	
	err := os.Remove(path)
	if err != nil {
		return nil, err
	}

	return result,nil
}

func main() {
	var path string
	fmt.Println("Input relative path directory : ")
	fmt.Scanln(&path)

	driver.Main(func(s screen.Screen) {
		// path := "./demo-image/" //Test path directory
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
			resizeImg = append(resizeImg, resize.Resize(500, 0, src, resize.Lanczos3))	
			// resizeImg = append(resizeImg, src)
		}

		w := resizeImg[0].Bounds().Max.X
		h := resizeImg[0].Bounds().Max.Y
		
		ws, err := s.NewWindow(&screen.NewWindowOptions{
			Width: w,
			Height: h,
		})
		if err != nil {
			panic(err)
			return
		}
		defer ws.Release()
		point := image.Pt(w,h)
		buffer, err := s.NewBuffer(point)
		if err != nil {
			panic(err)
			return
		}
		defer buffer.Release()
		source := buffer.RGBA()
		count := 0
		count = changeImg(&ws, count, source, resizeImg, &buffer)

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
						buffer.Release()
						fmt.Println("BYE!")
						return
					case CodeRightArrow :
						count = changeImg(&ws, count+1, source, resizeImg, &buffer)
					case CodeLeftArrow :
						count = changeImg(&ws, count-1, source, resizeImg, &buffer)
					case CodeDeleteForward :
						pathName := imgNames[count]
						resizeImg,err = deleteImg(resizeImg, count, pathName)
						if err != nil {
							log.Fatal(err)
						}
						count = changeImg(&ws, count, source, resizeImg, &buffer)
						fmt.Println("SUCCESS DELETE")
					}
				}
			}
		}
		
	})
}
