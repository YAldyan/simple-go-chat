package main

import (
	"github.com/gorilla/websocket"
)

// client represents a single chatting user.
type client struct {
	// socket is the web socket for this client.
	socket *websocket.Conn
	// send is a channel on which messages are sent.
	send chan *message
	// room is the room this client is chatting in.
	room *room
	// userData holds information about the user
	userData map[string]interface{}
}

/*
	The read method allows our client to read from the socket via the ReadMessage method,
	continually sending any received messages to the forward channel on the room type
*/
// func (c *client) read() {

// 	// fungsi defer sebaiknya digunakan untuk function
// 	// yang tidak diketahui kapan akan diakhiri atau di
// 	// tutup.
// 	defer c.socket.Close()
// 	for {

// 		_, msg, err := c.socket.ReadMessage()
// 		if err != nil {
// 			return
// 		}

// 		c.room.forward <- msg
// 	}
// }

/*
	the write method continually accepts messages from the send channel writing everything
	out of the socket via the WriteMessage method
*/
// func (c *client) write() {
// 	defer c.socket.Close()
// 	for msg := range c.send {
// 		err := c.socket.WriteMessage(websocket.TextMessage, msg)
// 		if err != nil {
// 			return
// 		}
// 	}
// }

/*
	komunikasi via web socket dengan JSON
*/
func (c *client) read() {
	defer c.socket.Close()
	for {
		var msg *message
		err := c.socket.ReadJSON(&msg)
		if err != nil {
			return
		}
		msg.When = time.Now()
		msg.Name = c.userData["name"].(string)

		// if avatarURL, ok := c.userData["avatar_url"]; ok {
		// 	msg.AvatarURL = avatarURL.(string)
		// }

		if avatarUrl, ok := c.userData["avatar_url"]; ok {
			msg.AvatarURL = avatarUrl.(string)
		}

		// msg.AvatarURL, _ = c.room.avatar.GetAvatarURL(c)

		c.room.forward <- msg
	}
}

func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.send {
		err := c.socket.WriteJSON(msg)
		if err != nil {
			break
		}
	}
}
