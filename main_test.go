package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"image"
	"image/color"
	"image/draw"
	"log"
	"os"
	"path/filepath"
	"testing"
)

var currentWD string

func getWD() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(fmt.Sprintf("Error getting the working directory: %v", err))
	}
	return dir
}

func copy(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()
	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func TestDecodeImage(t *testing.T) {
	currentWD = getWD()
	imgPath := fmt.Sprintf("%s/test_data/doge-1.jpg", currentWD)
	_, err := DecodeImage(imgPath)
	if err != nil {
		t.Errorf("DecodeImage failed, %v", err)
	}
}
func TestEncodeImage(t *testing.T){
	currentWD = getWD()
	imgPath := fmt.Sprintf("%s/test_create_image/rectangle.jpeg", currentWD)
	if _, err := os.Stat(fmt.Sprintf("%s/test_create_image",currentWD)); os.IsNotExist(err) {
		os.MkdirAll(fmt.Sprintf("%s/test_create_image",currentWD), 0755)
	}

    rectImage := image.NewRGBA(image.Rect(0, 0, 200, 200))
    green := color.RGBA{0, 100, 0, 255}
	draw.Draw(rectImage, rectImage.Bounds(), &image.Uniform{green}, image.ZP, draw.Src)
	
	err := EncodeImage(imgPath,rectImage)
	if err != nil{
		t.Errorf("EncodeImage failed, %v", err)
	}
	
	f, err := ioutil.ReadDir(fmt.Sprintf("%s/test_create_image/",currentWD))
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to read directory: %v", err))
	}
	if len(f) <= 0 {
		t.Errorf("EncodeImage failed. got: %d , want: 1",len(f))
	}
	err = os.RemoveAll(fmt.Sprintf("%s/test_create_image",currentWD))
	if err != nil {
		log.Fatal(fmt.Sprintf("Error delete copy data : %v", err))
	}
}

func TestDeleteArrayElement(t *testing.T){
	testArr := []string{
		"test1",
		"test2",
		"test3",
	}
	resultArr1 := DeleteArrayElement(testArr,1);
	if len(resultArr1)!=2 {
		t.Errorf("Error deleting last element in array,\nWanted : [test1 test3] , Current : %v",resultArr1)
	} 
	resultArr2 := DeleteArrayElement(testArr,2);
	if len(resultArr2)!= 2 {
		t.Errorf("Error deleting last element in array,\nWanted : [test1 test2] , Current : %v",resultArr2)
	} 
	resultArr3 := DeleteArrayElement(testArr,0);
	if len(resultArr3)!= 2 {
		t.Errorf("Error deleting first element in array,\nWanted : [test2 test3] , Current : %v",resultArr3)
	}
}

func TestReadFiles(t *testing.T) {
	currentWD := getWD()
	parentWD := fmt.Sprintf("%s/test_data", currentWD)
	childWD := fmt.Sprintf("%s/child_test_data", parentWD)
	files, err := ioutil.ReadDir(parentWD)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to read directory: %v", err))
	}
	beforeParentFiles := []string{}
	for _, file := range files {
		if filepath.Ext(file.Name()) != "" {
			beforeParentFiles = append(beforeParentFiles, file.Name())
		}
	}
	files, err = ioutil.ReadDir(childWD)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to read directory: %v", err))
	}
	beforeChildFiles := []string{}
	for _, file := range files {
		if filepath.Ext(file.Name()) != "" {
			beforeChildFiles = append(beforeChildFiles, file.Name())
		}
	}
	total := len(beforeParentFiles) + len(beforeChildFiles)
	afterFiles := ReadFiles(parentWD)
	if total != len(afterFiles) {
		log.Fatal(fmt.Sprintf("ReadFiles() is Failed, number of images, got: %d, want: %d", len(afterFiles), total))
	}
}

func TestCheckOutOfIndex(t *testing.T) {
	imgNames := []string{
		"asdlfalsdkjfsdf",
		"aldkfjalskdjfalsdf",
		"sdkfjalsdkfjalsdf",
	}
	curIndex1 := CheckOutOfIndex(len(imgNames), 10)
	curIndex2 := CheckOutOfIndex(len(imgNames), -1)
	curIndex3 := CheckOutOfIndex(len(imgNames), 1)
	if curIndex1 != 0 {
		t.Errorf("Incorrect index return value, should be 0, got: %d", curIndex1)
	}
	if curIndex2 != 2 {
		t.Errorf("Incorrect index return value, should be 2, got: %d", curIndex2)
	}
	if curIndex3 != 1 {
		t.Errorf("Incorrect index return value, should be 1, got: %d", curIndex3)
	}
}
