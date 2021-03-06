package main

import (
	"bytes"
	"github.com/deepch/vdk/format/rtmp"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
)

type Rtmp struct {
	Url   string
	Close bool
	Open  bool
}

const ffmpegGetRtmp = "-i {{.}} -vcodec libx264 -max_muxing_queue_size 1024 -movflags empty_moov+omit_tfhd_offset+frag_keyframe+default_base_moof -max_interleave_delta 0 -frag_duration 1 -b:v 1024k -c:a aac -f mp4 pipe:1"
const mp4boxHLS = "-i stdin -o file.m3u8"

func (self *Rtmp) runMP4Box(cmd *exec.Cmd) {

	err := cmd.Start()

	if err != nil {
		log.Err(err).Msg("gpac error")
	}
	cmd.Wait()
	self.Close = true
}

func main() {
	var stream Rtmp
	var err error
	baseUrl := os.Args[1]

	r, w := io.Pipe()

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
		if stream.Close {
			_, err = rtmp.Dial(baseUrl)
			if err == nil {
				stream.Close = false
			}
		}
		if !stream.Close && !stream.Open {
			stream.Open = false
			com := bytes.NewBuffer(nil)
			err = tmp.Execute(com, stream.Url)
			if err != nil {
				log.Error().Err(err).Msg("tmp.Execute")
				return
			}
			cmd := exec.Command("ffmpeg", strings.Split(com.String(), " ")...)
			cmd.Stdout = w
			cmdmp := exec.Command("gpac", strings.Split(mp4boxHLS, " ")...)
			cmdmp.Stdin = r

			go stream.runMP4Box(cmdmp)

			cmd.Stderr = os.Stdout
			err := cmd.Start()

			if err != nil {
				stream.Open = false
				log.Err(err).Msg("ffmpeg error")
			} else {
				stream.Open = true
			}
			for !stream.Close {
			}
		}
	}
}
