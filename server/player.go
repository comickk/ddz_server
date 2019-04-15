// player
package server

import (
	"UULoServer/logs"
	"UULoServer/model"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

type Packet struct {
	Id   uint16
	Msg  string
	Data interface{} //map[string]interface{}
}
type Player struct {
	//玩家所属房间
	room *Room

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	//玩家的位置
	pos int

	iscall int //是否叫地主 1 叫 0 不叫

	utype int //玩家类型 1 地主 2农名

	cards []int //分配的牌

	user *model.GameUser //玩家信息
}

func NewUser(ws *websocket.Conn) *Player {
	return &Player{room: nil, conn: ws, send: make(chan []byte, 512),
		pos: 0, iscall: 1, utype: 0, user: nil}
}

//读协程，从websocket中读取数据
func (this *Player) readPump() {
	go func() {
		defer func() {
			this.conn.Close()
		}()

		this.conn.SetReadLimit(maxMessageSize)
		this.conn.SetReadDeadline(time.Now().Add(pongWait))

		//收到pong回应后设置ReadDeadline
		this.conn.SetPongHandler(func(string) error { this.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

		//从websocket读数据
		for {
			_, message, err := this.conn.ReadMessage()
			if err != nil {
				if e, ok := err.(*websocket.CloseError); ok {
					logs.Info("websocket closed: %v", e.Error())
				} else {
					logs.Error("error: %v", err.Error())
				}
				break
			}
			fmt.Print("received a message ") //显示原始数据
			fmt.Println(string(message))

			var pk Packet
			err = json.Unmarshal(message, &pk)
			if err != nil {
				logs.Error("Unmarshal json error:%s", err.Error())
				continue
			}

			switch pk.Id {
			case 3001: //登录

				// msg["msg "]  send_AnonymityLogin msg["data"] {id:xxxxx}
				if data, ok := pk.Data.(map[string]interface{}); ok {
					if id, ok := data["id"].(string); ok {

						this.loginHandle(id)
					}
				}

			case 1010: //GetSysCfg: {sendId: 1010, msg: "send_GetSysCfg"},// 获取系统配置
				this.RoomCfg()

			case 1021: //EnterHome: {sendId: 1021, msg: "send_EnterHome"},// 进入房间 data :{ room: enterRoom.id }
				//EnterHome: {recvId: 2021, msg: "recv_EnterHome"},// 进入房间
				matchqueue.Put(this.user.ObjectId, this) //将该玩家加入匹配队列

				msg := make(map[string]interface{})

				msg["order"] = "2021"
				msg["code"] = "0"

				msg["data"] = "0"
				//--------------

				data, err := json.Marshal(&msg)
				if err != nil {
					logs.Error("json marshal error:%s", err.Error())
				}
				this.send <- data

			case 1050: //GetShopListInfo: {sendId: 1050, msg: "send_GetShopListInfo"},// 获取商品列表
				this.GoodsList()

			case 1056: //  GetRank: {sendId: 1056, msg: "send_GetRank"},// 获取排行
				this.RankInfo()

			case 1006: // SetHead: {sendId: 1006, msg: "send_SetHead"},// 设置自定义头像
				//SetHead: {recvId: 2006, msg: "recv_SetHead"},// 设置自定义头像

				msg := make(map[string]interface{})
				res := make(map[string]interface{})
				msg["order"] = "2006"
				code := "0"

				var roleid int = 0
				v, ok := pk.Data.(float64)
				if ok {
					roleid = (int)(v)
				} else {
					code = "1"
				}

				if code == "0" {
					bm := make(map[string]interface{})
					bm["img"] = roleid

					err := model.UpdateUser(this.user.ObjectId, bm)
					if err != nil {
						logs.Error("update db error:%s", err.Error())
						code = "2"
					}
					res["img"] = roleid
					res["costup"] = 0
					res["up"] = this.user.Up
					res["imsg"] = []int{1001, 1002, 1003, 1004}

					msg["data"] = res
					//data['imgs'];// 用户已经购买的所有角色
					//data['costup'];// 花费的金币
					//data['img'];// 更换后的头像
					//data['up'];// 最新的金币
				}

				msg["code"] = code
				data, err := json.Marshal(&msg)
				if err != nil {
					logs.Error("json marshal error:%s", err.Error())
				}
				this.send <- data

			//其余消息上传给房间
			default:
				// if this.user == nil || this.room == nil {
				// 	ret := fmt.Sprintf(`{"cmd":"%s","err":"request sequence error"}`, cmd)
				// 	this.send <- []byte(ret)
				// } else {
				// 	rmsg := Relay{pos: this.pos, msg: msg}
				// 	this.room.rmsg <- rmsg
				// }
			}
			/*
				case "joinroom":
					if this.user == nil {
						var res = `{"cmd":"joinroom","err":"request sequence error"}`
						this.send <- []byte(res)
					} else {
						//matchqueue.Put(this.user.ObjectId, this)
						var res = `{"cmd":"joinroom","err":""}`
						this.send <- []byte(res)
					}

				//离开房间，离开匹配队列
				case "leaveroom":
					if this.user == nil {
						var res = `{"cmd":"leaveroom","err":"request sequence error"}`
						this.send <- []byte(res)
					} else {
						if matchqueue.Contains(this.user.ObjectId) {
							matchqueue.Remove(this.user.ObjectId)
							var res = `{"cmd":"leaveroom","err":""}`
							this.send <- []byte(res)
							if this.room != nil {
								//告知房间用户离开
								rmsg := Relay{pos: this.pos, msg: msg}
								this.room.rmsg <- rmsg
							}

						} else {
							var res = `{"cmd":"leaveroom","err":"game already start"}`
							this.send <- []byte(res)
						}
					}
				//用户准备开始游戏,加入匹配队列
				case "userready":
					if this.user == nil {
						var res = `{"cmd":"userready","err":"request sequence error"}`
						this.send <- []byte(res)
					} else {
						matchqueue.Put(this.user.ObjectId, this)
						var res = `{"cmd":"userready","err":""}`
						this.send <- []byte(res)
					}
				//金币排名
				case "upranking":
					if this.user == nil {
						var res = `{"cmd":"userready","err":"request sequence error"}`
						this.send <- []byte(res)
					} else {
						this.RankingByUp()
					}

				//其余消息上传给房间
				default:
					if this.user == nil || this.room == nil {
						ret := fmt.Sprintf(`{"cmd":"%s","err":"request sequence error"}`, cmd)
						this.send <- []byte(ret)
					} else {
						rmsg := Relay{pos: this.pos, msg: msg}
						this.room.rmsg <- rmsg
					}
				}
			}*/
		}
	}()

}

//写协程，将消息写入websocket
func (this *Player) writePump() {
	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer func() {
			ticker.Stop()
			this.conn.Close()
		}()

		//向websocket写数据
		for {
			select {
			case message, ok := <-this.send:
				this.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					// closed the channel.
					this.conn.WriteMessage(websocket.CloseMessage, []byte{})
					logs.Info("send chan closed")
					return
				}

				w, err := this.conn.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}
				w.Write(message)

				if err := w.Close(); err != nil {
					logs.Error("writer close error:%v", err.Error())
					return
				}

			//ping
			case <-ticker.C:
				this.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := this.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					logs.Error("ping error:%v", err.Error())
					return
				}
			}
		}
	}()
}

