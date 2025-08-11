FROM ultralytics/ultralytics:latest

# Install missing dependencies for WebSocket server
RUN pip install websockets opencv-python numpy

# Copy server code and models
#COPY server.py /ultralytics/server.py
#COPY models /ultralytics/models

WORKDIR /ultralytics

CMD ["python", "server.py"]