# Frontend Integration Guide: Company Profile & Job Details

> **Status:** ✅ **Backend Implementation Completed**  
> All company profile information and logos are now fully integrated and populated across all Job endpoints.

---

## 1. Overview of Changes

The backend has been updated to automatically populate full company information into the `Job` response object. 

When you call `GET /api/v1/jobs/:id` or `GET /api/v1/jobs`, the response now includes:
1. `companyInfo` object with complete company details (Industry, About, Website, Location, Employees, Founded, Logo).
2. Root-level `logoUrl` on the job object for quick access.

---

## 2. API Endpoints Reference

### 2.1 Get Job Details (with Company Information)
Fetch single job details including complete company information for the Job Details page sidebar.

* **Method:** `GET`
* **Path:** `/api/v1/jobs/:id`
* **Auth:** Public
* **Sample Response (200 OK):**
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
    "requirements": "Requirements:\n- 5+ years of experience with React\n- TypeScript expertise",
    "benefits": "Benefits:\n- Health, dental, vision\n- Remote budget\n- 401(k) matching",
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
    "skills": ["React", "TypeScript", "Redux", "Tailwind CSS"],
    "vacancies": 2,
    "companyInfo": {
      "companyName": "InnoTech Solutions",
      "industry": "Information Technology",
      "about": "InnoTech Solutions builds cloud integrations for modern logistics and enterprise operations.",
      "website": "https://innotech.example.com",
      "location": "New York, NY, USA",
      "employees": "50-100 employees",
      "companySize": "50-100 employees",
      "companyType": "private",
      "founded": "Founded in 2018",
      "logoUrl": "https://job-portal-backend-1-dv1h.onrender.com/uploads/logos/60d5ecb8629ef31a98e07b0/logo.png"
    }
  }
}
```

---

### 2.2 Update Company Settings (Employer Dashboard)
When employer edits and saves their profile in **Company Dashboard > Company Settings** (`/company/settings`).

* **Method:** `PUT`
* **Path:** `/api/v1/company/settings`
* **Auth:** Required (`Bearer <token>`, Role: `company`)
* **Request Body (JSON):**
```json
{
  "companyName": "InnoTech Solutions",
  "industry": "Information Technology",
  "companySize": "50-100 employees",
  "companyType": "private",
  "website": "https://innotech.example.com",
  "founded": "2018",
  "about": "InnoTech Solutions builds cloud integrations for modern logistics and enterprise operations.",
  "city": "New York",
  "state": "NY",
  "country": "USA",
  "phone": "+1 (212) 555-9000",
  "hrEmail": "careers@innotech.com",
  "supportEmail": "support@innotech.com",
  "linkedin": "https://linkedin.com/company/innotech",
  "twitter": "https://twitter.com/innotech",
  "facebook": "https://facebook.com/innotech",
  "instagram": "https://instagram.com/innotech",
  "github": "https://github.com/innotech"
}
```
* **Sample Response (200 OK):**
```json
{
  "success": true,
  "message": "Company settings updated successfully",
  "data": {
    "companyName": "InnoTech Solutions",
    "industry": "Information Technology",
    "companySize": "50-100 employees",
    "companyType": "private",
    "website": "https://innotech.example.com",
    "founded": "2018",
    "about": "InnoTech Solutions builds cloud integrations for modern logistics and enterprise operations.",
    "city": "New York",
    "state": "NY",
    "country": "USA",
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

### 2.3 Get Current Company Settings
To prefill the **Company Settings** form when employer opens `/company/settings`.

* **Method:** `GET`
* **Path:** `/api/v1/company/settings`
* **Auth:** Required (`Bearer <token>`, Role: `company`)
* **Sample Response (200 OK):** Same schema as update settings response.

---

### 2.4 Upload & Delete Company Logo

#### Upload Logo:
* **Method:** `POST`
* **Path:** `/api/v1/company/logo`
* **Auth:** Required (`Bearer <token>`, Role: `company`)
* **Content-Type:** `multipart/form-data`
* **Form Field:** `logo` (File, max 2MB, formats: png, jpg, jpeg, svg, webp)
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

#### Delete Logo:
* **Method:** `DELETE`
* **Path:** `/api/v1/company/logo`
* **Auth:** Required (`Bearer <token>`, Role: `company`)
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Logo removed successfully",
  "data": null
}
```

---

## 3. TypeScript Interfaces for Frontend

You can add or update these TypeScript types in your frontend codebase:

```typescript
export interface CompanyInfo {
  companyName?: string;
  name?: string;
  industry: string;
  about: string;
  website: string;
  location: string;
  employees: string;
  companySize?: string;
  companyType?: string;
  founded: string;
  logoUrl: string;
}

export interface JobDetails {
  id: string;
  title: string;
  company: string;
  companyId: string;
  logoUrl?: string;
  location: string;
  postedAt: string;
  postedLabel: string;
  category: string;
  description: string;
  requirements?: string;
  benefits?: string;
  tags: string[];
  salary: string;
  salaryMin?: number;
  salaryMax?: number;
  salaryPeriod: string;
  applicants: number;
  jobType: string;
  workMode: string;
  experienceLevel: string;
  deadline: string;
  status: 'active' | 'draft' | 'closed' | 'expiring_soon';
  skills: string[];
  vacancies: number;
  companyInfo?: CompanyInfo;
}
```

---

## 4. Frontend Component Usage Example (React / Next.js)

```tsx
import React from 'react';
import { Building, MapPin, Users, Globe, Calendar } from 'lucide-react';

interface JobDetailsProps {
  job: JobDetails;
}

export const CompanyInfoSidebar: React.FC<JobDetailsProps> = ({ job }) => {
  const info = job.companyInfo;

  return (
    <div className="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-100 dark:border-gray-700">
      {/* Company Header */}
      <div className="flex items-center gap-4 mb-5">
        {info?.logoUrl || job.logoUrl ? (
          <img
            src={info?.logoUrl || job.logoUrl}
            alt={job.company}
            className="w-14 h-14 rounded-lg object-cover border"
          />
        ) : (
          <div className="w-14 h-14 rounded-lg bg-blue-50 dark:bg-blue-900/40 text-blue-600 flex items-center justify-center font-bold text-xl">
            {job.company.charAt(0)}
          </div>
        )}
        <div>
          <h3 className="font-semibold text-lg text-gray-900 dark:text-white">
            {job.company}
          </h3>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            {info?.industry || 'Technology & Software'}
          </p>
        </div>
      </div>

      {/* About Company */}
      {info?.about && (
        <p className="text-sm text-gray-600 dark:text-gray-300 mb-6 leading-relaxed">
          {info.about}
        </p>
      )}

      {/* Meta Info List */}
      <div className="space-y-3 text-sm text-gray-600 dark:text-gray-300">
        {info?.location && (
          <div className="flex items-center gap-3">
            <MapPin className="w-4 h-4 text-gray-400" />
            <span>{info.location}</span>
          </div>
        )}

        {info?.employees && (
          <div className="flex items-center gap-3">
            <Users className="w-4 h-4 text-gray-400" />
            <span>{info.employees}</span>
          </div>
        )}

        {info?.founded && (
          <div className="flex items-center gap-3">
            <Calendar className="w-4 h-4 text-gray-400" />
            <span>{info.founded}</span>
          </div>
        )}

        {info?.website && (
          <div className="flex items-center gap-3">
            <Globe className="w-4 h-4 text-gray-400" />
            <a
              href={info.website.startsWith('http') ? info.website : `https://${info.website}`}
              target="_blank"
              rel="noreferrer"
              className="text-blue-600 dark:text-blue-400 hover:underline break-all"
            >
              {info.website.replace(/^https?:\/\//, '')}
            </a>
          </div>
        )}
      </div>
    </div>
  );
};
```

---

## 5. Summary Field Mapping

| UI Field | Backend Property | Fallback Value |
| :--- | :--- | :--- |
| **Company Name** | `data.company` or `data.companyInfo.companyName` | `"Unknown Company"` |
| **Logo** | `data.companyInfo.logoUrl` or `data.logoUrl` | Default placeholder / Initial letter |
| **Industry** | `data.companyInfo.industry` | `"Technology & Software"` |
| **About / Bio** | `data.companyInfo.about` | `"No company bio provided yet."` |
| **Location** | `data.companyInfo.location` | Job location |
| **Employees / Size**| `data.companyInfo.employees` or `data.companyInfo.companySize` | `"1-50 employees"` |
| **Founded** | `data.companyInfo.founded` | `""` |
| **Website** | `data.companyInfo.website` | `""` |
