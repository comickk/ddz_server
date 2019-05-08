// Room
package server

import (
	"UULoServer/gamerule"
	"UULoServer/lib"
	"UULoServer/logs"
	"UULoServer/model"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"
)

//Player向Room转递过来的消息
type Relay struct {
	pos int
	msg Packet
	//msg map[string]interface{}
}

//玩家位置:      ua      uc
//                 ub

//ua的左侧是uc,右侧是ub
//ub的左侧是ua,右侧是uc
//uc的左侧是ub,右侧是ua

//游戏房间
type Room struct {
	roomid   string     //roomid
	roomtype int        //房间类型
	ua       *Player    //玩家A
	ub       *Player    //玩家B
	uc       *Player    //玩家C
	rmsg     chan Relay //通过Player中继过来的消息,有些消息在Player里处理，有些需要在Room中处理
	//readycout   int        // 准备好的用户个数，当为3时，服务器开始发牌

	//-----------
	mul      Mul //倍数
	mulstate int //加倍状态,计数,三人都加完倍后,进行下一步操作

	firstmingpai int //第一个明牌玩家位置
	landlord     int //地主位置

	spring int //春天   //0初始  1 地主春天  2 农民春天  3 没有春天
	//----------------

	verifycount int //发牌校验，当三个玩家都发来校验信息后，开始叫地主

	backcards [3]int    //底牌
	precards  []int     //上一个人出的牌
	propos    int       //用于记录上一个出牌的玩家
	stop      chan byte //关闭信号

	nextplayer int         //下一个说话的玩家
	waittimer  *time.Timer //等待下个说话玩家的计时器

	isplaying bool //游戏运行标志,用于判断是否为逃跑
}

type Mul struct {
	// 倍数情况
	Init   int //  -- init: 初始倍数
	Vc     int //  -- vc: 明牌倍数
	Grab   int //  -- grab: 抢地主倍数
	Bc     int //  -- bc: 底牌倍数
	Bomb   int //  -- bomb: 炸弹倍数
	Spring int //  -- spring: 春天倍数
	Lo     int //-- lo: 地主倍数
	U1     int //  -- u1 用户1农民倍数
	U2     int //  -- u2 用户2农民情况
	U3     int //  -- u3 用户3农民情况
}

func NewRoom(ua, ub, uc *Player, rt int) *Room {
	onlinenum[rt] += 3
	mul := Mul{Init: 1, Vc: 1, Grab: 1, Bc: 1, Bomb: 1, Spring: 1, Lo: 1, U1: 1, U2: 1, U3: 1}
	return &Room{roomid: lib.NewObjectId(), roomtype: rt, ua: ua, ub: ub, uc: uc, rmsg: make(chan Relay), mul: mul,
		verifycount: 0, firstmingpai: 0, landlord: 0, propos: 0, spring: 0, nextplayer: 0, waittimer: nil,
		mulstate: 0, isplaying: false, stop: make(chan byte, 1)}
}
func (this *Room) Run() {
	go func() {
		//发送匹配的用户
		this.sendMatchPlayer()
		for {
			select {
			case <-this.stop:
				//fmt.Println("---房间已关闭---")
				this.CloseRoom()
				return
			//Player中继过来了消息
			case rg, ok := <-this.rmsg:
				if ok {
					//fmt.Printf("房间处理数据 <---  %s \n", rg.msg.Id)
					//if cmd, ok := rg.msg["cmd"].(string); ok {
					if rg.msg.Id > 0 {
						switch rg.msg.Id {
						/* 该消息在player里处理，修改为userready时，玩家加入匹配队列
						//用户准备开始游戏
						case "userready":
							this.userReady(rg.pos)
							this.readycout++
							if this.readycout == 3 {
								this.dealCards(0)
							}*/

						//发牌验证
						case 0001: //"verifycards":
							this.verifycount++
							if this.verifycount == 3 {
								//三个玩家都发来了验证消息，开始叫地主
								this.sendFirstCaller()
							}

						case 1031:
							//DisplayPoker: {sendId: 1031, msg: "send_DisplayPoker"},// 明牌
							//data {"mul":倍数}
							if data, ok := rg.msg.Data.(map[string]interface{}); ok {
								if mul, ok := data["mul"].(float64); ok {
									this.DisplayPoker(rg.pos, (int)(mul))
								}
							}

						case 1032:
							//SendPokerOver: {sendId: 1032, msg: "send_SendPokerOver"},// 发完牌之后校验
							//分配第一个叫地主的玩家
							this.verifycount++
							if this.verifycount >= 3 {
								// if this.waittimer != nil {
								// 	this.waittimer.Stop()
								// }
								this.sendFirstCaller()
							}
						//ShoutPoker: {recvId: 2035, msg: "recv_ShoutPoker"},// 叫牌

						//叫地主
						case 1033: //"calllandlord":
							//ShoutLandlord: {sendId: 1033, msg: "send_ShoutLandlord"},// 叫地主
							//{"id":1033,"msg":"send_ShoutLandlord","data":0}

							if rg.pos == this.nextplayer {
								// if this.waittimer != nil {
								// 	this.waittimer.Stop()
								// }
								if cv, ok := rg.msg.Data.(float64); ok {
									iscall := int(cv)
									this.callLandlord(rg.pos, int(iscall))
								}
							} else {
								logs.Error("player %d  is not  next player", rg.pos)
							}

						//抢地主
						case 1034: //"grablandlord":
							//RobLandlord: {sendId: 1034, msg: "send_RobLandlord"},// 抢地主
							if rg.pos == this.nextplayer {
								// if this.waittimer != nil {
								// 	this.waittimer.Stop()
								// }
								if cv, ok := rg.msg.Data.(float64); ok {
									isgrab := int(cv)
									this.grabLandlord(rg.pos, isgrab)
								}
							} else {
								logs.Error("player %d  is not  next player", rg.pos)
							}

						case 1035: //{"id":1035,"msg":"send_ShoutDouble","data":0/1/2} //加倍

							if mul, ok := rg.msg.Data.(float64); ok {
								this.ShoutDouble(rg.pos, (int)(mul))
							}

						//出牌
						case 1036:
							//PutPoker: {sendId: 1036, msg: "send_PutPoker"},// 出牌
							//ex {"id":1036,"msg":"send_PutPoker","data":[305,205,105,404]}

							if rg.pos == this.nextplayer {
								// if this.waittimer != nil {
								// 	this.waittimer.Stop()
								// }
								if cv, ok := rg.msg.Data.([]interface{}); ok {
									this.PopCards(rg.pos, cv)
								} else {
									logs.Error("put poker card value param error")
								}
							} else {
								logs.Error("player %d  is not  next player", rg.pos)
							}

						//离开房间
						case 1022: //"leaveroom":
							this.LeaveRoom(rg.pos)

						//固定消息聊天
						case 1061:
							// Chat: {sendId: 1061, msg: "send_Chat"},// 聊天
							if data, ok := rg.msg.Data.(map[string]interface{}); ok {
								if id, ok := data["id"].(float64); ok {
									this.Chat(rg.pos, (int)(id))
								}
							}
						}

					} else {
						logs.Error("cmd type is not string")
					}
				}
			}
		}
	}()
}

