package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	obs_events "github.com/andreykaipov/goobs/api/events"
	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	"github.com/nobonobo/easportswrc/packet"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/icons"
	"golang.org/x/image/draw"

	"wrc-codriver-mod/obs"
	"wrc-codriver-mod/settings"
	"wrc-codriver-mod/telemetry"
)

var config = struct {
	ObsAddress        string `env:"OBS_ADDRESS" envDefault:"localhost:4455"`
	ObsPassword       string `env:"OBS_PASSWORD"`
	YouTubeUploader   string `env:"YOUTUBE_UPLOADER"`
	YouTubeSecret     string `env:"YOUTUBE_SECRET"`
	YouTubePublic     bool   `env:"YOUTUBE_PUBLIC" envDefault:"false"`
	ThumbnailTemplate string `env:"THUMBNAIL_TEMPLATE"`
	VoiceVoxCli       string `env:"VOICEVOX_CLI"`
}{}

var tempDir = os.TempDir()

//go:embed frontend/dist
var assets embed.FS

func init() {
	log.SetFlags(log.Lshortfile)
	fpath := filepath.Join(packet.WrcRoot, "wrc-codriver-mod.env")
	if err := godotenv.Load(fpath); err != nil {
		log.Fatal(err)
	}
	if err := env.Parse(&config); err != nil {
		log.Fatal(err)
	}
}

var descTemplate = template.Must(template.New("").Parse(`#EASportsWRC
game mode: {{.Mode}}
class: {{.Class}}
location: {{.Location}}
route: {{.Route}}
manufacturer: {{.Manufacturer}}
vehicle: {{.Vehicle}}
length: {{.Length}}
time: {{.Time}}
penalty: {{.Penalty}}
`))

func youtubeUpload(ctx context.Context, src string, info telemetry.Event) (string, error) {
	if len(config.YouTubeUploader) == 0 {
		return "", nil
	}
	buff := bytes.NewBuffer(nil)
	if err := descTemplate.Execute(buff, info); err != nil {
		return "", err
	}
	name := strings.Join([]string{
		"#EASportsWRC", info.Mode,
		info.Route, info.Vehicle,
	}, " - ")
	ssPath := filepath.Join(tempDir, name+".png")
	d := filepath.Dir(src)
	screenshot := filepath.Join(d, name+".jpg")
	if err := reduceJPEG(ssPath, screenshot); err != nil {
		return "", err
	}
	meta := filepath.Join(tempDir, "meta.json")
	if err := os.WriteFile(meta, []byte(`{"playlistIds": ["PLNhVzDfOlkDhEHJyp0VwUsYKZjyGDrj_i"]}`), 0644); err != nil {
		return "", err
	}
	args := []string{
		"-secrets", config.YouTubeSecret,
		"-title", strings.Join([]string{
			"#EASportsWRC", info.Mode,
			info.Route, info.Vehicle,
		}, " / "),
		"-description", buff.String(),
		"-language", "ja",
		"-quiet",
		"-filename", src,
		"-thumbnail", screenshot,
		"-metaJSON", meta,
	}
	if config.YouTubePublic {
		args = append(args, "-privacy", "public")
	}
	log.Println(config.YouTubeUploader, args)
	scripts := []string{
		config.YouTubeUploader,
	}
	for _, v := range args {
		if strings.Contains(v, " ") {
			scripts = append(scripts, `"`+v+`"`)
		} else {
			scripts = append(scripts, v)
		}
	}
	script := strings.Join(scripts, " ")
	cmd := exec.CommandContext(ctx, config.YouTubeUploader, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = packet.WrcRoot
	return script, cmd.Run()
}

func loadPNG(fname string) (image.Image, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}
func saveJPEG(fname string, img image.Image) (int64, error) {
	f, err := os.Create(fname)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: jpeg.DefaultQuality}); err != nil {
		return 0, err
	}
	if err := f.Sync(); err != nil {
		return 0, err
	}
	info, err := f.Stat()
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func reduceJPEG(input, output string) error {
	src, err := loadPNG(input)
	if err != nil {
		return err
	}
	bounds := src.Bounds()
	dst := image.NewRGBA(image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y))
	draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	size, err := saveJPEG(output, dst)
	if err != nil {
		return err
	}
	if size > 2*1024*1024 {
		return fmt.Errorf("file size too large: %d", size)
	}
	return nil
}

