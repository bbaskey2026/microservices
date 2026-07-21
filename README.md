Platform API Documentation
Version: 1.0
Base URL (Gateway): `http://<host>:8080`
Total Endpoints: 62 across 8 services
---
Table of Contents
Overview & Architecture
Authentication
Conventions
API Gateway
Auth Service
User Service
Card Service
Wallet Service
Payment Service
Notification Service
Reports Service
Error Handling
---
Overview & Architecture
This platform is composed of an API Gateway and 7 downstream microservices, each responsible for a distinct domain. All external traffic enters through the Gateway (`:8080`), which proxies requests to the appropriate service based on the URL path prefix.
```
Client
  │
  ▼
API Gateway (:8080)
  │
  ├──► Auth Service          (:3001)  — registration, login, tokens
  ├──► User Service          (:3002)  — profiles, KYC, account status
  ├──► Card Service          (:3003)  — virtual/physical card lifecycle
  ├──► Wallet Service        (:3004)  — balances, transfers, funding
  ├──► Payment Service       (:3005)  — bills, airtime, data bundles
  ├──► Notification Service  (:3006)  — in-app/push notifications
  └──► Reports Service       (:3007)  — analytics & summaries
```
Each service independently exposes `/health`, `/healthz`, and `/ready` for infrastructure-level monitoring, plus a `/api/v1/<service>/health` route reachable through the Gateway.
---
Authentication
Endpoints marked ✅ in the Auth column require a valid JWT Bearer token.
```
Authorization: Bearer <access_token>
```
Obtain a token via `POST /api/v1/auth/login` or `POST /api/v1/auth/register`.
Tokens expire; use `POST /api/v1/auth/refresh-token` to obtain a new access token without re-authenticating.
Unauthenticated requests to protected routes return `401 Unauthorized`.
---
Conventions
Convention	Detail
Content type	`application/json` for all request/response bodies (unless noted)
ID format	Path parameters like `:id` and `:ref` are strings (UUID or reference codes)
Pagination	List endpoints accept optional `?page=` and `?limit=` query params
Timestamps	ISO 8601 UTC, e.g. `2026-07-21T10:00:00Z`
Currency amounts	Numeric, smallest supported precision (e.g. `1500.00`)
---
1. API Gateway `:8080`
The Gateway performs routing and health aggregation only; it does not implement business logic.
`GET /health`
Gateway liveness check. No auth required.
Response `200`:
```json
{ "status": "ok" }
```
`GET /healthz`
Kubernetes-style alias for `/health`. No auth required.
`GET /health/services`
Aggregates and returns the health status of all downstream services. No auth required.
Response `200`:
```json
{
  "gateway": "ok",
  "services": {
    "auth": "ok",
    "user": "ok",
    "card": "ok",
    "wallet": "ok",
    "payment": "ok",
    "notification": "ok",
    "reports": "ok"
  }
}
```
Proxy Routes
All of the following forward any HTTP method and path suffix to the corresponding backend service, preserving headers and body:
Route Prefix	Forwards To
`ALL /api/v1/auth/*`	Auth Service
`ALL /api/v1/users/*`	User Service
`ALL /api/v1/cards/*`	Card Service
`ALL /api/v1/wallets/*`	Wallet Service
`ALL /api/v1/payments/*`	Payment Service
`ALL /api/v1/notifications/*`	Notification Service
`ALL /api/v1/reports/*`	Reports Service
---
2. Auth Service `:3001`
Handles registration, login, session/token lifecycle, and password management.
`GET /health` · `GET /healthz` · `GET /ready` · `GET /api/v1/auth/health`
Liveness/readiness checks. No auth required. `/ready` additionally verifies DB and Redis connectivity.
`POST /api/v1/auth/register`
Register a new user. No auth required.
Request:
```json
{
  "email": "jane.doe@example.com",
  "password": "StrongPassword123!",
  "firstName": "Jane",
  "lastName": "Doe",
  "phone": "+2348012345678"
}
```
Response `201`:
```json
{
  "user": { "id": "usr_123", "email": "jane.doe@example.com" },
  "accessToken": "<jwt>",
  "refreshToken": "<jwt>"
}
```
`POST /api/v1/auth/login`
Authenticate an existing user. No auth required.
Request:
```json
{ "email": "jane.doe@example.com", "password": "StrongPassword123!" }
```
Response `200`:
```json
{ "accessToken": "<jwt>", "refreshToken": "<jwt>" }
```
`POST /api/v1/auth/refresh-token`
Exchange a valid refresh token for a new access token. No auth required (uses refresh token in body).
Request:
```json
{ "refreshToken": "<jwt>" }
```
Response `200`:
```json
{ "accessToken": "<jwt>" }
```
`POST /api/v1/auth/logout` ✅
Invalidate the current session/refresh token.
Response `200`: `{ "message": "Logged out successfully" }`
`PUT /api/v1/auth/change-password` ✅
Change the authenticated user's password.
Request:
```json
{ "currentPassword": "OldPass123!", "newPassword": "NewPass456!" }
```
Response `200`: `{ "message": "Password updated successfully" }`
---
3. User Service `:3002`
Manages user profiles, account status, and KYC verification.
`GET /health` · `GET /healthz` · `GET /ready` · `GET /api/v1/users/health`
Liveness/readiness checks. `/ready` verifies DB connectivity.
`GET /api/v1/users/profile` ✅
Get the authenticated user's own profile.
Response `200`:
```json
{ "id": "usr_123", "email": "jane.doe@example.com", "firstName": "Jane", "kycStatus": "verified" }
```
`PUT /api/v1/users/profile` ✅
Update the authenticated user's own profile.
Request:
```json
{ "firstName": "Jane", "lastName": "Smith", "phone": "+2348012345678" }
```
`GET /api/v1/users/` ✅
List all users (admin scope).
Query params: `?page=1&limit=20`
`GET /api/v1/users/:id` ✅
Get a specific user by ID.
`PUT /api/v1/users/:id/status` ✅
Update a user's account status (e.g. active, suspended).
Request:
```json
{ "status": "suspended", "reason": "Suspicious activity" }
```
`PUT /api/v1/users/:id/kyc` ✅
Update a user's KYC verification status.
Request:
```json
{ "kycStatus": "verified" }
```
`DELETE /api/v1/users/:id` ✅
Delete a user account.
Response `200`: `{ "message": "User deleted successfully" }`
---
4. Card Service `:3003`
Manages virtual/physical card issuance and lifecycle.
`GET /health` · `GET /healthz` · `GET /ready` · `GET /api/v1/cards/health`
Liveness/readiness checks. `/ready` verifies DB connectivity.
`POST /api/v1/cards/` ✅
Create a new card for the authenticated user.
Request:
```json
{ "type": "virtual", "currency": "NGN" }
```
Response `201`:
```json
{ "id": "card_123", "last4": "4242", "status": "active", "type": "virtual" }
```
`GET /api/v1/cards/` ✅
List all cards belonging to the authenticated user.
`GET /api/v1/cards/:id` ✅
Get details of a specific card.
`PUT /api/v1/cards/:id` ✅
Update card metadata (e.g. nickname, spending limits).
Request:
```json
{ "label": "Travel Card", "monthlyLimit": 50000 }
```
`PUT /api/v1/cards/:id/pin` ✅
Change the card's PIN.
Request:
```json
{ "currentPin": "1234", "newPin": "5678" }
```
`PUT /api/v1/cards/:id/block` ✅
Temporarily block a card (e.g. lost/stolen).
Response `200`: `{ "id": "card_123", "status": "blocked" }`
`PUT /api/v1/cards/:id/unblock` ✅
Reactivate a previously blocked card.
Response `200`: `{ "id": "card_123", "status": "active" }`
`DELETE /api/v1/cards/:id` ✅
Permanently delete/cancel a card.
---
5. Wallet Service `:3004`
Manages balances, funding, transfers, and withdrawals.
`GET /health` · `GET /healthz` · `GET /ready` · `GET /api/v1/wallets/health`
Liveness/readiness checks. `/ready` verifies DB connectivity.
`POST /api/v1/wallets/` ✅
Create a wallet for the authenticated user.
Request:
```json
{ "currency": "NGN" }
```
`GET /api/v1/wallets/` ✅
Get the authenticated user's wallet details.
`GET /api/v1/wallets/balance` ✅
Get current wallet balance.
Response `200`:
```json
{ "balance": 125000.00, "currency": "NGN" }
```
`POST /api/v1/wallets/fund` ✅
Fund the wallet (e.g. via card or bank transfer).
Request:
```json
{ "amount": 10000.00, "source": "card", "cardId": "card_123" }
```
`POST /api/v1/wallets/transfer` ✅
Transfer funds to another wallet/user.
Request:
```json
{ "recipientId": "usr_456", "amount": 5000.00, "note": "Rent share" }
```
`POST /api/v1/wallets/withdraw` ✅
Withdraw funds to a linked bank account.
Request:
```json
{ "amount": 20000.00, "bankAccountId": "bank_789" }
```
`GET /api/v1/wallets/transactions` ✅
List wallet transaction history.
Query params: `?page=1&limit=20&from=2026-06-01&to=2026-07-21`
---
6. Payment Service `:3005`
Handles bill payments, airtime, and data bundle purchases.
`GET /health` · `GET /healthz` · `GET /ready` · `GET /api/v1/payments/health`
Liveness/readiness checks. `/ready` verifies DB connectivity.
`POST /api/v1/payments/bill` ✅
Pay a utility/service bill.
Request:
```json
{ "billerCode": "DSTV", "customerRef": "1234567890", "amount": 15000.00 }
```
`POST /api/v1/payments/airtime` ✅
Purchase mobile airtime.
Request:
```json
{ "network": "MTN", "phone": "+2348012345678", "amount": 1000.00 }
```
`POST /api/v1/payments/data` ✅
Purchase a mobile data bundle.
Request:
```json
{ "network": "MTN", "phone": "+2348012345678", "planCode": "1GB-30D" }
```
`GET /api/v1/payments/history` ✅
List past payments for the authenticated user.
Query params: `?page=1&limit=20`
`GET /api/v1/payments/:ref` ✅
Get a specific payment by its reference code.
---
7. Notification Service `:3006`
Manages in-app/push notifications and delivery preferences.
`GET /health` · `GET /healthz` · `GET /ready` · `GET /api/v1/notifications/health`
Liveness/readiness checks. `/ready` verifies DB and Redis connectivity.
`POST /api/v1/notifications/` ✅
Send a notification (typically system/admin-triggered).
Request:
```json
{ "userId": "usr_123", "title": "Payment Successful", "body": "Your bill payment was successful." }
```
`GET /api/v1/notifications/` ✅
List notifications for the authenticated user.
Query params: `?page=1&limit=20&unread=true`
`PUT /api/v1/notifications/read-all` ✅
Mark all notifications as read.
`GET /api/v1/notifications/unread-count` ✅
Get the count of unread notifications.
Response `200`: `{ "unreadCount": 4 }`
`GET /api/v1/notifications/settings` ✅
Get the user's notification delivery preferences.
`PUT /api/v1/notifications/settings` ✅
Update notification delivery preferences.
Request:
```json
{ "email": true, "push": true, "sms": false }
```
`PUT /api/v1/notifications/:id/read` ✅
Mark a single notification as read.
`DELETE /api/v1/notifications/:id` ✅
Delete a notification.
---
8. Reports Service `:3007`
Provides transaction analytics and summaries.
`GET /health` · `GET /healthz` · `GET /ready` · `GET /api/v1/reports/health`
Liveness/readiness checks. `/ready` verifies DB connectivity.
`GET /api/v1/reports/summary` ✅
Get an overall transaction summary for the authenticated user.
`GET /api/v1/reports/monthly` ✅
Get a monthly breakdown report.
Query params: `?month=2026-07`
`GET /api/v1/reports/payments` ✅
Get a summary of payments made (bills, airtime, data).
`GET /api/v1/reports/spending` ✅
Get spending analytics, typically broken down by category.
`GET /api/v1/reports/admin/dashboard` ✅
Get aggregate platform-wide metrics (admin scope).
---
Error Handling
All services follow a consistent error response shape:
```json
{
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Email or password is incorrect.",
    "statusCode": 401
  }
}
```
Status Code	Meaning
`400`	Bad Request — validation failure
`401`	Unauthorized — missing/invalid/expired token
`403`	Forbidden — insufficient permissions
`404`	Not Found — resource doesn't exist
`409`	Conflict — duplicate resource (e.g. email already registered)
`422`	Unprocessable Entity — semantic validation error
`500`	Internal Server Error
`503`	Service Unavailable — dependency (DB/Redis) unreachable
---
Generated from the internal API endpoint reference (62 endpoints, 8 services).
