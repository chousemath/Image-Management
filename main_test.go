package main

import (
	"errors"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/nfnt/resize"
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
		return errors.New(fmt.Sprintf("%s is not a regular file...", src))
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

func TestDeleteImage(t *testing.T) {
	currentWD = getWD()
	src := fmt.Sprintf("%s/test_data/doge-1.jpg", currentWD)
	dstDIR := fmt.Sprintf("%s/test_data/images_to_delete", currentWD)
	dst := fmt.Sprintf("%s/doge-1.jpg", dstDIR)
	err := copy(src, dst)
	filesBefore, err := ioutil.ReadDir(dstDIR)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to read directory: %v", err))
	}
	images := []image.Image{}
	imgSrc, err := DecodeImage(dst)
	if err != nil {
		log.Fatal(err)
	}
	images = append(images, resize.Resize(500, 0, imgSrc, resize.Lanczos3))
	images, err = DeleteImage(images, 0, dst)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to delete image: %v", err))
	}
	filesAfter, err := ioutil.ReadDir(dstDIR)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to read directory: %v", err))
	}
	numFilesAfter := len(filesBefore) - len(filesAfter)
	if numFilesAfter != 1 {
		t.Errorf("DeleteImage failed, number of files deleted, got: %d, want: 1", numFilesAfter)
	}
	if len(images) != 0 {
		t.Errorf("DeleteImage failed, number of images in slice should be 0, got: %d", len(images))
	}
}
