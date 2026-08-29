# Company Profile Update & Settings Integration Guide

> **Overview:**  
> Employers can update **all company information, branding, locations, contacts, and social links** at any time.  
> The **Account Login Email** is protected and read-only for security, while public contact emails (`hrEmail`, `supportEmail`) are fully editable.

---

## 1. Quick Summary of Endpoints

| Purpose | Method | Path | Auth |
| :--- | :--- | :--- | :--- |
| **Get Settings Form Data** | `GET` | `/api/v1/company/settings` | `Bearer <token>` (`company` role) |
| **Update Company Profile/Settings** | `PUT` | `/api/v1/company/settings` *(or `/api/v1/company/profile`)* | `Bearer <token>` (`company` role) |
| **Upload Company Logo** | `POST` | `/api/v1/company/logo` | `Bearer <token>` (`company` role) |
| **Remove Company Logo** | `DELETE` | `/api/v1/company/logo` | `Bearer <token>` (`company` role) |
| **View Own Public Profile** | `GET` | `/api/v1/company/profile` | `Bearer <token>` (`company` role) |

---

## 2. Editable vs Read-Only Fields

### ✅ Editable Fields (Can be updated at any time):
1. **Company Name:** `companyName` (or `name`)
2. **Industry:** `industry` (e.g. `"Information Technology"`, `"Fintech"`, `"Healthcare"`)
3. **Company Size / Employees:** `companySize` (or `size`, e.g. `"50-100 employees"`, `"500+"`)
4. **Company Type:** `companyType` (or `type`, e.g. `"private"`, `"public"`, `"startup"`, `"non-profit"`)
5. **Website:** `website` (e.g. `"https://example.com"`)
6. **Founded Year:** `founded` (e.g. `"2018"`)
7. **About / Description:** `about` (or `description`)
8. **Location Details:**
   - `city` (e.g. `"Dhaka"`, `"New York"`)
   - `state` (e.g. `"Dhaka"`, `"NY"`)
   - `country` (e.g. `"Bangladesh"`, `"USA"`)
9. **Public Contacts:**
   - `phone` (e.g. `"+8801700000000"`)
   - `hrEmail` (e.g. `"hr@company.com"`)
   - `supportEmail` (e.g. `"support@company.com"`)
10. **Social Links:**
    - `linkedin`
    - `twitter`
    - `facebook`
    - `instagram`
    - `github`
11. **Logo Image:** via `POST /api/v1/company/logo`

### 🔒 Read-Only Field:
* **Account Login Email:** `accountEmail` (Displayed in the settings form as a disabled/read-only field to prevent unauthorized account takeover).

---

## 3. Detailed API Documentation

### 3.1 Fetch Current Company Profile & Settings
Call this endpoint on page load when the employer navigates to the **Company Settings / Profile Edit** page.

* **Method:** `GET`
* **Path:** `/api/v1/company/settings`
* **Headers:** `Authorization: Bearer <accessToken>`
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Company settings fetched",
  "data": {
    "id": "60d5ecb8629ef31a98e078b0",
    "companyName": "InnoTech Solutions",
    "accountEmail": "employer@innotech.com",
    "industry": "Information Technology",
    "companySize": "50-100 employees",
    "companyType": "private",
    "website": "https://innotech.example.com",
    "founded": "2018",
    "about": "InnoTech Solutions is a cloud integration company building enterprise logistics tools.",
    "logoUrl": "https://job-portal-backend-1-dv1h.onrender.com/uploads/logos/60d5ecb8629ef31a98e078b0/logo.png",
    "city": "Dhaka",
    "state": "Dhaka",
    "country": "Bangladesh",
    "phone": "+8801700000000",
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

### 3.2 Update Company Profile & Settings
Submit this request when the employer clicks **Save Changes** in the form.

* **Method:** `PUT`
* **Path:** `/api/v1/company/settings` *(or `/api/v1/company/profile`)*
* **Headers:** 
  - `Authorization: Bearer <accessToken>`
  - `Content-Type: application/json`
* **Request Body (JSON):**
```json
{
  "companyName": "InnoTech Global Ltd",
  "industry": "Software & AI Solutions",
  "companySize": "100-250 employees",
  "companyType": "private",
  "website": "https://innotech-global.com",
  "founded": "2019",
  "about": "We empower logistics companies with intelligent automation and API integrations.",
  "city": "Dhaka",
  "state": "Dhaka Division",
  "country": "Bangladesh",
  "phone": "+8801811122233",
  "hrEmail": "recruiting@innotech-global.com",
  "supportEmail": "help@innotech-global.com",
  "linkedin": "https://linkedin.com/company/innotech-global",
  "twitter": "https://twitter.com/innotech_ai",
  "facebook": "https://facebook.com/innotech.global",
  "instagram": "https://instagram.com/innotech.life",
  "github": "https://github.com/innotech-global"
}
```
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Company settings updated successfully",
  "data": {
    "id": "60d5ecb8629ef31a98e078b0",
    "companyName": "InnoTech Global Ltd",
    "accountEmail": "employer@innotech.com",
    "industry": "Software & AI Solutions",
    "companySize": "100-250 employees",
    "companyType": "private",
    "website": "https://innotech-global.com",
    "founded": "2019",
    "about": "We empower logistics companies with intelligent automation and API integrations.",
    "logoUrl": "https://job-portal-backend-1-dv1h.onrender.com/uploads/logos/60d5ecb8629ef31a98e078b0/logo.png",
    "city": "Dhaka",
    "state": "Dhaka Division",
    "country": "Bangladesh",
    "phone": "+8801811122233",
    "hrEmail": "recruiting@innotech-global.com",
    "supportEmail": "help@innotech-global.com",
    "linkedin": "https://linkedin.com/company/innotech-global",
    "twitter": "https://twitter.com/innotech_ai",
    "facebook": "https://facebook.com/innotech.global",
    "instagram": "https://instagram.com/innotech.life",
    "github": "https://github.com/innotech-global"
  }
}
```

---

### 3.3 Upload Company Logo
To upload a brand new logo or replace existing logo:

* **Method:** `POST`
* **Path:** `/api/v1/company/logo`
* **Headers:** `Authorization: Bearer <accessToken>`
* **Body:** `FormData`
  * Key: `logo` (File: png, jpg, jpeg, webp, svg / Max size: 2MB)
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Logo uploaded successfully",
  "data": {
    "logoUrl": "https://job-portal-backend-1-dv1h.onrender.com/uploads/logos/60d5ecb8629ef31a98e078b0/logo.png"
  }
}
```

