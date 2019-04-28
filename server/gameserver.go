//Package server websocket.go
package server

import (
	"UULoServer/logs"
	"UULoServer/model"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var onlinenum map[int]int

var playerquenes map[int]*PlayerQueue

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

func matching(id int, queue *PlayerQueue) {
	for {
		//从匹配队列中取出三个玩家组成一个房间
		//玩家位置:      ua      uc
		//                 ub
		players := queue.Pop3()
		if players != nil {

			ua := players[0]
			ub := players[1]
			uc := players[2]
			ua.pos = 1
			ub.pos = 2
			uc.pos = 3

			//
			room := NewRoom(ua, ub, uc, id)
			ua.room = room
			ub.room = room
			uc.room = room
			room.Run()
		}

		time.Sleep(time.Second * 1)
	}
}

//Start
func Start() {

	//配置房间
	if l := model.InitRoomsCfg(); l != nil {
		fmt.Println("init room cfg ok")
		onlinenum = make(map[int]int)
		playerquenes = make(map[int]*PlayerQueue)
		for _, k := range l {
			onlinenum[k.Id] = 0
			playerquenes[k.Id] = NewPlayerQueue()
		}
	} else {
		fmt.Println("init room cfg filed")
	}

	//建立匹配队列
	for i, k := range playerquenes {
		go matching(i, k)
	}
	// go func() {
	// 	for {
	// 		//从匹配队列中取出三个玩家组成一个房间
	// 		//玩家位置:      ua      uc
	// 		//                 ub
	// 		players := matchqueue.Pop3()
	// 		if players != nil {

	// 			ua := players[0]
	// 			ub := players[1]
	// 			uc := players[2]
	// 			ua.pos = 1
	// 			ub.pos = 2
	// 			uc.pos = 3

	// 			//
	// 			room := NewRoom(ua, ub, uc)
	// 			ua.room = room
	// 			ub.room = room
	// 			uc.room = room
	// 			room.Run()
	// 		}

	// 		time.Sleep(time.Second * 1)
	// 	}
	// }()

	//建立网络
	logs.Info("server start, listen 192.168.0.126:8765/")
	fmt.Println("Server server start ...")

	http.HandleFunc("/", wsHandler)
	err := http.ListenAndServe("192.168.0.126:8765", nil)
	if err != nil {
		logs.Error("start server error:%s", err.Error())
	}

}
