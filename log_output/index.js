const http = require('http');
const fs = require('fs');
const path = require('path');

const port = process.env.PORT || 3000;
const logFile = path.join('/usr/src/app/shared', 'log.txt');

const server = http.createServer((req, res) => {
  if (req.url === '/') {
    if (!fs.existsSync(logFile)) {
      res.writeHead(200, { 'Content-Type': 'text/plain' });
      res.end('Log file not ready yet — writer is still starting up.\n');
      return;
    }
    const content = fs.readFileSync(logFile, 'utf8');
    res.writeHead(200, { 'Content-Type': 'text/plain' });
    res.end(content);
  } else {
    res.writeHead(404);
    res.end();
  }
});

server.listen(port, () => {
  console.log(`Reader server started on port ${port}`);
});