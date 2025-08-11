import asyncio
import websockets
import json
import cv2
import numpy as np
from ultralytics import YOLO

model = YOLO("models/yoloe-11s-seg.pt")

async def detect_frame(websocket):
    # Expect first message to be a JSON list of strings (classes)
    try:
        # Receive params
        params_msg = await websocket.recv()
        try:
            names = json.loads(params_msg)
            if not isinstance(names, list) or not all(isinstance(n, str) for n in names):
                raise ValueError
        except Exception:
            await websocket.send(json.dumps({"error": "Expected a JSON list of strings for params"}))
            return
        model.set_classes(names, model.get_text_pe(names))

        async for message in websocket:
            import time
            start_time = time.time()
            nparr = np.frombuffer(message, np.uint8)
            frame = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
            if frame is None:
                await websocket.send(json.dumps({"error": "Image decoding failed"}))
                continue
            
            results = model.predict(
                frame,
                imgsz=640,
                verbose=False,
                conf=0.5,
                iou=0.8
            )
            elapsed = time.time() - start_time

            detections = []
            for r in results:
                for box in r.boxes:
                    detections.append({
                        "class": model.names[int(box.cls)],
                        "confidence": float(box.conf),
                        "bbox": box.xyxy[0].tolist(),
                        "time": elapsed
                    })

            await websocket.send(json.dumps(detections))
            break
    except websockets.exceptions.ConnectionClosed:
        pass

async def main():
    print("Serveur WebSocket YOLOe démarré sur ws://0.0.0.0:8765")
    async with websockets.serve(detect_frame, "0.0.0.0", 8765):
        await asyncio.Future()

if __name__ == "__main__":
    asyncio.run(main())