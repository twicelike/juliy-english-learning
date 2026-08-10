import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '10s', target: 50 },  // Ramp up to 50 users over 10s
    { duration: '30s', target: 50 },  // Stay at 50 users for 30s
    { duration: '10s', target: 0 },   // Ramp down to 0 users over 10s
  ],
  thresholds: {
    http_req_duration: ['p(95)<200'], // 95% of requests must complete within 200ms
    http_req_failed: ['rate<0.01'],   // Error rate must be less than 1%
  },
};

export default function () {
  // Test dictionary detail API (which is cached)
  const res = http.get('http://host.docker.internal:8080/api/v1/words/detail/hello');
  
  check(res, {
    'is status 200': (r) => r.status === 200,
    'has word': (r) => r.json().word !== undefined,
  });

  sleep(1);
}
