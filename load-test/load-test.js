import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://auto-scaler.local';

export const options = {
  stages: [
    { duration: '30s', target: 10 },  // ramp up
    { duration: '2m',  target: 10 },  // hold
    { duration: '30s', target: 0  },  // ramp down
  ],
  thresholds: {
    http_req_failed: ['rate<0.01'],
  },
};

export default function () {
  const payload = JSON.stringify({ amount: 4, duration: 30 });
  const params  = { headers: { 'Content-Type': 'application/json' } };

  const res = http.post(`${BASE_URL}/load`, payload, params);

  check(res, {
    'status is 200': (r) => r.status === 200,
  });
}
