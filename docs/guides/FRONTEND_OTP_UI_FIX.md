# Frontend Instructions: Handling OTP Request Loader & Errors

This document provides instructions for the frontend developer on how to integrate the Forgot Password and Reset Password APIs, specifically focusing on managing the **"Sending OTP..."** loader state and handling backend error responses.

---

## 1. Backend API Update
The Go backend has been updated to implement a strict **10-second timeout** for the SMTP connection. 
* Previously, if there was an SMTP config error or firewall block, the backend would hang indefinitely, keeping the frontend loader spinning forever.
* **Now:** If the email fails to send within 10 seconds, the backend will return a standard `500 Internal Server Error` containing a JSON error message.

---

## 2. Frontend State Management Rules

To ensure a smooth user experience, the frontend **must**:
1. **Enable the Loader:** Set `loading = true` when the user clicks the "Send OTP" button.
2. **Disable the Button:** Set `disabled = true` on the button during the request to prevent duplicate submissions.
3. **Turn Off the Loader on Failure:** In the `catch` block of your API request, **you must set `loading = false`** so the user is not stuck on a loading screen.
4. **Display the Error Message:** Show the error message returned from the backend (e.g., in a toast notification).

---

## 3. Example Implementation (React / Axios)

Here is how the React component code should look to correctly handle the loader:

```jsx
import React, { useState } from 'react';
import axios from 'axios';

function ForgotPassword() {
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [step, setStep] = useState(1); // 1 = Email input, 2 = OTP + New Password

  const handleSendOTP = async (e) => {
    e.preventDefault();
    if (!email) return alert("Email is required");

    // 1. Turn on loader
    setLoading(true);

    try {
      // 2. Make API request to backend (Times out after 10s if SMTP fails)
      const response = await axios.post('http://localhost:8080/api/v1/auth/forgot-password', {
        email: email
      });

      // 3. Success: Show success message and proceed to next step
      alert(response.data.message || "OTP sent successfully!");
      setStep(2);
    } catch (error) {
      // 4. Failure: Extract and show backend error message
      const errorMsg = error.response?.data?.message || "Failed to send OTP. Please try again.";
      alert(errorMsg); 
    } finally {
      // 5. Critical: Turn off loader regardless of success or failure
      setLoading(false);
    }
  };

  return (
    <div>
      {step === 1 ? (
        <form onSubmit={handleSendOTP}>
          <input 
            type="email" 
            placeholder="Enter your email" 
            value={email} 
            onChange={(e) => setEmail(e.target.value)}
            disabled={loading}
          />
          <button type="submit" disabled={loading}>
            {loading ? "Sending OTP..." : "Send OTP"}
          </button>
        </form>
      ) : (
        <div>
          {/* Step 2 Form (Reset Password Form) goes here */}
          <p>Please enter the OTP sent to your email.</p>
        </div>
      )}
    </div>
  );
}

export default ForgotPassword;
```

## 4. Key Takeaways for the Frontend
* **Always use a `finally {}` block** or ensure the loader state is set to `false` in both the `.then()` and `.catch()` callbacks of your promise.
* Expect error HTTP status codes `404` (User not found), `400` (Bad validation), and `500` (SMTP connection failure). All of them return a JSON object with a `message` field.
