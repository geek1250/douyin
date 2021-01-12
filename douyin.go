package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/asticode/go-astisub"
)

var (
	helpFlag       bool
	quinnIniFlag   string
	versionFlag    bool
	startY         = 200
	videoPositionY = startY + 100*2

	IBMText2SpeechAPIKey = "THIS IS IBM Text to Speech API Key"

	outputPath = "../../douyin-works/"

	resourceFolder = "resource/"
	scriptFile     = resourceFolder + "scripts.srt"
	videoFolder    = resourceFolder + "videos/"
	tempFolder     = resourceFolder + "temp/"
	voicelistFile  = resourceFolder + "voicelist.txt"
	audioFile1     = tempFolder + "1.mp3"
	audioFile2     = tempFolder + "2.mp3"

	defaultImage      = tempFolder + "defaultImage.jpg"
	headerAudioFile1  = tempFolder + "1-formated.mp3"
	headerAudioFile2  = tempFolder + "2-formated.mp3"
	introductionAudio = tempFolder + "introduction.mp3"
	introductionVideo = tempFolder + "introductionVideo.mp4"
)

func usage() {
	fmt.Printf("Usage: quinn [-c filename]\n\nOptions:\n")
	flag.PrintDefaults()
}

func welcome() {
	//slant format
	//ASCII Art (AA) Generator : https://en.rakko.tools/tools/68/
	fmt.Println("")
	fmt.Println("  ____")
	fmt.Println("  / __ \\  __  __   (_)   ____    ____    ")
	fmt.Println(" / / / / / / / /  / /   / __ \\  / __ \\   ")
	fmt.Println("/ /_/ / / /_/ /  / /   / / / / / / / /   ")
	fmt.Println("\\___\\_\\ \\__,_/  /_/   /_/ /_/ /_/ /_/    ")
	fmt.Println("")

}

func main() {

	if helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	welcome()

	video()
}

func video() {
	// create folders and files
	exists := Exists(outputPath)
	if !exists {
		os.MkdirAll(outputPath, 0700)
	}

	videoFiles := getVideosFiles(videoFolder)
	colorRed := "\033[31m"
	if len(videoFiles) == 0 {
		fmt.Println(string(colorRed), errors.New("No video file in folder "+videoFolder))
		return
	}

	//parse scripts
	subtitles, err := astisub.OpenFile(scriptFile)

	fmt.Println(*subtitles)
	fmt.Println(len(subtitles.Items))
	if err != nil {
		fmt.Println(string(colorRed), errors.New("Script file error!"))
		return
	}
	if len(videoFiles) != len(subtitles.Items)-1 {
		fmt.Println(string(colorRed), errors.New("Video Script number does not match  video number!"))
		return
	}

	response, err := generateIntroductionVideo(subtitles, videoFiles)
	if err != nil {
		fmt.Println(response)
		return
	}

	// handle each video
	videoFiles = append([]string{introductionVideo}, videoFiles[0:]...)
	for i, file := range videoFiles {
		transformFileName := generateNewName(file, "-formated.mp4")
		//formate Video
		tempFileName := generateNewName(file, "-temp.mp4")
		response, err = formatVideo(file, tempFileName)
		if err != nil {
			return
		}
		textString, err := generateDrawText(subtitles, tempFileName, i)
		if err != nil {
			return
		}
		response, err = changeVideoHeigh(tempFileName, transformFileName)
		if err != nil {
			return
		}

		//add Text to Video
		textedFileName := generateNewName(file, "-new.mp4")
		response, err = addText2Video(textString, transformFileName, textedFileName)
		if err != nil {
			fmt.Println(err)
			return
		}

	}

	response, err = combineVideos()
	if err != nil {
		return
	}
	//remove videos elements
	for _, file := range videoFiles {
		removeFile(file)
	}

	BGreen := "\033[1;32m"
	fmt.Println(string(BGreen), "Success!")

}

func combineAudios() (response string, errname error) {
	removeFile(introductionAudio)
	response, err := setAudioRate()
	if err != nil {
		return
	}
	command := `ffmpeg -f concat -safe 0 -i ` + voicelistFile + ` -c copy ` + introductionAudio
	response, err = runCommand(command)
	return response, err
}

func setAudioRate() (response string, errname error) {
	removeFile(headerAudioFile1)
	removeFile(headerAudioFile2)

	command := "ffmpeg -i " + audioFile1 + " -ar 44100 " + headerAudioFile1
	response, err := runCommand(command)
	if err != nil {
		return response, err
	}
	command = "ffmpeg -i " + audioFile2 + " -ar 44100 " + headerAudioFile2
	response, err = runCommand(command)
	if err != nil {
		return response, err
	}
	removeFile(audioFile1)
	removeFile(audioFile2)
	return "", nil
}

func generateDefaultImage(videoFiles []string) (response string, errname error) {
	removeFile(defaultImage)
	video1FileName := videoFiles[0]
	command := "ffmpeg -i " + video1FileName + " -ss 00:00:01 -vframes 1 " + defaultImage
	response, err := runCommand(command)
	return defaultImage, err
}