func appMain(ctx context.Context, speak func(string)) {
	fpath := filepath.Join(packet.WrcRoot, "wrc-codriver-mod.config.json")
	settingsService := settings.New(fpath)
	obsClient := obs.New(config.ObsAddress, config.ObsPassword)
	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Bind' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running an macOS.
	app := application.New(application.Options{
		Name:        "wrc-codriver-mod",
		Description: "EA Sports™ WRC Codriver Mod",
		Services: []application.Service{
			application.NewService(obsClient),
			application.NewService(settingsService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	systemTray := app.NewSystemTray()

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	window := app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
		Title:     "EA Sports™ WRC Codriver Mod",
		Hidden:    false,
		MinWidth:  400,
		MinHeight: 400,
		Width:     400,
		Height:    400,
		ShouldClose: func(window *application.WebviewWindow) bool {
			window.Hide()
			return false
		},
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})
	log.Print(window)

	if runtime.GOOS == "darwin" {
		systemTray.SetTemplateIcon(icons.SystrayMacTemplate)
	}

	// building menu
	menu := app.NewMenu()
	item := menu.Add("Wails")
	item.SetBitmap(icons.WailsLogoBlackTransparent).OnClick(func(ctx *application.Context) {
		window.Show()
	})
	menu.AddSeparator()
	menu.Add("Quit").OnClick(func(ctx *application.Context) {
		app.Quit()
	})
	systemTray.SetMenu(menu)
	//systemTray.AttachWindow(window).WindowOffset(5)

	svc, err := telemetry.Start(ctx, "localhost:20777")
	if err != nil {
		log.Fatal(err)
	}
	log.Print(svc)
	if err := obsClient.Connect(); err != nil {
		log.Print(err)
	}
	defer obsClient.Disconnect()
	lastRecordState := telemetry.Event{}
	app.Events.On("obs-event", func(event *application.WailsEvent) {
		log.Printf("%s: %#v", event.Name, event.Data)
		switch e := event.Data.(type) {
		default:
			log.Println("obs-event:", event.Data)
		case *obs_events.CustomEvent:
			s, ok := e.EventData["id"]
			if ok && s == "connect" {
				speak("OBSと接続しました。")
			}
		case *obs_events.ExitStarted:
			speak("OBSが終了しました。")
		case *obs_events.RecordStateChanged:
			if e.OutputState != "OBS_WEBSOCKET_OUTPUT_STOPPED" {
				return
			}
			if !e.OutputActive && !lastRecordState.Recording {
				targetName := strings.Join([]string{
					"#EASportsWRC", lastRecordState.Mode,
					lastRecordState.Route, lastRecordState.Vehicle,
				}, " - ")
				dir := filepath.Dir(e.OutputPath)
				targetPath := filepath.Join(dir, targetName)
				if lastRecordState.Result != "finished" {
					unfinishedDir := filepath.Join(dir, "not_finished")
					os.MkdirAll(unfinishedDir, 0755)
					targetPath = filepath.Join(unfinishedDir, targetName)
				}
				if err := os.Rename(e.OutputPath, targetPath+".mkv"); err != nil {
					log.Print(err)
				}
				if lastRecordState.Result == "finished" {
					speak("録画を完了しました。")
					if settingsService.Get().AutoYouTubeUpload {
						script, err := youtubeUpload(ctx, targetPath+".mkv", lastRecordState)
						if err != nil {
							log.Println(err)
							speak("ユーチューブにアップロード失敗しました。")
							os.WriteFile(filepath.Join(dir, targetName+".bat"), []byte(script), 0755)
						} else {
							uploadDir := filepath.Join(dir, "uploaded")
							os.MkdirAll(uploadDir, 0755)
							if err := os.Rename(targetPath+".mkv", filepath.Join(uploadDir, targetName+".mkv")); err != nil {
								log.Print(err)
							}
							if err := os.Rename(targetPath+".jpg", filepath.Join(uploadDir, targetName+".png")); err != nil {
								log.Print(err)
							}
							speak("ユーチューブにアップロード完了しました。")
						}
					}
				}
			}
		}
	})
	paused := false
	app.Events.On("recording", func(event *application.WailsEvent) {
		lastRecordState = event.Data.(telemetry.Event)
		log.Printf("%s: %#v", event.Name, lastRecordState)
		if lastRecordState.Recording {
			if paused && lastRecordState.Progress > 0 {
				paused = false
				if err := obsClient.ResumeRecord(); err != nil {
					log.Print(err)
				}
			} else {
				if lastRecordState.Shakedown {
					speak("シェイクダウンです。安全運転をお願いします。")
					return
				}
				speak("録画を開始します。")
				status, err := obsClient.GetRecordStatus()
				if err != nil {
					log.Print(err)
				}
				if status.OutputActive {
					if err := obsClient.StopRecord(); err != nil {
						log.Print(err)
					}
				}
				name := strings.Join([]string{
					"#EASportsWRC", lastRecordState.Mode,
					lastRecordState.Route, lastRecordState.Vehicle,
				}, " - ")
				ssPath := filepath.Join(tempDir, name+".png")
				if _, err := os.Stat(ssPath); os.IsNotExist(err) {
					//speak("スクリーンショット")
					if err := obsClient.ScreenShot(ssPath); err != nil {
						log.Print(err)
					}
				}
				if err := obsClient.StartRecord(); err != nil {
					log.Print(err)
				}
			}
		} else {
			if lastRecordState.Result != "finished" {
				paused = true
				if err := obsClient.PauseRecord(); err != nil {
					log.Print(err)
				}
			} else {
				log.Println("finished!")
				if settingsService.Get().AutoStopRecording {
					if err := obsClient.StopRecord(); err != nil {
						log.Print(err)
					}
				}
			}
		}
	})
	app.Events.On("finished", func(event *application.WailsEvent) {
		state := event.Data.(telemetry.Event)
		log.Printf("%s: %#v", event.Name, state)
		speak("フィニッシュ！")
	})
	var lastTyreState telemetry.TyreState
	app.Events.On("tyre-state", func(event *application.WailsEvent) {
		state := event.Data.(telemetry.TyreState)
		defer func() {
			lastTyreState = state
		}()
		log.Printf("%s: %#v", event.Name, state)
		changed := state.ForwardLeft != lastTyreState.ForwardLeft ||
			state.ForwardRight != lastTyreState.ForwardRight ||
			state.BackwordLeft != lastTyreState.BackwordLeft ||
			state.BackwordRight != lastTyreState.BackwordRight
		if !changed {
			return
		}
		punctured := []string{}
		burst := []string{}
		names := []string{"左前", "右前", "左後ろ", "右後ろ"}
		for i, v := range []string{state.ForwardLeft,
			state.ForwardRight,
			state.BackwordLeft,
			state.BackwordRight,
		} {
			switch v {
			case "punctured":
				punctured = append(punctured, names[i])
			case "burst":
				burst = append(burst, names[i])
			}
		}
		text := []string{}
		if len(punctured) > 0 {
			text = append(text, strings.Join(punctured, "と")+"がパンク、")
		}
		if len(burst) > 0 {
			text = append(text, strings.Join(burst, "と")+"がバースト、")
		}
		if len(text) > 0 {
			text = append(text, "しちゃってます！")
		}
		time.AfterFunc(1500*time.Millisecond, func() { speak(strings.Join(text, "")) })
	})
	app.Events.On("packet", func(event *application.WailsEvent) {
		pkt := event.Data.(*packet.Packet)
		if pkt.PacketUID%600 == 0 {
			log.Printf("packet: %#v", pkt)
		}
	})
	// Run the application. This blocks until the application has been exited.
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reader, writer := io.Pipe()
	cmd := exec.CommandContext(ctx, config.VoiceVoxCli)
	cmd.Stdin = reader
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	appMain(ctx, func(text string) {
		log.Println("speak:", text)
		fmt.Fprintln(writer, text)
	})
}
