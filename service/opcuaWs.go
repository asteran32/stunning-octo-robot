package service

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	opcUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Client struct {
	// The websocket connection.
	conn *websocket.Conn

	// Buffered channels messages.
	write chan []byte // images and data to client
	read  chan []byte // commands from client
}

const (
	writeTimeout   = 10 * time.Second
	readTimeout    = 60 * time.Second
	pingPeriod     = (readTimeout * 9) / 10
	maxMessageSize = 512
)

// streamReader reads messages from the websocket connection and fowards them to the read channel
func (c *Client) streamReader() {
	defer func() {
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(readTimeout))
	// SetPongHandler sets the handler for pong messages received from the peer.
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(readTimeout)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// feed message to command channel
		c.read <- message
	}
}

// streamWriter writes messages from the write channel to the websocket connection
func (c *Client) streamWriter() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		// Go’s select lets you wait on multiple channel operations.
		// We’ll use select to await both of these values simultaneously.
		select {
		case message, ok := <-c.write:
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// NextWriter returns a writer for the next message to send.
			// The writer's Close method flushes the complete message to the network.
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.write)
			for i := 0; i < n; i++ {
				w.Write(<-c.write)
			}

			if err := w.Close(); err != nil {
				return
			}

		//a channel that will send the time with a period specified by the duration argument
		case <-ticker.C:
			// SetWriteDeadline sets the deadline for future Write calls
			// and any currently-blocked Write call.
			// Even if write times out, it may return n > 0, indicating that
			// some of the data was successfully written.
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
func ReadOPC(c *gin.Context) {
	configFile := "plcConfig.json"

	con, err := opcUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("err: Can not upgrade websocket")
	}
	con.EnableWriteCompression(true)

	// w := &threadSafeWriter{con, sync.Mutex{}}

	// create two channels for read write concurrency
	cWrite := make(chan []byte)
	cRead := make(chan []byte)

	client := &Client{conn: con, write: cWrite, read: cRead}

	// run opcua client application in separate go routine
	go LoadClientApp(cWrite, cRead, configFile)

	// run reader and writer in two different go routines
	// so they can act concurrently
	go client.streamReader()
	go client.streamWriter()

}