func generateIntroductionVideo(subtitles *astisub.Subtitles, videoFiles []string) (response string, errname error) {
	removeFile(introductionVideo)
	removeFile(headerAudioFile1)
	removeFile(headerAudioFile2)
	removeFile(introductionAudio)
	removeFile(defaultImage)

	//pepare raw meterials

	// generate line 1 voice
	textString := subtitles.Items[0].Lines[0].Items[0].Text
	//textToSpeech(IBMText2SpeechAPIKey, audioFile1, textString, 0) //For English
	textToSpeech(IBMText2SpeechAPIKey, audioFile1, textString, 1) // For Chinese

	// generate line 2 voice
	textString = subtitles.Items[0].Lines[1].Items[0].Text
	textToSpeech(IBMText2SpeechAPIKey, audioFile2, textString, 1)

	response, err := combineAudios()
	if err != nil {
		fmt.Println(response)
		return
	}

	response, err = generateDefaultImage(videoFiles)
	if err != nil {
		return
	}

	//generate video
	command := "ffmpeg -i " + defaultImage + " -i " + introductionAudio + " -filter:v fps=fps=30 -ac 2 " + introductionVideo
	response, err = runCommand(command)
	if err != nil {
		return response, err
	}
	removeFile(headerAudioFile1)
	removeFile(headerAudioFile2)
	removeFile(introductionAudio)
	removeFile(defaultImage)
	return "", nil
}

func generateNewName(originalName string, suffix string) string {
	name := filepath.Base(originalName)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	return tempFolder + name + suffix
}
func formatVideo(videoFileName string, tempFileName string) (response string, errname error) {
	//remove files
	removeFile(tempFileName)
	height, err := getVideoHeight(videoFileName)
	width, err := getVideoWidth(videoFileName)
	newHeight := int(float32(height) * (864.00 / float32(width)))
	if newHeight%2 != 0 {
		newHeight = newHeight - 1
	}
	command := "ffmpeg -i " + videoFileName + " -max_muxing_queue_size 9999 -vf scale=864:" + strconv.Itoa(newHeight) + " " + tempFileName
	response, err = runCommand(command)
	if err != nil {
		return response, err
	}
	removeFile(videoFileName)
	return "", nil
}

func getVideoHeight(videoFileName string) (height int, errname error) {
	heightCommand := "ffprobe -v error -select_streams v:0 -show_entries stream=height -of default=nw=1:nk=1 " + videoFileName
	fmt.Println(heightCommand)
	response, err := runCommand(heightCommand)
	if err != nil {
		return -1, err
	}
	response = strings.TrimSuffix(response, "\n")
	height, err = strconv.Atoi(response)
	if err != nil {
		return -1, err
	}
	return height, nil
}

func getVideoWidth(videoFileName string) (width int, errname error) {
	widthCommand := "ffprobe -v error -select_streams v:0 -show_entries stream=width -of default=nw=1:nk=1 " + videoFileName
	response, err := runCommand(widthCommand)
	if err != nil {
		return -1, err
	}
	response = strings.TrimSuffix(response, "\n")
	width, err = strconv.Atoi(response)
	if err != nil {
		return -1, err
	}
	return width, nil
}

