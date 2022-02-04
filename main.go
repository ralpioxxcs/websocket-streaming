package main

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"gocv.io/x/gocv"
)

const (
	// v4l2 interface
	v4l2PipelineFmt = `
	v4l2src device=/dev/video%d 
	! image/jpeg, format=MJPG, width=%d, height=%d, framerate=%d/1 
	! jpegdec 
	! videoconvert
	! video/x-raw, format=(string)BGR
	! appsink`

	// nvarguscamerasrc interface
	nvargusPipelineFmt = `
	nvarguscamerasrc device=%d
	! video/x-raw(memory:NVMM), width=%d, height=%d, framerate=%d/1 
	! nvvidconv !
	! video/x-raw, format=BGRx
	! videoconvert
	! video/x-raw, format=BGR
	! appsink`

	deviceNum = 0
	width     = 1920
	height    = 1080
	fps       = 30
)

var (
	WsConn   *websocket.Conn
	imgBytes []byte

	pauseLoop chan bool = make(chan bool)

	templates *template.Template
)

func openCam() error {
	pl := fmt.Sprintf(v4l2PipelineFmt, deviceNum, width, height, fps)

	cap, err := gocv.OpenVideoCaptureWithAPI(pl, gocv.VideoCaptureGstreamer)
	if err != nil {
		log.Fatal("failed to open capture device", err)
		return err
	}

	img := gocv.NewMat()
	defer img.Close()
	for {
		cap.Read(&img)
		bytes, _ := gocv.IMEncode(".jpg", img)
		imgBytes = bytes.GetBytes()
	}
}

func homePage(rw http.ResponseWriter, r *http.Request) {
	templates.Execute(rw, "Streaming")
}

func stream(rw http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	var err error
	WsConn, err = upgrader.Upgrade(rw, r, nil)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer WsConn.Close()

	// send
	go func() {
		for {
			select {
			case pause := <-pauseLoop:
				if pause {
					<-pauseLoop
				}
			default:
				enc := base64.StdEncoding.EncodeToString(imgBytes)
				err := WsConn.WriteMessage(websocket.TextMessage, []byte(enc))
				if err != nil {
					log.Printf("failed to write : %v", err)
				}
				// time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	pauseLoop <- true

	// recv
	for {
		_, bytes, err := WsConn.ReadMessage()
		if err != nil {
			fmt.Printf("read error : %v\n", err)
			break
		}

		msg := string(bytes)
		if msg == "start" {
			fmt.Println("start")
			pauseLoop <- true
		} else if msg == "stop" {
			fmt.Println("stop")
			pauseLoop <- true
		}
	}

}

func main() {
	go openCam()

	templates = template.Must(templates.ParseGlob("./public/index.html"))

	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/static/", http.StripPrefix("/static", fs)) // file serving

	http.HandleFunc("/", homePage) // '/' route
	http.HandleFunc("/ws", stream) // websocket '/ws' route

	fmt.Println("Listening on http://localhost:8080")
	log.Fatalln(http.ListenAndServe("[::]:8080", nil))
}
