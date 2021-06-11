package main

import (
	"bytes"
	"github.com/deepch/vdk/format/rtmp"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"regexp"
	"text/template"
)

type Rtmp struct {
	Url   string
	Close bool
	Open  bool
}

const ffmpegGetRtmp = "ffmpeg  -i {{.}} -vcodec libx264 -max_muxing_queue_size 1024 -movflags empty_moov+omit_tfhd_offset+frag_keyframe+default_base_moof -max_interleave_delta 0 -frag_duration 1 -b:v 1024k -c:a aac -f mp4 pipe:1 > mypipe"
const mp4boxHLS = "gpac -i pipe://mypipe:ext=avc -o file.m3u8"

func runMP4Box(stream Rtmp) {
	for stream.Open {
		cmd := exec.Command(mp4boxHLS)
		cmd.Run()
	}
}

func main() {
	var stream Rtmp
	var err error
	baseUrl := os.Args[1]

	rtmpHave, err := regexp.MatchString(`^rtmp*`, baseUrl)
	if err != nil {
		log.Print(err)
	}
	tmp, err := template.New("rtmp").Parse(ffmpegGetRtmp)
	if err != nil {
		log.Error().Err(err).Msg("template.New")
		return
	}
	if rtmpHave {
		stream.Url = baseUrl
		stream.Close = true
		stream.Open = false
		for {
			if stream.Close {
				_, err = rtmp.Dial(baseUrl)
				if err == nil {
					stream.Close = false
				}
			}
			if !stream.Close && !stream.Open {
				stream.Open = true
				com := bytes.NewBuffer(nil)
				err = tmp.Execute(com, stream.Url)
				if err != nil {
					log.Error().Err(err).Msg("tmp.Execute")
					continue
				}
				log.Print(com.String())
				cmd := exec.Command(com.String())
				go runMP4Box(stream)
				err := cmd.Run()
				if err != nil {
					stream.Open = false
					log.Print(err)
					log.Print("ОШИБКА ПОСЛЕ RUN")
				} else {
					log.Print("vse good")
				}
			}

		}
	}
}
