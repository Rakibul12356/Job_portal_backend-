# 💼 Online Job Portal REST API Backend

A production-grade, highly scalable REST API backend built in **Go (Gin)** and **MongoDB (Atlas)**. This backend powers candidate matching, resume uploads, real-time WebSocket chat between job seekers and employers, and job application workflows.

---

## 🌟 Key Features

*   **🔒 Secure Authentication:** JWT-based access/refresh token rotation, role-based authorization (`user` vs `company`), password hashing using `bcrypt`.
*   **✉️ Password Recovery (OTP):** Fully integrated password reset flow with securely generated 6-digit OTPs delivered via **Resend REST API** (designed to bypass cloud provider SMTP blocks).
*   **💬 Real-Time Chat System:** WebSockets (`/ws`) connection enabling real-time communications between job seekers and employers regarding job applications.
*   **📁 File Storage System:** Flexible storage layer supporting profile avatar uploads, company logos, and candidate resumes (supports local directory driver).
*   **⚡ Core Middlewares:** Rate limiting (requests per minute), CORS, Panic Recovery, Logger, and JWT Authentication validation.
*   **🛠️ Database Index Migrations:** Automatic Mongo index generation and verification on server boot to ensure optimal query execution speeds.
*   **🔄 Graceful Shutdown:** Safely drains active HTTP requests and closes MongoDB connections on OS interrupt signals.

---

## 🛠️ Tech Stack & Clean Architecture

*   **Backend:** Go (Golang) 1.22+
*   **HTTP Framework:** Gin Gonic
*   **Database:** MongoDB Atlas (Official Go Driver)
*   **Real-time Engine:** WebSockets (`gorilla/websocket`)
*   **Email Engine:** Resend REST API Client
*   **Testing:** Postman Collections

We enforce clean architectural separation of concerns to isolate side effects:

```text
Client Request ──▶ Handlers (Route Bindings, Payload Validation & DTOs)
                        │
                        ▼
                   Services (Core Business Logic, Transactions & Validation)
                        │
                        ▼
                   Repositories (Database Queries, Projections & Mutators)
                        │
                        ▼
                  MongoDB Atlas
```

---

## 📁 Project Structure

```text
├── cmd
│   └── api               # Main entrypoint (main.go) initializing all layers
├── internal
│   ├── config            # Environment configuration bindings
│   ├── database          # MongoDB connection and indexing scripts
│   ├── domain            # Core entities & schemas (User, Company, Job, Application, Room, Message)
│   ├── dto               # Data Transfer Objects for JSON binding & request validation
│   ├── handler           # Gin HTTP controllers parsing JSON payloads
│   ├── middleware        # CORS, Auth checks, Rate limits, Panic recoveries
│   ├── pkg
│   │   ├── errors        # Custom structured application error objects
│   │   ├── pagination    # Unified query paginator helper
│   │   ├── response      # Standardized JSON response envelope formatter
│   │   └── utils         # JWT generators, string formatters, and Resend email helpers
│   ├── repository        # MongoDB collection operations/queries
│   ├── router            # Endpoint groupings and middleware attachments
│   └── service           # Pure business logic and transaction coordinators
├── scripts               # Database seed scripts
└── docs                  # Postman collections & documentation guides
```

---

## 🚀 Getting Started

### Prerequisites
*   **Go 1.22+**
*   **MongoDB Atlas** Cluster (or Local MongoDB Server)
*   **Resend Account** (For OTP/Email sending)

### Installation & Run

1.  **Clone & Navigate:**
    ```bash
    git clone https://github.com/Rakibul12356/Job_portal_backend-.git
    cd Job_portal_backend-
    ```

2.  **Configure Environment Variables:**
    Copy the sample configuration file and add your credentials:
    ```bash
    cp .env.example .env
    ```

3.  **Install Dependencies:**
    ```bash
    go mod tidy
    ```

4.  **Seed Database (Optional):**
    Pre-fill your database with mock job seekers, company managers, and job openings:
    ```bash
    go run scripts/seed.go
    # or using Makefile
    make seed
    ```

5.  **Start REST API Server:**
    ```bash
    go run cmd/api/main.go
    # or using Makefile
    make run
    ```

6.  **Verify Status:**
    *   Health Check: `GET http://localhost:8080/health`
    *   Database Ping: `GET http://localhost:8080/ready`

---

## 📋 API Endpoints Summary

### 🔑 Authentication (`/api/v1/auth`)
*   `POST /register/seeker` - Sign up as a Job Seeker
*   `POST /register/employer` - Sign up as an Employer
*   `POST /login` - Sign in (returns JWT access token & HTTP-Only refresh cookie)
*   `POST /refresh` - Rotate expired access tokens
*   `POST /logout` - Clear refresh cookies and exit session
*   `POST /forgot-password` - Request a password reset OTP (delivered via Resend)
*   `POST /reset-password` - Reset password using OTP
*   `GET /me` - Get profile identity details of the current logged-in user

### 💼 Jobs Engine (`/api/v1/jobs`)
*   `GET /` - List public jobs (with search, category filters, and pagination)
*   `GET /:id` - Get specific job details
*   `GET /:id/similar` - Fetch relevant jobs matching current category
*   `POST /:id/report` - Report abuse/fake job openings
*   `POST /:id/applications` - Apply to a job (Seeker only)

### ✉️ Candidate Application Engine (`/api/v1/applications`)
*   `GET /me` - Get my job applications list (Seeker only)
*   `GET /:id` - View details of a specific application
*   `POST /:id/withdraw` - Withdraw an active application

### 💬 Chat Room & WS Messaging (`/api/v1/chats`)
*   `POST /` - Open/Create a chat room between seeker and employer
*   `GET /` - List active chat rooms for the authenticated user
*   `GET /:roomId/messages` - Load history of messages in a room
*   `GET /:roomId/ws` - Upgrade to WebSockets for real-time messaging

---

## ⚙️ Environment Variables Config

| Environment Variable | Default Value | Description |
| :--- | :--- | :--- |
| `PORT` / `APP_PORT` | `8080` | Server listening port |
| `APP_ENV` | `development` | Mode: `development` / `production` |
| `MONGO_URI` | `mongodb+srv://...` | MongoDB connection URI |
| `MONGO_DB` | `job_portal` | MongoDB database name |
| `JWT_ACCESS_SECRET` | `change-me` | Secret used to sign short-lived access tokens |
| `JWT_REFRESH_SECRET` | `change-me` | Secret used to sign refresh tokens |
| `JWT_ACCESS_TTL` | `60m` | Validity duration of access tokens |
| `JWT_REFRESH_TTL` | `168h` | Validity duration of refresh tokens |
| `RATE_LIMIT_RPM` | `60` | Client rate limit requests per minute |
| `RESEND_API_KEY` | `re_...` | Resend API Key for sending emails (Alternative: `SMTP_PASS`) |
| `SMTP_SENDER` | `onboarding@resend.dev` | Valid sender email address |

---

## 🧪 Testing with Postman

1.  Navigate to the `docs/postman/` directory.
2.  Import `Job_Portal_API.postman_collection.json` into Postman.
3.  Import `Job_Portal_Local.postman_environment.json` for local test variables.
4.  Default Test Users:
    *   **Job Seeker:** `you@example.com` / `password123`
    *   **Employer:** `company@example.com` / `password123`
