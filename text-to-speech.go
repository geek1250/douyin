package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/IBM/go-sdk-core/v4/core"
	"github.com/watson-developer-cloud/go-sdk/texttospeechv1"
)

// type TextToSpeechV1 struct {
// 	Service *core.BaseService
// }

var service *texttospeechv1.TextToSpeechV1

func textToSpeech(fileName string, text string, languageCode int) {
	if service == nil {
		keyName := "IBMText2SpeechAPIKey"
		key, ok := os.LookupEnv(keyName)
		if !ok {
			fmt.Printf("%s not set\n", keyName)
			fmt.Println("Please set IBM Text to Speech API Key in environment variables")
			return
		} else {
			fmt.Printf("%s=%s\n", keyName, key)
		}
		// Instantiate the Watson Text To Speech service
		authenticator := &core.IamAuthenticator{
			ApiKey: key,
		}
		var serviceErr error = nil
		service, serviceErr = texttospeechv1.
			NewTextToSpeechV1(&texttospeechv1.TextToSpeechV1Options{
				Authenticator: authenticator,
			})

		// Check successful instantiation
		if serviceErr != nil {
			panic(serviceErr)
		}
		fmt.Println("UserAgent: " + service.Service.UserAgent)
	}

	/* SYNTHESIZE */
	//https://cloud.ibm.com/docs/text-to-speech?topic=text-to-speech-voices
	//zh-CN_WangWeiVoice, zh-CN_ZhangJingVoice, zh-CN_LiNaVoice
	//en-US_AllisonVoice, en-GB_KateVoice
	synthesizeOptions := service.NewSynthesizeOptions(text).SetAccept("audio/mp3").SetVoice("zh-CN_LiNaVoice")
	if languageCode == 0 {
		synthesizeOptions = service.NewSynthesizeOptions(text).
			SetAccept("audio/mp3").
			SetVoice("en-US_AllisonVoice")
	} else {
		synthesizeOptions = service.NewSynthesizeOptions(text).
			SetAccept("audio/mp3").
			SetVoice("zh-CN_LiNaVoice")
	}
	// Call the textToSpeech Synthesize method
	synthesizeResult, _, responseErr := service.Synthesize(synthesizeOptions)

	// Check successful call
	if responseErr != nil {
		panic(responseErr)
	}

	// Check successful casting
	if synthesizeResult != nil {
		buff := new(bytes.Buffer)
		buff.ReadFrom(synthesizeResult)

		file, _ := os.Create(fileName)
		file.Write(buff.Bytes())
		file.Close()

		fmt.Println("Wrote synthesized text to " + fileName)
	}
	synthesizeResult.Close()

	/* SYNTHESIZE USING WEBSOCKET*/
	// // create a file for websocket output
	// fileName := "synthesize_ws_example_output.mp3"
	// file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0777)
	// if err != nil {
	// 	panic(err)
	// }

	// callback := myCallBack{f: file}

	// synthesizeUsingWebsocketOptions := service.
	// 	NewSynthesizeUsingWebsocketOptions("This is a <mark name=\"SIMPLE\"/>simple <mark name=\"EXAMPLE\"/> example.", callback)

	// synthesizeUsingWebsocketOptions.
	// 	SetAccept("audio/mp3").
	// 	SetVoice("en-US_AllisonVoice")
	// synthesizeUsingWebsocketOptions.SetTimings([]string{"words"})
	// err = service.SynthesizeUsingWebsocket(synthesizeUsingWebsocketOptions)
	// if err != nil {
	// 	fmt.Println(err)
	// }
}

type myCallBack struct {
	f *os.File
}

func (cb myCallBack) OnOpen() {
	fmt.Println("Handshake successful")
}

func (cb myCallBack) OnClose() {
	fmt.Println("Closing connection")
	cb.f.Close()
}

func (cb myCallBack) OnAudioStream(b []byte) {
	bytes, err := ioutil.ReadAll(bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
	_, err = cb.f.Write(bytes)
	if err != nil {
		panic(err)
	}
}

func (cb myCallBack) OnData(response *core.DetailedResponse) {}

func (cb myCallBack) OnError(err error) {
	fmt.Println("Received error")
	panic(err)
}

func (cb myCallBack) OnTimingInformation(timings texttospeechv1.Timings) {
	core.PrettyPrint(timings, "Timing information: ")
}

func (cb myCallBack) OnMarks(marks texttospeechv1.Marks) {
	core.PrettyPrint(marks, "Mark timings: ")
}

func (cb myCallBack) OnContentType(contentType string) {
	fmt.Println("The content type identified is:", contentType)
}
