package dto

import "time"

type CreateChatRoomDTO struct {
	JobID string `json:"jobId" binding:"required"`
}

type ChatRoomResponseDTO struct {
	ID              string    `json:"id"`
	JobID           string    `json:"jobId"`
	JobTitle        string    `json:"jobTitle"`
	OtherUserID     string    `json:"otherUserId"`
	OtherUserName   string    `json:"otherUserName"`
	OtherUserAvatar string    `json:"otherUserAvatar"`
	LastMessage     string    `json:"lastMessage"`
	UnreadCount     int64     `json:"unreadCount"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
