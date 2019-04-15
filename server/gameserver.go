//Package server websocket.go
package server

import (
	"UULoServer/logs"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		logs.Error("%s%s", "Cannot setup WebSocket connection:", err.Error())
		return
	}

	p := NewUser(ws)

	//read and write goroutine
	p.readPump()
	p.writePump()
}

//Start adadfasdf
func Start() {

	//玩家匹配
	go func() {
		for {

			//从匹配队列中取出三个玩家组成一个房间
			//玩家位置:      ua      uc
			//                 ub
			players := matchqueue.Pop3()
			if players != nil {

				ua := players[0]
				ub := players[1]
				uc := players[2]
				ua.pos = 1
				ub.pos = 2
				uc.pos = 3

				//
				room := NewRoom(ua, ub, uc)
				ua.room = room
				ub.room = room
				uc.room = room
				room.Run()
			}

			time.Sleep(time.Second * 1)
		}
	}()

	logs.Info("server start, listen 192.168.0.126:8765/")
	fmt.Println("Server server start ...")
	http.HandleFunc("/", wsHandler)
	err := http.ListenAndServe("192.168.0.126:8765", nil)
	if err != nil {
		logs.Error("start server error:%s", err.Error())
	}
}
