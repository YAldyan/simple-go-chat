package main

import {
	"simple-go-chat/trace"
}
/*
	We need a way for clients to join and leave rooms in order to ensure that
	the c.room.forward <- msg code in the preceding section actually forwards
	the message to all the clients. To ensure that we are not trying to access
	the same data at the same time, a sensible approach is to use two channels:
	one that will add a client to the room and another that will remove it
*/

type room struct {
	// forward is a channel that holds incoming messages
	// that should be forwarded to the other clients.
	// we will use to send the incoming messages to all other clients
	forward chan *message

	// join is a channel for clients wishing to join the room.
	join chan *client

	// leave is a channel for clients wishing to leave the room.
	leave chan *client

	// clients holds all current clients in this room.
	clients map[*client]bool

	// tracer will receive trace information of activity
	// in the room.
	tracer trace.Tracer

	// avatar is how avatar information will be obtained.
	// avatar Avatar
}

/*
	Now we get to use an extremely powerful feature of Go's concurrency offerings the
	select statement. We can use select statements whenever we need to synchronize or
	modify shared memory, or take different actions depending on the various activities
	within our channels.
*/
func (r *room) run() {

	/*
		The top for loop indicates that this method will run forever, until the program
		is terminated. This might seem like a mistake, but remember, if we run this code
		as a goroutine, it will run in the background, which won't block the rest of our
		application
	*/
	for {
		select {

		/*
			If we receive a message on the join channel, we simply update the r.clients map to
			keep a reference of the client that has joined the room. Notice that we are setting
			the value to true. We are using the map more like a slice, but do not have to worry
			about shrinking the slice as clients come and go through time setting the value to
			true is just a handy, low memory way of storing the reference.
		*/
		case client := <-r.join:
			// joining
			r.clients[client] = true

			r.tracer.Trace("New client joined")
		/*
			If we receive a message on the leave channel, we simply delete the client type from
			the map, and close its send channel
		*/
		case client := <-r.leave:
			// leaving
			delete(r.clients, client)
			close(client.send)

			r.tracer.Trace("Client left")

		/*
			If we receive a message on the forward channel, we iterate over all the clients and
			add  the message to each client's send channel. Then, the write method of our client
			type will pick it up and send it down the socket to the browser.
		*/
		case msg := <-r.forward:

			r.tracer.Trace("Message received: ", msg.Message)

			// forward message to all clients
			for client := range r.clients {
				client.send <- msg

				r.tracer.Trace(" -- sent to client")
			}
		}
	}
}

/*
	Now we are going to turn our room type into an http.Handler type like we did with the
	template handler earlier. As you will recall, to do this, we must simply add a method
	called ServeHTTP with the appropriate signature.
*/
const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

/*
	If you accessed the chat endpoint in a web browser, you would likely crash the program
	and see an error like ServeHTTPwebsocket: version != 13. This is because it is intended
	to be accessed via a web socket rather than a web browser.
*/
var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	/*
		In order to use web sockets, we must upgrade the HTTP connection using the websocket.
		Upgrader type, which is reusable so we need only create one. Then, when a request comes
		in via the ServeHTTP method, we get the socket by calling the upgrader.Upgrade method.
	*/
	socket, err := upgrader.Upgrade(w, req, nil)

	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	/*
		Melakukan pengecekan terkait user information
		dari Cookies
	*/
	authCookie, err := req.Cookie("auth")
	
	if err != nil {
		log.Fatal("Failed to get auth cookie:", err)
		return
	}

	/*
		All being well, we then create our client and pass it into the join channel for the current
		room. We also defer the leaving operation for when the client is finished, which will ensure
		everything is tidied up after a user goes away
	*/
	client := &client{
		socket: socket,
		send:   make(chan *message, messageBufferSize),
		room:   r,
		userData: objx.MustFromBase64(authCookie.Value),
	}
	r.join <- client
	defer func() { r.leave <- client }()

	// Go Routine sendiri, jalan di belakang - Asychoronous
	go client.write()

	client.read()
}

// newRoom makes a new room.
// func newRoom() *room {
// 	return &room{
// 		forward: make(chan *message),
// 		join:    make(chan *client),
// 		leave:   make(chan *client),
// 		clients: make(map[*client]bool),
// 		tracer: trace.Off(),
// 	}
// }

/*
	newRoom makes a new room that is ready to go.

	Room untuk menciptakan room dengan image avatar
	tertentu.
*/
func newRoom(avatar Avatar) *room {

	return &room{
		forward: make(chan *message),
		join: make(chan *client),
		leave: make(chan *client),
		clients: make(map[*client]bool),
		tracer: trace.Off(),
		// avatar: avatar,
	}
}
