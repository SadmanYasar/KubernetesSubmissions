const http = require('http');
const fs = require('fs');
const path = require('path');

const port = process.env.PORT || 3000;
const logFile = path.join('/usr/src/app/shared', 'log.txt');
const pongsFile = path.join('/usr/src/app/shared', 'pongs.txt');

const server = http.createServer((req, res) => {
  if (req.url === '/') {
    if (!fs.existsSync(logFile)) {
      res.writeHead(200, { 'Content-Type': 'text/plain' });
      res.end('Log file not ready yet — writer is still starting up.\n');
      return;
    }
    const content = fs.readFileSync(logFile, 'utf8');
    const lines = content.trim().split('\n');
    const lastLine = lines[lines.length - 1];

    let pongs = 0;
    if (fs.existsSync(pongsFile)) {
      pongs = parseInt(fs.readFileSync(pongsFile, 'utf8').trim(), 10) || 0;
    }

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