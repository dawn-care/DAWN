const { app, BrowserWindow, protocol } = require("electron");
const path = require("path");
const fs = require("fs");
const { pathToFileURL } = require("url");

// Register a custom "app://" protocol that serves files from the Expo web
// export with SPA fallback (unknown paths return index.html).  This gives
// client-side routing full History API support without an HTTP server.
protocol.registerSchemesAsPrivileged([
  {
    scheme: "app",
    privileges: {
      standard: true,   // enables History API, fetch, etc.
      secure: true,      // treated as a secure context
      supportFetchAPI: true,
      corsEnabled: true,
    },
  },
]);

function createWindow() {
  const win = new BrowserWindow({
    width: 1200,
    height: 800,
    webPreferences: {
      nodeIntegration: false,
    },
  });

  if (app.isPackaged) {
    const distDir = path.join(__dirname, "dist");

    // Handle every "app://" request by resolving it to a file in dist/
    protocol.handle("app", (request) => {
      const url = new URL(request.url);
      // Strip the leading slash so path.join works correctly
      const relativePath = decodeURIComponent(url.pathname).replace(/^\//, "");
      let filePath = path.join(distDir, relativePath);

      // SPA fallback: if the file doesn't exist, serve index.html
      if (!fs.existsSync(filePath) || fs.statSync(filePath).isDirectory()) {
        filePath = path.join(distDir, "index.html");
      }

      return new Response(fs.readFileSync(filePath), {
        headers: { "Content-Type": getMime(filePath) },
      });
    });

    win.loadURL("app://bundle/");
  } else {
    win.loadURL("http://localhost:8081");
  }
}

function getMime(filePath) {
  const MIME = {
    ".html": "text/html",
    ".js": "application/javascript",
    ".css": "text/css",
    ".json": "application/json",
    ".png": "image/png",
    ".jpg": "image/jpeg",
    ".svg": "image/svg+xml",
    ".ico": "image/x-icon",
    ".woff": "font/woff",
    ".woff2": "font/woff2",
    ".ttf": "font/ttf",
    ".mp3": "audio/mpeg",
    ".mp4": "video/mp4",
    ".webm": "video/webm",
  };
  const ext = path.extname(filePath).toLowerCase();
  return MIME[ext] || "application/octet-stream";
}

app.whenReady().then(createWindow);
