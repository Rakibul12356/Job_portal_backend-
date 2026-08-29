# Backend Guide: Forgot & Reset Password SMTP Integration

This guide explains the backend integration workflow for the **Forgot Password** and **Reset Password** endpoints. It specifically addresses why the `/auth/forgot-password` endpoint might hang (causing the frontend to stay in a "Sending OTP..." loading state) and how to configure a resilient Node.js email transporter.

---

## 1. Overview & Root Cause of the Hang

The frontend application successfully sends a `POST` request to `/api/v1/auth/forgot-password` with `{ email }`.
* **If the email does not exist:** The backend database check completes quickly, returning `404 Not Found` or `{ success: false, message: 'User not found' }`, which resolves the loading state immediately.
* **If the email exists:** The backend attempts to send an email with the OTP code using a library like **Nodemailer**. 

If the SMTP configuration is wrong, missing, or blocked by a firewall, the Nodemailer code will try to establish a socket connection to the mail server. By default, Node.js sockets do not timeout quickly, causing the endpoint response to hang indefinitely. This keeps the client-side HTTP request open, leaving the user stuck on the **"Sending OTP..."** loader.

---

## 2. Recommended Database Schema (`models/Otp.js`)

You need a temporary storage schema for OTP codes that automatically expires after 5 or 10 minutes using MongoDB TTL indexes.

```javascript
import mongoose from 'mongoose';

const otpSchema = new mongoose.Schema(
  {
    email: {
      type: String,
      required: true,
      index: true,
    },
    otp: {
      type: String,
      required: true,
    },
    expiresAt: {
      type: Date,
      required: true,
      default: () => new Date(Date.now() + 10 * 60 * 1000), // Expires in 10 minutes
    },
  },
  { timestamps: true }
);

// Create TTL index (automatically deletes document when expiresAt is reached)
otpSchema.index({ expiresAt: 1 }, { expireAfterSeconds: 0 });

export const Otp = mongoose.model('Otp', otpSchema);
```

---

## 3. Endpoints & Controllers Setup

### 3.1 Request OTP Endpoint (`POST /auth/forgot-password`)
Generates a 6-digit code, saves it to the database, and sends it via email.

```javascript
import crypto from 'crypto';
import { User } from '../models/User.js';
import { Otp } from '../models/Otp.js';
import { sendOTPEmail } from '../utils/emailService.js';

export const forgotPassword = async (req, res) => {
  try {
    const { email } = req.body;

    if (!email) {
      return res.status(400).json({ success: false, message: 'Email is required' });
    }

    // 1. Check if user exists
    const user = await User.findOne({ email });
    if (!user) {
      return res.status(404).json({ success: false, message: 'User not found with this email' });
    }

    // 2. Generate a random 6-digit numeric OTP
    const otp = Math.floor(100000 + Math.random() * 900000).toString();

    // 3. Save OTP to DB (Overwrites older OTPs for this email if any exist)
    await Otp.findOneAndDelete({ email });
    await Otp.create({ email, otp });

    // 4. Send the Email
    const emailSent = await sendOTPEmail(email, otp);
    if (!emailSent) {
      return res.status(500).json({
        success: false,
        message: 'Unable to send OTP at this time. Please check backend SMTP configs.',
      });
    }

    return res.status(200).json({
      success: true,
      message: 'OTP sent to your email successfully',
    });
  } catch (error) {
    return res.status(500).json({
      success: false,
      message: error.message || 'Error requesting password reset',
    });
  }
};
```

---

### 3.2 Reset Password Endpoint (`POST /auth/reset-password`)
Verifies the OTP, hashes the new password, updates the user document, and cleans up the OTP record.

