package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"unicode"
)

func outputInfo(tag string, value interface{}) {
	fmt.Printf("%-18s    %v\n", tag+":", value)
}

// Exists judge file or direction existence
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsExist(err) {
		return true
	}
	return false
}

func runCommand(command string) (response string, errname error) {
	fmt.Println(command)
	lastQuote := rune(0)
	f := func(c rune) bool {
		switch {
		case c == lastQuote:
			lastQuote = rune(0)
			return false
		case lastQuote != rune(0):
			return false
		case unicode.In(c, unicode.Quotation_Mark):
			lastQuote = c
			return false
		default:
			return unicode.IsSpace(c)
		}
	}
	parts := strings.FieldsFunc(command, f)

	cmd := exec.Command(parts[0], parts[1:]...)
	var out bytes.Buffer
	cmd.Stderr = os.Stderr
	//cmd.Stdout = os.Stdout
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func getVideosFiles(rootFolder string) []string {
	var videoFiles []string
	files, err := ioutil.ReadDir(rootFolder)
	if err != nil {
		panic(err)
	}
	// TODO: handle the error!
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".mp4") {
			videoFiles = append(videoFiles, rootFolder+file.Name())
		}
	}
	return videoFiles
}

func removeFile(fileName string) (response string, err error) {
	exists := Exists(fileName)
	if exists {
		e := os.Remove(fileName)
		if e != nil {
			log.Fatal(e)
			return "", e
		}
	}
	return "", nil
}

func rename(oldFileName string, newFileName string) (response string, err error) {
	return runCommand("mv " + oldFileName + " " + newFileName)
}

// CopyFile source to dest
func CopyFile(source string, dest string) {
	//Read the file
	temp, err := ioutil.ReadFile(source)
	if err == nil {
		ioutil.WriteFile(dest, temp, 0777)
	}
}
