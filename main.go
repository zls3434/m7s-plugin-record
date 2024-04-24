package record

import (
	_ "embed"
	"errors"
	"io"
	"sync"

	"github.com/zls3434/m7s-engine/v4"
	"github.com/zls3434/m7s-engine/v4/codec"
	"github.com/zls3434/m7s-engine/v4/config"
	"github.com/zls3434/m7s-engine/v4/util"
)

type RecordConfig struct {
	config.Subscribe
	Flv        Record `desc:"flv录制配置"`
	Mp4        Record `desc:"mp4录制配置"`
	Fmp4       Record `desc:"fmp4录制配置"`
	Hls        Record `desc:"hls录制配置"`
	Raw        Record `desc:"视频裸流录制配置"`
	RawAudio   Record `desc:"音频裸流录制配置"`
	recordings sync.Map
}

//go:embed default.yaml
var defaultYaml engine.DefaultYaml
var ErrRecordExist = errors.New("recorder exist")
var RecordPluginConfig = &RecordConfig{
	Flv: Record{
		Path:          "record/flv",
		Ext:           ".flv",
		GetDurationFn: getFLVDuration,
	},
	Fmp4: Record{
		Path: "record/fmp4",
		Ext:  ".mp4",
	},
	Mp4: Record{
		Path: "record/mp4",
		Ext:  ".mp4",
	},
	Hls: Record{
		Path: "record/hls",
		Ext:  ".m3u8",
	},
	Raw: Record{
		Path: "record/raw",
		Ext:  ".", // 默认h264扩展名为.h264,h265扩展名为.h265
	},
	RawAudio: Record{
		Path: "record/raw",
		Ext:  ".", // 默认aac扩展名为.aac,pcma扩展名为.pcma,pcmu扩展名为.pcmu
	},
}

var plugin = engine.InstallPlugin(RecordPluginConfig, defaultYaml)

func (conf *RecordConfig) OnEvent(event any) {
	switch v := event.(type) {
	case engine.FirstConfig, config.Config:
		conf.Flv.Init()
		conf.Mp4.Init()
		conf.Fmp4.Init()
		conf.Hls.Init()
		conf.Raw.Init()
		conf.RawAudio.Init()
	case engine.SEpublish:
		streamPath := v.Target.Path
		if conf.Flv.NeedRecord(streamPath) {
			go NewFLVRecorder().Start(streamPath)
		}
		if conf.Mp4.NeedRecord(streamPath) {
			go NewMP4Recorder().Start(streamPath)
		}
		if conf.Fmp4.NeedRecord(streamPath) {
			go NewFMP4Recorder().Start(streamPath)
		}
		if conf.Hls.NeedRecord(streamPath) {
			go NewHLSRecorder().Start(streamPath)
		}
		if conf.Raw.NeedRecord(streamPath) {
			go NewRawRecorder().Start(streamPath)
		}
		if conf.RawAudio.NeedRecord(streamPath) {
			go NewRawAudioRecorder().Start(streamPath)
		}
	}
}
func (conf *RecordConfig) getRecorderConfigByType(t string) (recorder *Record) {
	switch t {
	case "flv":
		recorder = &conf.Flv
	case "mp4":
		recorder = &conf.Mp4
	case "fmp4":
		recorder = &conf.Fmp4
	case "hls":
		recorder = &conf.Hls
	case "raw":
		recorder = &conf.Raw
	case "raw_audio":
		recorder = &conf.RawAudio
	}
	return
}

func getFLVDuration(file io.ReadSeeker) uint32 {
	_, err := file.Seek(-4, io.SeekEnd)
	if err == nil {
		var tagSize uint32
		if tagSize, err = util.ReadByteToUint32(file, true); err == nil {
			_, err = file.Seek(-int64(tagSize)-4, io.SeekEnd)
			if err == nil {
				_, timestamp, _, err := codec.ReadFLVTag(file)
				if err == nil {
					return timestamp
				}
			}
		}
	}
	return 0
}
