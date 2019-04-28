package server

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

//websocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	//允许跨域访问
	CheckOrigin: func(r *http.Request) bool { return true },
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	dealcard      = 15 //确认发牌完毕
	shoutLandlord = 20 // 叫地主
	robLandlord   = 20 // 抢地主
	double        = 15 // 加倍
	putCard       = 30 // 出牌
	noBig         = 15 // 要不起
)

//匹配队列
// var matchqueue *PlayerQueue

// func init() {
// 	matchqueue = NewPlayerQueue()
// }
