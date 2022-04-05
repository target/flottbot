package models

import (
	"time"

	"github.com/rs/xid"
)

// Message is the struct of the main data structure being passed around for each message generated
type Message struct {
	ID                string
	Type              MessageType
	Service           MessageService
	ChannelID         string
	ChannelName       string
	Input             string
	Output            string
	Error             string
	Timestamp         string
	ThreadID          string
	ThreadTimestamp   string
	BotMentioned      bool
	DirectMessageOnly bool
	Debug             bool
	IsEphemeral       bool
	StartTime         int64
	EndTime           int64
	Attributes        map[string]string
	Vars              map[string]string
	OutputToRooms     []string
	OutputToUsers     []string
	Remotes           Remotes
	SourceLink        string
}

// MessageType is used to differentiate between different message types
type MessageType int

// Supported MessageTypes
const (
	MsgTypeUnknown MessageType = iota
	MsgTypeDirect
	MsgTypeChannel
	MsgTypePrivateChannel
)

// MessageService is used to differentiate between different message services
type MessageService int

// Supported MessageServices
const (
	MsgServiceUnknown MessageService = iota
	MsgServiceChat
	MsgServiceCLI
	MsgServiceScheduler
)

// GenerateMessageID generates a random ID for a message
func GenerateMessageID() string {
	return xid.New().String()
}

// MessageTimestamp timestamps the message
func MessageTimestamp() int64 {
	return time.Now().Unix()
}

// NewMessage creates a new message with initialized fields
func NewMessage() Message {
	return Message{
		ID:            GenerateMessageID(),
		StartTime:     MessageTimestamp(),
		Attributes:    make(map[string]string),
		Vars:          make(map[string]string),
		OutputToRooms: []string{},
		OutputToUsers: []string{},
		Debug:         false,
		IsEphemeral:   false,
	}
}
