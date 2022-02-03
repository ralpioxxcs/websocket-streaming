package main

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

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
	Upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		}}

	imgBytes []byte

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

func home(rw http.ResponseWriter, r *http.Request) {
	templates.Execute(rw, "Streaming")
}

func handleStream(rw http.ResponseWriter, r *http.Request) {
	ws, err := Upgrader.Upgrade(rw, r, nil)
	if err != nil {
		// msg := fmt.Sprintf("failed to init websocket (%v)", err)
		// fmt.Println(msg)
		// http.Error(rw, msg, http.StatusInternalServerError)
		return
	}

	// recv
	go func() {
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				break
			}
		}
	}()

	res := map[string]interface{}{}

	// send
	for {
		enc := base64.StdEncoding.EncodeToString(imgBytes)
		res["enc"] = enc
		err := ws.WriteJSON(&res)
		if err != nil {
			log.Printf("failed to write : %v", err)
		}

		time.Sleep(100 * time.Millisecond)
	}

}

func main() {
	go openCam()

	templates = template.Must(templates.ParseGlob("public/index.html"))

	http.Handle("/static", http.FileServer(http.Dir("./public")))

	http.HandleFunc("/", home)
	http.HandleFunc("/stream", handleStream)

	fmt.Println("Listening on http://localhost:8080")
	log.Fatalln(http.ListenAndServe("[::]:8080", nil))
}
