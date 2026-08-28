# Job Portal API — Frontend Integration Guide

This document describes all API endpoints available for the frontend team to integrate with the backend. 

### 🌐 Base URL
* **Production/Live:** `https://job-portal-backend-1-dv1h.onrender.com/api/v1`
* **Local Development:** `http://localhost:8080/api/v1`

---

## 📋 Table of Contents
1. [Global API Conventions](#1-global-api-conventions)
2. [01 - Authentication & Session](#2-01---authentication--session)
3. [02 - Public Jobs](#3-02---public-jobs)
4. [03 - Job Seeker Features](#4-03---job-seeker-features)
   - [Job Application](#job-application)
   - [Applications Management](#applications-management)
   - [Saved Jobs](#saved-jobs)
   - [Seeker Profile](#seeker-profile)
   - [Dashboard](#seeker-dashboard)
5. [04 - Employer / Company Features](#5-04---employer--company-features)
   - [Manage Jobs](#manage-jobs)
   - [Applicants Management](#applicants-management-1)
   - [Company Settings & Profile](#company-settings--profile)
   - [Dashboard](#company-dashboard)
6. [05 - Public Companies](#6-05---public-companies)

---

## 1. Global API Conventions

### Success Envelope
All success API calls return a structured JSON response:
```json
{
  "success": true,
  "message": "OK",
  "data": { ... } // Single object or collection
}
```

### Paginated List Envelope
Endpoints returning lists use this structure:
```json
{
  "success": true,
  "message": "OK",
  "data": {
    "items": [ ... ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 24,
      "totalPages": 3
    }
  }
}
```

### Error Envelope
Validation, security, or server errors return a standard layout with HTTP status code:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR", // e.g. UNAUTHORIZED, BAD_REQUEST, NOT_FOUND, DUPLICATE_KEY
    "message": "Validation failed",
    "details": [
      {
        "field": "email",
        "message": "must be a valid email"
      }
    ]
  }
}
```

### Authentication Header
For all endpoints labeled **[Auth Required]**, the frontend must include the JWT access token in the headers:
```http
Authorization: Bearer <your_access_token>
```

---

## 2. 01 - Authentication & Session

### Register Job Seeker
Create a new user account with role `user`.
* **Method:** `POST`
* **Path:** `/auth/register/seeker`
* **Auth:** Public
* **Request Body (JSON):**
```json
{
  "name": "John Doe",
  "email": "john.doe@example.com",
  "phone": "+1 (555) 123-4567",
  "experience": "mid", // Options: "entry", "mid", "senior", "lead", "expert"
  "password": "securePassword123",
  "confirmPassword": "securePassword123",
  "termsAccepted": true
}
```
* **Response (201 Created):**
```json
{
  "success": true,
  "message": "Registration successful",
  "data": null
}
```

---

### Register Employer
Creates a company account along with its owner user (role `company`).
* **Method:** `POST`
* **Path:** `/auth/register/employer`
* **Auth:** Public
* **Request Body (JSON):**
```json
{
  "companyName": "InnoTech Solutions",
  "email": "careers@innotech.com",
  "website": "https://innotech.com",
  "industry": "technology",
  "companySize": "500", // Options: "1-10", "50", "200", "500", "1000", "5000", "10000", "10001+"
  "foundedYear": 2018,
  "location": "New York, NY",
  "description": "InnoTech Solutions is a leading provider of software development services. We build enterprise applications using modern cloud stacks.", // Minimum 100 characters required
  "password": "employerPassword123",
  "confirmPassword": "employerPassword123",
  "termsAccepted": true,
  "verifiedAuthorized": true,
  "marketingOptIn": false
}
```
* **Response (201 Created):**
```json
{
  "success": true,
  "message": "Registration successful",
  "data": null
}
```

---

### Login (Seeker or Employer)
Exchange credentials for JWT session tokens.
* **Method:** `POST`
* **Path:** `/auth/login`
* **Auth:** Public (Rate limited)
* **Request Body (JSON):**
```json
{
  "email": "john.doe@example.com",
  "password": "securePassword123"
}
```
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIsIn...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsIn...",
    "user": {
      "id": "60d5ecb8629ef31a98e078a1",
      "email": "john.doe@example.com",
      "name": "John Doe",
      "firstName": "John",
      "role": "user", // "user" or "company"
      "companyId": null // string ObjectID if role is "company", otherwise null
    }
  }
}
```

---

### Refresh Token
Obtain a new access token using a refresh token.
* **Method:** `POST`
* **Path:** `/auth/refresh`
* **Auth:** Public
* **Request Body (JSON):**
```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIsIn..."
}
```
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Token refreshed successfully",
  "data": {
    "accessToken": "new_eyJhbGciOiJIUzI1NiIsIn...",
    "refreshToken": "new_eyJhbGciOiJIUzI1NiIsIn..."
  }
}
```

---

### Logout
Invalidate current session.
* **Method:** `POST`
* **Path:** `/auth/logout`
* **Auth:** **[Auth Required]**
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Logged out successfully",
  "data": null
}
```

---

### Get Me
Get details of the currently authenticated user session.
* **Method:** `GET`
* **Path:** `/auth/me`
* **Auth:** **[Auth Required]**
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "User session fetched",
  "data": {
    "id": "60d5ecb8629ef31a98e078a1",
    "email": "john.doe@example.com",
    "name": "John Doe",
    "firstName": "John",
    "role": "user",
    "companyId": null
  }
}
```

---

## 3. 02 - Public Jobs

### List/Search Active Jobs
Search and list all active job postings with pagination and filters.
* **Method:** `GET`
* **Path:** `/jobs`
* **Auth:** Public
* **Query Parameters:**
  * `q` (string): Search terms matching job title, description, or tags/skills.
  * `jobType` (string): Options `full-time`, `part-time`, `contract`, `freelance`, `internship` (comma-separated for multiples, e.g. `full-time,contract`).
  * `workMode` (string): Options `on-site`, `remote`, `hybrid` (comma-separated).
  * `experienceLevel` (string): Options `entry`, `mid`, `senior`, `lead`, `expert` (comma-separated).
  * `category` (string): e.g. `engineering`, `design`, `product`, `marketing`.
  * `salaryMin` (number): Minimum salary constraint.
  * `salaryMax` (number): Maximum salary constraint.
  * `location` (string): Search location.
  * `sort` (string): Options: `newest` (default), `salary_desc`, `salary_asc`.
  * `page` (number): Default `1`.
  * `limit` (number): Default `10`, max `50`.

* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Jobs fetched successfully",
  "data": {
    "items": [
      {
        "id": "60d5ecb8629ef31a98e078b5",
        "title": "Senior React Developer",
        "company": "InnoTech Solutions",
        "companyId": "60d5ecb8629ef31a98e078b0",
        "location": "New York, NY (Hybrid)",
        "postedAt": "2026-08-20T10:00:00Z",
        "postedLabel": "5 days ago",
        "category": "engineering",
        "description": "We are looking for a Senior React Developer to join our growing tech team...",
        "tags": ["Full-time", "Hybrid", "Senior Level"],
        "salary": "$120k - $150k",
        "salaryMin": 120000,
        "salaryMax": 150000,
        "salaryPeriod": "yearly",
        "applicants": 12,
        "jobType": "full-time",
        "workMode": "hybrid",
        "experienceLevel": "senior",
        "deadline": "2026-10-31T00:00:00Z",
        "status": "active"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 1,
      "totalPages": 1
    }
  }
}
```

---

### Get Job Details
Get a single job posting details, requirements, and benefits.
* **Method:** `GET`
* **Path:** `/jobs/:id`
* **Auth:** Public
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Job details fetched successfully",
  "data": {
    "id": "60d5ecb8629ef31a98e078b5",
    "title": "Senior React Developer",
    "company": "InnoTech Solutions",
    "companyId": "60d5ecb8629ef31a98e078b0",
    "logoUrl": "https://job-portal-backend-1-dv1h.onrender.com/uploads/logos/60d5ecb8629ef31a98e07b0/logo.png",
    "location": "New York, NY (Hybrid)",
    "postedAt": "2026-08-20T10:00:00Z",
    "postedLabel": "5 days ago",
    "category": "engineering",
    "description": "We are looking for a Senior React Developer to join our growing tech team...",
    "requirements": "Requirements:\n- 5+ years of production experience with React.\n- Expertise in state management (Redux, Zustand).\n- Strong TypeScript skills.",
    "benefits": "Benefits:\n- Health, dental, and vision insurance.\n- Remote workspace setup budget.\n- 401(k) matching.",
    "tags": ["Full-time", "Hybrid", "Senior Level"],
    "salary": "$120k - $150k",
    "salaryMin": 120000,
    "salaryMax": 150000,
    "salaryPeriod": "yearly",
    "applicants": 12,
    "jobType": "full-time",
    "workMode": "hybrid",
    "experienceLevel": "senior",
    "deadline": "2026-10-31T00:00:00Z",
    "status": "active",
    "skills": ["React", "TypeScript", "Redux", "CSS"],
    "vacancies": 2,
    "companyInfo": {
      "companyName": "InnoTech Solutions",
      "industry": "Information Technology",
      "about": "InnoTech Solutions builds cloud integrations for logistics.",
      "website": "https://innotech.example.com",
      "location": "New York, New York, United States",
      "employees": "500",
      "companySize": "500",
      "companyType": "private",
      "founded": "Founded in 2018",
      "logoUrl": "https://job-portal-backend-1-dv1h.onrender.com/uploads/logos/60d5ecb8629ef31a98e07b0/logo.png"
    }
  }
}
```

---

### Get Similar Jobs
Retrieve a list of active postings similar to the target job (based on skills and category).
* **Method:** `GET`
* **Path:** `/jobs/:id/similar`
* **Auth:** Public
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Similar jobs fetched",
  "data": [
    {
      "id": "60d5ecb8629ef31a98e079c1",
      "title": "Front End Developer (React/Next)",
      "company": "PixelPerfect Design",
      "companyId": "60d5ecb8629ef31a98e079a0",
      "location": "Remote",
      "postedAt": "2026-08-22T08:00:00Z",
      "postedLabel": "3 days ago",
      "category": "engineering",
      "tags": ["Full-time", "Remote", "Mid Level"],
      "salary": "$90k - $110k",
      "salaryMin": 90000,
      "salaryMax": 110000,
      "salaryPeriod": "yearly",
      "applicants": 4,
      "jobType": "full-time",
      "workMode": "remote",
      "status": "active"
    }
  ]
}
```

---

### Report Job
Flag a job posting for terms of service violations.
* **Method:** `POST`
* **Path:** `/jobs/:id/report`
* **Auth:** Optional Auth
* **Request Body (JSON):**
```json
{
  "reason": "Spam / misleading advertisement",
  "details": "The job description redirects users to buy training materials."
}
```
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Job report submitted successfully",
  "data": null
}
```

---

## 4. 03 - Job Seeker Features

### Job Application

#### Apply to Job
Submit an application for a job. Expects file upload for the resume.
* **Method:** `POST`
* **Path:** `/jobs/:id/applications`
* **Auth:** **[Auth Required]** (Role must be `user`)
* **Content-Type:** `multipart/form-data`
* **Form Parameters:**
  * `resume` (File): PDF file, maximum size **5MB**.
  * `coverMessage` (Text): Brief cover note. Maximum **500 characters**.
* **Response (201 Created):**
```json
{
  "success": true,
  "message": "Application submitted successfully",
  "data": null
}
```
* **Response (409 Conflict - if already applied):**
```json
{
  "success": false,
  "error": {
    "code": "DUPLICATE_KEY",
    "message": "You have already applied to this job"
  }
}
```

---

### Applications Management

#### Get My Applications
List all applications submitted by the logged-in seeker.
* **Method:** `GET`
* **Path:** `/applications/me`
* **Auth:** **[Auth Required]** (Role must be `user`)
* **Query Parameters:**
  * `status` (string): Filter by `pending`, `shortlisted`, `interviewed`, `rejected`, `withdrawn`.
  * `date` (string): Filter by `all` (default), `7d`, `30d`, `3m`.
  * `sort` (string): `newest` (default) or `oldest`.
  * `page`, `limit` (numbers): Pagination controls.
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Applications fetched successfully",
  "data": {
    "items": [
      {
        "id": "60d5ecb8629ef31a98e07aaa",
        "jobId": "60d5ecb8629ef31a98e07b5",
        "jobTitle": "Senior React Developer",
        "company": "InnoTech Solutions",
        "companyId": "60d5ecb8629ef31a98e07b0",
        "companyLogo": "https://job-portal-backend-1-dv1h.onrender.com/uploads/logos/60d5ecb8629ef31a98e07b0/logo.png",
        "status": "pending",
        "appliedAt": "2026-08-24T12:30:00Z",
        "resumeUrl": "https://job-portal-backend-1-dv1h.onrender.com/uploads/resumes/60d5ecb8629ef31a98e078a1/resume.pdf",
        "resumeFilename": "john_doe_resume.pdf",
        "coverMessage": "I am excited to apply for the Senior React Developer role...",
        "location": "New York, NY",
        "jobType": "full-time"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 1,
      "totalPages": 1
    }
  }
}
```

---

#### Get One Application
Get detail of a specific application owned by the seeker.
* **Method:** `GET`
* **Path:** `/applications/:id`
* **Auth:** **[Auth Required]** (Role must be `user`)
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Application fetched",
  "data": {
    "id": "60d5ecb8629ef31a98e07aaa",
    "jobId": "60d5ecb8629ef31a98e07b5",
    "jobTitle": "Senior React Developer",
    "company": "InnoTech Solutions",
    "companyId": "60d5ecb8629ef31a98e07b0",
    "companyLogo": "https://job-portal-backend-1-dv1h.onrender.com/uploads/logos/60d5ecb8629ef31a98e07b0/logo.png",
    "status": "pending",
    "appliedAt": "2026-08-24T12:30:00Z",
    "resumeUrl": "https://job-portal-backend-1-dv1h.onrender.com/uploads/resumes/60d5ecb8629ef31a98e078a1/resume.pdf",
    "resumeFilename": "john_doe_resume.pdf",
    "coverMessage": "I am excited to apply for the Senior React Developer role...",
    "location": "New York, NY",
    "jobType": "full-time"
  }
}
```

---

#### Withdraw Application
Withdraw an active job application. Only valid if the status is `pending` or `shortlisted`.
* **Method:** `POST`
* **Path:** `/applications/:id/withdraw`
* **Auth:** **[Auth Required]** (Role must be `user`)
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Application withdrawn successfully",
  "data": null
}
```

---

### Saved Jobs

#### List Saved Jobs
Get bookmarks saved by the seeker.
* **Method:** `GET`
* **Path:** `/saved-jobs`
* **Auth:** **[Auth Required]** (Role must be `user`)
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Saved jobs fetched",
  "data": [
    {
      "id": "60d5ecb8629ef31a98e07b5",
      "title": "Senior React Developer",
      "company": "InnoTech Solutions",
      "companyId": "60d5ecb8629ef31a98e07b0",
      "location": "New York, NY (Hybrid)",
      "jobType": "full-time",
      "workMode": "hybrid",
      "status": "active"
    }
  ]
}
```

---

#### Save Job
Bookmark a job posting.
* **Method:** `POST`
* **Path:** `/saved-jobs`
* **Auth:** **[Auth Required]** (Role must be `user`)
* **Request Body (JSON):**
```json
{
  "jobId": "60d5ecb8629ef31a98e07b5"
}
```
* **Response (201 Created):**
```json
{
  "success": true,
  "message": "Job bookmarked successfully",
  "data": null
}
```

---

#### Unsave Job
Remove a bookmarked job.
* **Method:** `DELETE`
* **Path:** `/saved-jobs/:jobId`
* **Auth:** **[Auth Required]** (Role must be `user`)
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Job removed from bookmarks",
  "data": null
}
```

---

### Seeker Profile

#### Get Full Profile
Retrieve profile data for the current seeker.
* **Method:** `GET`
* **Path:** `/profile/me`
* **Auth:** **[Auth Required]** (Role must be `user`)
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Profile fetched successfully",
  "data": {
    "id": "60d5ecb8629ef31a98e07cba",
    "userId": "60d5ecb8629ef31a98e078a1",
    "title": "Full Stack Engineer",
    "phone": "+1 (555) 123-4567",
    "bio": "Passionate developer specialized in JavaScript and Go. Enjoy solving architectural bottlenecks.",
    "location": {
      "city": "New York",
      "state": "New York",
      "country": "United States",
      "zipcode": "10001"
    },
    "skills": ["JavaScript", "React", "Node.js", "Go", "MongoDB"],
    "experience": [
      {
        "id": "exp_1",
        "company": "Pixel Lab",
        "title": "Software Developer",
        "location": "Boston, MA",
        "startDate": "2022-01",
        "endDate": "2024-06",
        "current": false,
        "description": "Built responsive design templates and microservices."
      }
    ],
    "education": [
      {
        "id": "edu_1",
        "school": "NYU",
        "degree": "Bachelor of Science",
        "fieldOfStudy": "Computer Science",
        "startDate": "2018-09",
        "endDate": "2022-05",
        "description": "Graduated with honors."
      }
    ],
    "avatarUrl": "https://job-portal-backend-1-dv1h.onrender.com/uploads/avatars/60d5ecb8629ef31a98e078a1/avatar.jpg",
    "resume": {
      "url": "https://job-portal-backend-1-dv1h.onrender.com/uploads/resumes/60d5ecb8629ef31a98e078a1/resume.pdf",
      "filename": "john_doe_resume.pdf",
      "uploadedAt": "2026-08-20T12:00:00Z"
    },
    "social": {
      "linkedin": "https://linkedin.com/in/johndoe",
      "github": "https://github.com/johndoe",
      "portfolio": "https://johndoe.dev"
    }
  }
}
```

---

#### Update Basic Profile Info
Update core demographics, biography, and social links.
* **Method:** `PUT`
* **Path:** `/profile/me`
* **Auth:** **[Auth Required]** (Role must be `user`)
* **Request Body (JSON):**
```json
{
  "name": "John Doe",
  "email": "john.doe@example.com",
  "phone": "+1 (555) 765-4321",
  "title": "Lead Software Architect",
  "city": "Austin",
  "state": "Texas",
  "country": "United States",
  "zipcode": "73301",
  "bio": "Lead architect focused on scalable architectures.",
  "skills": ["JavaScript", "Go", "Docker", "GraphQL"],
  "linkedin": "https://linkedin.com/in/johnarchitect",
  "github": "https://github.com/johndev",
  "portfolio": "https://john.tech"
}
```
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Profile updated successfully",
  "data": { ... } // Updated profile object
}
```

---

#### Upload Profile Avatar
* **Method:** `POST`
* **Path:** `/profile/me/avatar`
* **Auth:** **[Auth Required]**
* **Content-Type:** `multipart/form-data`
* **Form Parameters:**
  * `avatar` (File): JPG, PNG, or GIF file, maximum size **5MB**.
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Avatar uploaded successfully",
  "data": {
    "avatarUrl": "https://job-portal-backend-1-dv1h.onrender.com/uploads/avatars/60d5ecb8629ef31a98e078a1/avatar.jpg"
  }
}
```

---

#### Delete Profile Avatar
* **Method:** `DELETE`
* **Path:** `/profile/me/avatar`
* **Auth:** **[Auth Required]**
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Avatar removed successfully",
  "data": null
}
```

---

#### Upload Profile Resume
Save a default resume file associated with the profile.
* **Method:** `POST`
* **Path:** `/profile/me/resume`
* **Auth:** **[Auth Required]**
* **Content-Type:** `multipart/form-data`
* **Form Parameters:**
  * `resume` (File): PDF, DOC, or DOCX file, maximum size **5MB**.
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Resume uploaded successfully",
  "data": {
    "resume": {
      "url": "https://job-portal-backend-1-dv1h.onrender.com/uploads/resumes/60d5ecb8629ef31a98e078a1/resume.pdf",
      "filename": "john_doe_resume.pdf",
      "uploadedAt": "2026-08-25T15:00:00Z"
    }
  }
}
```

---

#### Delete Profile Resume
* **Method:** `DELETE`
* **Path:** `/profile/me/resume`
* **Auth:** **[Auth Required]**
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Resume removed successfully",
  "data": null
}
```

---

#### Add Seeker Experience
Add work experience entry to the profile list.
* **Method:** `POST`
* **Path:** `/profile/me/experience`
* **Auth:** **[Auth Required]**
* **Request Body (JSON):**
```json
{
  "company": "Big Tech Co",
  "title": "Senior Engineer",
  "location": "Seattle, WA",
  "startDate": "2024-07",
  "endDate": "",
  "current": true,
  "description": "Leading the backend migrations crew."
}
```
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Experience added successfully",
  "data": { ... } // Updated profile object containing the new experience (with generated string ID)
}
```

---

#### Update Seeker Experience
* **Method:** `PUT`
* **Path:** `/profile/me/experience/:expId`
* **Auth:** **[Auth Required]**
* **Request Body (JSON):** Same fields as [Add Experience](#add-seeker-experience).
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Experience updated successfully",
  "data": { ... } // Updated profile object
}
```

---

#### Delete Seeker Experience
* **Method:** `DELETE`
* **Path:** `/profile/me/experience/:expId`
* **Auth:** **[Auth Required]**
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Experience deleted successfully",
  "data": { ... } // Updated profile object
}
```

---

#### Add Seeker Education
* **Method:** `POST`
* **Path:** `/profile/me/education`
* **Auth:** **[Auth Required]**
* **Request Body (JSON):**
```json
{
  "school": "Stanford University",
  "degree": "Master of Science",
  "fieldOfStudy": "Software Engineering",
  "startDate": "2022-09",
  "endDate": "2024-06",
  "description": "Specialized in distributed systems."
}
```
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Education added successfully",
  "data": { ... } // Updated profile object (with generated string ID)
}
```

---

#### Update Seeker Education
* **Method:** `PUT`
* **Path:** `/profile/me/education/:eduId`
* **Auth:** **[Auth Required]**
* **Request Body (JSON):** Same fields as [Add Education](#add-seeker-education).
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Education updated successfully",
  "data": { ... } // Updated profile
}
```

---

#### Delete Seeker Education
* **Method:** `DELETE`
* **Path:** `/profile/me/education/:eduId`
* **Auth:** **[Auth Required]**
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Education deleted successfully",
  "data": { ... } // Updated profile
}
```

---

### Seeker Dashboard

#### Get Seeker Dashboard Data
Fetch counters, recently applied jobs, and matching job recommendations.
* **Method:** `GET`
* **Path:** `/dashboard/seeker`
* **Auth:** **[Auth Required]** (Role must be `user`)
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Seeker dashboard data fetched",
  "data": {
    "stats": {
      "totalApplications": 12,
      "shortlisted": 3,
      "rejected": 2,
      "pendingReviews": 7,
      "savedJobs": 5
    },
    "recentApplied": [
      {
        "id": "60d5ecb8629ef31a98e07aaa",
        "jobId": "60d5ecb8629ef31a98e07b5",
        "jobTitle": "Senior React Developer",
        "company": "InnoTech Solutions",
        "companyId": "60d5ecb8629ef31a98e07b0",
        "companyLogo": "https://job-portal-backend-1-dv1h.onrender.com/uploads/logos/60d5ecb8629ef31a98e07b0/logo.png",
        "status": "pending",
        "appliedAt": "2026-08-24T12:30:00Z"
      }
    ],
    "recommendedJobs": [
      {
        "id": "60d5ecb8629ef31a98e07c12",
        "title": "Go Backend Developer",
        "company": "InnoTech Solutions",
        "companyId": "60d5ecb8629ef31a98e07b0",
        "location": "Remote",
        "postedAt": "2026-08-24T09:00:00Z",
        "postedLabel": "Today",
        "category": "engineering",
        "tags": ["Full-time", "Remote", "Senior Level"],
        "salary": "$130k - $160k",
        "status": "active"
      }
    ]
  }
}
```

---

## 5. 04 - Employer / Company Features

### Manage Jobs

#### List Owned Jobs
List postings created by this employer for the backend dashboard.
* **Method:** `GET`
* **Path:** `/company/jobs`
* **Auth:** **[Auth Required]** (Role must be `company`)
* **Query Parameters:**
  * `q` (string): Text filter.
  * `status` (string): Filter by `draft`, `active`, `expiring_soon`, `closed`.
  * `sort` (string): `newest` (default) or `oldest`.
  * `page`, `limit` (numbers): Pagination parameters.
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Jobs fetched successfully",
  "data": {
    "items": [
      {
        "id": "60d5ecb8629ef31a98e07b5",
        "title": "Senior React Developer",
        "company": "InnoTech Solutions",
        "companyId": "60d5ecb8629ef31a98e07b0",
        "location": "New York, NY",
        "postedAt": "2026-08-20T10:00:00Z",
        "postedLabel": "5 days ago",
        "category": "engineering",
        "applicants": 12,
        "jobType": "full-time",
        "workMode": "hybrid",
        "experienceLevel": "senior",
        "deadline": "2026-10-31T00:00:00Z",
        "status": "active"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 1,
      "totalPages": 1
    }
  }
}
```

---

#### Create Job Posting
Post a new job as `draft` or `active`. If status is `active`, all schema validations apply.
* **Method:** `POST`
* **Path:** `/company/jobs`
* **Auth:** **[Auth Required]** (Role must be `company`)
* **Request Body (JSON):**
```json
{
  "title": "Go Backend Architect",
  "jobType": "full-time", // Options: "full-time", "part-time", "contract", "freelance", "internship"
  "workMode": "remote", // Options: "on-site", "remote", "hybrid"
  "category": "engineering", // Options: "engineering", "design", "product", "marketing", "sales", "hr", "finance", "other"
  "experienceLevel": "lead", // Options: "entry", "mid", "senior", "lead", "expert"
  "location": "Remote (USA/Canada)",
  "salaryMin": 150000,
  "salaryMax": 200000,
  "salaryPeriod": "yearly", // Options: "yearly", "monthly", "hourly"
  "description": "We are seeking a Go expert to oversee the design of our microservice architecture.",
  "requirements": "Key Qualifications:\n- 8+ years experience in software engineering.\n- 4+ years shipping Go services in production.\n- Experience with high throughput messaging streams (Kafka/NSQ).",
  "benefits": "Key Benefits:\n- Fully remote.\n- Unlimited PTO.\n- Hardware allowance.",
  "skills": ["Go", "Distributed Systems", "gRPC", "Kafka"],
  "vacancies": 1,
  "deadline": "2026-12-31T23:59:59Z",
  "status": "active" // "active" or "draft". If blank, defaults to "active"
}
```
* **Response (201 Created):**
```json
{
  "success": true,
  "message": "Job created successfully",
  "data": {
    "id": "60d5ecb8629ef31a98e07fff" // Created Job ID
  }
}
```

---

#### Get Own Job Details (for Editing)
* **Method:** `GET`
* **Path:** `/company/jobs/:id`
* **Auth:** **[Auth Required]** (Role must be `company` + owner check)
* **Response (200 OK):** Same schema as [Get Job Details](#get-job-details).

---

#### Update Job Posting
* **Method:** `PUT`
* **Path:** `/company/jobs/:id`
* **Auth:** **[Auth Required]** (Role must be `company` + owner check)
* **Request Body (JSON):** Same fields as [Create Job Posting](#create-job-posting). All fields are optional.
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Job updated successfully",
  "data": null
}
```

---

#### Delete Job Posting
* **Method:** `DELETE`
* **Path:** `/company/jobs/:id`
* **Auth:** **[Auth Required]** (Role must be `company` + owner check)
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Job deleted successfully",
  "data": null
}
```

---

#### Publish Draft Job
Change job status from `draft` to `active`.
* **Method:** `POST`
* **Path:** `/company/jobs/:id/publish`
* **Auth:** **[Auth Required]** (Role must be `company` + owner check)
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Job published successfully",
  "data": null
}
```

---

#### Close Active Job
Set job status to `closed`. Applications will be disabled.
* **Method:** `POST`
* **Path:** `/company/jobs/:id/close`
* **Auth:** **[Auth Required]** (Role must be `company` + owner check)
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Job closed successfully",
  "data": null
}
```

---

#### Reactivate Closed Job
Reopen a closed job posting (sets status back to `active`).
* **Method:** `POST`
* **Path:** `/company/jobs/:id/reactivate`
* **Auth:** **[Auth Required]** (Role must be `company` + owner check)
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Job reactivated successfully",
  "data": null
}
```

---

#### Bulk Actions
Apply status changes or deletion to multiple jobs in one call.
* **Method:** `POST`
* **Path:** `/company/jobs/bulk`
* **Auth:** **[Auth Required]** (Role must be `company`)
* **Request Body (JSON):**
```json
{
  "jobIds": ["60d5ecb8629ef31a98e07b5", "60d5ecb8629ef31a98e07c12"],
  "action": "deactivate" // Options: "activate", "deactivate", "delete"
}
```
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Bulk action applied successfully",
  "data": null
}
```

