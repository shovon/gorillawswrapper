# Gorilla WebSocket Wrapper

Package `gorillawswrapper` is a wrapper library for the most common WebSocket usage, involving the [Gorilla WebSocket library](https://github.com/gorilla/websocket).

The following are the boilerplates that this library removes:

- thread-safe message reads
- pings and pongs
- helper functions to write common message types to the clients.

> This library is not a commentary about the utility of Gorilla WebSocket. There are very few libraries like Gorilla, and they should be cherished for what they have to offer.
> 
> With all that said, there were boilerplates involved in writing software that utilizes Gorilla, and this library encapsulates the most common boilerplates that I had to write.

## Usage

```go
package main

import (
  "http"
)

var upgrader = websocket.Upgrader{}

func main() {
  http.HandleFunc(func (w http.ResponseWriter, r *http.Request) {
    c, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
      return
    }
    defer c.Close()

    conn := gorillawswrapper.NewWrapper(c)
    // conn.Stop() does NOT stop the connection; just the ping/pong and read
    // loop!
    defer conn.Stop()

    for msg := range conn.MessagesChannel() {
      fmt.Println(string(msg.Message))
      if err := conn.WriteTextMessage("Got message"); err != nil {
        return
      }
    }
  })
}
```