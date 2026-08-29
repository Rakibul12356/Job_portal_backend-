# Deploying Online Job Portal Backend (Render / Railway)

This guide walks you through deploying the Go + MongoDB REST API backend to production on **Render** or **Railway** for free.

---

## 🛠️ Step 1: Prepare Codebase for Production

1.  **GOPATH & Port Binding Support:**
    We have updated the code to dynamically listen to the `PORT` environment variable injected by Render/Railway:
    ```go
    AppPort: getEnv("PORT", getEnv("APP_PORT", "8080"))
    ```

2.  **Git Integration:**
    Make sure your project is committed to a GitHub repository. Initialize Git if you haven't:
    ```bash
    git init
    git add .
    git commit -m "feat: initial commit for deployment"
    ```
    Create a repository on GitHub and push your code:
    ```bash
    git remote add origin <your-github-repo-url>
    git branch -M main
    git push -u origin main
    ```

> [!WARNING]
> **Ephemeral File Storage Limitation:**
> Both Render and Railway use ephemeral file systems. Files uploaded to the local `./uploads` directory will be **wiped out** whenever the container restarts or redeploys.
> - **Workaround for v1:** It is fine to use `./uploads` for quick demos, but data will be reset.
> - **Production solution:** In `internal/service/storage_service.go`, swap the local driver with an AWS S3, Cloudinary, or Supabase Storage adapter.

---

## 🚀 Option A: Deploying on Render (Free)

Render offers free hosting for web services.

### Steps:
1.  Go to [Render](https://render.com/) and log in/register.
2.  Click **New +** and select **Web Service**.
3.  Connect your GitHub account and select your repository.
4.  Configure the service details:
    -   **Name:** `job-portal-api` (or any name)
    -   **Region:** Select closest to you (e.g., Singapore or Oregon)
    -   **Branch:** `main`
    -   **Runtime:** `Go`
    -   **Build Command:**
        ```bash
        go build -o bin/api cmd/api/main.go
        ```
    -   **Start Command:**
        ```bash
        ./bin/api
        ```
    -   **Instance Type:** `Free`
5.  Click **Advanced** to add **Environment Variables** (see below).
6.  Click **Create Web Service**. Render will compile and deploy your app.

---

## 🚀 Option B: Deploying on Railway (Free Trial)

Railway offers easy deployments and sets up Go automatically.

### Steps:
1.  Go to [Railway](https://railway.app/) and register.
2.  Click **New Project** -> **Deploy from GitHub repo**.
3.  Select your repository.
4.  Add variables by clicking the **Variables** tab (see table below).
5.  Railway will detect the Go project, build, and deploy it automatically.

---

## 🔑 Environment Variables to Configure (Production)

You must configure these variables in the **Variables** panel of Render or Railway:

| Key | Example Value | Description |
|---|---|---|
| `APP_ENV` | `production` | Switches log formats and Gin release mode |
| `MONGO_URI` | `mongodb+srv://Job_portal_db:rakib74@...` | Your Atlas MongoDB URL (already cloud hosted) |
| `MONGO_DB` | `job_portal` | Database name |
| `JWT_ACCESS_SECRET` | `generate-a-secure-random-hash-here` | Secret key for signing access tokens |
| `JWT_REFRESH_SECRET` | `generate-another-secure-random-hash` | Secret key for signing refresh tokens |
| `JWT_ACCESS_TTL` | `60m` | Access token lifetime |
| `JWT_REFRESH_TTL` | `168h` | Refresh token lifetime |
| `CORS_ORIGINS` | `https://your-frontend.vercel.app` | **CRITICAL:** Set to your live React frontend URL |
| `UPLOAD_DRIVER` | `local` | Upload adaptor |
| `UPLOAD_DIR` | `./uploads` | Temporary upload folder |
| `RATE_LIMIT_RPM` | `60` | Client request rate limit per minute |

---

## 🧪 Seeding Production Database

If you want to seed your production database:
1.  Temporary change your local `.env` `MONGO_URI` to your production MongoDB URI (if different).
2.  Run the seed script from your local machine:
    ```bash
    go run scripts/seed.go
    ```
    Since MongoDB Atlas is already cloud-hosted, this will write directly to your database cluster.
