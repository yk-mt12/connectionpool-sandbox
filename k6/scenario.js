import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  scenarios: {
    with_pool: {
      executor: 'constant-vus',
      vus: 20,
      duration: '60s',
      exec: 'withPool',
      tags: { scenario: 'with_pool' },
    },
    without_pool: {
      executor: 'constant-vus',
      vus: 20,
      duration: '60s',
      exec: 'withoutPool',
      startTime: '70s',
      tags: { scenario: 'without_pool' },
    },
  },
  thresholds: {
    'http_req_duration{scenario:with_pool}': ['p(95)<500'],
    'http_req_duration{scenario:without_pool}': ['p(95)<5000'],
  },
};

export function withPool() {
  const res = http.get(`${BASE_URL}/with-pool`, {
    tags: { name: 'with-pool' },
  });
  check(res, { 'status 200': (r) => r.status === 200 });
  sleep(0.05);
}

export function withoutPool() {
  const res = http.get(`${BASE_URL}/without-pool`, {
    tags: { name: 'without-pool' },
  });
  check(res, { 'status 200': (r) => r.status === 200 });
  sleep(0.05);
}
