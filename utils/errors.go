package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

// WriteErr writes the error to a local file
func WriteErr(f *os.File, myErr error, tag string, fatal bool) {
	if _, err := f.WriteString(fmt.Sprintf("[%s]<%d>: %s\n", tag, time.Now().Unix(), myErr.Error())); err != nil {
		log.Fatal(fmt.Sprintf("Crashed while writing error to file: %v", err))
	}
	if fatal {
		log.Fatal(fmt.Sprintf("[%s]<%d>: %s\n", tag, time.Now().Unix(), myErr.Error()))
	}
	fmt.Println(myErr.Error())
}