//将要发送的数据打包成客户端一样的格式
func (this *Room) packData(order int64, code string, res *map[string]interface{}) []byte {
	msg := make(map[string]interface{})
	msg["order"] = order
	msg["code"] = code
	msg["data"] = res

	data, err := json.Marshal(&msg)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
		return nil
	}
	return data
}

func (this *Room) SendToRoom(order int64, code string, res *map[string]interface{}) {
	msg := make(map[string]interface{})
	msg["order"] = order
	msg["code"] = code
	msg["data"] = res

	data, err := json.Marshal(&msg)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}

	this.ua.send <- data
	this.ub.send <- data
	this.uc.send <- data
}

//快速匹配后向玩家推送匹配的玩家
func (this *Room) sendMatchPlayer() {
	//向三个玩家推送匹配的用户
	//DeskEnterUser: {recvId: 2031, msg: "recv_DeskEnterUser"},//牌桌匹配到用户

	this.Reset()

	code := "0"
	res := make(map[string]interface{})
	//res["room"] = this.roomid
	res["room"] = this.roomtype

	//向ua推送,左侧是uc,右侧是ub
	wr := fmt.Sprintf("%.1f/%%", float32(this.uc.user.Wn)/float32(this.uc.user.Ct)*100)
	res["u1"] = map[string]interface{}{"name": this.uc.user.Nickname,
		"up":   this.uc.user.Up,
		"ct":   this.uc.user.Ct,
		"wr":   wr,
		"cw":   this.uc.user.Cw,
		"isr":  1,
		"img":  this.uc.user.Img, //------------
		"pos":  this.uc.pos,      //------------------------
		"head": this.uc.user.Headimgurl}

	wr = fmt.Sprintf("%.1f/%%", float32(this.ub.user.Wn)/float32(this.ub.user.Ct)*100)
	res["u2"] = map[string]interface{}{"name": this.ub.user.Nickname,
		"up":   this.ub.user.Up,
		"ct":   this.ub.user.Ct,
		"wr":   wr,
		"cw":   this.ub.user.Cw,
		"isr":  1,
		"img":  this.ub.user.Img,
		"pos":  this.ub.pos, //-----------------------------
		"head": this.ub.user.Headimgurl}

	//this.SendToRoom(2031, code, &res)
	this.ua.send <- this.packData(2031, code, &res)

	//向ub推送,左侧是ua,右侧是uc
	wr = fmt.Sprintf("%.1f/%%", float32(this.ua.user.Wn)/float32(this.ua.user.Ct)*100)
	res["u1"] = map[string]interface{}{"name": this.ua.user.Nickname, "up": this.ua.user.Up,
		"ct": this.ua.user.Ct, "wr": wr, "cw": this.ua.user.Cw, "isr": 0, "img": this.ua.user.Img, "pos": this.ua.pos, "head": this.ua.user.Headimgurl}

	wr = fmt.Sprintf("%.1f/%%", float32(this.uc.user.Wn)/float32(this.uc.user.Ct)*100)
	res["u2"] = map[string]interface{}{"name": this.uc.user.Nickname, "up": this.uc.user.Up,
		"ct": this.uc.user.Ct, "wr": wr, "cw": this.uc.user.Cw, "isr": 0, "img": this.uc.user.Img, "pos": this.uc.pos, "head": this.uc.user.Headimgurl}

	this.ub.send <- this.packData(2031, code, &res)

	//向uc推送,左侧是ub,右侧是ua
	wr = fmt.Sprintf("%.1f/%%", float32(this.ub.user.Wn)/float32(this.ub.user.Ct)*100)
	res["u1"] = map[string]interface{}{"name": this.ub.user.Nickname, "up": this.ub.user.Up,
		"ct": this.ub.user.Ct, "wr": wr, "cw": this.ub.user.Cw, "isr": 0, "img": this.ub.user.Img, "pos": this.ub.pos, "head": this.ub.user.Headimgurl}

	wr = fmt.Sprintf("%.1f/%%", float32(this.ua.user.Wn)/float32(this.ua.user.Ct)*100)
	res["u2"] = map[string]interface{}{"name": this.ua.user.Nickname, "up": this.ua.user.Up,
		"ct": this.ua.user.Ct, "wr": wr, "cw": this.ua.user.Cw, "isr": 0, "img": this.ua.user.Img, "pos": this.ua.pos, "head": this.ua.user.Headimgurl}

	this.uc.send <- this.packData(2031, code, &res)

	//开始发牌
	this.dealCards(0)
}