func generateDrawText(subtitles *astisub.Subtitles, fileName string, videoID int) (response string, errname error) {
	height, err := getVideoHeight(fileName)
	if err != nil {
		return response, err
	}
	textString := ""
	if videoID == 0 {
		text1 := subtitles.Items[videoID].Lines[0].Items[0].Text
		text2 := subtitles.Items[videoID].Lines[1].Items[0].Text

		boxX := 100
		boxheight := 100
		textString := ""
		textString = textString + "drawbox=y=" + strconv.Itoa(startY) + ":color=yellow@0.8:width=iw:height=100:t=fill,drawtext=fontcolor=black:fontsize=60:fontfile=/Library/Fonts/SimHei.ttf" + ":x=20:y=" + strconv.Itoa(startY+20) + ":text=" + "\"" + text1 + "\""
		text2 = replaceSpecialChars(text2)
		textString = textString + "," + "drawbox=x=" + strconv.Itoa(boxX) + ":y=" + strconv.Itoa(startY+boxheight) + ":color=green@0.8:width=" + strconv.Itoa(864-2*boxX) + ":height=100:t=fill,drawtext=fontcolor=black:fontsize=50:fontfile=/Library/Fonts/SimHei.ttf" + ":x=" + strconv.Itoa(boxX+20) + ":y=" + strconv.Itoa(startY+30+boxheight) + ":text=" + "\"" + text2 + "\""
		return textString, nil
	}

	if (len(subtitles.Items[0].Lines)) > 2 {
		script := subtitles.Items[0].Lines[2].Items[0].Text
		script = replaceSpecialChars(script)
		textString = textString + "drawbox=y=" + strconv.Itoa(startY) + ":color=yellow@0.8:width=iw:height=100:t=fill,drawtext=fontcolor=black:fontsize=40:fontfile=/Library/Fonts/SimHei.ttf" + ":x=20:y=" + strconv.Itoa(startY+40) + ":text=" + "\"" + script + "\""

	}
	if (len(subtitles.Items[0].Lines)) > 3 {
		script := subtitles.Items[0].Lines[3].Items[0].Text
		script = replaceSpecialChars(script)
		textString = textString + ","
		textString = textString + "drawbox=y=" + strconv.Itoa(startY+100) + ":color=green@0.8:width=iw:height=100:t=fill,drawtext=fontcolor=black:fontsize=40:fontfile=/Library/Fonts/SimHei.ttf" + ":x=20:y=" + strconv.Itoa(startY+100+40) + ":text=" + "\"" + script + "\""
	}
	length := len(subtitles.Items[videoID].Lines)
	if length > 0 {
		textString = textString + ","
	}
	for i, lineText := range subtitles.Items[videoID].Lines {
		script := lineText.Items[0].Text
		fmt.Println(script)
		script = replaceSpecialChars(script)
		fmt.Println(script)
		y := videoPositionY + height + i*60
		textString = textString + "drawbox=y=" + strconv.Itoa(y) + ":color=yellow@0.8:width=iw:height=60:t=fill,drawtext=fontcolor=black:fontsize=40:fontfile=/Library/Fonts/SimHei.ttf" + ":x=20:y=" + strconv.Itoa(y+20) + ":text=" + "\"" + script + "\""
		if i < length-1 {
			textString = textString + ","
		}
	}
	return textString, nil
}

func changeVideoHeigh(fileName string, transformFileName string) (response string, errname error) {
	fmt.Println(fileName)
	fmt.Println(transformFileName)
	removeFile(transformFileName)
	//increase video height
	command := "ffmpeg -i " + fileName + " -max_muxing_queue_size 9999 " + " -vf pad=width=iw:height=1536:x=0:y=" + strconv.Itoa(videoPositionY) + ":color=violet,setdar=9/16 " + transformFileName
	response, err := runCommand(command)
	if err != nil {
		return response, err
	}
	removeFile(fileName)
	return response, nil
}

func addText2Video(textString string, inputFileName string, outputFileName string) (response string, errname error) {
	removeFile(outputFileName)
	command := `ffmpeg  -i ` + inputFileName + ` -vf [in]` + textString + `[out] -strict -2 -max_muxing_queue_size 1024 ` + outputFileName
	fmt.Println(command)
	cmd := exec.Command("bash", "-c", command)
	var out bytes.Buffer
	cmd.Stderr = os.Stderr
	//cmd.Stdout = os.Stdout
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", err
	}
	introductionVideoTransformFileName := generateNewName(introductionVideo, "-formated.mp4")
	if inputFileName == introductionVideoTransformFileName {
		removeFile(inputFileName)
	}
	copyFileName := generateNewName(inputFileName, "-copy.mp4")
	CopyFile(inputFileName, copyFileName)
	return "", nil
}
func replaceSpecialChars(script string) (reponse string) {
	script = strings.Replace(script, `'`, `\\\\\'`, -1)
	script = strings.Replace(script, `,`, `\\\\\,`, -1)
	script = strings.Replace(script, `:`, `\\\\\:`, -1)
	return script
}
func combineVideos() (response string, errname error) {
	var output = getOutputFileName()
	//get all formated videos and generate command
	videosFiles := getVideosFiles(tempFolder)
	n := len(videosFiles)
	command := `ffmpeg `
	for _, file := range videosFiles {
		command = command + ` -i ` + file + ` `
	}
	command = command + ` -filter_complex '`
	for i := 0; i < n; i++ {
		command = command + `[` + strconv.Itoa(i) + `v] `
		command = command + `[` + strconv.Itoa(i) + `a] `
	}
	command = command + `concat=n=` + strconv.Itoa(n)
	command = command + `:v=1:a=1 [v] [a]'`
	command = command + ` -map "[v]" -map "[a]" ` + output
	fmt.Println(command)
	cmd := exec.Command("bash", "-c", command)

	var out bytes.Buffer
	cmd.Stderr = os.Stderr
	//cmd.Stdout = os.Stdout
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", err
	}
	for _, file := range videosFiles {
		removeFile(file)
	}
	return out.String(), nil
}

func getOutputFileName() string {
	t := time.Now()
	var month int = int(t.Month())
	var filename = getTimeString(t.Year()) + "-" + getTimeString(month) + "-" + getTimeString(t.Day()) + "_" + getTimeString(t.Hour()) + "_" + getTimeString(t.Minute()) + "_" + getTimeString(t.Second())
	return outputPath + filename + ".mp4"
}
func getTimeString(num int) string {
	if num < 10 {
		return "0" + strconv.Itoa(num)
	}
	return strconv.Itoa(num)
}
