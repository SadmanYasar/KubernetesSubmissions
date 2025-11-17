const { randomUUID } = require('crypto');

// Generate a random string (UUID) on startup and store it in memory
const randomString = randomUUID();

const printWithTimestamp = () => {
  const timestamp = new Date().toISOString();
  console.log(`${timestamp}: ${randomString}`);
}

// Output immediately and then every 5 seconds
printWithTimestamp();
setInterval(printWithTimestamp, 5000);