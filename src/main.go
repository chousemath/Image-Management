package main

//golang.org package download cmd command : go get golang.org/..
import (
	"fmt"
	"os"
	"io/ioutil"
	"log"
	"image"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/key"
	// "golang.org/x/exp/shiny/widget"
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

func main() {
	// var path string
	// fmt.Println("Input File Path : ")
	// fmt.Scanln(&path)

	driver.Main(func(s screen.Screen) {
		path := "../demo-image/"
		imgFiles, err := ioutil.ReadDir(path)
		if err != nil {
			log.Fatal(err)
		}	
		imgNames := []string{}
		for _, file := range imgFiles {
			imgNames = append(imgNames,path + file.Name())
		}
		
		// Image Decode
		src, err := decode(imgNames[0])
		if err != nil {
			log.Fatal(err)
		}
		// Resize one image
		resizeImg := resize.Resize(300, 0, src, resize.Lanczos3)
		
		// Show one image on the screen
		// w := widget.NewSheet(widget.NewImage(resizeImg, resizeImg.Bounds()))
		// if err := widget.RunWindow(s, w, &widget.RunWindowOptions{
		// 	NewWindowOptions: screen.NewWindowOptions{
		// 		Width: resizeImg.Bounds().Max.X,
		// 		Height: resizeImg.Bounds().Max.Y,
		// 	},
		// }); err != nil {
		// 	log.Fatal(err)
		// }

		// TODO : NEW WINDOW로 변경해서 해보기
		w, err := s.NewWindow(nil)
		if err != nil {
			panic(err)
			return
		}
		defer w.Release()

		w.Publish()
		for {
			switch e := w.NextEvent().(type){
			case lifecycle.Event:
				if e.To == lifecycle.StageDead{
					return
				}
			case key.Event:
				if e.Direction ==  DirRelease {
					switch e.Code {
					case CodeEscape :
						return
					case CodeRightArrow :
						fmt.Println("right")
					case CodeLeftArrow :
						fmt.Println("left")
					case CodeDeleteForward :
						fmt.Println("delete")
					}
				}
			}
		}
		
	})
}