func (this *Player) loginHandle(unionid string) {

	/*
		1 从游戏用户表中查找 userid
		2 找到后返回数据
		3 未找到,从 app  user 表中查询
		4 未找到返回错误
		5 找到 则将数据添加进游戏表中并返回
	*/

	var u *model.GameUser
	var err error
	var has bool

	var code string
	u, has = model.GetUser(unionid) //1 从游戏用户表中查找 userid

	msg := make(map[string]interface{})
	res := make(map[string]interface{})
	if err != nil {
		code = err.Error()
	} else {
		code = "0"
	}

	//AnonymityLogin:{recvId:4001,msg:"recv_AnonymityLogin"},// 匿名登录
	//GetUserInfo: {recvId: 2001, msg: "recv_GetUserInfo"},// 获取用户消息
	msg["order"] = "2001"
	//msg["msg"] = "recv_GetUserInfo"

	if has { //2 找到后返回数据
		this.user = u
		res["name"] = u.Nickname
		res["img"] = u.Img
		res["head"] = u.Headimgurl
		res["up"] = u.Up
		res["cw"] = u.Cw
		res["wn"] = u.Wn
		res["ct"] = u.Ct
		res["ud"] = u.Ud
		res["type"] = u.Tp
		res["br"] = u.Br
		res["guide"] = u.Guide
		res["scene"] = u.Scene

		msg["data"] = res

	} else { //3 未找到,从 app  user 表中查询
		var bu *model.User
		bu, has = model.GetBaseUser(unionid)

		if !has { //4 未找到返回错误
			code = "1" //用户不存在

		} else { //5 找到 则将数据添加进游戏表中并返回
			provice := ""
			city := "郑州"
			sex := 1
			country := "中国"
			headimgurl := ""

			u, err = model.NewUser(bu.Username, provice, city, country,
				headimgurl, bu.Id, sex)
			this.user = u

			res["name"] = u.Nickname
			res["img"] = u.Img
			res["head"] = u.Headimgurl
			res["up"] = u.Up
			res["cw"] = u.Cw
			res["wn"] = u.Wn
			res["ct"] = u.Ct
			res["ud"] = u.Ud
			res["type"] = u.Tp
			res["br"] = u.Br
			res["guide"] = u.Guide
			res["scene"] = u.Scene

			msg["data"] = res
		}
	}

	msg["code"] = code

	data, err := json.Marshal(&msg)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
		//code =err.Error()
	}

	//fmt.Println(string(data))
	this.send <- data
}

