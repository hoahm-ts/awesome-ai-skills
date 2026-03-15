# RESTful API Conventions

> **When writing or reviewing any HTTP API code, read this file first.**
>
> These guidelines apply to all services that expose HTTP APIs.
> Follow them alongside `golang-conventions.md` and `logging-conventions.md`.

---

## Table of Contents

- [Standard Response Envelope](#standard-response-envelope)
- [Predefined Verdict Codes](#predefined-verdict-codes)
- [HTTP Headers](#http-headers)
- [HTTP Methods](#http-methods)
- [Request Handling](#request-handling)
- [Naming Conventions](#naming-conventions)
- [Security](#security)
- [Performance](#performance)
- [Observability](#observability)
- [Documentation](#documentation)

---

## Standard Response Envelope

Every API response — success or error — must be wrapped in a consistent JSON envelope:

```go
type Response struct {
    Verdict string `json:"verdict"`
    Message string `json:"message"`
    Data    any    `json:"data"`
    Time    string `json:"time"`
}
```

Example success response:

```json
{
  "verdict": "success",
  "time": "2026-03-10T09:00:00+07:00",
  "message": "",
  "data": {}
}
```

Example error response:

```json
{
  "verdict": "invalid_parameters",
  "time": "2026-03-10T09:00:00+07:00",
  "message": "field 'email' is required",
  "data": null
}
```

- `verdict` — machine-readable outcome code (see table below).
- `message` — human-readable description; empty string on success. Never include stack traces or internal details.
- `data` — response payload; `null` on error.
- `time` — response timestamp formatted as RFC 3339 with timezone offset (e.g. `time.Now().Format(util.DateTimeLayout)`).

---

## Predefined Verdict Codes

| HTTP Status | Verdict |
|---|---|
| 200 | `success` |
| 400 | `invalid_parameters` \| `missing_parameters` \| `forbidden_parameters` \| `invalid_msisdn` \| `malformed_json` |
| 401 | `missing_authorization` \| `unknown_authorization` \| `invalid_credential` \| `invalid_token` \| `expired_token` |
| 403 | `permission_denied` |
| 404 | `invalid_route` \| `record_not_found` \| `unknown_client` |
| 405 | `invalid_method` |
| 409 | `invalid_state` |
| 429 | `limit_exceeded` |
| 500 | `failure` |

Map every error path in your handler to exactly one verdict. Never invent new verdict strings without updating this table.

---

## HTTP Headers

Standard headers required on every request/response:

| Header | Usage |
|---|---|
| `Content-Type: application/json` | All requests with a body and all responses |
| `Authorization` | Authentication token (bearer or custom scheme) |
| `X-Request-ID` | Request tracking and log correlation |

Predefined custom headers (reserved for mobile clients):

| Header | Description |
|---|---|
| `X-APP-VERSION` | Client app version |
| `X-TIME-ZONE-OFFSET` | Client time zone offset |
| `X-TIME-ZONE` | Client time zone name |
| `X-DEVICE-ID` | Device fingerprint / unique ID |
| `X-DEVICE-MODEL` | Device model identifier |
| `X-PLATFORM` | Client platform (`android`, `ios`, `web`) |
| `X-OS-VERSION` | Mobile OS version |
| `X-APP-BUILD-NUMBER` | Mobile app build number |
| `X-OS-BUILD-NUMBER` | Mobile OS build number |
| `X-LANGUAGE` | Client language/locale |

Always read `X-Request-ID` at the handler boundary, store it in the request context, and include it in all log entries and outbound calls.

---

## HTTP Methods

Use **GET**, **POST**, and **PATCH** only. Do not use `DELETE` (use soft-delete via PATCH instead) and do not use `PUT` (PATCH is preferred for all mutations: it enables partial updates, aligns naturally with soft-delete patterns, and removes the risk of a client accidentally overwriting fields it did not intend to change).

| | GET | POST | PATCH |
|---|---|---|---|
| Use case | Retrieve data | Create new resources | Partial update |
| Request body | No | Yes | Yes |
| Success HTTP status | 200 | 200 | 200 |
| Error HTTP statuses | 404 | 400, 404, 409 | 400, 404 |

---

## Request Handling

- Use `snake_case` for all request parameters and body fields.
- Validate and sanitize all incoming data before passing it to the domain layer.
- Do not mix query-string parameters (GET) and body parameters (POST/PATCH) in the same endpoint.
- Return `malformed_json` (400) when the request body cannot be decoded.
- Return `missing_parameters` (400) when a required field is absent; `invalid_parameters` (400) when a field fails validation; `forbidden_parameters` (400) when a field is present but not allowed for the caller.

---

## Naming Conventions

Use `snake_case` for all JSON request and response fields.

Apply the following type-specific rules:

| Data type | Rule | Example |
|---|---|---|
| Boolean | Prefix with `is_`, `has_`, or `can_`. Never use a negative form. | `is_active: true`, `has_access: true` |
| Numeric ID | Suffix with `_id` | `user_id: 1` |
| Count / total | Suffix with `_count` or `_total` | `item_count: 5` |
| Financial amount | Suffix with `_amount` | `transaction_amount: 10` |
| Rate / percentage | Suffix with `_rate` or `_percentage` | `success_rate: 0.98` |
| Level | Suffix with `_level` | `risk_level: 2` |
| Currency-specific | Suffix with the ISO 4217 code in lowercase | `price_vnd: 10000` |
| Array (general) | Use plural names; suffix with `_items`, `_ids`, or `_history` as appropriate | `user_ids: [1, 2, 3]`, `activities_history: []` |

Additional response rules:

- Return only the fields the caller needs — no auto-increment IDs, `created_at`, or full objects unless explicitly required.
- Error messages must be generic — no stack traces, no internal field names, no database error strings.

---

## Security

- **Authentication required by default.** Unauthenticated endpoints must be explicitly opted in via a decorator, middleware bypass list, or equivalent mechanism — not by omitting auth middleware.
- **PII encrypted at rest** using AES-CTR. Never return PII in any API response (authenticated or not).
- Apply rate limiting at two layers: network layer (WAF/API Gateway) and application layer (user-level, endpoint-specific). Use `limit_exceeded` (429) when a limit is breached.
- Always use HTTPS for all public-facing APIs.
- Protect critical endpoints (login, OTP, password reset) with WAF rules, CAPTCHA, and device fingerprinting.

---

## Performance

- Implement pagination, filtering, and sorting for any endpoint that returns a list. Use keyset (cursor) pagination for large datasets.
- Cache frequently-read, rarely-changing data in Redis with an appropriate TTL.
- Use asynchronous processing (background workers, queues) for long-running tasks; return an acknowledgement immediately and allow the client to poll or receive a webhook.
- Support batch operations where applicable to reduce round-trips.
- Enable gzip compression (`Content-Encoding: gzip`) on the API gateway or application layer to reduce payload size.

---

## Observability

- Log every inbound request and response using the structured logging patterns described in `logging-conventions.md` (`marker: "[api]"`).
- Include `X-Request-ID` in every log entry for correlation.
- Mask / never log raw PII data in request or response logs.
- Emit metrics for request count, error rate, and latency per endpoint.

---

## Documentation

| API type | Documentation standard |
|---|---|
| User-facing API | Use Swagger (OpenAPI). |
| Partner API | Write a versioned API spec. Must be reviewed and approved by the tech lead before sharing. Maintain a separate spec per partner. |
