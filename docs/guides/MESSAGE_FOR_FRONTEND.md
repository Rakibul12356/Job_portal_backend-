# Messages for Frontend Team

You can copy and send one of the messages below to your frontend team to update them about the Forgot Password SMTP timeout fix and loader state handling.

---

## Option 1: English (Professional / Slack / Discord)

```text
Hey team,

I have fixed the forgot-password API hang issue on the backend. 

What was the issue:
Render was blocking outbound traffic on port 587 (SMTP), causing the mailer connection to hang indefinitely on production. This is why the frontend stayed stuck on the "Sending OTP..." loading state.

What has been updated:
1. Backend Timeout & Fix: I have added a strict 10-second connection timeout to the Go backend and migrated the SMTP port to 465 (SSL), which is allowed by Render. Emails are sending successfully now.
2. Loader / Error Handling: If the SMTP connection ever fails or times out in the future, the backend will now return a 500 Internal Server Error instead of hanging. 

Action item for Frontend:
Please check the updated integration guide: "FRONTEND_OTP_UI_FIX.md" in the repository. Make sure you handle the API errors in a catch (or finally) block to turn off the loading spinner/button state (e.g., setLoading(false)) if the API returns a 500 or 404 error.

Let me know if you face any issues!
```

---

## Option 2: Benglish (Informal / Messenger / WhatsApp)

```text
Hey guys,

Forgot Password-এর ওটিপি সেন্ড হওয়ার সময় যে লোডার আটকে থাকত, ঐ হ্যাং হওয়ার প্রবলেমটা ব্যাকএন্ড থেকে সলভ করা হয়েছে।

কী প্রবলেম ছিল:
রেন্ডার (Render) হোস্টিং থেকে ডিফল্টভাবে SMTP পোর্ট ৫৮৭ ব্লক থাকায় কানেকশন হচ্ছিল না এবং টাইমআউট না থাকায় রিকোয়েস্টটি ঝুলে থাকত। যার জন্য ফ্রন্টএন্ডে লোডার অফ হতো না।

কী আপডেট করা হয়েছে:
১. ব্যাকএন্ডে ১০ সেকেন্ডের কানেকশন টাইমআউট সেট করা হয়েছে এবং পোর্ট পরিবর্তন করে ৪৬৫ (SSL) করা হয়েছে। এখন ওটিপি মেইল সাকসেসফুলি যাচ্ছে।
২. যদি পরবর্তীতে কখনো SMTP কানেকশন ফেইল হয়, তবে ব্যাকএন্ড এখন হ্যাং না করে সাথে সাথে 500 Internal Server Error রেসপন্স ব্যাক করবে।

ফ্রন্টএন্ডের জন্য রিকোয়েস্ট:
রেপোতে নতুন যুক্ত করা "FRONTEND_OTP_UI_FIX.md" ফাইলটি একটু দেখে নিও। ওটিপি রিকোয়েস্ট ফেইল বা এরর আসলেও যাতে লোডিং স্পিনার অফ হয়ে যায় (যেমন: setLoading(false)), সেটার জন্য try-catch বা finally ব্লকে হ্যান্ডলিং অ্যাড করে নিও।

কোনো সমস্যা হলে জানিও!
```
