# Backend Specification: Advanced Chat WebSocket & Status Syncing

This guide details the backend changes required to support correct status transitions (Sent / Delivered / Seen ticks) and a User-level (Global) WebSocket connection for real-time inbox notifications.

---

## 1. Sent vs Delivered vs Seen Status Transitions

The message tick mark status (`sent`, `delivered`, `seen`) requires the backend to track whether the recipient is active/online and when they read the messages.

### 1.1 Message Flow & Status Mapping
When User A sends a message to User B:

1. **Check Recipient Online Status:**
   * If User B is **online** (connected to the WebSocket server - either globally or in the room):
     * Set the message status in the DB to `"delivered"`.
     * Send a `"message_status"` update event of type `"delivered"` to User A.
   * If User B is **offline**:
     * Set the message status in the DB to `"sent"`.
     * (Do not send any status updates until User B logs in).

2. **Recipient Connects (Delivered Sync):**
   * When User B connects to the WebSocket server (either global or room-specific):
     * The server must scan all unread messages sent to User B by other users.
     * Update their status in the DB from `"sent"` to `"delivered"`.
     * Broadcast a `"message_status"` update event with `status: "delivered"` to the respective senders (so they get double grey ticks).

3. **Recipient Reads Messages (Seen Trigger):**
   * When User B sends a `{"type": "seen"}` event to the server for a specific room (or when they open the room):
     * **CRITICAL BUG FIX:** The backend must **ONLY** update messages where `senderId != UserB` (meaning only User B's unread messages from the other user are marked as seen).
     * **DO NOT** mark messages sent by User B as seen.
     * Update the status of these incoming messages in the DB to `"seen"`.
     * Broadcast a `"message_status"` event to User A (the sender of those messages):
       ```json
       {
         "type": "message_status",
         "roomId": "room_id_here",
         "status": "seen",
         "seenBy": "UserB"
       }
       ```

---

## 2. User-level (Global) WebSocket Connection

Currently, the WebSocket connection is room-level (`/chats/:roomId/ws`). If the user is on the main `/chat` screen with no room open (or browsing other pages), they do not receive real-time updates (incoming messages or unread count increments in the sidebar).

### 2.1 Proposed Global WebSocket Endpoint
Implement a global WebSocket endpoint:
* **Endpoint:** `/api/v1/chats/ws` (or a general `/api/v1/ws`)
* **Handshake Parameter:** `?token=<access_token>`

### 2.2 Global Broadcast Rules
When a WebSocket connection is established at this user-level endpoint:
1. **User Status (Online):** Register the user as `"online"` globally. Broadcast `{"type": "user_status", "userId": "...", "status": "online"}` to all active rooms this user belongs to.
2. **Real-time Inbox Updates:** Whenever a new message is sent in *any* room this user belongs to, the server must broadcast the `"message"` event to this global socket connection.
3. This allows the frontend to:
   * Dynamically increment `unreadCount` on the room list in the sidebar.
   * Show notification banners/alerts on the screen.
   * Sort the room list instantly to bring the latest conversation to the top.
