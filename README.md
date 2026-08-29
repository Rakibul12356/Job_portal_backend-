# 💼 Online Job Portal REST API Backend

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-blue?logo=go&logoColor=white)](https://golang.org)
[![Gin Framework](https://img.shields.io/badge/Framework-Gin--Gonic-00ADD8?logo=go&logoColor=white)](https://gin-gonic.com)
[![MongoDB](https://img.shields.io/badge/Database-MongoDB%20Atlas-47A248?logo=mongodb&logoColor=white)](https://www.mongodb.com)
[![WebSockets](https://img.shields.io/badge/RealTime-WebSockets-010101?logo=socket.io&logoColor=white)](https://github.com/gorilla/websocket)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A production-grade, highly scalable REST API backend built in **Go (Gin)** and **MongoDB (Atlas)**. This backend powers candidate matching, profile/resume uploads, real-time WebSocket chat between job seekers and employers, role-based analytics dashboards, background matching alert email digests, and complex job application workflows.

---

## 🌟 Key Features

*   **🔒 Secure Authentication:** JWT-based access/refresh token rotation (with cookies), role-based authorization (`user` vs `company`), password hashing using `bcrypt`.
*   **✉️ Password Recovery (OTP via Gmail API):** Fully integrated password reset flow with securely generated 6-digit OTPs delivered via **Google Gmail REST API (OAuth 2.0)** (bypasses hosting provider/cloud SMTP blocks by routing requests over HTTPS port 443).
*   **💬 Real-Time Chat System:** WebSockets enabling real-time communications between job seekers and employers. Features a global activity socket (`/ws`) and room-specific sockets.
*   **📁 File Storage System:** Flexible storage layer supporting profile avatar uploads, company logos, and candidate resumes (PDF/Doc files) with local directory storage.
*   **⭐ Saved Jobs Center:** Job Seekers can save jobs to apply later, unsave them, and retrieve lists of saved openings.
*   **👤 Seeker Profile Management:** Manage resumes, profile pictures/avatars, professional experience, and educational background.
*   **🏢 Employer Job Control:** Full CRUD, draft publishing, closing, and bulk actions (bulk close, bulk delete) on job postings.
*   **👥 Candidate Tracker & Status Pipeline:** View and filter applicants who applied to company jobs, download candidate resumes directly, and update application status (`Reviewed`, `Shortlisted`, `Accepted`, `Rejected`).
*   **💡 Insights Dashboards:** Custom analytics dashboards for seekers (application and saved job stats) and employers (active jobs, candidate count, application status pipeline metrics).
*   **🔔 Real-Time Notifications:** Live in-app notifications system alerting users about status updates on applications, new chats, and matches.
*   **⏰ Job Alerts Daemon:** A native background ticker service running concurrently. In development, scans every 2 minutes; in production, runs weekly to match active jobs with user skills and dispatches matching HTML digests via email.
*   **⚡ Core Middlewares:** Rate limiting (requests per minute), CORS (origins configurable), Panic Recovery, Logger, and JWT Authentication.
*   **🛠️ Database Index Migrations:** Automatic Mongo index generation and verification (indexes for search, pagination, and fields) on server boot to ensure optimal query execution speeds.
*   **🔄 Graceful Shutdown:** Safely drains active HTTP requests and closes MongoDB connections on OS interrupt signals.

---

## 🛠️ Tech Stack & Clean Architecture

*   **Backend Language:** Go (Golang) 1.22+
*   **HTTP Framework:** Gin Gonic
*   **Database:** MongoDB Atlas (Official Go Driver)
*   **Real-time Engine:** WebSockets (`gorilla/websocket`)
*   **Email Engine:** Google Gmail API (via OAuth 2.0 authorization)
*   **Testing:** Postman Collections, native Go unit/integration tests

We enforce clean architectural separation of concerns to isolate side effects and maintain a maintainable codebase:

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
                   MongoDB Atlas (Database Persistence Layer)
```

---

## 📁 Project Structure

```text
├── cmd
│   └── api               # Main entrypoint (main.go) initializing all layers & workers
├── internal
│   ├── config            # Environment configuration bindings and parser
│   ├── database          # MongoDB connection and indexing scripts
│   ├── domain            # Core entities, structs, & schemas (User, Company, Job, Application, Room, Message, etc.)
│   ├── dto               # Data Transfer Objects for JSON binding & request validation
│   ├── handler           # Gin HTTP controllers parsing JSON payloads & returning responses
│   ├── middleware        # CORS, Auth validation, Rate limits, Panic recoveries
│   ├── pkg
│   │   ├── errors        # Custom structured application error objects
│   │   ├── pagination    # Unified query paginator helper
│   │   ├── response      # Standardized JSON response envelope formatter
│   │   └── utils         # JWT generators, string formatters, and Gmail OAuth mailer helpers
│   ├── repository        # MongoDB collection operations/queries
│   ├── router            # Endpoint groupings and middleware attachments
│   └── service           # Pure business logic, transaction coordinators, and background workers
├── scripts               # Database seed scripts and load testing scripts
└── docs                  # Postman collections, DEPLOY guides, and technical specifications
```

---

## 🚀 Getting Started

### Prerequisites
*   **Go 1.22+**
*   **MongoDB Atlas** Cluster (or Local MongoDB Server)
*   **Gmail OAuth 2.0 Credentials** (Client ID, Client Secret, Refresh Token)

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
    Ensure you fill in your database connections and your Gmail OAuth client credentials in `.env`.

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
*   `POST /forgot-password` - Request a password reset OTP (delivered via Gmail OAuth 2.0)
*   `POST /reset-password` - Reset password using OTP
*   `GET /me` - Get profile identity details of the current logged-in user

### 💼 Jobs Engine (`/api/v1/jobs`)
*   `GET /` - List public jobs (with search, category, location, salary filters, and pagination)
*   `GET /:id` - Get specific job details
*   `GET /:id/similar` - Fetch relevant jobs matching current category
*   `POST /:id/report` - Report abuse/fake job openings
*   `POST /:id/applications` - Apply to a job (Seeker only)

### ⭐ Saved Jobs (`/api/v1/saved-jobs`) - Job Seeker Only
*   `GET /` - Get list of saved jobs for the authenticated seeker
*   `POST /` - Save a job (body: `{ "jobId": "string" }`)
*   `DELETE /:jobId` - Unsave a job

### 👤 Seeker Profile (`/api/v1/profile/me`) - Job Seeker Only
*   `GET /` - Get details of current seeker profile (skills, education, experience)
*   `PUT /` - Update seeker profile details
*   `POST /avatar` - Upload avatar image (multipart form-data)
*   `DELETE /avatar` - Remove avatar image
*   `POST /resume` - Upload resume PDF/Doc (multipart form-data)
*   `DELETE /resume` - Remove resume PDF/Doc
*   *Seeker Experience:*
    *   `POST /experience` - Add professional experience item
    *   `PUT /experience/:expId` - Update specific experience details
    *   `DELETE /experience/:expId` - Delete specific experience
*   *Seeker Education:*
    *   `POST /education` - Add educational details
    *   `PUT /education/:eduId` - Update specific education details
    *   `DELETE /education/:eduId` - Delete specific education

### 🏢 Company & Jobs Management (`/api/v1/company` & `/api/v1/company/jobs`) - Employer Only
*   `GET /company/profile` - Get own company profile details
*   `PUT /company/profile` - Update company profile settings
*   `GET /company/settings` - Get company settings
*   `PUT /company/settings` - Update company settings
*   `POST /company/logo` - Upload company logo
*   `DELETE /company/logo` - Remove company logo
*   `GET /company/jobs` - List jobs posted by this employer
*   `POST /company/jobs` - Create/Post a new job (Draft or Active status)
*   `GET /company/jobs/:id` - Get specific job details with application metrics
*   `PUT /company/jobs/:id` - Update job details
*   `DELETE /company/jobs/:id` - Delete a job
*   `POST /company/jobs/:id/publish` - Publish a draft job
*   `POST /company/jobs/:id/close` - Close an active job
*   `POST /company/jobs/:id/reactivate` - Reactivate a closed job
*   `POST /company/jobs/bulk` - Perform bulk actions (close/delete/publish)

### 👥 Company Applicants Management (`/api/v1/company/applicants`) - Employer Only
*   `GET /` - List all applicants for company's job postings
*   `GET /:id` - Get details of a specific candidate application
*   `PATCH /:id/status` - Update applicant's pipeline status (`Reviewed`, `Shortlisted`, `Accepted`, `Rejected`)
*   `GET /:id/resume` - Download applicant's resume

### 🏢 Public Companies Info (`/api/v1/companies`) - Public
*   `GET /` - List public details of all registered companies
*   `GET /:id` - Get details of a specific company

### 💬 Chat Rooms & WebSockets (`/api/v1/chats`) - Authenticated
*   `POST /` - Open/Create a chat room between seeker and employer
*   `GET /` - List active chat rooms for the user
*   `GET /:roomId/messages` - Load history of messages in a room
*   `GET /ws` - Upgrade to global WebSocket for notifications / online status
*   `GET /:roomId/ws` - Upgrade to room-specific WebSocket for real-time messaging

### 🔔 Notifications (`/api/v1/notifications`) - Authenticated
*   `GET /` - Fetch all notifications for the user
*   `PATCH /:id/read` - Mark notification as read

### 💡 Dashboards Insights (`/api/v1/dashboard`) - Authenticated
*   `GET /seeker` - Seeker dashboard insights (application statuses count, saved jobs count)
*   `GET /company` - Employer dashboard insights (active jobs, candidate count, application status counts)

---

## ⚙️ Environment Variables Config

| Environment Variable | Default Value | Description |
| :--- | :--- | :--- |
| `APP_ENV` | `development` | Mode: `development` / `production` |
| `APP_PORT` | `8080` | Server listening port |
| `APP_BASE_URL` | `http://localhost:8080` | Server base url for storage path constructions |
| `MONGO_URI` | `mongodb+srv://...` | MongoDB connection URI |
| `MONGO_DB` | `job_portal` | MongoDB database name |
| `JWT_ACCESS_SECRET` | `change-me-access...` | Secret used to sign short-lived access tokens |
| `JWT_REFRESH_SECRET` | `change-me-refresh...`| Secret used to sign refresh tokens |
| `JWT_ACCESS_TTL` | `60m` | Validity duration of access tokens |
| `JWT_REFRESH_TTL` | `168h` | Validity duration of refresh tokens |
| `CORS_ORIGINS` | `http://localhost:5173` | Allowed frontend origins (comma separated) |
| `UPLOAD_DRIVER` | `local` | Upload storage driver (`local` supported) |
| `UPLOAD_DIR` | `./uploads` | Destination folder for uploaded files |
| `RATE_LIMIT_RPM` | `60` | Client rate limit requests per minute (set to `0` to disable) |
| **`GMAIL_CLIENT_ID`** | *Required* | Google OAuth 2.0 Client ID for Gmail API |
| **`GMAIL_CLIENT_SECRET`**| *Required* | Google OAuth 2.0 Client Secret for Gmail API |
| **`GMAIL_REFRESH_TOKEN`**| *Required* | Google OAuth 2.0 Refresh Token |
| **`GMAIL_SENDER_EMAIL`** | *Required* | Configured Gmail address used to dispatch OTPs/Alerts |

---

## 🧪 Testing and Verification

### Postman Collections
1.  Navigate to the `docs/postman/` directory.
2.  Import `Job_Portal_API.postman_collection.json` into Postman.
3.  Import `Job_Portal_Local.postman_environment.json` for local test variables.
4.  Use default test users created by the seeding script:
    *   **Job Seeker:** `you@example.com` / `password123`
    *   **Employer:** `company@example.com` / `password123`

### Command Line / Makefile
The project provides a simplified workflow via Makefile:
*   **Run Server:** `make run`
*   **Run Tests:** `make test`
*   **Seed Database:** `make seed`
