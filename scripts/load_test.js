import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Trend, Rate } from 'k6/metrics';

// =============================================================================
// CUSTOM METRICS & TRENDS
// =============================================================================
const publicJobsDuration = new Trend('waiting_time_jobs_api');
const loginDuration = new Trend('waiting_time_login_api');
const profileMeDuration = new Trend('waiting_time_profile_api');
const seekerDashboardDuration = new Trend('waiting_time_seeker_dashboard_api');

const publicJobsSuccessRate = new Rate('success_rate_jobs_api');
const loginSuccessRate = new Rate('success_rate_login_api');
const profileMeSuccessRate = new Rate('success_rate_profile_api');
const seekerDashboardSuccessRate = new Rate('success_rate_seeker_dashboard_api');

// =============================================================================
// CONFIGURATION & STAGES
// =============================================================================
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const SEEKER_EMAIL = __ENV.SEEKER_EMAIL || 'you@example.com';
const SEEKER_PASSWORD = __ENV.SEEKER_PASSWORD || 'password123';

export const options = {
  stages: [
    { duration: '30s', target: 500 },   // ৩০ সেকেন্ডে ৫০০ VU
    { duration: '1m',  target: 2000 },  // ১ মিনিটে ২,০০০ VU
    { duration: '1m',  target: 5000 },  // ১ মিনিটে ৫,০০০ VU
    { duration: '1m',  target: 10000 }, // ১ মিনিটে ১০,০০০ VU (Peak Spike)
    { duration: '30s', target: 0 },     // ৩০ সেকেন্ডে ০ তে নেমে আসবে
  ],
  thresholds: {
    http_req_failed: ['rate<0.05'],     // ৯৫% রিকোয়েস্ট পাস হতে হবে
    http_req_duration: ['p(95)<3000'],  // p95 রেসপন্স টাইম ৩ সেকেন্ডের নিচে
  },
};

function thinkTime() {
  sleep(Math.random() * 1 + 1); // 1 to 2 seconds pause
}

// =============================================================================
// LOAD TEST SCENARIO
// =============================================================================
export default function () {
  let accessToken = '';

  // --- Step 1: Public Jobs Feed ---
  group('1. Public Jobs Feed', function () {
    const url = `${BASE_URL}/api/v1/jobs?page=1&limit=10&sort=newest`;
    const res = http.get(url, {
      headers: {
        'Accept': 'application/json',
      },
    });

    publicJobsDuration.add(res.timings.waiting);

    const isOk = check(res, {
      'status is 200': (r) => r.status === 200,
      'response has success flag': (r) => {
        try {
          return JSON.parse(r.body).success === true;
        } catch (e) {
          return false;
        }
      },
      'response data has items array': (r) => {
        try {
          const body = JSON.parse(r.body);
          // Fixed: List envelope uses data.items
          return body.data && Array.isArray(body.data.items);
        } catch (e) {
          return false;
        }
      },
    });

    publicJobsSuccessRate.add(isOk);
  });

  thinkTime();

  // --- Step 2: Authentication Login ---
  group('2. Authentication Login', function () {
    const url = `${BASE_URL}/api/v1/auth/login`;
    const payload = JSON.stringify({
      email: SEEKER_EMAIL,
      password: SEEKER_PASSWORD,
    });
    const params = {
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
    };

    const res = http.post(url, payload, params);

    loginDuration.add(res.timings.waiting);

    const isOk = check(res, {
      'status is 200': (r) => r.status === 200,
      'login response success flag is true': (r) => {
        try {
          return JSON.parse(r.body).success === true;
        } catch (e) {
          return false;
        }
      },
      'accessToken exists in envelope data': (r) => {
        try {
          const body = JSON.parse(r.body);
          return (
            body.data &&
            typeof body.data.accessToken === 'string' &&
            body.data.accessToken.length > 0
          );
        } catch (e) {
          return false;
        }
      },
    });

    loginSuccessRate.add(isOk);

    if (isOk) {
      try {
        const body = JSON.parse(res.body);
        accessToken = body.data.accessToken;
      } catch (e) {
        // Extraction failed
      }
    }
  });

  if (!accessToken) {
    return; // Skip authenticated calls if login failed
  }

  thinkTime();

  // --- Step 3: Authenticated Profile Endpoint ---
  group('3. Get Seeker Profile', function () {
    const url = `${BASE_URL}/api/v1/profile/me`;
    const params = {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Accept': 'application/json',
      },
    };

    const res = http.get(url, params);

    profileMeDuration.add(res.timings.waiting);

    const isOk = check(res, {
      'status is 200': (r) => r.status === 200,
      'profile response success is true': (r) => {
        try {
          return JSON.parse(r.body).success === true;
        } catch (e) {
          return false;
        }
      },
      'profile data object exists': (r) => {
        try {
          const body = JSON.parse(r.body);
          // Fixed: Checks if data object exists (handles both flat & nested email)
          return body.data && typeof body.data === 'object';
        } catch (e) {
          return false;
        }
      },
    });

    profileMeSuccessRate.add(isOk);
  });

  thinkTime();

  // --- Step 4: Authenticated Seeker Dashboard ---
  group('4. Get Seeker Dashboard', function () {
    const url = `${BASE_URL}/api/v1/dashboard/seeker`;
    const params = {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Accept': 'application/json',
      },
    };

    const res = http.get(url, params);

    seekerDashboardDuration.add(res.timings.waiting);

    const isOk = check(res, {
      'status is 200': (r) => r.status === 200,
      'dashboard response success is true': (r) => {
        try {
          return JSON.parse(r.body).success === true;
        } catch (e) {
          return false;
        }
      },
      'dashboard has valid data object': (r) => {
        try {
          const body = JSON.parse(r.body);
          return body.data !== null && typeof body.data === 'object';
        } catch (e) {
          return false;
        }
      },
    });

    seekerDashboardSuccessRate.add(isOk);
  });

  thinkTime();
}