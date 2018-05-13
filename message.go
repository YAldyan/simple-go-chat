package main

import (
	"time"
)

// message represents a single message
/*
	pesan yang terkirim membawa 3 variabel di bawah ini
	1. Name
	2. Message
	3. When
	4. AvatarURL (Profile Picture)
*/
type message struct {
	Name      string
	Message   string
	When      time.Time
	AvatarURL string
}
