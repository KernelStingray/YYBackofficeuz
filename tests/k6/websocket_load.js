import ws from 'k6/ws';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 5000 }, // Ramp up to 5000 WebSocket connections
    { duration: '3m', target: 5000 }, // Hold 5000 connections
    { duration: '1m', target: 0 },    // Ramp down
  ],
};

const WS_URL = __ENV.WS_URL || 'ws://localhost:8080/ws';
const TOKEN = __ENV.TOKEN || 'TEST_TOKEN';

export default function () {
  const url = `${WS_URL}?token=${TOKEN}`;
  
  const res = ws.connect(url, {}, function (socket) {
    socket.on('open', () => {
      // Send a simple ping message depending on backend WS protocol
      socket.send(JSON.stringify({ type: 'ping' }));
    });

    socket.on('message', (msg) => {
      // Optional: Verify message structure
      const data = JSON.parse(msg);
      check(data, {
        'message received': (d) => d !== null,
      });
    });

    socket.on('close', () => {
      // Socket closed
    });

    // Keep connection alive by sending ping every 30 seconds
    socket.setInterval(function () {
      socket.send(JSON.stringify({ type: 'ping' }));
    }, 30000); 

    // We let the connection stay open for 30s to 3 minutes depending on test duration
  });

  check(res, { 'status is 101': (r) => r && r.status === 101 });
  
  sleep(1);
}
