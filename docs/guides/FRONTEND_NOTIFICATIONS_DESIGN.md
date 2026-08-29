# Frontend Integration Guide: In-App Notifications & Alerts

This guide explains how to implement the in-app notification system (with a notification bell, unread badge counts, and read status updates) in the frontend application.

---

## 1. API Endpoints

### 1.1 Fetch Notifications (Seeker Dashboard)
Retrieve all notifications for the logged-in candidate, sorted from newest to oldest.
*   **URL:** `/api/v1/notifications`
*   **Method:** `GET`
*   **Headers:** `Authorization: Bearer <access_token>`
*   **Response Format (200 OK):**
    ```json
    {
      "status": 200,
      "message": "Success",
      "data": [
        {
          "id": "60d5ec4b1234567890123456",
          "userId": "60d5ec4b9012345678901234",
          "title": "Application Shortlisted!",
          "message": "Congratulations! You have been shortlisted for the role of Go Developer at Innovate Tech Solutions.",
          "type": "shortlist",
          "isRead": false,
          "createdAt": "2026-08-29T11:45:00Z"
        }
      ]
    }
    ```

### 1.2 Mark Notification as Read
Update a specific notification's status to "read" when the user clicks on it or expands the list.
*   **URL:** `/api/v1/notifications/:id/read`
*   **Method:** `PATCH`
*   **Headers:** `Authorization: Bearer <access_token>`
*   **Response Format (200 OK):**
    ```json
    {
      "status": 200,
      "message": "Notification marked as read successfully",
      "data": null
    }
    ```

---

## 2. Recommended Frontend UI Layout & Components

To build an intuitive user experience, create a **Notification Bell** component inside your Navigation Bar (Navbar).

### 2.1 Navigation Bar Bell Component
*   **Visual Elements:**
    *   A bell icon (e.g., from `lucide-react` or `font-awesome`).
    *   A red floating circle (badge) showing the count of **unread** notifications (`isRead === false`).
    *   *Behavior:* Clicking the bell toggles a dropdown panel displaying the list of notifications.

```text
+--------------------------------------------------------+
|  Online Job Portal                     [Search]  (3) 🔔|  <-- Bell with badge count
+--------------------------------------------------|-----|
                                                   | [ ] Application Shortlisted!
                                                   |     Congratulations! You...
                                                   |     2 minutes ago
                                                   | 
                                                   | [x] New Job Alerts Matching...
                                                   |     We found new jobs matching...
                                                   |     1 day ago
                                                   +-----------------------------+
```

### 2.2 Dropdown Item Highlights
*   **Unread Items:** Should have a light grey or blue background highlight, with a small blue indicator dot to denote that it is unread.
*   **Interaction:** Clicking a notification item should trigger the `PATCH /api/v1/notifications/:id/read` API call to update the DB state and remove the highlight/badge count instantly.

---

## 3. Example React Component Implementation

```jsx
import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Bell, Check } from 'lucide-react';

function NotificationBell() {
  const [notifications, setNotifications] = useState([]);
  const [isOpen, setIsOpen] = useState(false);

  // Fetch notifications from backend
  const fetchNotifications = async () => {
    try {
      const token = localStorage.getItem('accessToken');
      const response = await axios.get('http://localhost:8080/api/v1/notifications', {
        headers: { Authorization: `Bearer ${token}` }
      });
      setNotifications(response.data.data);
    } catch (error) {
      console.error("Failed to load notifications:", error);
    }
  };

  useEffect(() => {
    fetchNotifications();
    // Poll notifications every 30 seconds for live updates
    const interval = setInterval(fetchNotifications, 30000);
    return () => clearInterval(interval);
  }, []);

  const unreadCount = notifications.filter(n => !n.isRead).length;

  // Mark notification as read
  const handleMarkAsRead = async (id) => {
    try {
      const token = localStorage.getItem('accessToken');
      await axios.patch(`http://localhost:8080/api/v1/notifications/${id}/read`, {}, {
        headers: { Authorization: `Bearer ${token}` }
      });
      // Update local state instantly
      setNotifications(prev => 
        prev.map(n => n.id === id ? { ...n, isRead: true } : n)
      );
    } catch (error) {
      console.error("Failed to mark as read:", error);
    }
  };

  return (
    <div className="relative">
      {/* Bell Trigger */}
      <button onClick={() => setIsOpen(!isOpen)} className="relative p-2 text-gray-600 hover:text-blue-600 transition-colors">
        <Bell size={24} />
        {unreadCount > 0 && (
          <span className="absolute top-0 right-0 inline-flex items-center justify-center px-2 py-1 text-xs font-bold leading-none text-white transform translate-x-1/2 -translate-y-1/2 bg-red-500 rounded-full">
            {unreadCount}
          </span>
        )}
      </button>

      {/* Notifications Dropdown Panel */}
      {isOpen && (
        <div className="absolute right-0 mt-2 w-80 bg-white border border-gray-200 rounded-lg shadow-xl overflow-hidden z-50">
          <div className="p-3 border-b border-gray-100 font-semibold text-gray-700 flex justify-between items-center bg-gray-50">
            <span>Notifications</span>
            {unreadCount > 0 && <span className="text-xs text-blue-600">{unreadCount} unread</span>}
          </div>
          <div className="max-h-64 overflow-y-auto">
            {notifications.length === 0 ? (
              <p className="p-4 text-center text-sm text-gray-400">No notifications yet.</p>
            ) : (
              notifications.map(notif => (
                <div 
                  key={notif.id}
                  onClick={() => !notif.isRead && handleMarkAsRead(notif.id)}
                  className={`p-3 border-b border-gray-50 hover:bg-gray-50 cursor-pointer transition-colors ${!notif.isRead ? 'bg-blue-50/40' : ''}`}
                >
                  <div className="flex justify-between items-start">
                    <h4 className={`text-sm ${!notif.isRead ? 'font-semibold text-gray-900' : 'text-gray-700'}`}>
                      {notif.title}
                    </h4>
                    {!notif.isRead && <span className="w-2 h-2 bg-blue-600 rounded-full mt-1.5"></span>}
                  </div>
                  <p className="text-xs text-gray-500 mt-1">{notif.message}</p>
                  <span className="text-[10px] text-gray-400 mt-2 block">
                    {new Date(notif.createdAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                  </span>
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}

export default NotificationBell;
```