//取得房间配置信息
func (this *Player) RoomCfg() {
	//goldRoom: [
	// 场次ID + | + 底分 + | + 房间初始倍数 + | + 入场最小值 + | + 入场最大值 + | + 门票 + | + 输赢封顶
	// {id: 0, underPoint: 0, initMul: 0, minEnterPoint: 0, maxEnterPoint: 0, ticket: 0, maxEarn: 0},

	// U钻场数据结构
	//gemRoom: [
	// 场次ID + | + 底分 + | + 房间初始倍数 + | + 可赢钻石 + | + 入场金币 + | + 金币购买的积分数量
	// {id: 0, underPoint: 0, initMul: 0, gem: 0, ticket: 0, gamePoint: 0},

	// GetSysCfg: {recvId: 2010, msg: "recv_GetSysCfg"},// 获取系统配置

	msg := make(map[string]interface{})
	res := make(map[string]interface{})

	msg["order"] = "2010"
	msg["code"] = "0"

	res["cp"] = []string{"1|10|5|100|10000|1|1000000", "2|10|5|100|10000|1|1000000"}
	res["sp"] = []string{"10|10|5|100|10000|100000", "11|20|5|100|10000|100000"}

	msg["data"] = res
	data, err := json.Marshal(&msg)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.send <- data
}

//获取排行榜
func (this *Player) RankInfo() {

	msg := make(map[string]interface{})
	res := make(map[string]interface{})

	//GetRank: {recvId: 2055, msg: "recv_GetRank"},// 获取排行
	msg["order"] = "2055"
	msg["code"] = "0"

	//--------------
	res["gold"] = []string{"t1|10|5|100|11|1000000|222", "t1|10|5|100|11|1000000|222"}
	res["ud"] = []string{"t1|10|5|100|11|1000000|222", "t1|10|5|100|11|1000000|222"}

	msg["data"] = res
	//--------------

	data, err := json.Marshal(&msg)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.send <- data
}

//获取商品列表
func (this *Player) GoodsList() {

	msg := make(map[string]interface{})

	//GetShopListInfo: {recvId: 2050, msg: "recv_GetShopListInfo"},// 获取商品列表
	msg["order"] = "2050"
	msg["code"] = "0"

	//--------------
	res := make(map[string][]string)

	res["list"] = []string{"10|100|100|1|10",
		"11|100|100|0|10",
		"12|100|100|0|10",
		"50|100|100|1|11",
		"51|100|100|1|11",
		"52|100|100|1|11"}

	msg["data"] = res
	//--------------
	data, err := json.Marshal(&msg)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.send <- data
}

//离开房间
func (this *Player) LeaveRoom() {
	//重置用户pos: 0, iscall: 1, utype: 0
	this.pos = 0
	this.iscall = 1
	this.utype = 0
	this.room = nil
	this.cards = nil
}

//金币排名
func (this *Player) RankingByUp() {
	res := make(map[string]interface{})
	res["cmd"] = "upranking"
	res["err"] = ""

	//获取前20名排名
	rank, err := model.RankingByUp()
	if err != nil {
		res["err"] = err.Error()
	}
	res["ranking"] = rank

	data, err := json.Marshal(&res)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.send <- data
}
