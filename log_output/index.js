const { randomUUID } = require('crypto');
const http = require('http');

// Generate a random string (UUID) on startup and store it in memory
const randomString = randomUUID();
const port = process.env.PORT || 3000;

const printWithTimestamp = () => {
  const timestamp = new Date().toISOString();
  console.log(`${timestamp}: ${randomString}`);
  return `${timestamp}: ${randomString}`;
}

// Output immediately and then every 5 seconds
printWithTimestamp();
setInterval(printWithTimestamp, 5000);

const server = http.createServer((req, res) => {
  if (req.url === '/') {
    const status = printWithTimestamp();
    res.writeHead(200, { 'Content-Type': 'text/plain' });
    res.end(`${status}`);
  } else {
    res.writeHead(404);
    res.end();
  }
});

server.listen(port, () => {
  console.log(`Server started in port ${port}`);
});