//某个玩家准备好了, 1为左侧用户，2为右侧用户，3为自己 (已废,不用准备,匹配到三名玩家后直接开始)
func (this *Room) userReady(pos int) {
	var res string
	res = `{"cmd":"userready","err":"",user":`
	switch pos {
	//ua玩家准备好了
	case 1:
		//向ua发送
		ret := fmt.Sprintf("%s%d}", res, 3)
		this.ub.send <- []byte(ret)
		//向ub发送
		ret = fmt.Sprintf("%s%d}", res, 1)
		this.ub.send <- []byte(ret)
		//向uc发送
		ret = fmt.Sprintf("%s%d}", res, 2)
		this.uc.send <- []byte(ret)

	//ub玩家准备好了
	case 2:
		//向ua发送
		ret := fmt.Sprintf("%s%d}", res, 2)
		this.ub.send <- []byte(ret)
		//向ub发送
		ret = fmt.Sprintf("%s%d}", res, 3)
		this.ub.send <- []byte(ret)
		//向uc发送
		ret = fmt.Sprintf("%s%d}", res, 1)
		this.uc.send <- []byte(ret)

	//uc玩家准备好了
	case 3:
		//向ua发送
		ret := fmt.Sprintf("%s%d}", res, 1)
		this.ub.send <- []byte(ret)
		//向ub发送
		ret = fmt.Sprintf("%s%d}", res, 2)
		this.ub.send <- []byte(ret)
		//向uc发送
		ret = fmt.Sprintf("%s%d}", res, 3)
		this.uc.send <- []byte(ret)
	}

}

//三个玩家都准备好了，服务器发牌
//again 是否是从发的牌
func (this *Room) dealCards(again int) {

	//{num: 进行局数,用于钻场,金币场不需要,}
	//
	// SendPoker: {recvId: 2032, msg: "recv_SendPoker"},// 发牌
	//ReSendPoker: {recvId: 2042, msg: "recv_ReSendPoker"},// 重新发牌
	msg := make(map[string]interface{})

	mi := 1 //初始倍数
	cfg := model.GetDetialCfg(this.roomtype)
	mi = cfg.Initmul
	this.mul.Init = mi

	if again == 0 {
		msg["order"] = "2032"
		//扣除门票
		ticket := cfg.Ticket

		this.ua.user.Up -= ticket
		this.ub.user.Up -= ticket
		this.uc.user.Up -= ticket

		this.isplaying = true //游戏开始了,此后再退出 视为 逃跑

		//更新数据库
		//--------------------
	} else {
		msg["order"] = "2042"
	}
	msg["code"] = "0"

	//{"order":  ,"code":,"data":{"u1":,"u2":,"u3":,"type":,"up":{}}}

	this.ua.cards = make([]int, 17)
	this.ub.cards = make([]int, 17)
	this.uc.cards = make([]int, 17)

	up := make(map[string]interface{})
	cards := gamerule.GenRandCards()
	//每个人17张，留三张作为底牌
	res := make(map[string]interface{})
	res["type"] = "1"

	res["num"] = 0
	res["u1"] = cards[0:0]
	res["u2"] = cards[0:0]

	m := make(map[string]interface{})
	m["init"] = mi
	res["mul"] = m

	//向ua发牌----------------------------------
	copy(this.ua.cards, cards[0:17])
	res["u3"] = cards[0:17]

	up["u1"] = this.uc.user.Up
	up["u2"] = this.ub.user.Up
	up["u3"] = this.ua.user.Up
	res["up"] = up
	msg["data"] = res
	data, err := json.Marshal(&msg)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.ua.send <- data

	//向ub发牌--------------------------------
	//this.ub.cards = cards[17:34]
	copy(this.ub.cards, cards[17:34])
	res["u3"] = cards[17:34]

	up["u1"] = this.ua.user.Up
	up["u2"] = this.uc.user.Up
	up["u3"] = this.ub.user.Up
	res["up"] = up
	msg["data"] = res
	data, err = json.Marshal(&msg)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.ub.send <- data

	//向uc发牌------------------------------
	//this.uc.cards = cards[34:51]
	copy(this.uc.cards, cards[34:51])
	res["u3"] = cards[34:51]

	up["u1"] = this.ub.user.Up
	up["u2"] = this.ua.user.Up
	up["u3"] = this.uc.user.Up
	res["up"] = up
	msg["data"] = res
	data, err = json.Marshal(&msg)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.uc.send <- data

	//设置底牌
	this.backcards[0] = cards[51]
	this.backcards[1] = cards[52]
	this.backcards[2] = cards[53]

	//是否有人明牌开始
	if this.ua.isvcstart {
		this.DisplayPoker(this.ua.pos, 5)
	}
	if this.ub.isvcstart {
		this.DisplayPoker(this.ub.pos, 5)
	}
	if this.uc.isvcstart {
		this.DisplayPoker(this.uc.pos, 5)
	}

	//超时处理
	//this.waittimer = time.AfterFunc(time.Second*dealcard, this.sendFirstCaller)
}

//玩家明牌
func (this *Room) DisplayPoker(pos, mul int) {
	//type 1,2,3,4 card:    user 坐位 1左  2 右  3自己
	// data = {type:1,user: 2, mul:1,card: [107, 207, 307, 407, 108, 208, 308, 408, 206]};
	//DisplayPoker: {recvId: 2033, msg: "recv_DisplayPoker"},// 明牌

	if this.firstmingpai == 0 {
		this.firstmingpai = pos
	}

	this.mul.Vc *= mul

	code := "0"
	var res = make(map[string]interface{})
	res["type"] = 1

	var m = make(map[string]interface{})
	m["vc"] = this.mul.Vc

	res["mul"] = m

	res["card"] = this.GetPlayer(pos).cards
	res["user"] = pos
	this.SendToRoom(2033, code, &res)
}