---

### Applicants Management

#### List Job Applicants
View applications submitted to your company jobs.
* **Method:** `GET`
* **Path:** `/company/applicants`
* **Auth:** **[Auth Required]** (Role must be `company`)
* **Query Parameters:**
  * `status` (string): `pending`, `shortlisted`, `interviewed`, `rejected`.
  * `experienceLevel` (string): `entry`, `mid`, `senior`, `lead`, `expert`.
  * `date` (string): Filter by applied time: `24h`, `7d`, `30d`, `all` (default).
  * `jobId` (string): Optional filter to show applicants only for one job.
  * `page`, `limit` (numbers): Pagination values.
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Applicants fetched successfully",
  "data": {
    "items": [
      {
        "id": "60d5ecb8629ef31a98e07aaa",
        "jobId": "60d5ecb8629ef31a98e07b5",
        "jobTitle": "Senior React Developer",
        "userId": "60d5ecb8629ef31a98e078a1",
        "seekerName": "John Doe",
        "seekerEmail": "john.doe@example.com",
        "seekerTitle": "Full Stack Engineer",
        "seekerPhone": "+1 (555) 123-4567",
        "seekerAvatar": "https://job-portal-backend-1-dv1h.onrender.com/uploads/avatars/60d5ecb8629ef31a98e078a1/avatar.jpg",
        "seekerSkills": ["JavaScript", "React", "Node.js", "Go"],
        "status": "pending",
        "appliedAt": "2026-08-24T12:30:00Z",
        "resumeUrl": "https://job-portal-backend-1-dv1h.onrender.com/uploads/resumes/60d5ecb8629ef31a98e078a1/resume.pdf",
        "resumeFilename": "john_doe_resume.pdf",
        "coverMessage": "I am excited to apply for the Senior React Developer role..."
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 1,
      "totalPages": 1
    }
  }
}
```

---

#### Get Applicant Details
Get detail snapshot of an application.
* **Method:** `GET`
* **Path:** `/company/applicants/:id`
* **Auth:** **[Auth Required]** (Role must be `company` + owner check)
* **Response (200 OK):** Same structure as a single item in [List Applicants Response](#list-job-applicants).

---

#### Update Applicant Status
Progress or reject an applicant in the pipeline.
* **Method:** `PATCH`
* **Path:** `/company/applicants/:id/status`
* **Auth:** **[Auth Required]** (Role must be `company` + owner check)
* **Request Body (JSON):**
```json
{
  "status": "shortlisted" // Options: "shortlisted", "interviewed", "rejected"
}
```
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Applicant status updated successfully",
  "data": null
}
```

