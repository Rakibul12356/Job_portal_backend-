# Frontend Integration Guide: Candidate Interview Scheduling

This guide details the API updates, request/response formats, and validation rules required for the frontend team to integrate the new candidate interview scheduling feature.

---

## 1. Updating Applicant Status (Schedule Interview)

When an employer/company schedules an interview for a candidate, you must call the existing status update endpoint but supply additional properties.

* **Endpoint:** `PATCH /api/v1/company/applicants/:id/status`
* **Headers:** 
  * `Authorization: Bearer <JWT_ACCESS_TOKEN>`
  * `Content-Type: application/json`

* **Request Body Payload (JSON):**
  If `status` is set to `"interviewed"`, you must provide the scheduling details in the request body.

```json
{
  "status": "interviewed",
  "interviewDate": "2026-09-15",
  "interviewTime": "14:30",
  "notes": "The interview will be conducted via Google Meet. Link: https://meet.google.com/abc-defg-hij. Please bring your portfolio."
}
```

### Validation Rules (Enforced by Backend)
1. **Status:** Must be `"interviewed"`.
2. **interviewDate:** 
   * **Required** when status is `interviewed`.
   * Must be in `YYYY-MM-DD` format.
   * Must be **today or in the future** (past dates are rejected with a Validation Error).
3. **interviewTime:** 
   * **Required** when status is `interviewed`.
   * Must be in `HH:MM` format (24-hour representation).
4. **notes:** 
   * **Optional** string field containing descriptions, links, or instructions.

---

## 2. Rendering Interview Scheduling Details

The interview scheduling properties are returned in both Seeker-facing and Company-facing application detail and list endpoints. You should render these details in the UI (e.g., in dashboards, application detail cards, or event timelines) if they are present.

### Exposed Fields in Responses
The following fields are conditionally returned in the response when an interview is scheduled:
* `interviewDate`: Date string (e.g. `"2026-09-15"`)
* `interviewTime`: Time string (e.g. `"14:30"`)
* `interviewNotes`: Text description / links (e.g. `"Google Meet Link: ..."`)

### Affected Endpoints

#### A. Seeker Dashboards and Applications:
* **Endpoints:**
  * `GET /api/v1/applications/me` (List Seeker Applications)
  * `GET /api/v1/applications/:id` (Get Seeker Application Details)
* **Response Payload Example:**
```json
{
  "id": "64e0abc...",
  "jobId": "64e0def...",
  "jobTitle": "Frontend Engineer",
  "company": "Tech Solutions Inc.",
  "status": "interviewed",
  "appliedAt": "2026-08-25T10:00:00Z",
  "interviewDate": "2026-09-15",
  "interviewTime": "14:30",
  "interviewNotes": "The interview will be conducted via Google Meet. Link: https://meet.google.com/abc-defg-hij. Please bring your portfolio."
}
```

#### B. Company Applicant Review Dashboards:
* **Endpoints:**
  * `GET /api/v1/company/applicants` (List Company Applicants)
  * `GET /api/v1/company/applicants/:id` (Get Company Applicant Details)
* **Response Payload Example:**
```json
{
  "id": "64e0abc...",
  "jobId": "64e0def...",
  "jobTitle": "Frontend Engineer",
  "seekerName": "John Doe",
  "status": "interviewed",
  "interviewDate": "2026-09-15",
  "interviewTime": "14:30",
  "interviewNotes": "The interview will be conducted via Google Meet. Link: https://meet.google.com/abc-defg-hij. Please bring your portfolio."
}
```

---

## 3. In-App Notifications

When the status is updated to `"interviewed"`, a real-time notification is automatically generated for the job seeker. 

* **Notification List Endpoint:** `GET /api/v1/notifications`
* **Trigger Details:**
  * **type:** `"interview"`
  * **title:** `"Interview Scheduled!"`
  * **message:** `"Congratulations! Your interview for the role of [Job Title] at [Company Name] has been scheduled on [Interview Date] at [Interview Time]."`

> **Tip:** You can check the `type` attribute in the notifications array. If the type is `"interview"`, you can render a custom icon (such as a calendar 📅 or clock 🕒) to improve the user experience.
