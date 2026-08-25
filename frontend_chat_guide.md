# Frontend Implementation Guide - Real-Time Chat Feature

This document explains the steps required to implement the real-time chat feature on the Frontend. You can share this file directly with your Frontend Developer.

---

## 💰 Do We Need to Purchase Anything?
**No. There is $0 cost for this feature.**
- We do **not** use third-party services like Firebase, Pusher, or Twilio.
- The Backend handles WebSockets natively using Go.
- The Frontend can use the browser's built-in native HTML5 **`WebSocket` API** (no extra client-side packages are strictly required, though libraries like `socket.io-client` are also NOT needed because this is a standard RFC 6455 WebSocket server).

---

## 🛠️ Frontend Tasks Overview

### 1. UI Components to Build
- **"Chat" / "Message" Buttons:**
  - **Job Seeker Side:** A "Chat with Employer" button on the Job Details or Applied Jobs detail screen.
  - **Employer Side:** A "Chat with Candidate" button on the Applicant Details screen.
- **Chat Page / Dashboard (`/chat`):**
  - **Sidebar:** Lists all active chat rooms (showing the other user's name, avatar, last message snippet, and unread count badge).
  - **Chat Area:**
    - Header showing the other user's name, avatar, and active status.
    - Message history feed (scrollable, auto-scrolls to the bottom on new messages).
    - Different visual bubble styles for Sent (e.g., right side, blue/green background) and Received (e.g., left side, gray background) messages.
    - Input box to type messages with a "Send" button (supporting `Enter` key to send).

---

## 🔄 API Integration Protocol

### Step 1: Open/Get Chat Room ID
Before starting a chat, call the Backend to check if a room already exists or create a new one.

- **Endpoint:** `POST /api/v1/chats`
- **Headers:** `Authorization: Bearer <your_jwt_token>`
- **Request Body (JSON):**
  ```json
  {
    "jobId": "60d5ecb8629ef31a98e078a1"
  }
  ```
- **Response (200 OK):**
  ```json
  {
    "success": true,
    "message": "Chat room retrieved",
    "data": {
      "id": "60d5ecc8629ef31a98e07cba",
      "jobId": "60d5ecb8629ef31a98e078a1",
      "seekerId": "60d5ecb8629ef31a98e078a1",
      "employerId": "60d5ecb8629ef31a98e078a2",
      "createdAt": "2026-08-25T12:00:00Z"
    }
  }
  ```

---

### Step 2: Fetch Chat History (Messages)
Load existing messages before opening the live connection.

- **Endpoint:** `GET /api/v1/chats/:roomId/messages`
- **Headers:** `Authorization: Bearer <your_jwt_token>`
- **Response (200 OK):**
  ```json
  {
    "success": true,
    "data": [
      {
        "id": "60d5ecc8629ef31a98e07cbb",
        "roomId": "60d5ecc8629ef31a98e07cba",
        "senderId": "60d5ecb8629ef31a98e078a1",
        "message": "Hello, is this position still open?",
        "isRead": true,
        "createdAt": "2026-08-25T12:05:00Z"
      }
    ]
  }
  ```

---

### Step 3: Establish Real-Time WebSocket Connection
Once you have the `roomId`, connect using the browser's native `WebSocket` API.

- **WS Protocol URL:**
  - Local Dev: `ws://localhost:8080/api/v1/chats/:roomId/ws?token=<bearerToken>`
  - Production: `wss://job-portal-backend-1-dv1h.onrender.com/api/v1/chats/:roomId/ws?token=<bearerToken>`
  *(Pass your JWT token in the query string `token` parameter for authentication, as WebSocket handshakes cannot easily contain Custom Headers in standard browser clients).*

#### Code Example (JavaScript):
```javascript
const token = localStorage.getItem("accessToken"); // Your JWT token
const roomId = "60d5ecc8629ef31a98e07cba";
const wsUrl = `ws://localhost:8080/api/v1/chats/${roomId}/ws?token=${token}`;

const ws = new WebSocket(wsUrl);

// 1. Connection opened
ws.onopen = () => {
    console.log("Connected to live chat room!");
};

// 2. Listen for messages from backend (real-time)
ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    
    // Append this message object to your local messages array/state
    console.log("New message received:", data);
    appendMessageToFeed(data);
};

// 3. Send message to backend
function sendMessage(text) {
    const payload = {
        message: text
    };
    ws.send(JSON.stringify(payload));
}

// 4. Handle connection closed
ws.onclose = () => {
    console.log("Disconnected from chat. Implement auto-reconnect logic here if needed.");
};
```

---

### Step 4: Fetch All Chat Rooms
To show a list of all active conversations the user has in their inbox:

- **Endpoint:** `GET /api/v1/chats`
- **Headers:** `Authorization: Bearer <your_jwt_token>`
- **Response (200 OK):**
  ```json
  {
    "success": true,
    "data": [
      {
        "id": "60d5ecc8629ef31a98e07cba",
        "jobId": "60d5ecb8629ef31a98e078a1",
        "otherUserName": "Acme Corp (HR)",
        "otherUserAvatar": "https://domain.com/uploads/logos/default.png",
        "lastMessage": "We would like to schedule an interview.",
        "unreadCount": 2,
        "updatedAt": "2026-08-25T12:15:00Z"
      }
    ]
  }
  ```

---

## 💡 Best Practices for Frontend Developer
1. **Auto-Scroll:** Use a helper function to automatically scroll the chat container to the bottom when a new message is appended or when the user first opens the chat room.
2. **Auto-Reconnect:** If the WebSocket connection drops unexpectedly (`ws.onclose`), retry the connection with an exponential backoff (e.g. try after 2s, 4s, 8s, 16s).
3. **Sound/Push Notification:** Play a subtle sound or trigger a browser notification if a message arrives and the tab is currently inactive.
