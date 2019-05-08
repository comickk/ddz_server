// server
package server

import (
	"UULoServer/logs"
	"UULoServer/model"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

//Packet 客户端发来的数据包格式
type Packet struct {
	Id   uint16
	Msg  string
	Data interface{} //map[string]interface{}
}
type Player struct {
	//玩家所属房间
	queue *PlayerQueue //所在队列
	room  *Room

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	//玩家的位置
	pos int

	iscall int //-1 未叫 是否叫地主 2 抢 1 叫 0 不叫/不抢

	utype     int  //玩家类型 1 地主 2农名
	isvcstart bool //是否明牌开始

	mul   int   //玩家倍数
	cards []int //分配的牌

	user *model.GameUser //玩家信息

	noping int //无心跳检测次数
}

func NewUser(ws *websocket.Conn) *Player {
	return &Player{queue: nil, room: nil, conn: ws, send: make(chan []byte, 512),
		pos: 0, iscall: -1, utype: 0, user: nil, mul: 1, noping: 0, isvcstart: false}
}

//读协程，从websocket中读取数据
func (this *Player) readPump() {
	go func() {
		defer func() {
			this.Close()
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
			//fmt.Print("接收到消息  <--- ") //显示原始数据
			//fmt.Println(string(message))

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

			// case 1900: //取得token 并开启心跳
			// 	res := `{"order":"2900","code":0,"data":{"token":"0","key":"0","iv":"0"}}`
			// 	this.send <- []byte(res)

			// case 1100:
			// 	// HeartBeat: {sendId: 1100, msg: "send_HeartBeat"},// 心跳
			// 	this.HeartBeat()

			case 1010: //GetSysCfg: {sendId: 1010, msg: "send_GetSysCfg"},// 获取系统配置
				this.RoomCfg()

			case 1002:
				//{name: strNam, img: roleId}
				if data, ok := pk.Data.(map[string]interface{}); ok {
					if img, ok1 := data["img"].(float64); ok1 {
						if name, ok2 := data["name"].(string); ok2 {

							this.SeNameAndHead(name, (int)(img))
						}
					}
				}

			case 1011: //UpdateDeskPeople: {sendId: 1011, msg: "send_UpdateDeskPeople"},// 获取场次人数
				this.UpdateDeskPeople()

			case 1021: //EnterHome: {sendId: 1021, msg: "send_EnterHome"},// 进入房间 data :{ room: enterRoom.id }
				//EnterHome: {recvId: 2021, msg: "recv_EnterHome"},// 进入房间
				this.EnterRoom()

			case 1050: //GetShopListInfo: {sendId: 1050, msg: "send_GetShopListInfo"},// 获取商品列表
				//data: 1  豆    2   钻石
				if v, ok := pk.Data.(float64); ok {
					this.GoodsList((int)(v))
				}

			case 1051: //{"id":1051,"msg":"send_PayShop","data":{"id":"11","platform":3}}
				if data, ok := pk.Data.(map[string]interface{}); ok {
					if idstr, ok := data["id"].(string); ok {
						//id, err := strconv.Atoi(idstr)
						this.PayShop(idstr)
					}
				}

			case 1056: //  GetRank: {sendId: 1056, msg: "send_GetRank"},// 获取排行
				this.RankInfo()

			case 1006: // SetHead: {sendId: 1006, msg: "send_SetHead"},// 设置自定义头像
				//SetHead: {recvId: 2006, msg: "recv_SetHead"},// 设置自定义头像
				var roleid int
				v, ok := pk.Data.(float64)
				if ok {
					roleid = (int)(v)
					this.SetHead(roleid)
				} else {
					continue
				}

			case 1030: // BeganGame: {sendId: 1030, msg: "send_BeganGame"},// 开始游戏
				//NetSocketMgr.send(GameNetMsg.send.BeganGame, {vc: 0, room: roomID});
				if data, ok := pk.Data.(map[string]interface{}); ok {

					if room, ok := data["room"].(string); ok {

						if vc, ok := data["vc"].(float64); ok {

							id, err := strconv.Atoi(room)
							if err != nil {
								//fmt.Println("string to in  error")
							}
							this.BeginGame((int)(vc), (int)(id))
						}
					}
				}

			case 1022:
				//LeaveHome: {sendId: 1022, msg: "send_LeaveHome"},// 离开房间
				//this.LeaveRoom()
				if this.user == nil || this.room == nil {
					//fmt.Println("将消息转入房间时失败")
					this.LeaveRoom()
				} else {
					rmsg := Relay{pos: this.pos, msg: pk}
					this.room.rmsg <- rmsg
				}

			case 1027: //FindPlayerOver:{sendId:1027,msg:"send_FindPlayerOver"},//匹配超时
				this.FindPlayerOver()
			case 1037: //ChangeDesk: {sendId: 1037, msg: "send_ChangeDesk"},// 换桌
				this.ChangeDesk()

			case 1007: //切换场景   SetGameScene: {sendId: 1007, msg: "send_SetGameScene"},//
				var sid int
				v, ok := pk.Data.(float64)
				if ok {
					sid = (int)(v)
					this.SetGameScene(sid)
				} else {
					continue
				}

			//其余消息上传给房间
			default:
				if this.user == nil || this.room == nil {
					//fmt.Println("将消息转入房间时失败")
					//ret := fmt.Sprintf(`{"cmd":"%s","err":"request sequence error"}`, cmd)
					//this.send <- []byte(ret)
				} else {
					rmsg := Relay{pos: this.pos, msg: pk}
					this.room.rmsg <- rmsg
				}
			}
		}
	}()

}

