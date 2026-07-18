import http from "k6/http";
import { check, sleep } from "k6";
import { Trend } from "k6/metrics";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";

const productDuration = new Trend("product_duration", true);

export const options = {
  stages: [
    { duration: "10s", target: 50 }, // ramp up to 50 VU in 10 seconds
    { duration: "50s", target: 100 }, // ramp up/hold to 100 VU for 50 seconds
    { duration: "10s", target: 0 }, // ramp down to 0 VU in 10 seconds
  ],
  thresholds: {
    // p(95) response time must be <= 200ms
    http_req_duration: ["p(95)<=200"],
    // ensure that there are no failed requests
    http_req_failed: ["rate<0.01"],
  },
};

export default function () {
  const res = http.get(`${BASE_URL}/v1/product`);

  productDuration.add(res.timings.duration);

  check(res, {
    "status is 200": (r) => r.status === 200,
    "response time <= 200ms": (r) => r.timings.duration <= 200,
    "response body is not empty": (r) => r.body.length > 0,
    "content type is JSON": (r) =>
      r.headers["Content-Type"]?.includes("application/json"),
  });

  // sleep(1);
}
