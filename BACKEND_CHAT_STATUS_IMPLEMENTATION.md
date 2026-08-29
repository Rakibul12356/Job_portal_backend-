# Backend Implementation Checklist: Chat Message Status & Global WebSocket

This document outlines the step-by-step changes required in the backend (Go service) to support message delivery indicators (Sent / Delivered / Seen ticks) and real-time inbox updates.

---

## Checklist for Backend Team

### 1. Message Model & Database Fields
* [ ] Ensure the Message schema/struct has a `status` field:
  * **Type:** String (or enum)
  * **Allowed Values:** `"sent"`, `"delivered"`, `"seen"`
  * **Default Value:** `"sent"`

### 2. WebSocket Event: Outbound Message Sending Flow
When a user (Sender) sends a message over the WebSocket connection:
* [ ] Retrieve the recipient user ID for the room.
* [ ] Check if the recipient is currently **online/active** (connected to the WebSocket server - either in the room or via the global inbox socket):
  * **If Recipient is Online:**
    * Save the message to the DB with `status: "delivered"`.
    * Broadcast the new message containing `"status": "delivered"` to both participants.
  * **If Recipient is Offline:**
    * Save the message to the DB with `status: "sent"`.
    * Broadcast/send the new message containing `"status": "sent"` to the sender.

### 3. Recipient Connection Flow (Delivered Sync)
* [ ] When a user establishes a WebSocket connection (either globally or room-specific):
  * Find all messages in the DB sent to this user that have `status: "sent"`.
  * Update their status to `"delivered"` in the DB.
  * For each updated message, broadcast a `"message_status"` update event to the original sender of that message to update their ticks:
    ```json
    {
      "type": "message_status",
      "roomId": "room_id_here",
      "status": "delivered"
    }
    ```

### 4. Seen Event Handler Flow (Read Receipts)
When the server receives a read receipt event (`{"type": "seen"}`) from a client (User A) inside a room:
* [ ] **CRITICAL:** Update the database to set `status: "seen"` **ONLY** for messages in that room where `senderId != UserA` (do not mark messages sent by User A as seen).
* [ ] Broadcast a `"message_status"` update event to the other participant (User B) to change their ticks to blue double ticks:
  ```json
  {
    "type": "message_status",
    "roomId": "room_id_here",
    "status": "seen",
    "seenBy": "UserA"
  }
  ```

### 5. Global WebSocket Support
* [ ] Expose a global WebSocket endpoint at `/api/v1/chats/ws` authenticated via `?token=<access_token>`.
* [ ] Register connection to track the user's online status globally.
* [ ] When a new message is created in any room, check if the participants are connected to the global WebSocket. If they are connected (but not in the active room socket), broadcast the `"message"` payload to their global socket so their sidebar unread counts increment in real-time.
