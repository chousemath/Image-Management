// C:\Users\Kwak\Desktop\Trive\Image-Management\test_data
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"

	"github.com/chousemath/Image-Management/utils"
	"github.com/disintegration/imaging"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
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
func ReadFiles(imgDir string) []string {
	files, err := ioutil.ReadDir(imgDir)
	if err != nil {
		log.Fatal(fmt.Sprintf("ReadFiles Error: %v\n", err))
	}
	re := regexp.MustCompile("[.]")
	for _, file := range files {
		fullPath := fmt.Sprintf("%s/%s", imgDir, file.Name())
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
	imgDir string,
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
func DeleteFile(imgDir string) error {
	err := os.Remove(imgDir)
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

func main() {
	config, err := os.Open("./config.json")
	if err != nil {
		log.Fatal("Unable to open the config.json file")
	}
	defer config.Close()

	configBytes, err := ioutil.ReadAll(config)
	if err != nil {
		log.Fatal("Unable to convert config file into bytes")
	}

	conf := new(utils.Configuration)
	json.Unmarshal(configBytes, conf)

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: conf.GithubPersonalAccessToken,
	})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Open (and create if necessary) a simple text file to hold all errors
	// os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0666)
	errFile, err := os.OpenFile("errors.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error: %v", err))
	}
	defer errFile.Close()

	var imgDir string
	fmt.Println("Input imgDir directory : ")
	fmt.Scanln(&imgDir)
	err = os.RemoveAll(path.Join(imgDir, ".DS_Store"))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	driver.Main(func(s screen.Screen) {
		ws, err := s.NewWindow(nil)
		if err != nil {
			utils.CreateGithubIssue(ctx, conf, client, &[]string{"NewWindow"}, "Error creating a new window", err.Error())
			utils.WriteErr(errFile, err, "NewWindow", true)
		}
		defer ws.Release()

		buffer, err := s.NewBuffer(image.Pt(maxWidth, maxHeight))
		if err != nil {
			utils.CreateGithubIssue(ctx, conf, client, &[]string{"NewBuffer"}, "Error creating a new buffer", err.Error())
			utils.WriteErr(errFile, err, "NewBuffer", true)
		}
		defer buffer.Release()

		ReadFiles(imgDir)
		curIndex := 0
		curCopyDir := GetCopyDir(imgNames[curIndex], fmt.Sprintf("%s/copy_data", imgDir))
		curCopyImage, err := InitCopyData(imgNames, curIndex, curCopyDir, fmt.Sprintf("%s/copy_data", imgDir))
		if err != nil {
			utils.WriteErr(errFile, err, "InitCopyData1", false)
		}
		newTrashDir := fmt.Sprintf("%s/../trash_data/", imgDir)
		if _, err := os.Stat(newTrashDir); os.IsNotExist(err) {
			utils.WriteErr(errFile, err, "Stat", false)
			os.MkdirAll(newTrashDir, 0755)
		}
		// Draw Copy Image on window
		err = DrawImage(&ws, &buffer, curCopyDir, curCopyImage)
		if err != nil {
			utils.WriteErr(errFile, err, "DrawImage1", true)
		}

		brightClicks := 0
		contrastClicks := 0
		var copyPath string

		for {
			switch e := ws.NextEvent().(type) {
			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					copyPath = fmt.Sprintf("%s/copy_data/", imgDir)
					err = os.RemoveAll(copyPath)
					if err != nil {
						utils.WriteErr(errFile, err, "RemoveAll1", true)
					}
					return
				}

			case key.Event:
				if e.Direction == key.DirRelease {
					switch e.Code {
					case key.CodeEscape:
						buffer.Release()
						copyPath = fmt.Sprintf("%s/copy_data/", imgDir)
						err := os.RemoveAll(copyPath)
						if err != nil {
							utils.WriteErr(errFile, err, "RemoveAll2", true)
						}
						return
					case key.CodeRightArrow, key.CodeLeftArrow:
						// change original image
						err = EncodeImage(imgNames[curIndex], curCopyImage)
						if err != nil {
							utils.WriteErr(errFile, err, "EncodeImage1", true)
						}
						if e.Code == key.CodeRightArrow {
							curIndex++
						} else {
							curIndex--
						}
						curIndex = CheckOutOfIndex(len(imgNames), curIndex)
						err = os.Remove(curCopyDir)
						if err != nil {
							utils.WriteErr(errFile, err, "Remove1", true)
						}
						copyPath = fmt.Sprintf("%s/copy_data/", imgDir)
						curCopyDir := GetCopyDir(imgNames[curIndex], copyPath)
						curCopyImage, err = InitCopyData(imgNames, curIndex, curCopyDir, copyPath)
						if err != nil {
							utils.WriteErr(errFile, err, "InitCopyData2", false)
						}
					case key.CodeDeleteForward, key.CodeDeleteBackspace:
						trashDataDir := GetCopyDir(imgNames[curIndex], newTrashDir)
						copyPath = fmt.Sprintf("%s/copy_data/", imgDir)
						err = CopyImage(imgNames[curIndex], trashDataDir, copyPath)
						if err != nil {
							utils.WriteErr(errFile, err, "GetCopyDir1", true)
						}
						// Delete copy data
						err := DeleteFile(curCopyDir)
						if err != nil {
							utils.WriteErr(errFile, err, "DeleteFile1", true)
						}
						// Delete original data
						err = DeleteFile(imgNames[curIndex])
						if err != nil {
							utils.WriteErr(errFile, err, "DeleteFile2", true)
						}
						imgNames = DeleteArrayElement(imgNames, curIndex)
						curCopyDir := GetCopyDir(imgNames[curIndex], copyPath)
						curCopyImage, err = InitCopyData(imgNames, curIndex, curCopyDir, copyPath)
						if err != nil {
							utils.WriteErr(errFile, err, "InitCopyData3", false)
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
								utils.WriteErr(errFile, err, "DecodeImage1", true)
							}
						}
						err := EncodeImage(curCopyDir, curCopyImage)
						if err != nil {
							utils.WriteErr(errFile, err, "EncodeImage2", true)
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
								utils.WriteErr(errFile, err, "DecodeImage2", true)
							}
						}
						err := EncodeImage(curCopyDir, curCopyImage)
						if err != nil {
							utils.WriteErr(errFile, err, "EncodeImage3", true)
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
							utils.WriteErr(errFile, err, "EncodeImage4", true)
						}
					}
				}
				copyPath = fmt.Sprintf("%s/copy_data/", imgDir)
				curCopyDir = GetCopyDir(imgNames[curIndex], copyPath)
				err = DrawImage(&ws, &buffer, curCopyDir, curCopyImage)
				if err != nil {
					utils.WriteErr(errFile, err, "EncodeImage5", true)
				}
				brightClicks = 0
				contrastClicks = 0
			}
		}
	})
}
