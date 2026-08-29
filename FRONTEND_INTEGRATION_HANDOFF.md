# Frontend Chat & WebSocket Integration Guide

This document summarizes the backend updates, endpoint details, and event schemas for the frontend team to integrate message ticks (Sent / Delivered / Seen) and real-time sidebar notifications.

---

## 1. WebSocket Endpoints

### 1.1 Room-Level WebSocket
For real-time chat *inside* a specific room:
- **URL**: `ws://localhost:8080/api/v1/chats/:roomId/ws?token=<access_token>`
- **Authentication**: Pass the `access_token` as a query parameter (`token`).

### 1.2 User-Level (Global) WebSocket
For real-time sidebar unread counts, status updates, and notification alerts when a room is not active:
- **URL**: `ws://localhost:8080/api/v1/chats/ws?token=<access_token>`
- **Authentication**: Pass the `access_token` as a query parameter (`token`).

---

## 2. Inbound Events (Frontend -> Backend)

Events sent by the client over the **Room-Level WebSocket**:

### 2.1 Send Message
```json
{
  "type": "message",
  "message": "Hello World!"
}
```

### 2.2 Typing Indicator
```json
{
  "type": "typing",
  "isTyping": true // or false
}
```

### 2.3 Seen Event (Read Receipt)
Send this event when the user opens a room or receives a message while inside the room:
```json
{
  "type": "seen"
}
```

---

## 3. Outbound Events (Backend -> Frontend)

Events received by the client over the WebSockets:

### 3.1 New Message (`type: "message"`)
Received on the **Room-Level Socket** (and on the **Global Socket** if the user is *not* in the active room):
```json
{
  "type": "message",
  "message": {
    "id": "message_object_id",
    "roomId": "room_object_id",
    "senderId": "sender_object_id",
    "message": "Hello World!",
    "status": "delivered", // "sent" (offline) | "delivered" (online) | "seen"
    "isRead": false,
    "createdAt": "2026-08-29T16:00:00Z"
  }
}
```

### 3.2 Message Status Ticks (`type: "message_status"`)
Received on both Room and Global sockets when message ticks change status.
- **Delivered Ticks (Double Grey Ticks)**: Sent when the recipient connects.
- **Seen Ticks (Double Blue Ticks)**: Sent when the recipient reads messages in the room.

```json
{
  "type": "message_status",
  "roomId": "room_id_here",
  "status": "delivered" // or "seen"
}
```

### 3.3 User Status Updates (`type: "user_status"`)
Received when a participant's online/offline status changes (triggered by connecting/disconnecting to global or room sockets):
```json
{
  "type": "user_status",
  "userId": "user_id_here",
  "status": "online" // or "offline"
}
```

### 3.4 Typing Indicator (`type: "typing"`)
Received on the Room-Level Socket when the other participant is typing:
```json
{
  "type": "typing",
  "senderId": "sender_id_here",
  "senderName": "Sender Name",
  "isTyping": true // or false
}
```

---

## 4. Summary of Status Ticks Flow

1. **Sent (Single Tick)**: If you send a message and the recipient is offline, the message payload status is `"sent"`.
2. **Delivered (Double Grey Ticks)**: 
   - If the recipient is online globally when you send the message, the status is `"delivered"`.
   - When the recipient goes online, the backend broadcasts a `"message_status"` event with `status: "delivered"`.
3. **Seen (Double Blue Ticks)**: 
   - When the recipient opens the room or views the message, they send `{"type": "seen"}`.
   - The backend sets all incoming messages to `"seen"` in the DB and broadcasts a `"message_status"` event with `status: "seen"`.
