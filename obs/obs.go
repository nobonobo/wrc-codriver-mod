package obs

import (
	"encoding/base64"
	"image"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/events"
	"github.com/andreykaipov/goobs/api/requests/record"
	"github.com/andreykaipov/goobs/api/requests/sources"
	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.org/x/image/draw"
)

type Client struct {
	addr, password string
	client         *goobs.Client
}

func New(addr, password string) *Client {
	return &Client{addr: addr, password: password}
}

func (s *Client) Connect() error {
	if s.client != nil {
		return nil
	}
	client, err := goobs.New(s.addr, goobs.WithPassword(s.password))
	if err != nil {
		return err
	}
	s.client = client
	application.Get().EmitEvent("obs-event", &events.CustomEvent{
		EventData: map[string]any{
			"id": "connect",
		},
	})
	go client.Listen(func(event any) {
		switch event.(type) {
		case *events.ExitStarted:
			log.Println("ExitStarted")
			s.client.Disconnect()
			s.client = nil
		default:
			application.Get().EmitEvent("obs-event", event)
		}
	})
	return nil
}

func (s *Client) Disconnect() {
	if s.client == nil {
		return
	}
	s.client.Disconnect()
	s.client = nil
}

func (s *Client) GetRecordStatus() (*record.GetRecordStatusResponse, error) {
	if err := s.Connect(); err != nil {
		return nil, err
	}
	return s.client.Record.GetRecordStatus()
}

func (s *Client) StartRecord() error {
	if err := s.Connect(); err != nil {
		return err
	}
	if _, err := s.client.Record.StartRecord(); err != nil {
		return err
	}
	return nil
}

func (s *Client) PauseRecord() error {
	if err := s.Connect(); err != nil {
		return err
	}
	if _, err := s.client.Record.PauseRecord(); err != nil {
		return err
	}
	return nil
}

func (s *Client) ResumeRecord() error {
	if err := s.Connect(); err != nil {
		return err
	}
	if _, err := s.client.Record.ResumeRecord(); err != nil {
		return err
	}
	return nil
}

func (s *Client) StopRecord() error {
	if err := s.Connect(); err != nil {
		return err
	}
	if _, err := s.client.Record.StopRecord(); err != nil {
		return err
	}
	return nil
}

func (s *Client) ScreenShot(output string) error {
	if err := s.Connect(); err != nil {
		return err
	}
	params := sources.NewGetSourceScreenshotParams().
		WithSourceName("Game Capture").
		WithImageCompressionQuality(-1).
		WithImageFormat("png")
	res, err := s.client.Sources.GetSourceScreenshot(params)
	if err != nil {
		return err
	}
	data := res.ImageData[strings.IndexByte(res.ImageData, ',')+1:]
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
	img, err := png.Decode(reader)
	if err != nil {
		return err
	}
	bounds := img.Bounds()
	width := bounds.Dy() * 1920 / 1080
	offsetX := (bounds.Dx() - width) / 2
	log.Printf("bounds: %#v", bounds)
	rect := image.Rect(offsetX, bounds.Min.Y, bounds.Max.X-offsetX, bounds.Max.Y)
	// Crop the image
	croppedImage := img.(interface {
		SubImage(image.Rectangle) image.Image
	}).SubImage(rect)
	resizedImage := image.NewRGBA(image.Rect(0, 0, 1920, 1080))
	draw.BiLinear.Scale(resizedImage, resizedImage.Bounds(), croppedImage, croppedImage.Bounds(), draw.Over, nil)
	fp, err := os.Create(output)
	if err != nil {
		return err
	}
	defer fp.Close()
	if err := png.Encode(fp, resizedImage); err != nil {
		return err
	}
	return nil
}