//写协程，将消息写入websocket
func (this *Player) writePump() {
	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer func() {
			ticker.Stop()
			this.Close()
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
				//fmt.Print("输出消息  ---> ") //显示原始数据
				//fmt.Println(string(message))
				w.Write(message)

				if err := w.Close(); err != nil {
					logs.Error("writer close error:%v", err.Error())
					return
				}

			//ping  定时发送心跳检测,只要无错误  即联接正常
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

	code := 0
	u, has = model.GetUser(unionid) //1 从游戏用户表中查找 userid

	msg := make(map[string]interface{})
	res := make(map[string]interface{})
	if err != nil {
		code = 9015
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
		//var bu *model.User
		//bu, has = model.GetBaseUser(unionid)

		username := model.GetBaseUser(unionid)

		fmt.Printf("-----------------------  %s    \n", username)
		if username == "" { //4 未找到返回错误
			code = 9015 //用户不存在

		} else { //5 找到 则将数据添加进游戏表中并返回
			provice := ""
			city := "郑州"
			sex := 1
			country := "中国"
			headimgurl := ""

			u, err = model.NewUser(username, provice, city, country,
				headimgurl, unionid, sex)
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

	//res["cp"] = []string{"1|100|1|10000|100000|1|1000000", "2|1000|5|10000|10000000|1|10000000"}
	res["cp"] = model.GetRoomsCfg()
	res["sp"] = []string{"10|10|5|100|10000|100000", "11|20|5|100|10000|100000"}

	msg["data"] = res
	data, err := json.Marshal(&msg)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.send <- data
}

//设置用户昵称和虚拟形象
func (this *Player) SeNameAndHead(name string, roleid int) {
	//SetNameAndHead: {sendId: 1002, msg: "send_SetNameAndHead"},// 设置用户昵称形象
	//SetNameAndHead: {recvId: 2002, msg: "recv_SetNameAndHead"},// 设置用户昵称形象

	//设置昵称和   形象id

	code := 0

	this.user.Nickname = name
	this.user.Img = roleid

	bm := make(map[string]interface{})
	bm["nickname"] = name
	bm["img"] = roleid

	err := model.UpdateUser(this.user.ObjectId, bm)
	if err != nil {
		logs.Error("set name and head update db error :%s", err.Error())
		code = 1
	}

	if code == 0 {
		//
		res := `{"order":"2002","code":0,"data":""}`
		this.send <- []byte(res)
	}
}

//获取排行榜
func (this *Player) RankInfo() {

	msg := make(map[string]interface{})
	res := make(map[string]interface{})

	//GetRank: {recvId: 2055, msg: "recv_GetRank"},// 获取排行
	msg["order"] = "2055"
	msg["code"] = "0"

	//--------------
	// //   名字 |  胜局  | 连胜  |  3   |  4  | 头像
	// //获取前20名排名
	if model.Uprank == nil {
		e := model.GetRank(1)
		res["err"] = e.Error()
	}

	if model.Udrank == nil {
		e := model.GetRank(1)
		res["err"] = e.Error()
	}

	// res["ranking"] = rank
	//res["gold"] = []string{"t1|10|5|100|11|1000000|222", "t1|10|5|100|11|1000000|222"}
	res["gold"] = model.Uprank
	res["ud"] = model.Udrank //[]string{"t1|10|5|100|11|1000000|222", "t1|10|5|100|11|1000000|222"}

	msg["data"] = res
	//--------------

	data, err := json.Marshal(&msg)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.send <- data
}

//获取商品列表
func (this *Player) GoodsList(goodtype int) {

	msg := make(map[string]interface{})

	//GetShopListInfo: {recvId: 2050, msg: "recv_GetShopListInfo"},// 获取商品列表
	msg["order"] = "2050"
	msg["code"] = "0"

	//--------------
	res := make(map[string][]string)

	// 商品ID + | + 金币 + | + 对应人民币 + | + 是否热卖 + | + 多送百分比
	// res["list"] = []string{"10|100|100|1|10",
	// 	"11|100|100|0|10",
	// 	"12|100|100|0|10",
	// 	"50|100|100|1|11",
	// 	"51|100|100|1|11",
	// 	"52|100|100|1|11"}

	res["list"], _ = model.GetGoodList(goodtype)

	msg["data"] = res
	//--------------
	data, err := json.Marshal(&msg)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.send <- data
}

//玩家进入游戏
func (this *Player) EnterRoom() {

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
}

//玩家开始游戏 开始匹配 用户
func (this *Player) BeginGame(vc, id int) {

	//BeyondMax: 9007, //金豆不在该房间范围之内 超出最高限制
	if this.user.Up > model.GetDetialCfg(id).Maxenterpoint {
		res := `{"order":"","code":9007,"data":""}`
		this.send <- []byte(res)
		return
	}

	//是否破产
	if this.user.Up <= model.GetDetialCfg(id).Minenterpoint {
		this.Bankrupt(id)
	}

	//GoldNotEnough: 9000,// 金豆不足 低于最低限制
	if this.user.Up < model.GetDetialCfg(id).Minenterpoint {
		res := `{"order":"","code":9000,"data":""}`
		this.send <- []byte(res)
		return
	}

	this.queue = playerquenes[id]
	this.queue.Put(this.user.ObjectId, this) //将该玩家加入匹配队列
	//matchqueue.Put(this.user.ObjectId, this) //将该玩家加入匹配队列
	//fmt.Printf("当前队列人数:  %d  \n", this.queue.Len())

	//是否明牌开始
	if vc > 0 {
		this.isvcstart = true
	}

	msg := make(map[string]interface{})

	msg["order"] = "2030"
	msg["code"] = "0"

	//res := make(map[string]interface{})
	//res["pos"] = this.pos
	//res["room"] = this.room.roomtype
	msg["data"] = ""
	//--------------

	data, err := json.Marshal(&msg)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.send <- data
}

//设置形像
func (this *Player) SetHead(img int) {
	msg := make(map[string]interface{})
	res := make(map[string]interface{})
	msg["order"] = "2006"
	code := "0"

	bm := make(map[string]interface{})
	bm["img"] = img

	err := model.UpdateUser(this.user.ObjectId, bm)
	if err != nil {
		logs.Error("update db error:%s", err.Error())
		code = "1"
	}
	res["img"] = img
	res["costup"] = 0
	res["up"] = this.user.Up
	res["imsg"] = []int{1001, 1002, 1003, 1004}

	msg["data"] = res
	//data['imgs'];// 用户已经购买的所有角色
	//data['costup'];// 花费的金币
	//data['img'];// 更换后的头像
	//data['up'];// 最新的金币

	msg["code"] = code
	data, err := json.Marshal(&msg)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.send <- data
}

//匹配对手超时,移出匹配队列
func (this *Player) FindPlayerOver() {
	//matchqueue.Remove(this.user.ObjectId) //
	if this.queue != nil {
		this.queue.Remove(this.user.ObjectId) //
	}
}

func (this *Player) ChangeDesk() {
	//ChangeDesk: {recvId: 2044, msg: "recv_ChangeDesk"},// 换桌
	msg := make(map[string]interface{})
	msg["order"] = "2044"
	msg["code"] = "0"
	msg["data"] = "0"

	data, err := json.Marshal(&msg)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.send <- data
}

//离开房间
func (this *Player) LeaveRoom() {
	//LeaveHome: {recvId: 2022, msg: "recv_LeaveHome"},// 离开房间

	//matchqueue.Remove(this.user.ObjectId) //移出匹配队列
	if this.queue != nil {
		this.queue.Remove(this.user.ObjectId)
		this.queue = nil
	}
	msg := make(map[string]interface{})

	msg["order"] = "2022"
	msg["code"] = "0"
	msg["data"] = "0"
	//--------------

	data, err := json.Marshal(&msg)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.send <- data
}

func (this *Player) UpdateDeskPeople() {
	//UpdateDeskPeople: {recvId: 2011, msg: "recv_UpdateDeskPeople"},// 获取场次人数
	msg := make(map[string]interface{})
	msg["order"] = "2011"
	msg["code"] = "0"
	msg["data"] = onlinenum

	data, err := json.Marshal(&msg)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.send <- data
}

// 破产
func (this *Player) Bankrupt(id int) {
	//	OnBankrupt: {recvId: 2060, msg: "recv_OnBankrupt"},
	//data.getup;
	add := 0 //初助数

	t, err := time.Parse("2006-01-02 15:04:05", this.user.Lastsubsidy.Format("2006-01-02 15:04:05"))

	if err != nil {
		//fmt.Println(err)
	} else {
		d := time.Now().Sub(t)
		if d.Hours() > 24 {
			//fmt.Println("可以补助")
			add = 3000

			bm := make(map[string]interface{})
			bm["up"] = this.user.Up + add
			bm["br"] = this.user.Br + 1
			bm["lastsubsidy"] = time.Now()

			err := model.UpdateUser(this.user.ObjectId, bm)
			if err != nil {
				logs.Error("update db error:%s", err.Error())
				return
			}

			this.user.Up += add //补助
			this.user.Br += 1   //破产数
			this.user.Lastsubsidy = time.Now()
			res := make(map[string]interface{})
			res["up"] = this.user.Up
			res["num"] = this.user.Br
			res["getup"] = add

			msg := make(map[string]interface{})
			msg["order"] = 2060
			msg["code"] = 0
			msg["data"] = res

			data, err := json.Marshal(&msg)
			if err != nil {
				logs.Error("json marshal error:%s", err.Error())
			}
			this.send <- data
		}
	}
}

//购买商品
func (this *Player) PayShop(id string) {
	//返回订单号 和 商品id
	// PayShop: {recvId: 2051, msg: "recv_PayShop"},// 购买商品
	//PaySuccess: {recvId: 2052, msg: "recv_PaySuccess"},// 支付成功

	//data:{type    up   getup    ud    getud   }

	goodid, err := strconv.Atoi(id)

	code := 0
	gt := 1 // 货物类型  1 豆  2 钻
	if goodid > 20 {
		gt = 2
	}

	//修改数据,生成订单-----
	r, _ := model.NewRecord(goodid, this.user.Unionid)

	if r != nil { //成功生成订单后,修改数据
		bm := make(map[string]interface{})
		if gt == 1 {
			if this.user.Ud >= r.Price/100 { //扣除 钻石  添加金豆
				this.user.Ud -= (r.Price / 100)
				this.user.Up += r.Num

				bm["up"] = this.user.Up
				bm["ud"] = this.user.Ud

				err := model.UpdateUser(this.user.ObjectId, bm)
				if err != nil {
					logs.Error("exchagne update db error:%s", err.Error())
					code = 1
				}

			} else {
				//返回错误
				logs.Error("user  %s  ud  %d is not enongh!  %d \n", this.user.ObjectId, this.user.Ud, (r.Price / 100))
				errdata := `{"order":"","code":9014,"data":""}`
				this.send <- []byte(errdata)
				return
			}

		} else {
			//扣除 用户对应的虚拟币

			//--------------------
			this.user.Ud += r.Num
			bm["ud"] = this.user.Ud

			err := model.UpdateUser(this.user.ObjectId, bm)
			if err != nil {
				logs.Error("exchagne update db error:%s", err.Error())
				code = 1
			}
		}

	} else {
		//返回失败错误
		logs.Error("record order  created  failed! \n")
		errdata := `{"order":"","code":9014,"data":""}`
		this.send <- []byte(errdata)
		return
	}

	//----------------------

	//返回数据
	msg := make(map[string]interface{})
	msg["order"] = 2052
	msg["code"] = code

	res := make(map[string]interface{})
	// res["id"] = id
	// res["no"] = 100000 //模拟一个订单号
	res["type"] = gt
	if gt == 1 { //修改豆数
		res["up"] = this.user.Up
		res["ud"] = this.user.Ud
		res["getup"] = r.Num
	} else { //修改钻数
		res["ud"] = this.user.Ud
		res["getud"] = r.Num
	}

	msg["data"] = res

	data, err := json.Marshal(&msg)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.send <- data
}

//金币排名
func (this *Player) RankingByUp() {
	// res := make(map[string]interface{})
	// res["cmd"] = "upranking"
	// res["err"] = ""

	// //获取前20名排名
	// rank, err := model.RankingByUp()
	// if err != nil {
	// 	res["err"] = err.Error()
	// }
	// res["ranking"] = rank

	// data, err := json.Marshal(&res)
	// if err != nil {
	// 	logs.Error("json marshal error:%s", err.Error())
	// }
	// this.send <- data
}

// func (this *Player) HeartBeat{

// }

func (this *Player) SetGameScene(id int) {

	msg := make(map[string]interface{})
	res := make(map[string]interface{})
	msg["order"] = "2007"
	code := "0"

	// bm := make(map[string]interface{})
	// bm["img"] = img
	// err := model.UpdateUser(this.user.ObjectId, bm)
	// if err != nil {
	// 	logs.Error("update db error:%s", err.Error())
	// 	code = "1"
	// }
	res["scene"] = id                 //选中的场景
	res["costup"] = 0                 // 花费的金币
	res["up"] = this.user.Up          // 最新的金币
	res["scenes"] = []int{3001, 3002} // 用户已经购买的所有场景

	msg["data"] = res
	msg["code"] = code
	data, err := json.Marshal(&msg)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.send <- data
}

func (this *Player) Close() {
	if this.user != nil {
		//向服务器发送 断开消息
		//fmt.Printf("========  player [ %s ] is close conn ...   ========\n", this.user.Nickname)
		if this.room != nil { //转发房间  有人退出
			this.room.Escape(this.pos)
		}
	} else {
		//fmt.Printf("========  player is close conn ...   ========\n")
	}
}