---

### 3.4 Delete Company Logo
* **Method:** `DELETE`
* **Path:** `/api/v1/company/logo`
* **Headers:** `Authorization: Bearer <accessToken>`
* **Response (200 OK):**
```json
{
  "success": true,
  "message": "Logo removed successfully",
  "data": null
}
```

---

## 4. Frontend Implementation Example (React)

```tsx
import React, { useState, useEffect } from 'react';
import axios from 'axios';

const API_BASE = 'https://job-portal-backend-1-dv1h.onrender.com/api/v1';

export function CompanySettingsForm() {
  const [formData, setFormData] = useState({
    companyName: '',
    accountEmail: '', // Read only
    industry: '',
    companySize: '',
    companyType: '',
    website: '',
    founded: '',
    about: '',
    city: '',
    state: '',
    country: '',
    phone: '',
    hrEmail: '',
    supportEmail: '',
    linkedin: '',
    twitter: '',
    facebook: '',
    instagram: '',
    github: '',
    logoUrl: '',
  });

  const [loading, setLoading] = useState(false);
  const [logoLoading, setLogoLoading] = useState(false);
  const token = localStorage.getItem('accessToken');

  // 1. Fetch settings on mount
  useEffect(() => {
    async function loadSettings() {
      try {
        const res = await axios.get(`${API_BASE}/company/settings`, {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (res.data?.data) {
          setFormData((prev) => ({ ...prev, ...res.data.data }));
        }
      } catch (err) {
        console.error('Failed to load settings', err);
      }
    }
    loadSettings();
  }, [token]);

  // 2. Handle input changes
  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  // 3. Handle logo file upload
  const handleLogoUpload = async (e) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const data = new FormData();
    data.append('logo', file);

    setLogoLoading(true);
    try {
      const res = await axios.post(`${API_BASE}/company/logo`, data, {
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'multipart/form-data',
        },
      });
      setFormData((prev) => ({ ...prev, logoUrl: res.data?.data?.logoUrl }));
      alert('Logo uploaded successfully!');
    } catch (err) {
      alert('Failed to upload logo');
    } finally {
      setLogoLoading(false);
    }
  };

  // 4. Submit updated data
  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    try {
      const res = await axios.put(`${API_BASE}/company/settings`, formData, {
        headers: { Authorization: `Bearer ${token}` },
      });
      alert('Company profile updated successfully!');
    } catch (err) {
      alert(err.response?.data?.message || 'Failed to update company profile');
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6 max-w-4xl mx-auto p-6 bg-white rounded-xl shadow-sm">
      <h2 className="text-2xl font-bold text-gray-900">Company Profile Settings</h2>

      {/* Logo Section */}
      <div className="flex items-center gap-4">
        {formData.logoUrl ? (
          <img src={formData.logoUrl} alt="Logo" className="w-20 h-20 rounded-lg object-cover border" />
        ) : (
          <div className="w-20 h-20 rounded-lg bg-gray-100 flex items-center justify-center text-gray-400">
            No Logo
          </div>
        )}
        <div>
          <label className="block text-sm font-medium text-gray-700">Company Logo</label>
          <input type="file" accept="image/*" onChange={handleLogoUpload} disabled={logoLoading} className="mt-1" />
        </div>
      </div>

      {/* Basic Info */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700">Company Name</label>
          <input
            type="text"
            name="companyName"
            value={formData.companyName}
            onChange={handleChange}
            required
            className="w-full mt-1 p-2.5 border rounded-lg"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700">Account Login Email (Read Only)</label>
          <input
            type="email"
            value={formData.accountEmail}
            disabled
            className="w-full mt-1 p-2.5 bg-gray-100 border rounded-lg text-gray-500 cursor-not-allowed"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700">Industry</label>
          <input
            type="text"
            name="industry"
            value={formData.industry}
            onChange={handleChange}
            className="w-full mt-1 p-2.5 border rounded-lg"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700">Company Size</label>
          <input
            type="text"
            name="companySize"
            value={formData.companySize}
            onChange={handleChange}
            placeholder="e.g. 50-100 employees"
            className="w-full mt-1 p-2.5 border rounded-lg"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700">Company Type</label>
          <input
            type="text"
            name="companyType"
            value={formData.companyType}
            onChange={handleChange}
            placeholder="e.g. Private, Public, Startup"
            className="w-full mt-1 p-2.5 border rounded-lg"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700">Website</label>
          <input
            type="url"
            name="website"
            value={formData.website}
            onChange={handleChange}
            className="w-full mt-1 p-2.5 border rounded-lg"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700">Founded Year</label>
          <input
            type="text"
            name="founded"
            value={formData.founded}
            onChange={handleChange}
            placeholder="e.g. 2018"
            className="w-full mt-1 p-2.5 border rounded-lg"
          />
        </div>
      </div>

      {/* About Company */}
      <div>
        <label className="block text-sm font-medium text-gray-700">About Company</label>
        <textarea
          name="about"
          rows={4}
          value={formData.about}
          onChange={handleChange}
          className="w-full mt-1 p-2.5 border rounded-lg"
        />
      </div>

      {/* Location */}
      <h3 className="text-lg font-semibold text-gray-900 pt-4">Location</h3>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700">City</label>
          <input
            type="text"
            name="city"
            value={formData.city}
            onChange={handleChange}
            className="w-full mt-1 p-2.5 border rounded-lg"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700">State / Division</label>
          <input
            type="text"
            name="state"
            value={formData.state}
            onChange={handleChange}
            className="w-full mt-1 p-2.5 border rounded-lg"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700">Country</label>
          <input
            type="text"
            name="country"
            value={formData.country}
            onChange={handleChange}
            className="w-full mt-1 p-2.5 border rounded-lg"
          />
        </div>
      </div>

      {/* Contact Emails & Phone */}
      <h3 className="text-lg font-semibold text-gray-900 pt-4">Public Contacts</h3>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700">Phone</label>
          <input
            type="text"
            name="phone"
            value={formData.phone}
            onChange={handleChange}
            className="w-full mt-1 p-2.5 border rounded-lg"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700">HR Email</label>
          <input
            type="email"
            name="hrEmail"
            value={formData.hrEmail}
            onChange={handleChange}
            className="w-full mt-1 p-2.5 border rounded-lg"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700">Support Email</label>
          <input
            type="email"
            name="supportEmail"
            value={formData.supportEmail}
            onChange={handleChange}
            className="w-full mt-1 p-2.5 border rounded-lg"
          />
        </div>
      </div>

      {/* Social Links */}
      <h3 className="text-lg font-semibold text-gray-900 pt-4">Social Profiles</h3>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700">LinkedIn URL</label>
          <input
            type="url"
            name="linkedin"
            value={formData.linkedin}
            onChange={handleChange}
            className="w-full mt-1 p-2.5 border rounded-lg"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700">Twitter URL</label>
          <input
            type="url"
            name="twitter"
            value={formData.twitter}
            onChange={handleChange}
            className="w-full mt-1 p-2.5 border rounded-lg"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700">Facebook URL</label>
          <input
            type="url"
            name="facebook"
            value={formData.facebook}
            onChange={handleChange}
            className="w-full mt-1 p-2.5 border rounded-lg"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700">GitHub URL</label>
          <input
            type="url"
            name="github"
            value={formData.github}
            onChange={handleChange}
            className="w-full mt-1 p-2.5 border rounded-lg"
          />
        </div>
      </div>

      {/* Submit Button */}
      <button
        type="submit"
        disabled={loading}
        className="w-full py-3 bg-slate-900 text-white font-semibold rounded-lg hover:bg-slate-800 transition"
      >
        {loading ? 'Saving Changes...' : 'Save Profile Changes'}
      </button>
    </form>
  );
}
```
