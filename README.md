# poc-go-python-websocket

A demo project for real-time object detection using a Python WebSocket server (YOLO) and a Go client with OpenCV.

## Features

- Python server with YOLO model for detection
- Go client captures webcam frames, sends to server, and displays results
- Detection boxes, labels, FPS, and process time drawn on frames
- Docker Compose setup for easy deployment

## Usage with Docker Compose

1. **Build and start all services:**
   ```sh
   docker-compose up --build
   ```

2. **Python server** runs in its own container and exposes WebSocket on port 8765.

3. **Go client** runs in a separate container and connects to the server.

## Requirements

- Docker and Docker Compose installed

## Repository

https://github.com/deadelus/poc-go-python-websocket