```javascript
import bcrypt from 'bcryptjs';
import { User } from '../models/User.js';
import { Otp } from '../models/Otp.js';

export const resetPassword = async (req, res) => {
  try {
    const { email, otp, password, confirmPassword } = req.body;

    if (!email || !otp || !password || !confirmPassword) {
      return res.status(400).json({ success: false, message: 'All fields are required' });
    }

    if (password !== confirmPassword) {
      return res.status(400).json({ success: false, message: 'Passwords do not match' });
    }

    if (password.length < 8) {
      return res.status(400).json({ success: false, message: 'Password must be at least 8 characters long' });
    }

    // 1. Find and verify the OTP
    const otpRecord = await Otp.findOne({ email, otp });
    if (!otpRecord) {
      return res.status(400).json({ success: false, message: 'Invalid or expired OTP code' });
    }

    // 2. Hash new password
    const salt = await bcrypt.genSalt(10);
    const hashedPassword = await bcrypt.hash(password, salt);

    // 3. Update user password
    const user = await User.findOneAndUpdate(
      { email },
      { password: hashedPassword },
      { new: true }
    );

    if (!user) {
      return res.status(404).json({ success: false, message: 'User not found' });
    }

    // 4. Clean up OTP record
    await Otp.deleteOne({ _id: otpRecord._id });

    return res.status(200).json({
      success: true,
      message: 'Password reset successfully',
    });
  } catch (error) {
    return res.status(500).json({
      success: false,
      message: error.message || 'Error resetting password',
    });
  }
};
```

---

## 4. Configuring Nodemailer Service with Connection Timeouts

To prevent the backend from hanging forever on bad SMTP credentials, you **MUST** configure `connectionTimeout` and `socketTimeout` limits in your transporter options.

Create `utils/emailService.js`:

```javascript
import nodemailer from 'nodemailer';

// Configure Transporter with timeouts
const transporter = nodemailer.createTransport({
  host: process.env.SMTP_HOST || 'smtp.gmail.com',
  port: parseInt(process.env.SMTP_PORT || '465'),
  secure: process.env.SMTP_SECURE === 'true' || true, // true for port 465, false for 587
  auth: {
    user: process.env.SMTP_USER, // e.g. careers@company.com or Gmail address
    pass: process.env.SMTP_PASS, // App password (not regular Gmail password)
  },
  connectionTimeout: 10000, // 10 seconds timeout to establish connection
  socketTimeout: 10000,     // 10 seconds timeout for data transmission
});

export const sendOTPEmail = async (email, otp) => {
  try {
    const mailOptions = {
      from: `"Online Job Portal" <${process.env.SMTP_USER}>`,
      to: email,
      subject: 'Password Reset OTP Code',
      html: `
        <div style="font-family: Arial, sans-serif; padding: 20px; border: 1px solid #eee; border-radius: 8px; max-width: 600px; margin: 0 auto;">
          <h2 style="color: #0f172a; text-align: center;">Reset Your Password</h2>
          <p>Hello,</p>
          <p>You requested a password reset for your account on Online Job Portal. Use the following OTP code to proceed:</p>
          <div style="background-color: #f1f5f9; text-align: center; padding: 15px; margin: 20px 0; border-radius: 8px;">
            <span style="font-size: 28px; font-weight: bold; letter-spacing: 5px; color: #0f172a;">${otp}</span>
          </div>
          <p style="color: #ef4444; font-size: 14px;"><strong>Note:</strong> This OTP is valid for 10 minutes only. Do not share this code with anyone.</p>
          <p>If you did not request this, please ignore this email.</p>
          <hr style="border: 0; border-top: 1px solid #eee; margin: 20px 0;" />
          <p style="font-size: 12px; color: #64748b; text-align: center;">Online Job Portal &copy; 2026. All rights reserved.</p>
        </div>
      `,
    };

    await transporter.sendMail(mailOptions);
    return true;
  } catch (error) {
    console.error('Error sending SMTP email:', error);
    return false; // Returns false so controller can send 500 error instead of hanging
  }
};
```

---

## 5. Troubleshooting Nodemailer on Render/Heroku

1. **Port 25 is Blocked:** Hosting services (like Render, Heroku, AWS EC2) block SMTP Port `25` by default to prevent spam. Use port `465` (SSL) or `587` (TLS/STARTTLS).
2. **Gmail App Password:** If using a Gmail account, you **cannot** use your regular Google Password. You must enable Two-Factor Authentication (2FA) and generate an **App Password** from Google Account Settings > Security.
3. **Configure Environment Variables:** Make sure you set these on your Render Environment Dashboard:
   - `SMTP_HOST=smtp.gmail.com`
   - `SMTP_PORT=465`
   - `SMTP_SECURE=true`
   - `SMTP_USER=your_email@gmail.com`
   - `SMTP_PASS=your_app_password`
