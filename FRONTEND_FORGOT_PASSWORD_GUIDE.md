# Frontend Integration Guide: Forgot & Reset Password

This guide explains how to implement the "Forgot Password" and "Reset Password" flow in the frontend application using the newly created backend REST API endpoints.

---

## 1. API Endpoints

### Step 1: Request Password Reset (Send OTP)
* **URL:** `/api/v1/auth/forgot-password`
* **Method:** `POST`
* **Headers:** `Content-Type: application/json`
* **Body:**
  ```json
  {
    "email": "user@example.com"
  }
  ```
* **Success Response (200 OK):**
  ```json
  {
    "status": 200,
    "message": "OTP sent to your email successfully"
  }
  ```

---

### Step 2: Verify OTP and Set New Password
* **URL:** `/api/v1/auth/reset-password`
* **Method:** `POST`
* **Headers:** `Content-Type: application/json`
* **Body:**
  ```json
  {
    "email": "user@example.com",
    "otp": "123456",
    "password": "newSuperPassword123",
    "confirmPassword": "newSuperPassword123"
  }
  ```
* **Success Response (200 OK):**
  ```json
  {
    "status": 200,
    "message": "Password reset successfully"
  }
  ```

---

## 2. Frontend User Flow & State Management

A clean user experience can be built with a **2-Step View/State** on a single "Forgot Password" page:

### Step 1 View: Email Input
* User enters their email and clicks **"Send OTP"**.
* Disable the button and show a loading spinner while the API call is in progress.
* If successful, transition the screen to **Step 2 View** (OTP and Password form) and save the email in state.

### Step 2 View: OTP & Password Reset
* Show fields for:
  1. **OTP Code** (6-digit text/number input).
  2. **New Password** (Password input).
  3. **Confirm Password** (Password input).
* User clicks **"Reset Password"**.
* On success, show a success toast or notification and redirect the user to the `/login` page.

---

## 3. Example Implementation (React/Vue/Axios)

Here is a sample JS integration code:

```javascript
import axios from 'axios';

// API Base configuration
const API = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
});

// Step 1: Request OTP
async function handleRequestOTP(email) {
  try {
    const response = await API.post('/auth/forgot-password', { email });
    alert(response.data.message); // "OTP sent to your email successfully"
    return true; // Proceed to step 2 view
  } catch (error) {
    const errorMsg = error.response?.data?.message || "Failed to send OTP";
    alert(errorMsg);
    return false;
  }
}

// Step 2: Reset Password
async function handleResetPassword(email, otp, password, confirmPassword) {
  try {
    const response = await API.post('/auth/reset-password', {
      email,
      otp,
      password,
      confirmPassword
    });
    alert(response.data.message); // "Password reset successfully"
    // Redirect to login page
  } catch (error) {
    const errorMsg = error.response?.data?.message || "Failed to reset password";
    alert(errorMsg);
  }
}
```

---

## 4. Best Practices for Frontend UI
1. **OTP Formatting:** Make the OTP field a simple 6-digit text input with `maxLength={6}`.
2. **Resend Timer:** Implement a 60-second cooldown timer for the "Resend OTP" button to prevent spamming the backend.
3. **Password Validation:** Ensure the new password matches basic security constraints (min 8 chars) on the frontend before submitting to the server.
