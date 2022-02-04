const endpoint = '/ws';
const wsUrl = 'ws://' + window.location.hostname + ':' + window.location.port + endpoint;
let ws = new WebSocket(wsUrl);

let canvas = document.getElementById("canvas");
let canvasCtx = canvas.getContext("2d");
let canvasImg = new Image();

ws.onmessage = (ev) => {
  canvasImg.onload = () => {
    canvasCtx.drawImage(canvasImg, 0, 0);
  };
  canvasImg.src = `data:image/jpeg;base64,` + ev.data; // base64 encoded 
};

ws.onopen = (ev) => {
  console.log("websocket connected");

  ws.send("connection establised");
};

ws.onclose = (ev) => {
  if (ev.wasClean) {
    console.log("websocket closed gracefully", ev.code);
  } else {
    console.log("websocket closed", ev.code)
  }
};

ws.onerror = (err) => {
  console.log('websocket error: ', err);
}

const playCamera = () => {
  ws.send("start")
  toggleButton(true);
}

const stopCamera = () => {
  ws.send("stop")
  toggleButton(false);
}

const toggleButton = (start) => {
  if (start) {
    document.getElementById("playBtn").disabled = true;
    document.getElementById("stopBtn").disabled = false;
  } else {
    document.getElementById("playBtn").disabled = false;
    document.getElementById("stopBtn").disabled = true;
  }
};

window.playCamera = playCamera;
window.stopCamera = stopCamera;