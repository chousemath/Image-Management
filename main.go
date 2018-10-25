// C:\Users\Kwak\Desktop\Trive\Image-Management\test_data
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"path/filepath"
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

	"github.com/disintegration/imaging"
)

const (
	maxWidth  = 1920
	maxHeight = 1080
	cropSizeUnit = 100
	brightUnit = 10	
	contrastUnit = 15
)

var currentWD string
var imgNames []string

// getWD get current working directory path
func getWD() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(fmt.Sprintf("Error getting the working directory: %v", err))
	}
	return dir
}

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

// EncodeImage encodes a single image by its name
func EncodeImage(filename string,src image.Image)(error){
	f, err := os.Create(filename)
	if err!=nil{
		return err
	}
	defer f.Close()
	jpeg.Encode(f, src, nil)
	return nil	
}

// ReadFiles recursively searches an entire directory for all the files in that directory
func ReadFiles(path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	re := regexp.MustCompile("[.]")
	for _, file := range files {
		fullPath := fmt.Sprintf("%s/%s", path, file.Name())
		if re.MatchString(file.Name()) {
			imgNames = append(imgNames, fullPath)
		} else {
			ReadFiles(fullPath)
		}
	}
}

// GetCopyDir returns copy image directory
func GetCopyDir(path string) string {
	_, fileName := filepath.Split(path)
	copyDir := fmt.Sprintf("%s/copy_data/", currentWD)
	dstDir := fmt.Sprintf("%s/copy_data/%s", currentWD, fileName)
	if _, err := os.Stat(copyDir); os.IsNotExist(err) {
		os.MkdirAll(copyDir, 0755)
	}
	return dstDir
}

// CopyImage copy image in working directory
func CopyImage(srcDir string, dstDir string) error{
	src, err := DecodeImage(srcDir)
	if err != nil {
		return err
	}

	err = EncodeImage(dstDir, src)
	if err != nil {
		return err
	}
	return nil
}

// DrawImage draw a single image on window
func DrawImage(
	ws *screen.Window,
	buffer *screen.Buffer,
	path string,
	src image.Image) (error){	
	source := (*buffer).RGBA()
	// draw background
	black := color.RGBA{0, 0, 0, 0}
	draw.Draw(source, (*buffer).Bounds(), &image.Uniform{black}, image.ZP, 1)
	// draw data image
	draw.Draw(source, src.Bounds(), src, image.ZP, 1)
	// upload image on screen
	(*ws).Upload(image.ZP, *buffer, (*buffer).Bounds())
	(*ws).Publish()
	return nil
}

// DeleteFile deletes a single file path
func DeleteFile(path string) (error) {
	err := os.Remove(path)
	if err != nil {
		return err
	}
	return nil
}

// DeleteArrayElement deletes array element by index
func DeleteArrayElement(arr[] string, index int)([]string){
	if len(arr) == 1 {
		arr = []string{}
		return nil
	}
	switch index {
	case len(arr) - 1:
		// you have reached the end of the list
		arr = arr[:index]
	case 0:
		// you are at the start of the list
		arr = arr[1:]
	default:
		// you are somewhere between the end and the start of the list
		arr = append(arr[:index], arr[index+1:]...)
	}
	return arr
}

// CheckOutOfIndex checks for index out of bounds errors
func CheckOutOfIndex(sliceLength int, index int) int {
	switch {
	case index >= sliceLength:
		return 0
	case index < 0:
		return sliceLength - 1
	default:
		return index
	}
}

func InitCopyData(arr[] string,index int, dir string)(image.Image, error){
	err := CopyImage(arr[index], dir)
	if err != nil {
		return nil, err
	}
	curCopyImage,err := DecodeImage(dir)
	if err != nil {
		return nil, err
	}
	return curCopyImage, nil
}

