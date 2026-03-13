const { randomUUID } = require('crypto');
const fs = require('fs');
const path = require('path');

// Generate a random string (UUID) once on startup
const randomString = randomUUID();
const logFile = path.join('/usr/src/app/shared', 'log.txt');

console.log(`Writer started. Random string: ${randomString}`);

// Ensure the shared directory exists
fs.mkdirSync(path.dirname(logFile), { recursive: true });

const writeLog = () => {
  const line = `${new Date().toISOString()}: ${randomString}\n`;
  fs.appendFileSync(logFile, line);
  process.stdout.write(line);
};

// Write immediately on startup, then every 5 seconds
writeLog();
setInterval(writeLog, 5000);
