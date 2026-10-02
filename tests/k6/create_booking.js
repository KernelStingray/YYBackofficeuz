import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 50 },
    { duration: '1m', target: 200 }, // Booking creates write load, keep max concurrent lower
    { duration: '2m', target: 200 },
    { duration: '30s', target: 0 },
  ],
};

const BASE_URL = __ENV.API_URL || 'http://localhost:8080/api/v1';
const RIDE_ID = __ENV.RIDE_ID || '00000000-0000-0000-0000-000000000000'; // Target ride
const TOKEN = __ENV.TOKEN || 'TEST_TOKEN';

export default function () {
  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${TOKEN}`,
    },
  };

  const payload = JSON.stringify({
    seats: 1,
    is_full_car_buyout: false
  });

  const res = http.post(`${BASE_URL}/rides/${RIDE_ID}/book`, payload, params);

  // Status 201 Created or 400 Bad Request (if seats run out during test)
  check(res, {
    'status is 201 or 400': (r) => r.status === 201 || r.status === 400,
    'response time < 1000ms': (r) => r.timings.duration < 1000,
  });

  sleep(2); // Wait a bit before next attempt
}