//sendFirstCaller  服务器随机第一个叫地主的玩家
func (this *Room) sendFirstCaller() {
	//this.verifycount=0
	rand.Seed(time.Now().Unix())
	n := rand.Intn(3) + 1

	//有人明牌,明牌玩家先叫
	if this.firstmingpai > 0 {
		n = this.firstmingpai
	}

	this.nextplayer = n

	//this.waittimer = time.AfterFunc(time.Second*shoutLandlord, func() { this.callLandlord(n, 0) })

	var res string
	//res = `{"cmd":"caller","err":"","user":`
	res = fmt.Sprintf(`{"order":2035,"code":"0","data":%d}`, n)

	this.ua.send <- []byte(res)
	this.ub.send <- []byte(res)
	this.uc.send <- []byte(res)
}

// callLandlord  玩家叫地主
// pos 具体哪个玩家
// call 该玩家是否叫地主 1 叫 0 不叫
func (this *Room) callLandlord(pos, call int) {
	//ShoutLandlord: {recvId: 2036, msg: "recv_ShoutLandlord"},// 叫地主
	//{order :  , code:  , data: {user:   ,call:   ,nt:  ,ntdo:  }}

	code := "0"
	var res = make(map[string]interface{})
	res["call"] = call

	curr := this.GetPlayer(pos)
	curr.iscall = call

	if call == 1 { //叫
		this.landlord = pos //确定临时地主位
		res["ntdo"] = "g"   //进入抢地主阶段

		//查看有抢地主机会的玩家

		if this.ua.iscall > -1 && this.ub.iscall > -1 && this.uc.iscall > -1 { //无抢地主机会
			this.sendLandlordOwer()
			return
		} else {
			//有抢地主机会时,重置地主 iscall 为-1,使其能反抢一次
			curr.iscall = -1
		}

	} else { //未叫
		res["ntdo"] = "c" //继续叫地主

		//都未叫
		if this.ua.iscall == 0 && this.ub.iscall == 0 && this.uc.iscall == 0 {

			if this.firstmingpai > 0 { //如有明牌玩家,明牌玩家为地主
				this.landlord = this.firstmingpai
				this.sendLandlordOwer()
				return

			} else { //无明牌玩家 重置牌局
				this.Reset()
				this.dealCards(1) //重新发牌
				return
			}
		}
	}

	next := this.GetNextPlayer(pos)
	res["user"] = pos    //叫地主的是ua
	res["nt"] = next.pos //ub执行下一步操作
	this.SendToRoom(2036, code, &res)

	this.nextplayer = next.pos
	//this.waittimer = time.AfterFunc(time.Second*robLandlord, func() { this.grabLandlord(this.nextplayer, 0) })
}

//grabLandlord  玩家抢地主
// pos 具体哪个玩家
// grab 该玩家是否抢地主 1 抢 0 不抢
func (this *Room) grabLandlord(pos, grab int) {

	//RobLandlord: {recvId: 2037, msg: "recv_RobLandlord"},// 抢地主

	var res = make(map[string]interface{})
	code := "0"
	res["grab"] = grab

	curr := this.GetPlayer(pos)
	if grab == 1 { //抢地主
		this.landlord = pos
		curr.iscall = 1
		this.mul.Grab *= 2
	} else { //不抢
		curr.iscall = 0
	}

	//取得下一个可抢地主的人
	n := this.getNextOperator(pos)

	if n == 0 { //没人有抢的机会,确定地主
		this.sendLandlordOwer()
		return
	}

	if grab == 0 && n == this.landlord { //没人抢地主,下一个可抢的是地主,确定地主
		this.sendLandlordOwer()
		return
	}

	//还有人能抢,则继续抢地主

	var mul = make(map[string]interface{})
	mul["grab"] = this.mul.Grab
	res["mul"] = mul

	res["user"] = pos //抢地主的是
	res["nt"] = n

	this.SendToRoom(2037, code, &res)

	this.nextplayer = n
	//this.waittimer = time.AfterFunc(time.Second*robLandlord, func() { this.grabLandlord(this.nextplayer, 0) })
}

//getNextOperator   判断下一步操作的人
//返回值 1 ua 2 ub 3 uc 0 没有
func (this *Room) getNextOperator(pos int) int {
	//var n = 0

	next := this.GetNextPlayer(pos)
	if next.iscall == -1 {
		return next.pos
	} else {
		next = this.GetNextPlayer(next.pos)
		if next.iscall == -1 {
			return next.pos
		} else {
			return 0
		}
	}
}

//sendLandlordOwer 亮底牌，通知谁是地主
func (this *Room) sendLandlordOwer() {
	//EnsureLandlord: {recvId: 2038, msg: "recv_EnsureLandlord"},// 确定地主
	var res = make(map[string]interface{})
	code := "0"
	res["bc"] = this.backcards

	//确定底牌倍数
	this.mul.Bc = this.GetBackCardMul()

	mul := make(map[string]interface{})
	mul["bc"] = this.mul.Bc
	res["mul"] = mul

	var lordpos = this.landlord
	res["lo"] = lordpos //who是地主

	this.ua.utype = 2
	this.ub.utype = 2
	this.uc.utype = 2

	lord := this.GetPlayer(lordpos)
	lord.utype = 1
	lord.cards = append(lord.cards, this.backcards[:]...)

	this.SendToRoom(2038, code, &res)

	//处理确定 地主后  超时操作
	this.nextplayer = lordpos

	//处理超时加倍
	// this.waittimer = time.AfterFunc(time.Second*(double), func() {
	// 	this.mulstate = 3
	// 	this.ShoutDouble(0, 0)
	// })
	//this.waittimer = time.AfterFunc(time.Second*(double+putCard), func() { this.PopCards(this.nextplayer, this.AutoPutCard(this.nextplayer)) })
}

