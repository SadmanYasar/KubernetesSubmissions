const http = require('http');
const fs = require('fs');
const path = require('path');

const port = process.env.PORT || 3000;
const pingPongUrl = process.env.PING_PONG_URL || 'http://ping-pong-svc:80/pings';
const logFile = path.join('/usr/src/app/shared', 'log.txt');

async function getPongs() {
  try {
    const res = await fetch(pingPongUrl);
    if (!res.ok) {
      console.error(`Ping-pong service responded with status: ${res.status}`);
      return 0;
    }
    const text = await res.text();
    const match = text.match(/\d+/);
    return match ? parseInt(match[0], 10) : 0;
  } catch (err) {
    console.error(`Error connecting to ping-pong service at ${pingPongUrl}:`, err.message);
    return 0;
  }
}

const server = http.createServer(async (req, res) => {
  if (req.url === '/') {
    if (!fs.existsSync(logFile)) {
      res.writeHead(200, { 'Content-Type': 'text/plain' });
      res.end('Log file not ready yet — writer is still starting up.\n');
      return;
    }
    const content = fs.readFileSync(logFile, 'utf8');
    const lines = content.trim().split('\n');
    const lastLine = lines[lines.length - 1];

    const pongs = await getPongs();

    // Ensure the lastLine ends with a period if it doesn't already
    const formattedLastLine = lastLine.endsWith('.') ? lastLine : `${lastLine}.`;

    res.writeHead(200, { 'Content-Type': 'text/plain' });
    res.end(`${formattedLastLine}\nPing / Pongs: ${pongs}\n`);
  } else {
    res.writeHead(404);
    res.end();
  }
});

server.listen(port, () => {
  console.log(`Reader server started on port ${port}`);
});