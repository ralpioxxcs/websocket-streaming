let ws = new WebSocket("localhost:8080/stream")

const imgElem = document.getElementById("imgBox");

ws.onmessage = (ev) => {
  var msg = ev.data;

  imgElem.setAttribute('src', `data:image/jpeg;base64.` + msg);
}

ws.onopen = () => {
  console.log("websocket connected");
}

ws.onclose = () => {
  console.log("websocket closed");
}

ws.onerror = (err) => {
  console.log("websocket err", err);
}