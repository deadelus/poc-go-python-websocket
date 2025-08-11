package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log"

	"github.com/gorilla/websocket"
	"gocv.io/x/gocv"
)

func sendFrameFromCamera() {
	webcam, err := gocv.OpenVideoCapture(0)
	if err != nil {
		log.Fatalf("cannot open camera: %v", err)
	}
	defer webcam.Close()

	window := gocv.NewWindow("Detections")
	defer window.Close()

	var lastTime = gocv.GetTickCount()
	var fps float64

	for {
		img := gocv.NewMat()
		defer img.Close()
		currTime := gocv.GetTickCount()
		elapsed := float64(currTime-lastTime) / gocv.GetTickFrequency()
		if elapsed > 0 {
			fps = 1.0 / elapsed
		}
		lastTime = currTime

		if ok := webcam.Read(&img); !ok || img.Empty() {
			log.Fatal("cannot read frame from camera")
		}

		buf, _ := gocv.IMEncode(".jpg", img)
		imgDecoded, _, err := image.Decode(bytes.NewReader(buf.GetBytes()))
		if err != nil {
			log.Fatalf("decode error: %v", err)
		}
		jpegBuf := new(bytes.Buffer)
		jpeg.Encode(jpegBuf, imgDecoded, nil)

		ws, _, err := websocket.DefaultDialer.Dial("ws://localhost:8765", nil)
		if err != nil {
			log.Fatal(err)
		}
		defer ws.Close()

		// Send params as first message (JSON list of strings)
		params := []string{"watch", "mug", "person", "keys"}
		paramsMsg, _ := json.Marshal(params)
		err = ws.WriteMessage(websocket.TextMessage, paramsMsg)
		if err != nil {
			log.Fatal(err)
		}

		// Send frame as second message
		err = ws.WriteMessage(websocket.BinaryMessage, jpegBuf.Bytes())
		if err != nil {
			log.Fatal(err)
		}

		_, resp, err := ws.ReadMessage()
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("Detections: %s\n", string(resp))

		// Parse detections and draw boxes
		var detections []struct {
			Class      string    `json:"class"`
			Confidence float64   `json:"confidence"`
			BBox       []float64 `json:"bbox"`
			Time       float64   `json:"time"`
		}
		err = json.Unmarshal(resp, &detections)
		if err != nil {
			fmt.Printf("Failed to parse detections: %v\n", err)
		} else {
			// Draw boxes on frame
			mat, _ := gocv.IMDecode(jpegBuf.Bytes(), gocv.IMReadColor)
			for _, det := range detections {
				if len(det.BBox) == 4 {
					rect := image.Rect(int(det.BBox[0]), int(det.BBox[1]), int(det.BBox[2]), int(det.BBox[3]))
					gocv.Rectangle(&mat, rect, color.RGBA{0, 255, 0, 0}, 2)
					label := fmt.Sprintf("%s %.2f", det.Class, det.Confidence)
					gocv.PutText(&mat, label, rect.Min, gocv.FontHersheySimplex, 0.7, color.RGBA{0, 255, 0, 0}, 2)
				}
			}
			window.IMShow(mat)
			gocv.WaitKey(1)
			// Show FPS on top right
			fpsLabel := fmt.Sprintf("FPS: %.2f", fps)
			size := mat.Size()
			pt := image.Pt(size[1]-150, 30)
			gocv.PutText(&mat, fpsLabel, pt, gocv.FontHersheySimplex, 0.7, color.RGBA{255, 0, 0, 0}, 2)

			// Show elapsed time (process time) below FPS
			if len(detections) > 0 {
				timeLabel := fmt.Sprintf("Process: %.4fs", detections[0].Time)
				ptTime := image.Pt(size[1]-150, 60)
				gocv.PutText(&mat, timeLabel, ptTime, gocv.FontHersheySimplex, 0.7, color.RGBA{255, 0, 0, 0}, 2)
			}

			window.IMShow(mat)
			gocv.WaitKey(1)
		}
	}
}

func main() {
	sendFrameFromCamera()
}
