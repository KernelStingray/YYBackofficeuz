import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 50 }, // Ramp up to 50 users
    { duration: '1m', target: 500 }, // Ramp up to 500 users
    { duration: '2m', target: 500 }, // Stay at 500 users
    { duration: '30s', target: 0 },  // Ramp down to 0
  ],
};

const BASE_URL = __ENV.API_URL || 'http://localhost:8080/api/v1';

export default function () {
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  // Пример параметров поиска из Ташкента в Самарканд
  const payload = JSON.stringify({
    origin_lat: 41.2995,
    origin_lon: 69.2401,
    dest_lat: 39.6542,
    dest_lon: 66.9597,
    departure_time: new Date(Date.now() + 86400000).toISOString(),
    seats_needed: 1
  });

  const res = http.post(`${BASE_URL}/rides/search`, payload, params);

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(1);
}
