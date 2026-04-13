import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const BASE_URL = __ENV.TARGET_URL || 'http://localhost:8080';

const writeLatency = new Trend('write_latency', true);
const readLatency = new Trend('read_latency', true);
const writeErrors = new Rate('write_errors');
const readErrors = new Rate('read_errors');

export const options = {
  scenarios: {
    seed: {
      executor: 'shared-iterations',
      vus: 10,
      iterations: 500,
      exec: 'seedUrls',
      maxDuration: '30s',
    },
    write: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 50 },
        { duration: '1m', target: 100 },
        { duration: '1m', target: 100 },
        { duration: '30s', target: 0 },
      ],
      exec: 'writeTest',
      startTime: '35s',
    },
    read: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 500 },
        { duration: '1m', target: 1000 },
        { duration: '1m', target: 1000 },
        { duration: '30s', target: 0 },
      ],
      exec: 'readTest',
      startTime: '35s',
    },
  },
  thresholds: {
    'write_latency': ['p(95)<500'],
    'read_latency': ['p(95)<200'],
    'write_errors': ['rate<0.01'],
    'read_errors': ['rate<0.05'],
  },
};

const seededCodes = [];

export function setup() {
  const codes = [];
  for (let i = 0; i < 200; i++) {
    const res = http.post(`${BASE_URL}/shorten`, JSON.stringify({
      url: `https://example.com/seed-${i}-${Date.now()}`,
    }), { headers: { 'Content-Type': 'application/json' } });

    if (res.status === 201) {
      try {
        const body = JSON.parse(res.body);
        codes.push(body.short_url.split('/').pop());
      } catch (e) {}
    }
  }
  console.log(`Seeded ${codes.length} URLs for read tests`);
  return { codes };
}

export function seedUrls() {
  const res = http.post(`${BASE_URL}/shorten`, JSON.stringify({
    url: `https://example.com/extra-${__VU}-${__ITER}-${Date.now()}`,
  }), { headers: { 'Content-Type': 'application/json' } });
}

export function writeTest() {
  const res = http.post(`${BASE_URL}/shorten`, JSON.stringify({
    url: `https://example.com/${__VU}-${__ITER}-${Date.now()}`,
  }), { headers: { 'Content-Type': 'application/json' } });

  writeLatency.add(res.timings.duration);
  const success = check(res, { 'write 201': (r) => r.status === 201 });
  writeErrors.add(!success);
}

export function readTest(data) {
  if (!data.codes || data.codes.length === 0) {
    sleep(1);
    return;
  }

  const code = data.codes[Math.floor(Math.random() * data.codes.length)];
  const res = http.get(`${BASE_URL}/${code}`, { redirects: 0 });

  readLatency.add(res.timings.duration);
  const success = check(res, { 'read 302': (r) => r.status === 302 });
  readErrors.add(!success);
}
