# Online Job Portal REST API Backend

A production-grade REST API backend built in **Go (Gin)** and **MongoDB (Atlas)** powering the Candidate Matching & Job Application portal.

---

## 🚀 Quick Start (Local Run Only)

This project runs locally without Docker, connecting directly to a MongoDB Atlas cluster.

### Prerequisites
- **Go 1.22+** (Installed)
- **Postman Desktop** (For endpoint testing)

### Setup Steps

1. **Clone & Open Project Directory:**
   ```bash
   cd job_portal_backend
   ```

2. **Configure Environment Variables:**
   Copy the example environment configuration. The Atlas URI is already set by default:
   ```bash
   cp .env.example .env
   ```

3. **Install Dependencies:**
   ```bash
   go mod tidy
   ```

4. **Seed Database:**
   Populate MongoDB Atlas with seed job listings, candidate profiles, and application statuses:
   ```bash
   go run scripts/seed.go
   # or: make seed
   ```

5. **Start REST API Server:**
   ```bash
   go run cmd/api/main.go
   # or: make run
   ```

6. **Verify Server Status:**
   - Health Status: [http://localhost:8080/health](http://localhost:8080/health)
   - MongoDB Ping Readiness: [http://localhost:8080/ready](http://localhost:8080/ready)

---

## 🛠️ Architecture request flow

We enforce clean architectural layers to isolate side effects:
```text
Client request ──▶ Handler (HTTP validation & DTOs)
                    │
                    ▼
                  Service (Business logics & validations)
                    │
                    ▼
                  Repository (MongoDB driver execution)
```

---

## 🧪 Postman & Swagger Testing

### Postman Collection
Import the collection located in [docs/postman/](file:///c:/Users/rakib/job_portal_backend/docs/postman/):
1. **Environment:** `Job_Portal_Local.postman_environment.json`
2. **Collection:** `Job_Portal_API.postman_collection.json`
3. Refer to [docs/postman/README.md](file:///c:/Users/rakib/job_portal_backend/docs/postman/README.md) for step-by-step endpoint executions.

### Seed Accounts

| Role | Email | Password |
|---|---|---|
| Job Seeker | `you@example.com` | `password123` |
| Company Owner | `company@example.com` | `password123` |

---

## ⚙️ Environment Variables Config

| Key | Default | Purpose |
|---|---|---|
| `APP_PORT` | `8080` | Server listening port |
| `MONGO_URI` | `mongodb+srv://Job_portal_db:...` | MongoDB Atlas cluster URI |
| `MONGO_DB` | `job_portal` | Database collection name |
| `JWT_ACCESS_SECRET` | `change-me-access-super-secret` | Access token sign secret |
| `JWT_REFRESH_SECRET` | `change-me-refresh-super-secret` | Refresh token sign secret |
| `UPLOAD_DIR` | `./uploads` | Local upload directory path |
| `RATE_LIMIT_RPM` | `60` | Client rate limit per minute |
