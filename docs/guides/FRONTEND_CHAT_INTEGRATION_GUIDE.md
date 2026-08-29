# Frontend Integration Guide: Advanced Chat Features

This guide details how the frontend team can integrate the new advanced WebSocket-based chat features: online status green dots, message read receipts (sent/delivered/seen ticks), and real-time typing indicators.

---

## 1. WebSocket Event Protocol (Inbound & Outbound)

All WebSocket communications are now event-driven JSON payloads. You will send and receive objects containing a `"type"` property to distinguish between different event types.

### Outbound Event Formats (Client to Server)

#### A. Sending a Chat Message
To send a chat message, construct a payload of type `"message"`:
```json
{
  "type": "message",
  "message": "Hello! How are you?"
}
```

#### B. Triggering a Typing Indicator
Send this payload when the user starts or stops typing. Make sure to implement debouncing (see Section 2).
```json
{
  "type": "typing",
  "isTyping": true // true when typing starts, false when typing stops
}
```

#### C. Sending a Read Receipt ("Seen")
Send this event when the user opens a chat room, scrolls to view unread messages, or focuses on the input field of an active chat room:
```json
{
  "type": "seen"
}
```

---

### Inbound Event Formats (Server to Client)

Your client WebSocket listener must parse the incoming JSON payload and check the `"type"` field.

#### A. New Message Received
* **type:** `"message"`
* **Payload Structure:**
```json
{
  "type": "message",
  "message": {
    "id": "64e0abc...",
    "roomId": "64e0def...",
    "senderId": "64e0ghi...",
    "message": "Hello! How are you?",
    "status": "sent", // "sent" | "delivered" | "seen"
    "createdAt": "2026-08-29T15:24:00Z"
  }
}
```

#### B. Message Status Update (Sent / Delivered / Seen)
This event notifies the client when messages have transitioned their status (e.g. from single tick to double tick, or double tick to blue seen tick).
* **type:** `"message_status"`
* **Payload Structure:**
```json
{
  "type": "message_status",
  "roomId": "64e0def...",
  "status": "seen", // "delivered" | "seen"
  "seenBy": "64e0ghi..." // User ID who read the messages
}
```
* **Handling Rule:** When you receive this event, update the visual tick marks for all messages in the room sent by the active user:
  * If `status` is `"delivered"`, render double checkmarks (`✔✔` gray).
  * If `status` is `"seen"`, render blue double checkmarks (`✔✔` blue).

#### C. Typing Indicator Status
* **type:** `"typing"`
* **Payload Structure:**
```json
{
  "type": "typing",
  "senderId": "64e0ghi...",
  "senderName": "Jane Doe",
  "isTyping": true // true if they are typing, false if they stopped
}
```
* **Handling Rule:** If `isTyping` is `true`, render a notification indicator at the bottom of the message container or under the header (e.g. `"Jane Doe is typing..."`). Remove it when `isTyping` is `false`.

#### D. Participant User Status (Online / Offline Green Dot)
This event notifies you when the other participant in the active chat room connects or disconnects.
* **type:** `"user_status"`
* **Payload Structure:**
```json
{
  "type": "user_status",
  "userId": "64e0ghi...",
  "status": "online" // "online" | "offline"
}
```
* **Handling Rule:** Toggle the visual green status dot next to the user's avatar or username in the chat header or sidebar based on this status.

---

## 2. Best Practices for Keystroke Debouncing (Typing Indicators)

To prevent flooding the WebSocket connection with packets on every single keypress, the client should debounce keystrokes:

1. **Start Typing:** On the first keypress, if the client is not already marked as typing, send the `{"type": "typing", "isTyping": true}` payload.
2. **Debounce Timer:** Start a local timeout timer (e.g. `1.5 seconds`).
3. **Subsequent Keypresses:** Reset/clear the existing timeout timer and start a new one. Do not send another socket event.
4. **Stop Typing:** When the timeout timer finally fires (meaning the user has paused typing for 1.5 seconds), send the `{"type": "typing", "isTyping": false}` payload.

---

## 3. Tick Mark Rendering UI Guidelines

When rendering sent/received message status tick marks:
* **Message sent by another user:** Never render tick marks.
* **Message sent by current user:** Render indicators based on the message's `status` field:
  * `status === "sent"`: Gray single tick (`✔`)
  * `status === "delivered"`: Gray double tick (`✔✔`)
  * `status === "seen"`: Blue double tick (`✔✔` colored)