---

#### Download Applicant Resume
Stream/download the PDF resume file.
* **Method:** `GET`
* **Path:** `/company/applicants/:id/resume`
* **Auth:** **[Auth Required]** (Role must be `company` + owner check)
* **Response:** File download stream (`application/pdf`).

---

### Company Settings & Profile

#### Get Company Settings
Load settings form data.
* **Method:** `GET`
* **Path:** `/company/settings`
* **Auth:** **[Auth Required]** (Role must be `company`)
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Company settings fetched",
  "data": {
    "companyName": "InnoTech Solutions",
    "industry": "Information Technology",
    "companySize": "500",
    "companyType": "private", // Options: "startup", "private", "public", "non-profit", "government", "educational", "partnership"
    "website": "https://innotech.example.com",
    "founded": "2018",
    "about": "InnoTech Solutions builds cloud integrations for logistics.",
    "city": "New York",
    "state": "New York",
    "country": "United States",
    "phone": "+1 (212) 555-9000",
    "hrEmail": "careers@innotech.com",
    "supportEmail": "support@innotech.com",
    "linkedin": "https://linkedin.com/company/innotech",
    "twitter": "https://twitter.com/innotech",
    "facebook": "https://facebook.com/innotech",
    "instagram": "https://instagram.com/innotech",
    "github": "https://github.com/innotech"
  }
}
```

---

#### Update Company Settings
* **Method:** `PUT`
* **Path:** `/company/settings`
* **Auth:** **[Auth Required]** (Role must be `company`)
* **Request Body (JSON):** Same fields as [Get Company Settings Response](#get-company-settings).
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Company settings updated successfully",
  "data": null
}
```

