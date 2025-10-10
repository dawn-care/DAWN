const { app, BrowserWindow } = require("electron");
const path = require("path");

function createWindow() {
  const win = new BrowserWindow({
    width: 1200,
    height: 800,
    webPreferences: {
      nodeIntegration: false,
    },
  });

  // Point to Expo web build
  win.loadURL("http://localhost:8081");
}

app.whenReady().then(createWindow);
