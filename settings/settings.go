package settings

import (
	"encoding/json"
	"log"
	"os"
)

type Info struct {
	AutoStartRecording bool `json:"auto_start_recording"`
	AutoStopRecording  bool `json:"auto_stop_recording"`
	AutoYouTubeUpload  bool `json:"auto_youtube_upload"`
}

type Settings struct {
	fpath    string
	settings Info
}

func New(fpath string) *Settings {
	settings := Info{
		AutoStartRecording: true,
		AutoStopRecording:  true,
		AutoYouTubeUpload:  false,
	}
	b, err := os.ReadFile(fpath)
	if err != nil {
		log.Print(err)
	} else {
		if err := json.Unmarshal(b, &settings); err != nil {
			log.Print(err)
		}
	}
	return &Settings{
		fpath:    fpath,
		settings: settings,
	}
}

func (s *Settings) Get() Info {
	return s.settings
}

func (s *Settings) Update(settings Info) {
	s.settings = settings
	b, err := json.Marshal(settings)
	if err != nil {
		log.Print(err)
	} else {
		if err := os.WriteFile(s.fpath, b, 0644); err != nil {
			log.Print(err)
		}
	}
}
