import http from "k6/http";
import { check, sleep } from "k6";
import { Counter, Rate } from "k6/metrics";

// ====== CONFIG ======
const BASE_URL = __ENV.BASE_URL || "http://localhost:8080/v1";
const VU_COUNT = 100;
const DURATION = __ENV.DURATION || "1m";
const PASSWORD = "password123"; // matches the seeder

// ====== CUSTOM METRICS ======
const loginFailures = new Counter("login_failures");
const unexpectedCount = new Counter("unexpected_status");
// From 2 requests per batch (same token, sent concurrently), this rate is
// computed over total requests (not per batch), so the ideal target is 0.5
const successRate = new Rate("profile_success_rate");
const rateLimitedRate = new Rate("profile_rate_limited_rate");

// ====== K6 OPTIONS ======
export const options = {
  scenarios: {
    profile_load: {
      executor: "constant-vus",
      vus: VU_COUNT,
      duration: DURATION,
      exec: "hitProfile",
    },
  },
  thresholds: {
    // Each VU sends 2 requests concurrently (1 token, limit of 1 req/sec) ->
    // we expect EXACTLY 1 out of 2 to succeed (200) and 1 to be rate limited
    // (429), i.e. a 50/50 split. Allow a ±5% tolerance for timing jitter
    // between VUs.
    profile_success_rate: ["rate>=0.45", "rate<=0.55"],
    profile_rate_limited_rate: ["rate>=0.45", "rate<=0.55"],
    login_failures: ["count==0"],
    unexpected_status: ["count==0"], // no status other than 200/429 allowed
  },
};

// ====== SETUP: log in once at the start for all VUs, each gets a unique token ======
export function setup() {
  const tokens = [];

  for (let i = 1; i <= VU_COUNT; i++) {
    const email = `user${i}@example.com`;
    const res = http.post(
      `${BASE_URL}/auth/login`,
      JSON.stringify({ email, password: PASSWORD }),
      { headers: { "Content-Type": "application/json" } },
    );

    const ok = check(res, {
      [`login vu${i} status 200`]: (r) => r.status === 200,
      [`login vu${i} has token`]: (r) => {
        try {
          return !!JSON.parse(r.body).data.token;
        } catch (e) {
          return false;
        }
      },
    });

    if (!ok) {
      console.error(
        `Login failed for ${email}: status=${res.status} body=${res.body}`,
      );
      tokens.push(null);
      continue;
    }

    const body = JSON.parse(res.body);
    tokens.push(body.data.token);
  }

  return { tokens };
}

// Safety margin above 1 second to absorb latency jitter & JS overhead.
// This keeps the gap BETWEEN batches (not between requests within a batch)
// consistent at ~1s, so the rate limiter window is always "fresh" per iteration.
const SAFETY_MARGIN_MS = 50; // 1.05s per iteration

// ====== MAIN SCENARIO: each VU sends 2 requests CONCURRENTLY (http.batch) ======
// using the same token. Since the limit is 1 req/sec per token, exactly
// 1 of the 2 requests should succeed (200) and the other should be rate
// limited (429).
export function hitProfile(data) {
  const iterationStart = Date.now(); // wall-clock time, not just HTTP duration

  // __VU starts at 1, matching the token index in the array
  const vuIndex = __VU - 1;
  const token = data.tokens[vuIndex];

  if (!token) {
    loginFailures.add(1);
    sleep(1);
    return;
  }

  const authHeaders = { headers: { Authorization: `Bearer ${token}` } };

  // Send 2 concurrent GET /user/profile requests using the same token
  const responses = http.batch([
    ["GET", `${BASE_URL}/user/profile`, null, authHeaders],
    ["GET", `${BASE_URL}/user/profile`, null, authHeaders],
  ]);

  responses.forEach((res, idx) => {
    const is200 = res.status === 200;
    const is429 = res.status === 429;

    check(res, {
      [`request ${idx + 1}: status 200 or 429`]: () => is200 || is429,
    });

    if (is200) {
      check(res, {
        [`request ${idx + 1}: response has profile data`]: (r) => {
          try {
            return !!JSON.parse(r.body).data.id;
          } catch (e) {
            return false;
          }
        },
      });
    } else if (!is429) {
      // Any status other than 200/429 is considered an anomaly
      // (outside the expected scenario).
      unexpectedCount.add(1);
      console.error(
        `VU ${__VU} got unexpected status: ${res.status} body=${res.body}`,
      );
    }

    // rate() is computed per request (not per batch), so the ideal is 50/50
    successRate.add(is200);
    rateLimitedRate.add(is429);
  });

  // Compute elapsed wall-clock time since the start of the iteration,
  // so the gap between batches stays consistently around 1.05s.
  const elapsedMs = Date.now() - iterationStart;
  const targetMs = 1000 + SAFETY_MARGIN_MS;
  const sleepMs = Math.max(targetMs - elapsedMs, 0);
  sleep(sleepMs / 1000);
}

// ====== TEARDOWN (optional, prints a short summary) ======
export function teardown(data) {
  console.log(
    `Test finished. Total tokens used: ${data.tokens.filter((t) => t).length}/${VU_COUNT}`,
  );
}
