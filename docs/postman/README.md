# Job Portal API Postman Guide

This directory contains the ready-to-import Postman collection and local environment files to test **all API endpoints** without opening the frontend.

## Files
1. `Job_Portal_API.postman_collection.json` — The endpoints collection.
2. `Job_Portal_Local.postman_environment.json` — Local testing environment variables.

## How to Import & Run
1. Open Postman.
2. Click **Import** in the top-left corner.
3. Drag and drop both files:
   - `Job_Portal_API.postman_collection.json`
   - `Job_Portal_Local.postman_environment.json`
4. Confirm import.

## Select Environment
In the top-right corner of Postman, select the environment **Job Portal Local**.

## Local Server Startup
Run the server locally:
```bash
go run ./cmd/api
# or make run
```

## Seed Credentials
The environment comes configured with the following credentials to authenticate:

| Role | Email | Password |
|---|---|---|
| Job Seeker | `you@example.com` | `password123` |
| Company Owner | `company@example.com` | `password123` |

## Test Sequence
To ensure IDs are automatically saved and referenced across variables:
1. Fire `00 Health` -> `GET Health` and `GET Ready` to check database/server status.
2. Fire `01 Auth` -> `POST Login Seeker`. This automatically updates your `{{accessToken}}` and `{{userId}}`.
3. Fire `02 Public Jobs` -> `GET List Jobs`. This saves `{{jobId}}` automatically.
4. Fire `03 Seeker — Apply & Applications` -> `POST Apply to Job`. Attach a dummy PDF file in the `resume` form-data key. This saves `{{applicationId}}`.
5. Fire `01 Auth` -> `POST Login Company` to switch authentication to Company mode. This saves `{{companyId}}`.
6. Fire `06 Company — Jobs` -> `POST Create Job (Publish)`.
7. Fire `07 Company — Applicants` -> `GET Applicants` -> `PATCH Shortlist`.
8. Fire `08 Company — Profile & Settings` -> `GET Company Dashboard`.