//ShoutDouble  玩家加倍
func (this *Room) ShoutDouble(pos int, mul int) {
	//return ShoutDouble: {recvId: 2039, msg: "recv_ShoutDouble"},// 加倍
	//{user   type  ntdo  mul}

	this.mulstate++
	var res = make(map[string]interface{})
	code := "0"

	var m = make(map[string]interface{})

	if mul == 0 {
		res["type"] = 1
	} else {
		res["type"] = 2
	}
	mul++

	if this.mulstate >= 3 {
		// if this.waittimer != nil {
		// 	this.waittimer.Stop()
		// }

		res["ntdo"] = 6
	} else {
		res["ntdo"] = 0
	}

	if pos > 0 {
		res["user"] = pos //加倍是

		if pos == this.landlord {
			this.mul.Lo = mul
			m["lo"] = this.mul.Lo
		} else {
			switch pos {
			case 1:
				this.mul.U1 = mul
				m["u1"] = mul
			case 2:
				this.mul.U2 = mul
				m["u2"] = mul
			case 3:
				this.mul.U3 = mul
				m["u3"] = mul
			}
		}

		res["mul"] = m
	}

	this.SendToRoom(2039, code, &res)

	//地主出牌超时处理
	//this.waittimer = time.AfterFunc(time.Second*putCard, func() { this.PopCards(this.nextplayer, this.AutoPutCard(this.nextplayer)) })
}

//玩家出牌
//cv 玩家出的牌
func (this *Room) PopCards(pos int, cv []interface{}) {
	//PlayerOutPoker: {recvId: 2040, msg: "recv_PlayerOutPoker"},

	var res = make(map[string]interface{})
	//{user   card   nt  num 剩余牌数  mul}
	code := "0"

	var mul = make(map[string]interface{})

	//将[]interface{} 转为 []int
	cards := make([]int, len(cv))
	for i, v := range cv {
		cards[i] = int(v.(float64))
	}

	//自动出牌标志
	af := 0
	curr := this.GetPlayer(pos)
	if this.propos == pos { //上次出牌的人是自己,此次是重新出牌
		this.precards = nil
		af = 0 // 下家自动 出牌 时  直接不出
	}

	nt := this.GetNextPlayerPos(pos) //该who出牌

	pcn := len(cards)
	if pcn != 0 {

		//判断牌类型
		myType, myValue, myNv := gamerule.GetCardsType(cards)

		if myType == gamerule.ERROR_CARD {
			code = "牌类型错误"

			curr.send <- this.packData(2040, code, &res)
			return
		} else {
			//判断是否大于上家出的牌
			if len(this.precards) > 0 {
				prevType, prevValue, preNv := gamerule.GetCardsType(this.precards)
				over := gamerule.IsOvercomePrev(myType, myValue, myNv, prevType, prevValue, preNv)

				if !over {
					//没有上家的牌大
					code = "没有上家的牌大"
					curr.send <- this.packData(2040, code, &res)
					return
				}
			}
		}

		if myType == gamerule.BOMB_CARD || myType == gamerule.ROCKET_CARD {
			this.mul.Bomb *= 2
			mul["bomb"] = this.mul.Bomb
			res["mul"] = mul
		}
	}

	//设置上次玩家出的牌
	if pcn != 0 {
		this.precards = cards

		if nt == this.propos { //如果上次出牌的人是下一个人
			af = 1 //此人至少要出一张
		}

		this.propos = pos

		if curr.utype == 1 { //地主出牌
			this.spring += 100
		} else { //农民出牌
			this.spring += 1
		}
	}

	//通知其他人出牌情况-----------------------
	res["cards"] = cards

	hcn := len(curr.cards)

	//num := len(curr.cards) - len(cards) //手中剩余牌数
	num := hcn - pcn //手中剩余牌数

	curr.cards = gamerule.Difference(curr.cards, cards)

	res["num"] = num
	res["user"] = pos //who出的牌

	res["nt"] = nt

	if num == 0 {
		res["nt"] = 0 //结束出牌
		this.SendToRoom(2040, code, &res)

		//结算
		this.Settlement(pos)
	} else {
		this.SendToRoom(2040, code, &res)

		//下一个 的超时处理
		this.nextplayer = nt
		if af == 0 {
			//非第一手牌 不要,

			//this.waittimer = time.AfterFunc(time.Second*putCard, func() { this.PopCards(this.nextplayer, nil) })

		} else {
			//第一手牌 出一张
			//this.waittimer = time.AfterFunc(time.Second*putCard, func() { this.PopCards(this.nextplayer, this.AutoPutCard(this.nextplayer)) })
		}
	}
}

