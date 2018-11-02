// C:\Users\Kwak\Desktop\Trive\Image-Management\test_data
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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
	maxWidth     = 1920
	maxHeight    = 1080
	cropSizeUnit = 100
	brightUnit   = 10
	contrastUnit = 15
)

var imgNames []string

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
func EncodeImage(filename string, src image.Image) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	jpeg.Encode(f, src, nil)
	return nil
}

// ReadFiles recursively searches an entire directory for all the files in that directory
func ReadFiles(path string) []string {
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
	return imgNames
}

// GetCopyDir returns copy image directory
func GetCopyDir(src string, dst string) string {
	_, fileName := filepath.Split(src)
	dstDir := fmt.Sprintf("%s/%s", dst, fileName)
	return dstDir
}

// CopyImage copy image in working directory
func CopyImage(srcDir string, dstDir string, copyPath string) error {
	if _, err := os.Stat(copyPath); os.IsNotExist(err) {
		os.MkdirAll(copyPath, 0755)
	}

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
	src image.Image) error {
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
func DeleteFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		return err
	}
	return nil
}

// DeleteArrayElement deletes array element by index
func DeleteArrayElement(arr []string, index int) []string {
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

// InitCopyData xxxxx
func InitCopyData(arr []string, index int, dir string, copyPath string) (image.Image, error) {
	// err := os.RemoveAll(fmt.Sprintf("%s/.",copyPath))
	// if err != nil {
	// 	log.Fatal(fmt.Sprintf("Error delete copy data : %v", err))
	// }
	err := CopyImage(arr[index], dir, copyPath)
	if err != nil {
		return nil, err
	}
	curCopyImage, err := DecodeImage(dir)
	if err != nil {
		return nil, err
	}
	return curCopyImage, nil
}

func writeErr(f *os.File, myErr error, fatal bool) {
	if _, err := f.WriteString(fmt.Sprintf("%s\n", myErr.Error())); err != nil {
		log.Fatal(fmt.Sprintf("Crashed while writing error to file: %v", err))
	}
	if fatal {
		log.Fatal(myErr.Error())
	}
	fmt.Println(myErr.Error())
}

func main() {

	// Open (and create if necessary) a simple text file to hold all errors
	// os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0666)
	errFile, err := os.OpenFile("errors.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error: %v", err))
	}
	defer errFile.Close()

	var path string
	fmt.Println("Input path directory : ")
	fmt.Scanln(&path)

	driver.Main(func(s screen.Screen) {
		ws, err := s.NewWindow(nil)
		if err != nil {
			writeErr(errFile, err, true)
		}
		defer ws.Release()

		buffer, err := s.NewBuffer(image.Pt(maxWidth, maxHeight))
		if err != nil {
			writeErr(errFile, err, true)
		}
		defer buffer.Release()

		ReadFiles(path)
		curIndex := 0
		curCopyDir := GetCopyDir(imgNames[curIndex], fmt.Sprintf("%s/copy_data", path))
		curCopyImage, err := InitCopyData(imgNames, curIndex, curCopyDir, fmt.Sprintf("%s/copy_data", path))
		if err != nil {
			writeErr(errFile, err, true)
		}
		newTrashDir := fmt.Sprintf("%s/../trash_data/", path)
		if _, err := os.Stat(newTrashDir); os.IsNotExist(err) {
			writeErr(errFile, err, false)
			os.MkdirAll(newTrashDir, 0755)
		}
		// Draw Copy Image on window
		err = DrawImage(&ws, &buffer, curCopyDir, curCopyImage)
		if err != nil {
			writeErr(errFile, err, true)
		}

		brightClicks := 0
		contrastClicks := 0

		for {
			switch e := ws.NextEvent().(type) {
			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					err = os.RemoveAll(fmt.Sprintf("%s/copy_data/", path))
					if err != nil {
						writeErr(errFile, err, true)
					}
					return
				}

			case key.Event:
				if e.Direction == key.DirRelease {
					switch e.Code {
					case key.CodeEscape:
						buffer.Release()
						err := os.RemoveAll(fmt.Sprintf("%s/copy_data/", path))
						if err != nil {
							writeErr(errFile, err, true)
						}
						return
					case key.CodeRightArrow, key.CodeLeftArrow:
						// change original image
						err = EncodeImage(imgNames[curIndex], curCopyImage)
						if err != nil {
							writeErr(errFile, err, true)
						}
						if e.Code == key.CodeRightArrow {
							curIndex++
						} else {
							curIndex--
						}
						curIndex = CheckOutOfIndex(len(imgNames), curIndex)
						err = os.Remove(curCopyDir)
						if err != nil {
							writeErr(errFile, err, true)
						}
						curCopyDir := GetCopyDir(imgNames[curIndex], fmt.Sprintf("%s/copy_data/", path))
						curCopyImage, err = InitCopyData(imgNames, curIndex, curCopyDir, fmt.Sprintf("%s/copy_data", path))
						if err != nil {
							writeErr(errFile, err, true)
						}
					case key.CodeDeleteForward, key.CodeDeleteBackspace:
						trashDataDir := GetCopyDir(imgNames[curIndex], newTrashDir)
						err = CopyImage(imgNames[curIndex], trashDataDir, fmt.Sprintf("%s/copy_data/", path))
						if err != nil {
							writeErr(errFile, err, true)
						}
						// Delete copy data
						err := DeleteFile(curCopyDir)
						if err != nil {
							writeErr(errFile, err, true)
						}
						// Delete original data
						err = DeleteFile(imgNames[curIndex])
						if err != nil {
							writeErr(errFile, err, true)
						}
						imgNames = DeleteArrayElement(imgNames, curIndex)
						curCopyDir := GetCopyDir(imgNames[curIndex], fmt.Sprintf("%s/copy_data/", path))
						curCopyImage, err = InitCopyData(imgNames, curIndex, curCopyDir, fmt.Sprintf("%s/copy_data", path))
						if err != nil {
							writeErr(errFile, err, true)
						}
					case key.CodeDownArrow, key.CodeUpArrow:
						if e.Code == key.CodeUpArrow {
							brightClicks++
						} else if e.Code == key.CodeDownArrow {
							brightClicks--
						}
						if brightClicks < 0 {
							curCopyImage = imaging.AdjustBrightness(curCopyImage, (-1)*brightUnit)
						} else if brightClicks > 0 {
							curCopyImage = imaging.AdjustBrightness(curCopyImage, brightUnit)
						} else {
							curCopyImage, err = DecodeImage(imgNames[curIndex])
							if err != nil {
								writeErr(errFile, err, true)
							}
						}
						err := EncodeImage(curCopyDir, curCopyImage)
						if err != nil {
							writeErr(errFile, err, true)
						}
					case key.CodeZ, key.CodeX:
						if e.Code == key.CodeZ {
							contrastClicks++
						} else if e.Code == key.CodeX {
							contrastClicks--
						}
						if contrastClicks < 0 {
							curCopyImage = imaging.AdjustContrast(curCopyImage, (-1)*contrastUnit)
						} else if contrastClicks > 0 {
							curCopyImage = imaging.AdjustContrast(curCopyImage, contrastUnit)
						} else {
							curCopyImage, err = DecodeImage(imgNames[curIndex])
							if err != nil {
								writeErr(errFile, err, true)
							}
						}
						err := EncodeImage(curCopyDir, curCopyImage)
						if err != nil {
							writeErr(errFile, err, true)
						}
					case key.CodeA, key.CodeW, key.CodeD, key.CodeS:
						width := curCopyImage.Bounds().Max.X
						height := curCopyImage.Bounds().Max.Y
						if e.Code == key.CodeA {
							curCopyImage = imaging.Crop(curCopyImage, image.Rect(0, 0, width-25, height))
						} else if e.Code == key.CodeD {
							curCopyImage = imaging.Crop(curCopyImage, image.Rect(25, 0, width, height))
						} else if e.Code == key.CodeW {
							curCopyImage = imaging.Crop(curCopyImage, image.Rect(0, 0, width, height-25))
						} else if e.Code == key.CodeS {
							curCopyImage = imaging.Crop(curCopyImage, image.Rect(0, 25, width, height))
						}

						err := EncodeImage(curCopyDir, curCopyImage)
						if err != nil {
							writeErr(errFile, err, true)
						}
					}
				}
				curCopyDir = GetCopyDir(imgNames[curIndex], fmt.Sprintf("%s/copy_data", path))
				err = DrawImage(&ws, &buffer, curCopyDir, curCopyImage)
				if err != nil {
					writeErr(errFile, err, true)
				}
				brightClicks = 0
				contrastClicks = 0
			}
		}
	})
}