---

#### Upload Company Logo
* **Method:** `POST`
* **Path:** `/company/logo`
* **Auth:** **[Auth Required]** (Role must be `company`)
* **Content-Type:** `multipart/form-data`
* **Form Parameters:**
  * `logo` (File): JPG, PNG, or SVG image file, maximum size **2MB**.
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Logo uploaded successfully",
  "data": {
    "logoUrl": "https://job-portal-backend-1-dv1h.onrender.com/uploads/logos/60d5ecb8629ef31a98e07b0/logo.png"
  }
}
```

---

#### Remove Company Logo
* **Method:** `DELETE`
* **Path:** `/company/logo`
* **Auth:** **[Auth Required]** (Role must be `company`)
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Logo removed successfully",
  "data": null
}
```

---

#### Get Own Company Public Profile
Get public-view preview of the employer's profile page, including their open jobs.
* **Method:** `GET`
* **Path:** `/company/profile`
* **Auth:** **[Auth Required]** (Role must be `company`)
* **Response (200 OK):** Same structure as [Get Public Company Profile](#get-public-company-profile).

---

### Company Dashboard

#### Get Company Dashboard Data
Fetch counters, posted jobs list, and recent applicants.
* **Method:** `GET`
* **Path:** `/company/dashboard`
* **Auth:** **[Auth Required]** (Role must be `company`)
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Company dashboard data fetched",
  "data": {
    "stats": {
      "activeJobs": 5,
      "totalApplicants": 48,
      "pendingReviews": 16,
      "shortlisted": 8
    },
    "recentJobs": [
      {
        "id": "60d5ecb8629ef31a98e07b5",
        "title": "Senior React Developer",
        "location": "New York, NY",
        "postedAt": "2026-08-20T10:00:00Z",
        "postedLabel": "5 days ago",
        "applicants": 12,
        "jobType": "full-time",
        "status": "active"
      }
    ],
    "recentApplicants": [
      {
        "id": "60d5ecb8629ef31a98e07aaa",
        "jobId": "60d5ecb8629ef31a98e07b5",
        "jobTitle": "Senior React Developer",
        "userId": "60d5ecb8629ef31a98e078a1",
        "seekerName": "John Doe",
        "seekerEmail": "john.doe@example.com",
        "status": "pending",
        "appliedAt": "2026-08-24T12:30:00Z"
      }
    ]
  }
}
```

---

## 6. 05 - Public Companies

### List Companies
Get a paginated directory of registered companies.
* **Method:** `GET`
* **Path:** `/companies`
* **Auth:** Public
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Companies fetched successfully",
  "data": {
    "items": [
      {
        "id": "60d5ecb8629ef31a98e07b0",
        "name": "InnoTech Solutions",
        "industry": "Information Technology",
        "logoUrl": "https://job-portal-backend-1-dv1h.onrender.com/uploads/logos/60d5ecb8629ef31a98e07b0/logo.png",
        "city": "New York",
        "country": "United States"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 1,
      "totalPages": 1
    }
  }
}
```