//结算
// pos 赢的玩家
func (this *Room) Settlement(pos int) {
	//GameOver_NormalMatch: {recvId: 2041, msg: "recv_GameOver_NormalMatch"},

	winner := this.GetPlayer(pos)
	var res = make(map[string]interface{})
	code := "0"

	reward := make(map[string]interface{})

	var rua, rub, ruc int

	//查看春天
	if this.spring < 200 || this.spring%100 < 1 {
		this.mul.Spring = 2
	}

	//取得倍数
	mul := this.mul.Init * this.mul.Vc * this.mul.Bc * this.mul.Grab * this.mul.Spring * this.mul.Bomb
	//this.mul.Lo *this.mul.U1 * this.mul.U2 * this.mul.U3

	// Underpoint    int `xorm:"underpoint"`    //底分
	// Initmul       int `xorm:"initmul"`       //初始倍数
	// Ticket        int `xorm:"ticket"`        //门票
	// Maxearn       int `xorm:"maxearn"`       //收益封顶

	cfg := model.GetDetialCfg(this.roomtype)
	base := cfg.Underpoint * mul * this.mul.Lo

	//计算金币
	//连胜次数更新
	if winner.utype == 1 { //地主胜
		switch winner.pos {
		case 1:
			//top := this.ua.user.Up / 2 //计算上限
			top := this.ua.user.Up //计算上限
			if top > cfg.Maxearn {
				top = cfg.Maxearn
			}

			rub = -UpLimit(base*this.mul.U2, this.ub.user.Up, top)
			ruc = -UpLimit(base*this.mul.U3, this.uc.user.Up, top)
			rua = 0 - rub - ruc

			this.ua.user.Cw++
			this.ub.user.Cw = 0
			this.uc.user.Cw = 0

		case 2:
			//top := this.ub.user.Up / 2 //计算上限
			top := this.ub.user.Up //计算上限
			if top > cfg.Maxearn {
				top = cfg.Maxearn
			}

			rua = -UpLimit(base*this.mul.U1, this.ua.user.Up, top)
			ruc = -UpLimit(base*this.mul.U3, this.uc.user.Up, top)
			rub = 0 - ruc - rua

			this.ua.user.Cw = 0
			this.ub.user.Cw++
			this.uc.user.Cw = 0

		case 3:
			//top := this.uc.user.Up / 2 //计算上限
			top := this.uc.user.Up //计算上限
			if top > cfg.Maxearn {
				top = cfg.Maxearn
			}
			rua = -UpLimit(base*this.mul.U1, this.ua.user.Up, top)
			rub = -UpLimit(base*this.mul.U2, this.ub.user.Up, top)
			ruc = 0 - rub - rua

			this.ua.user.Cw = 0
			this.ub.user.Cw = 0
			this.uc.user.Cw++
		}
	} else { //地主负 农民胜
		//
		switch this.landlord {
		case 1:
			top := this.ua.user.Up //计算赔付上限
			if top > cfg.Maxearn {
				top = cfg.Maxearn
			}
			//
			v := base*this.mul.U2 + base*this.mul.U3                    //理论赔付额
			losttop := UpLimit(v, this.ub.user.Up+this.uc.user.Up, top) //实际赔付额

			if v > losttop { //赔付额不足 按各自倍数比例分配
				rub = int(math.Floor(float64(losttop) * (float64(this.mul.U2) / float64(this.mul.U2+this.mul.U3)))) //* ((mul*this.mul.U2) /(mul*this.mul.U2 * this.mul.U3 )) )
				ruc = losttop - rub
			} else {
				rub = UpLimit(base*this.mul.U2, this.ub.user.Up, losttop)
				ruc = losttop - rub //UpLimit(base*this.mul.U3, this.uc.user.Up, losttop)
			}
			rua = 0 - rub - ruc

			this.ua.user.Cw = 0
			this.ub.user.Cw++
			this.uc.user.Cw++

		case 2:
			top := this.ub.user.Up //计算赔付上限
			if top > cfg.Maxearn {
				top = cfg.Maxearn
			}
			//
			v := base*this.mul.U1 + base*this.mul.U3
			losttop := UpLimit(v, this.ua.user.Up+this.uc.user.Up, top)

			if v > losttop { //赔付额不足 平分//按各自倍数比例分配
				rua = int(math.Floor(float64(losttop) * (float64(this.mul.U1) / float64(this.mul.U1+this.mul.U3)))) //* ((mul*this.mul.U2) /(mul*this.mul.U2 * this.mul.U3 )) )
				ruc = losttop - rua
			} else {
				rua = UpLimit(base*this.mul.U1, this.ua.user.Up, losttop)
				ruc = losttop - rua //UpLimit(base*this.mul.U3, this.uc.user.Up, losttop)
			}
			rub = 0 - rua - ruc

			this.ua.user.Cw++
			this.ub.user.Cw = 0
			this.uc.user.Cw++

		case 3:
			top := this.uc.user.Up //计算赔付上限
			if top > cfg.Maxearn {
				top = cfg.Maxearn
			}
			//
			v := base*this.mul.U1 + base*this.mul.U2
			losttop := UpLimit(v, this.ua.user.Up+this.ub.user.Up, top)

			if v > losttop { //赔付额不足 平分//按各自倍数比例分配
				rua = int(math.Floor(float64(losttop) * (float64(this.mul.U1) / float64(this.mul.U2+this.mul.U1)))) //* ((mul*this.mul.U2) /(mul*this.mul.U2 * this.mul.U3 )) )
				rub = losttop - rua
			} else {
				rua = UpLimit(base*this.mul.U1, this.ua.user.Up, losttop)
				rub = losttop - rua //UpLimit(base*this.mul.U2, this.ub.user.Up, losttop)
			}
			ruc = 0 - rua - rub

			this.ua.user.Cw++
			this.ub.user.Cw++
			this.uc.user.Cw = 0
		}
	}

	this.ua.user.Up += rua
	this.ub.user.Up += rub
	this.uc.user.Up += ruc

	//胜利次数
	winner.user.Wn++

	//对局次数
	this.ua.user.Ct++
	this.ub.user.Ct++
	this.uc.user.Ct++

	//更新数据库
	//更新ua
	bm := make(map[string]interface{})
	bm["up"] = this.ua.user.Up
	bm["ct"] = this.ua.user.Ct
	bm["cw"] = this.ua.user.Cw
	bm["wn"] = this.ua.user.Wn

	err := model.UpdateUser(this.ua.user.ObjectId, bm)
	if err != nil {
		logs.Error("update db error:%s", err.Error())
	}
	//更新ub
	bm["up"] = this.ub.user.Up
	bm["ct"] = this.ub.user.Ct
	bm["cw"] = this.ub.user.Cw
	bm["wn"] = this.ub.user.Wn

	err = model.UpdateUser(this.ub.user.ObjectId, bm)
	if err != nil {
		logs.Error("gameover update db error 1:%s", err.Error())
	}
	//更新uc
	bm["up"] = this.uc.user.Up
	bm["ct"] = this.uc.user.Ct
	bm["cw"] = this.uc.user.Cw
	bm["wn"] = this.uc.user.Wn

	err = model.UpdateUser(this.uc.user.ObjectId, bm)
	if err != nil {
		logs.Error("gameover update db error 2:%s", err.Error())
	}

	m := make(map[string]interface{})
	m["spring"] = this.mul.Spring
	res["mul"] = m

	res["win"] = winner.utype
	reward["u1"] = rua
	reward["u2"] = rub
	reward["u3"] = ruc
	res["reward"] = reward
	res["c1"] = this.ua.cards
	res["c2"] = this.ub.cards
	res["c3"] = this.uc.cards
	res["u1"] = this.ua.user.Up
	res["u2"] = this.ub.user.Up
	res["u3"] = this.uc.user.Up

	this.SendToRoom(2041, code, &res)

	this.isplaying = false
	//房间重置
	//	this.readycout = 0
	this.ResetRoom()
}

