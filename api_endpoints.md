# Complete API Endpoint Reference

> **Total: 62 endpoints** across 8 services

---

## API Gateway `:8080`

| # | Method | Endpoint | Auth | Description |
|---|--------|----------|------|-------------|
| 1 | `GET` | `/health` | No | Gateway liveness check |
| 2 | `GET` | `/healthz` | No | Gateway liveness (K8s alias) |
| 3 | `GET` | `/health/services` | No | Aggregate downstream service status |
| 4 | `ALL` | `/api/v1/auth/*` | — | Proxy → Auth Service |
| 5 | `ALL` | `/api/v1/users/*` | — | Proxy → User Service |
| 6 | `ALL` | `/api/v1/cards/*` | — | Proxy → Card Service |
| 7 | `ALL` | `/api/v1/wallets/*` | — | Proxy → Wallet Service |
| 8 | `ALL` | `/api/v1/payments/*` | — | Proxy → Payment Service |
| 9 | `ALL` | `/api/v1/notifications/*` | — | Proxy → Notification Service |
| 10 | `ALL` | `/api/v1/reports/*` | — | Proxy → Reports Service |

---

## Auth Service `:3001`

| # | Method | Endpoint | Auth | Description |
|---|--------|----------|------|-------------|
| 1 | `GET` | `/health` | No | Liveness check |
| 2 | `GET` | `/healthz` | No | Liveness (K8s alias) |
| 3 | `GET` | `/ready` | No | Readiness (DB + Redis) |
| 4 | `GET` | `/api/v1/auth/health` | No | Service health via API prefix |
| 5 | `POST` | `/api/v1/auth/register` | No | Register new user |
| 6 | `POST` | `/api/v1/auth/login` | No | User login |
| 7 | `POST` | `/api/v1/auth/refresh-token` | No | Refresh JWT token |
| 8 | `POST` | `/api/v1/auth/logout` | ✅ | User logout |
| 9 | `PUT` | `/api/v1/auth/change-password` | ✅ | Change password |

---

## User Service `:3002`

| # | Method | Endpoint | Auth | Description |
|---|--------|----------|------|-------------|
| 1 | `GET` | `/health` | No | Liveness check |
| 2 | `GET` | `/healthz` | No | Liveness (K8s alias) |
| 3 | `GET` | `/ready` | No | Readiness (DB) |
| 4 | `GET` | `/api/v1/users/health` | No | Service health via API prefix |
| 5 | `GET` | `/api/v1/users/profile` | ✅ | Get user profile |
| 6 | `PUT` | `/api/v1/users/profile` | ✅ | Update user profile |
| 7 | `GET` | `/api/v1/users/` | ✅ | Get all users |
| 8 | `GET` | `/api/v1/users/:id` | ✅ | Get user by ID |
| 9 | `PUT` | `/api/v1/users/:id/status` | ✅ | Update user status |
| 10 | `PUT` | `/api/v1/users/:id/kyc` | ✅ | Update KYC status |
| 11 | `DELETE` | `/api/v1/users/:id` | ✅ | Delete user |

---

## Card Service `:3003`

| # | Method | Endpoint | Auth | Description |
|---|--------|----------|------|-------------|
| 1 | `GET` | `/health` | No | Liveness check |
| 2 | `GET` | `/healthz` | No | Liveness (K8s alias) |
| 3 | `GET` | `/ready` | No | Readiness (DB) |
| 4 | `GET` | `/api/v1/cards/health` | No | Service health via API prefix |
| 5 | `POST` | `/api/v1/cards/` | ✅ | Create card |
| 6 | `GET` | `/api/v1/cards/` | ✅ | Get user's cards |
| 7 | `GET` | `/api/v1/cards/:id` | ✅ | Get card by ID |
| 8 | `PUT` | `/api/v1/cards/:id` | ✅ | Update card |
| 9 | `PUT` | `/api/v1/cards/:id/pin` | ✅ | Change PIN |
| 10 | `PUT` | `/api/v1/cards/:id/block` | ✅ | Block card |
| 11 | `PUT` | `/api/v1/cards/:id/unblock` | ✅ | Unblock card |
| 12 | `DELETE` | `/api/v1/cards/:id` | ✅ | Delete card |

