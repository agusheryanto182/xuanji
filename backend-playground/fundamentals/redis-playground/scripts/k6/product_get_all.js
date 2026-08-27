import http from "k6/http";
import { check } from "k6";

export const options = {
  stages: [
    { duration: "10s", target: 50 },
    { duration: "50s", target: 100 },
    { duration: "10s", target: 0 },
  ],

  thresholds: {
    http_req_duration: ["p(95)<100"],
  },
};

const BASE_URL = "http://localhost:8080";

const TOKEN = __ENV.BEARER_TOKEN;

export default function () {
  const response = http.get(`${BASE_URL}/v1/product`, {
    headers: {
      Authorization: `Bearer ${TOKEN}`,
      Accept: "application/json",
    },
  });

  check(response, {
    "status is 200": (r) => r.status === 200,
    "response time < 100ms": (r) => r.timings.duration < 100,
  });
}