//离开房间
func (this *Room) LeaveRoom(pos int) {
	//{"id":1022,"msg":"send_LeaveHome","data":{}}
	//LeaveHome: {recvId: 2022, msg: "recv_LeaveHome"}

	this.ua.isvcstart = false
	this.ub.isvcstart = false
	this.uc.isvcstart = false

	if this.isplaying {
		this.Escape(pos)
		return
	}
	var res string
	//res = `{"cmd":"caller","err":"","user":`
	//DeskUserLeave: {recvId: 2043, msg: "recv_DeskUserLeave"},// 用户离开

	switch pos {
	case 1:
		this.ua.LeaveRoom()
		res = `{"order":2043,"code":"0","data":{"user":1}}`
		this.ub.send <- []byte(res)
		res = `{"order":2043,"code":"0","data":{"user":2}}`
		this.uc.send <- []byte(res)
	case 2:
		this.ub.LeaveRoom()
		res = `{"order":2043,"code":"0","data":{"user":2}}`
		this.ua.send <- []byte(res)
		res = `{"order":2043,"code":"0","data":{"user":1}}`
		this.uc.send <- []byte(res)
	case 3:
		this.uc.LeaveRoom()
		res = `{"order":2043,"code":"0","data":{"user":1}}`
		this.ua.send <- []byte(res)
		res = `{"order":2043,"code":"0","data":{"user":2}}`
		this.ub.send <- []byte(res)
	}
}

//玩家逃跑
func (this *Room) Escape(pos int) {
	var res string
	//res = `{"cmd":"caller","err":"","user":`
	//DeskUserLeave: {recvId: 2043, msg: "recv_DeskUserLeave"},// 用户离开

	p := this.GetPlayer(pos)

	name := p.user.Nickname

	//返还门票
	ticket := model.GetDetialCfg(this.roomtype).Ticket

	switch pos {
	case 1:
		this.ua.LeaveRoom()

		res = fmt.Sprintf(`{"order":2043,"code":"0","data":{"user":1,"escape":"%s","ticket":%d}}`, name, ticket)
		this.ub.user.Up += ticket
		this.ub.send <- []byte(res)
		res = fmt.Sprintf(`{"order":2043,"code":"0","data":{"user":2,"escape":"%s","ticket":%d}}`, name, ticket)
		this.uc.user.Up += ticket
		this.uc.send <- []byte(res)
	case 2:
		this.ub.LeaveRoom()

		res = fmt.Sprintf(`{"order":2043,"code":"0","data":{"user":2,"escape":"%s","ticket":%d}}`, name, ticket)
		this.ua.user.Up += ticket
		this.ua.send <- []byte(res)
		res = fmt.Sprintf(`{"order":2043,"code":"0","data":{"user":1,"escape":"%s","ticket":%d}}`, name, ticket)
		this.uc.user.Up += ticket
		this.uc.send <- []byte(res)
	case 3:
		this.uc.LeaveRoom()

		res = fmt.Sprintf(`{"order":2043,"code":"0","data":{"user":1,"escape":"%s","ticket":%d}}`, name, ticket)
		this.ua.user.Up += ticket
		this.ua.send <- []byte(res)
		res = fmt.Sprintf(`{"order":2043,"code":"0","data":{"user":2,"escape":"%s","ticket":%d}}`, name, ticket)
		this.ub.user.Up += ticket
		this.ub.send <- []byte(res)
	}

	//只扣除 逃跑玩家的 门票   保存至数据库

	bm := make(map[string]interface{})
	bm["up"] = p.user.Up
	bm["cw"] = 0

	err := model.UpdateUser(p.user.ObjectId, bm)
	if err != nil {
		logs.Error("gameover update db error 3:%s", err.Error())
	}

	// this.ua.room = nil
	// this.ub.room = nil
	// this.uc.room = nil

	// this.ua = nil
	// this.ub = nil
	// this.uc = nil
	//结束对局

	onlinenum[this.roomtype] -= 3
	this.stop <- 1
}

//固定消息聊天
func (this *Room) Chat(pos, id int) {
	//Chat: {recvId: 2061, msg: "recv_Chat"},// 聊天
	//{id:说的什么,user 谁说的 1左,2右}

	msg := make(map[string]interface{})
	msg["order"] = "2061"
	msg["code"] = "0"

	var res = make(map[string]interface{})
	res["id"] = id

	switch pos {
	case 1:
		//发送给ub
		res["user"] = 1
		msg["data"] = res
		data, err := json.Marshal(&msg)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.ub.send <- []byte(data)
		//发送给uc
		res["user"] = 2
		msg["data"] = res
		data, err = json.Marshal(&msg)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.uc.send <- []byte(data)
	case 2:
		//发送给ua
		res["user"] = 2
		msg["data"] = res
		data, err := json.Marshal(&msg)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.ua.send <- []byte(data)
		//发送给uc
		res["user"] = 1
		msg["data"] = res
		data, err = json.Marshal(&msg)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.uc.send <- []byte(data)
	case 3:
		//发送给ua
		res["user"] = 1
		msg["data"] = res
		data, err := json.Marshal(&msg)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.ua.send <- []byte(data)
		//发送给ub
		res["user"] = 2
		msg["data"] = res
		data, err = json.Marshal(&msg)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.ub.send <- []byte(data)
	}
}

