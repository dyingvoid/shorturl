import http from 'k6/http';
import { check, sleep } from 'k6';

http.setResponseCallback(http.expectedStatuses(200, 201, 302));

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8082';
const PASSWORD = 'Password1';
const URLS_PER_USER = 10;

export const options = {
  vus: 100,
  duration: '30s',
  thresholds: {
    checks: ['rate>0.99'],
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(50)<300', 'p(95)<1000', 'p(99)<2000'],
    'http_req_duration{name:register}': ['p(50)<300', 'p(95)<1000', 'p(99)<2000'],
    'http_req_duration{name:login}': ['p(50)<300', 'p(95)<1000', 'p(99)<2000'],
    'http_req_duration{name:create_url}': ['p(50)<300', 'p(95)<1000', 'p(99)<2000'],
    'http_req_duration{name:redirect}': ['p(50)<300', 'p(95)<1000', 'p(99)<2000'],
  },
};

function jsonHeaders(token) {
  const headers = { 'Content-Type': 'application/json' };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  return headers;
}

function uniqueEmail() {
  return `k6-${__VU}-${__ITER}-${Date.now()}@example.com`;
}

export default function () {
  const email = uniqueEmail();

  const registerRes = http.post(
    `${BASE_URL}/api/v1/auth/register`,
    JSON.stringify({ email, password: PASSWORD }),
    { headers: jsonHeaders(), tags: { name: 'register' } },
  );
  check(registerRes, {
    'register status is 201': (r) => r.status === 201,
  });

  const loginRes = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({ email, password: PASSWORD }),
    { headers: jsonHeaders(), tags: { name: 'login' } },
  );
  const loggedIn = check(loginRes, {
    'login status is 200': (r) => r.status === 200,
    'login returns access token': (r) => Boolean(r.json('access_token')),
  });

  if (!loggedIn) {
    return;
  }

  const accessToken = loginRes.json('access_token');
  const shortCodes = [];

  for (let i = 0; i < URLS_PER_USER; i++) {
    const createRes = http.post(
      `${BASE_URL}/api/v1/urls`,
      JSON.stringify({
        original_url: `https://example.com/load/${__VU}/${__ITER}/${i}`,
      }),
      { headers: jsonHeaders(accessToken), tags: { name: 'create_url' } },
    );
    check(createRes, {
      'create url status is 201': (r) => r.status === 201,
      'create url returns short code': (r) => Boolean(r.json('short_code')),
    });

    const shortCode = createRes.json('short_code');
    if (shortCode) {
      shortCodes.push(shortCode);
    }
  }

  for (const shortCode of shortCodes) {
    const redirectRes = http.get(`${BASE_URL}/${shortCode}`, {
      redirects: 0,
      tags: { name: 'redirect' },
    });
    check(redirectRes, {
      'redirect status is 302': (r) => r.status === 302,
      'redirect has location header': (r) => Boolean(r.headers.Location),
    });
  }

  sleep(1);
}
