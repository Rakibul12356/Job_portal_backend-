# Backend Specification: Candidate Interview Scheduling

This specification outlines the backend updates required to support scheduling interviews for job seekers, sending in-app notifications, and sending email notifications when an employer schedules an interview.

---

## 1. API Changes

### Update Endpoint: Update Applicant Status
* **Endpoint:** `PATCH /api/v1/company/applicants/:id/status`
* **Controller / Handler Context:** Currently, this endpoint updates the status of an applicant (e.g., to `shortlisted`, `rejected`, `interviewed`). It must be updated to accept additional optional fields when the status is set to `interviewed`.
* **Request Body Payload (JSON):**
  ```json
  {
    "status": "interviewed",
    "interviewDate": "2026-09-15",
    "interviewTime": "14:30",
    "notes": "The interview will be conducted via Google Meet. Link: https://meet.google.com/abc-defg-hij. Please bring your portfolio."
  }
  ```

#### Validation Rules:
* If `status` is `"interviewed"`, then:
  * `interviewDate` is **required** and must be a valid date in `YYYY-MM-DD` format (should be in the future).
  * `interviewTime` is **required** and must be a valid time in `HH:MM` format.
  * `notes` is **optional** (string).

---

## 2. Database Schema Updates

We need to store the interview schedule details in the database. 

### Option A: Add Columns to `applications` (or `applicants`) Table
Add the following columns to the existing applications table:
* `interview_date`: `DATE` (or `VARCHAR/TEXT` depending on the database driver)
* `interview_time`: `TIME` (or `VARCHAR/TEXT`)
* `interview_notes`: `TEXT` (nullable)

### Option B: Create a `job_interviews` Table (Recommended if supporting reschedule history)
```sql
CREATE TABLE job_interviews (
    id VARCHAR(36) PRIMARY KEY,
    application_id VARCHAR(36) NOT NULL,
    interview_date DATE NOT NULL,
    interview_time TIME NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
);
```

---

## 3. Email and Notification Triggers

When an application's status is successfully updated to `"interviewed"`:

### 3.1 In-App Notification Trigger
An in-app notification record must be created for the candidate (Job Seeker):
* **Recipient:** Candidate User ID (`seekerId` / `userId` associated with the application).
* **Title:** "Interview Scheduled!"
* **Message:** "Congratulations! Your interview for the role of **[Job Title]** at **[Company Name]** has been scheduled on **[Interview Date]** at **[Interview Time]**."
* **Type:** `interview` or `general` (should render in the seeker's notification bell).

### 3.2 Email Notification Trigger
Send an email to the candidate's email address containing the interview details:
* **Subject:** "Interview Scheduled - [Company Name]"
* **Template Body (HTML):**
  ```html
  <p>Dear [Candidate Name],</p>
  <p>Your application for the position of <strong>[Job Title]</strong> at <strong>[Company Name]</strong> has been reviewed, and we would like to invite you for an interview.</p>
  <p><strong>Interview Details:</strong></p>
  <ul>
    <li><strong>Date:</strong> [Interview Date]</li>
    <li><strong>Time:</strong> [Interview Time]</li>
    <li><strong>Format / Notes:</strong> [Interview Notes / Instructions]</li>
  </ul>
  <p>Best regards,<br/>[Company Name] Hiring Team</p>
  ```
