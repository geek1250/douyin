## Usage
go mod init douyinService 

before run:
1. add video in resource/videos folder
2. edit content of scripts.srt

run:
go run douyin.go utils.go text-to-speech.go


generate binary executable file:

go build -o douyinService douyin.go utils.go text-to-speech.go
will generate an executable file named douyinService

go build douyin.go utils.go text-to-speech.go
will generate an executable file named main


Please note: 
ffprobe is required and installed in /usr/local/bin/