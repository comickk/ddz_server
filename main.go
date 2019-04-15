// UULoServer project main.go
package main

//依赖模块
//github.com/deckarep/golang-set
//github.com/go-sql-driver/mysql
//github.com/go-xorm/xorm
//github.com/gorilla/websocket

import (
	"UULoServer/logs"
	"UULoServer/server"
)

//恢复panic
func RecoverPanic() {
	if p := recover(); p != nil {
		err, ok := p.(error)
		if ok {
			logs.Error(err.Error())
		} else {
			logs.Error("%v\n", p)
		}
	}

}

func main() {
	logs.SetLogFile("uulandlord")
	defer logs.CloseFile()

	//go web.Start()
	server.Start()
}