---

## Wallet Service `:3004`

| # | Method | Endpoint | Auth | Description |
|---|--------|----------|------|-------------|
| 1 | `GET` | `/health` | No | Liveness check |
| 2 | `GET` | `/healthz` | No | Liveness (K8s alias) |
| 3 | `GET` | `/ready` | No | Readiness (DB) |
| 4 | `GET` | `/api/v1/wallets/health` | No | Service health via API prefix |
| 5 | `POST` | `/api/v1/wallets/` | ✅ | Create wallet |
| 6 | `GET` | `/api/v1/wallets/` | ✅ | Get wallet |
| 7 | `GET` | `/api/v1/wallets/balance` | ✅ | Get balance |
| 8 | `POST` | `/api/v1/wallets/fund` | ✅ | Fund wallet |
| 9 | `POST` | `/api/v1/wallets/transfer` | ✅ | Transfer funds |
| 10 | `POST` | `/api/v1/wallets/withdraw` | ✅ | Withdraw funds |
| 11 | `GET` | `/api/v1/wallets/transactions` | ✅ | Get transactions |

---

## Payment Service `:3005`

| # | Method | Endpoint | Auth | Description |
|---|--------|----------|------|-------------|
| 1 | `GET` | `/health` | No | Liveness check |
| 2 | `GET` | `/healthz` | No | Liveness (K8s alias) |
| 3 | `GET` | `/ready` | No | Readiness (DB) |
| 4 | `GET` | `/api/v1/payments/health` | No | Service health via API prefix |
| 5 | `POST` | `/api/v1/payments/bill` | ✅ | Pay bill |
| 6 | `POST` | `/api/v1/payments/airtime` | ✅ | Buy airtime |
| 7 | `POST` | `/api/v1/payments/data` | ✅ | Buy data bundle |
| 8 | `GET` | `/api/v1/payments/history` | ✅ | Get payment history |
| 9 | `GET` | `/api/v1/payments/:ref` | ✅ | Get payment by reference |

---

## Notification Service `:3006`

| # | Method | Endpoint | Auth | Description |
|---|--------|----------|------|-------------|
| 1 | `GET` | `/health` | No | Liveness check |
| 2 | `GET` | `/healthz` | No | Liveness (K8s alias) |
| 3 | `GET` | `/ready` | No | Readiness (DB + Redis) |
| 4 | `GET` | `/api/v1/notifications/health` | No | Service health via API prefix |
| 5 | `POST` | `/api/v1/notifications/` | ✅ | Send notification |
| 6 | `GET` | `/api/v1/notifications/` | ✅ | Get user notifications |
| 7 | `PUT` | `/api/v1/notifications/read-all` | ✅ | Mark all as read |
| 8 | `GET` | `/api/v1/notifications/unread-count` | ✅ | Get unread count |
| 9 | `GET` | `/api/v1/notifications/settings` | ✅ | Get notification settings |
| 10 | `PUT` | `/api/v1/notifications/settings` | ✅ | Update notification settings |
| 11 | `PUT` | `/api/v1/notifications/:id/read` | ✅ | Mark single as read |
| 12 | `DELETE` | `/api/v1/notifications/:id` | ✅ | Delete notification |

---

## Reports Service `:3007`

| # | Method | Endpoint | Auth | Description |
|---|--------|----------|------|-------------|
| 1 | `GET` | `/health` | No | Liveness check |
| 2 | `GET` | `/healthz` | No | Liveness (K8s alias) |
| 3 | `GET` | `/ready` | No | Readiness (DB) |
| 4 | `GET` | `/api/v1/reports/health` | No | Service health via API prefix |
| 5 | `GET` | `/api/v1/reports/summary` | ✅ | Transaction summary |
| 6 | `GET` | `/api/v1/reports/monthly` | ✅ | Monthly report |
| 7 | `GET` | `/api/v1/reports/payments` | ✅ | Payment summary |
| 8 | `GET` | `/api/v1/reports/spending` | ✅ | Spending analytics |
| 9 | `GET` | `/api/v1/reports/admin/dashboard` | ✅ | Admin dashboard |
