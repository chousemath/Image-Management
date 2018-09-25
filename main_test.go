package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"path/filepath"
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

func TestDeleteFile(t *testing.T){
	currentWD = getWD()
	src := fmt.Sprintf("%s/test_data/doge-1.jpg", currentWD)
	dstDIR := fmt.Sprintf("%s/test_data/images_to_delete", currentWD)
	dst := fmt.Sprintf("%s/doge-1.jpg",dstDIR)
	err := copy(src,dst)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed copy image: %v",err))
	}
	filesBefore, err := ioutil.ReadDir(dstDIR)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to read directory: %v",err))
	}
	imgNames := []string{}
	for i, file := range filesBefore{
		imgNames = append(imgNames,fmt.Sprintf("%s/%s",dstDIR,file.Name()))
		fmt.Sprintf("Before imgNames[%d] : %s",i,imgNames[i])
	}
	imgNames,err = DeleteFile(imgNames, 0)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to delete directory: %v",err))
	}
	filesAfter, err := ioutil.ReadDir(dstDIR)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to read directory: %v",err))
	}
	numFilesAfter := len(filesBefore) - len(filesAfter)
	if numFilesAfter != 1 {
		t.Errorf("DeleteImage failed, number of files deleted, got: %d, want: 1", numFilesAfter)
	}
	if len(imgNames) != 0 {
		t.Errorf("DeleteImage failed, number of imgNames in slice should be 0, got: %d", len(imgNames))
	}	
}

func TestDrawImage(t *testing.T){
	
}

func TestReadFiles(t *testing.T) {
	currentWD := getWD()
	parentWD := fmt.Sprintf("%s/test_data",currentWD)
	childWD := fmt.Sprintf("%s/child_test_data",parentWD)
	files, err := ioutil.ReadDir(parentWD)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to read directory: %v",err))
	}
	beforeParentFiles := []string{}
	for _, file := range files {
		if filepath.Ext(file.Name()) != "" {
			beforeParentFiles = append(beforeParentFiles,file.Name())
		} 
	}
	files, err = ioutil.ReadDir(childWD)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to read directory: %v",err))
	}
	beforeChildFiles := []string{}
	for _, file := range files {
		if filepath.Ext(file.Name()) != "" {
			beforeChildFiles = append(beforeChildFiles,file.Name())
		} 
	}
	total := len(beforeParentFiles) + len(beforeChildFiles)
	afterFiles := ReadFiles(parentWD)
	if total != len(afterFiles){
		log.Fatal(fmt.Sprintf("ReadFiles() is Failed, number of images, got: %d, want: %d",len(afterFiles),total))
	}
}