func main() {
	currentWD = getWD()
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

		ReadFiles(path)
		curIndex := 0
		curCopyDir := GetCopyDir(imgNames[curIndex])
		curCopyImage,err := InitCopyData(imgNames, curIndex, curCopyDir)
		if err!=nil{
			log.Fatal(err)
		}
		// Draw Copy Image on window
		err = DrawImage(&ws, &buffer, curCopyDir, curCopyImage)
		if err!=nil {
			log.Fatal(fmt.Sprintf("Error draw image : %v", err))
		}

		brightClicks := 0
		contrastClicks := 0

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
						err := os.RemoveAll(fmt.Sprintf("%s/copy_data/", currentWD))
						if err != nil {
							log.Fatal(fmt.Sprintf("Error delete copy data : %v", err))
						}
						return
					case key.CodeRightArrow, key.CodeLeftArrow:
						// change original image
						err = EncodeImage(imgNames[curIndex], curCopyImage)
						if e.Code == key.CodeRightArrow {
							curIndex++	
						} else {
							curIndex--
						}
						curIndex = CheckOutOfIndex(len(imgNames),curIndex)
						curCopyImage,err = InitCopyData(imgNames, curIndex, curCopyDir)
						if err!=nil{
							log.Fatal(err)
						}
					case key.CodeDeleteForward, key.CodeDeleteBackspace:
						// Delete copy data
						err := DeleteFile(curCopyDir)
						if err != nil {
							log.Fatal(fmt.Sprintf("Error deleteing a copy data file : %v", err))
						}
						// Delete original data
						err = DeleteFile(imgNames[curIndex])
						if err != nil {
							log.Fatal(fmt.Sprintf("Error deleteing a original data file : %v", err))
						}
						imgNames = DeleteArrayElement(imgNames,curIndex)
						curCopyImage,err = InitCopyData(imgNames, curIndex, curCopyDir)
						if err!=nil{
							log.Fatal(err)
						}
					case key.CodeDownArrow, key.CodeUpArrow:
						if e.Code == key.CodeUpArrow{
							brightClicks++
						}else if e.Code == key.CodeDownArrow{
							brightClicks--
						}
						fmt.Println("click : ",brightClicks)
						if brightClicks < 0 {
							curCopyImage = imaging.AdjustBrightness(curCopyImage, (-1)*brightUnit)
						} else if brightClicks > 0 {
							curCopyImage = imaging.AdjustBrightness(curCopyImage, brightUnit)
						} else {
							curCopyImage,err = DecodeImage(imgNames[curIndex])
							if err != nil {
								log.Fatal(err)
							}	
						}
						err := EncodeImage(curCopyDir, curCopyImage)
						if err != nil {
							log.Fatal(fmt.Sprintf("Error encode changed image : %v", err))
						}
					case key.CodePageUp, key.CodePageDown:
						if e.Code == key.CodePageUp{
							contrastClicks++
						}else if e.Code == key.CodePageDown{
							contrastClicks--
						}
						if contrastClicks < 0 {
							curCopyImage = imaging.AdjustContrast(curCopyImage, (-1)*contrastUnit)
						} else if contrastClicks > 0 {
							curCopyImage = imaging.AdjustContrast(curCopyImage, contrastUnit)
						} else {
							curCopyImage,err = DecodeImage(imgNames[curIndex])
							if err != nil {
								log.Fatal(err)
							}	
						}
						err := EncodeImage(curCopyDir, curCopyImage)
						if err != nil {
							log.Fatal(fmt.Sprintf("Error encode changed image : %v", err))
						}
					case  key.CodeS:
						width := curCopyImage.Bounds().Max.X
						height := curCopyImage.Bounds().Max.Y
						curCopyImage = imaging.Crop(curCopyImage,image.Rect(25,25,width-25,height-25))
						err := EncodeImage(curCopyDir, curCopyImage)
						if err != nil {
							log.Fatal(fmt.Sprintf("Error encode changed image : %v", err))
						}
					}
				}
				curCopyDir = GetCopyDir(imgNames[curIndex])
				err = DrawImage(&ws, &buffer, curCopyDir, curCopyImage)
				if err!=nil {
					log.Fatal(fmt.Sprintf("Error draw image : %v", err))
				}
			}
		}
	})
}