//房间 结束房间
func (this *Room) ResetRoom() {

	var res string

	res = `{"order":2043,"code":"0","data":{"user":2}}`
	this.ua.send <- []byte(res)
	res = `{"order":2043,"code":"0","data":{"user":1}}`
	this.ua.send <- []byte(res)

	res = `{"order":2043,"code":"0","data":{"user":1}}`
	this.ub.send <- []byte(res)
	res = `{"order":2043,"code":"0","data":{"user":2}}`
	this.ub.send <- []byte(res)

	res = `{"order":2043,"code":"0","data":{"user":2}}`
	this.uc.send <- []byte(res)
	res = `{"order":2043,"code":"0","data":{"user":1}}`
	this.uc.send <- []byte(res)

	// if this.waittimer != nil {
	// 	this.waittimer.Stop()
	// }

	this.ua.isvcstart = false
	this.ub.isvcstart = false
	this.uc.isvcstart = false

	//this.Reset()

	onlinenum[this.roomtype] -= 3
	this.stop <- 1
}

func (this *Room) Reset() {

	/*
		roomid   string     //roomid
		rmsg     chan Relay //通过Player中继过来的消息,有些消息在Player里处理，有些需要在Room中处理

		backcards [3]int    //底牌
		precards  []int     //上一个人出的牌

		stop      chan byte //关闭信号
		waittimer  *time.Timer //等待下个说话玩家的计时器
		**/

	//this.roomtype = 0
	this.verifycount = 0

	this.propos = 0

	this.landlord = 0
	this.spring = 0

	this.ua.iscall = -1
	this.ub.iscall = -1
	this.uc.iscall = -1

	this.ua.utype = 0
	this.ub.utype = 0
	this.uc.utype = 0

	// this.ua.isvcstart = false
	// this.ub.isvcstart = false
	// this.uc.isvcstart = false

	this.mulstate = 0
	this.mul.Rest()

	this.nextplayer = 0

	this.isplaying = false

	this.firstmingpai = 0

	//this.backcards = nil //底牌
	this.precards = nil //上一个人出的牌

	//this.waittimer.Stop()
}

func (this *Room) CloseRoom() {
	this.ua.room = nil
	this.ub.room = nil
	this.uc.room = nil

	this.ua = nil
	this.ub = nil
	this.uc = nil
}

//计算底牌倍数
func (this *Room) GetBackCardMul() int {
	//取得底牌
	//计算牌型
	//按牌型计算倍数
	if this.backcards[0] > 517 || this.backcards[1] > 517 || this.backcards[2] > 517 {
		//大小王
		return 2
	}

	var d [3]int
	var f [3]int

	for i, v := range this.backcards {
		d[i] = (int)(v / 100.0)
		f[i] = v % 100
	}

	if d[0] == d[1] && d[0] == d[2] { //清一色
		return 3
	}

	if f[0] == f[1] && f[0] == f[2] { //豹子
		return 3
	}

	//排序
	if f[0] > f[1] {
		t := f[0]
		f[0] = f[1]
		f[1] = t
	}

	if f[1] > f[2] {
		t := f[1]
		f[1] = f[2]
		f[2] = t
	}

	if f[0] > f[1] {
		t := f[0]
		f[0] = f[1]
		f[1] = t
	}

	//顺子
	if f[0]+1 == f[1] && f[0]+2 == f[2] {
		return 3
	}

	return 1
}

func (this *Room) GetPlayer(pos int) *Player {
	switch pos {
	case 1:
		return this.ua
	case 2:
		return this.ub
	case 3:
		return this.uc
	}
	return nil
}
func (this *Room) GetNextPlayer(currpos int) *Player {
	switch currpos {
	case 1:
		return this.ub
	case 2:
		return this.uc
	case 3:
		return this.ua
	}
	return nil
}

func (this *Room) GetNextPlayerPos(currpos int) int {
	next := (currpos + 1) % 3
	if next == 0 {
		next = 3
	}
	return next
}

func (this *Room) AutoPutCard(pos int) []interface{} {

	c := make([]interface{}, 1)
	switch pos {
	case 1:
		c[0] = (float64)(this.ua.cards[0])
	case 2:
		c[0] = (float64)(this.ub.cards[0])
	case 3:
		c[0] = (float64)(this.uc.cards[0])
	}

	return c
}

func (this *Mul) Rest() {
	this.Init = 1
	this.Vc = 1     //  -- vc: 明牌倍数
	this.Grab = 1   //  -- grab: 抢地主倍数
	this.Bc = 1     //  -- bc: 底牌倍数
	this.Bomb = 1   //  -- bomb: 炸弹倍数
	this.Spring = 1 //  -- spring: 春天倍数
	this.Lo = 1     //-- lo: 地主倍数
	this.U1 = 1     //  -- u1 用户1农民倍数
	this.U2 = 1     //  -- u2 用户2农民情况
	this.U3 = 1     //  -- u3 用户3农民情况
}

func UpLimit(v, own, max int) int { //v 理论,own 拥有的豆  max  封顶豆   tex 门票(税)

	if v > own {
		if own > max {
			return max
		} else {
			return own
		}
	} else {
		if v > max {
			return max
		} else {
			return v
		}
	}
}