---

### Get Public Company Profile
Retrieve details of a company and its list of active job openings.
* **Method:** `GET`
* **Path:** `/companies/:id`
* **Auth:** Public
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Company profile fetched successfully",
  "data": {
    "id": "60d5ecb8629ef31a98e07b0",
    "name": "InnoTech Solutions",
    "industry": "Information Technology",
    "website": "https://innotech.example.com",
    "size": "500",
    "type": "private",
    "founded": "2018",
    "about": "InnoTech Solutions builds cloud integrations for logistics.",
    "logoUrl": "https://job-portal-backend-1-dv1h.onrender.com/uploads/logos/60d5ecb8629ef31a98e07b0/logo.png",
    "city": "New York",
    "state": "New York",
    "country": "United States",
    "linkedin": "https://linkedin.com/company/innotech",
    "twitter": "https://twitter.com/innotech",
    "facebook": "https://facebook.com/innotech",
    "instagram": "https://instagram.com/innotech",
    "github": "https://github.com/innotech",
    "openJobs": [
      {
        "id": "60d5ecb8629ef31a98e07b5",
        "title": "Senior React Developer",
        "company": "InnoTech Solutions",
        "companyId": "60d5ecb8629ef31a98e07b0",
        "location": "New York, NY",
        "postedAt": "2026-08-20T10:00:00Z",
        "postedLabel": "5 days ago",
        "category": "engineering",
        "tags": ["Full-time", "Hybrid", "Senior Level"],
        "salary": "$120k - $150k",
        "status": "active"
      }
    ]
  }
}
```
