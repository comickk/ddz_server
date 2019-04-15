// Room
package server

import (
	"UULoServer/gamerule"
	"UULoServer/lib"
	"UULoServer/logs"
	"UULoServer/model"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

//Player向Room转递过来的消息
type Relay struct {
	pos int
	msg map[string]interface{}
}

//玩家位置:      ua      uc
//                 ub

//ua的左侧是uc,右侧是ub
//ub的左侧是ua,右侧是uc
//uc的左侧是ub,右侧是ua

//游戏房间
type Room struct {
	roomid string     //roomid
	ua     *Player    //玩家A
	ub     *Player    //玩家B
	uc     *Player    //玩家C
	rmsg   chan Relay //通过Player中继过来的消息,有些消息在Player里处理，有些需要在Room中处理
	//readycout   int        // 准备好的用户个数，当为3时，服务器开始发牌
	verifycount   int       //发牌校验，当三个玩家都发来校验信息后，开始叫地主
	notcalllocout int       //不叫地主的玩家个数，如果为3，服务器重新发牌
	grabcount     int       //进行抢地主玩家的个数，抢和不抢的都记录
	firstcaller   int       //第一个叫地主的玩家
	firstgraber   int       //第一个抢地主的玩家
	backcards     [3]int    //底牌
	precards      []int     //上一个人出的牌
	propos        int       //用于记录上一个出牌的玩家
	stop          chan byte //关闭信号
}

func NewRoom(ua, ub, uc *Player) *Room {
	return &Room{roomid: lib.NewObjectId(), ua: ua, ub: ub, uc: uc, rmsg: make(chan Relay),
		verifycount: 0, notcalllocout: 0, grabcount: 0, firstcaller: 0, firstgraber: 0, propos: 0,
		stop: make(chan byte, 1)}
}
func (this *Room) Run() {
	go func() {
		//发送匹配的用户
		this.sendMatchPlayer()
		for {
			select {
			case <-this.stop:
				return
			//Player中继过来了消息
			case rg, ok := <-this.rmsg:
				if ok {
					if cmd, ok := rg.msg["cmd"].(string); ok {
						switch cmd {
						/* 该消息在player里处理，修改为userready时，玩家加入匹配队列
						//用户准备开始游戏
						case "userready":
							this.userReady(rg.pos)
							this.readycout++
							if this.readycout == 3 {
								this.dealCards(0)
							}*/

						//发牌验证
						case "verifycards":
							this.verifycount++
							if this.verifycount == 3 {
								//三个玩家都发来了验证消息，开始叫地主
								this.sendFirstCaller()
							}

						//叫地主
						case "calllandlord":
							if cv, ok := rg.msg["call"].(float64); ok {
								iscall := int(cv)

								if iscall == 0 {
									this.notcalllocout++
								} else {
									if this.firstcaller == 0 {
										//记下第一个叫地主的玩家
										this.firstcaller = rg.pos
									}
								}

								//三个玩家都未叫地主
								if this.notcalllocout == 3 {
									//房间重置
									this.ResetRoom()
									//服务器从新发牌
									this.dealCards(1)
								} else if this.notcalllocout == 2 && this.firstcaller != 0 {
									//两个不叫一个叫的，可以直接确定地主
									this.firstgraber = this.firstcaller
									this.sendLandlordOwer()
								} else {
									this.callLandlord(rg.pos, int(iscall))
								}

							} else {
								logs.Error("calllandlord call param error")
							}

						//抢地主
						case "grablandlord":
							if cv, ok := rg.msg["grab"].(float64); ok {
								this.grabcount++
								isgrab := int(cv)

								//记下第一个抢地主的人
								if isgrab == 1 && this.firstgraber == 0 {
									this.firstgraber = rg.pos
								}

								//如果叫地主的人也抢地主,地主是叫地主的玩家
								if isgrab == 1 && this.firstcaller == rg.pos {
									this.firstgraber = this.firstcaller

								}

								//其余两个玩家没有抢地主,地主是叫地主的玩家
								if this.firstgraber == 0 && this.grabcount == 2 {
									this.firstgraber = this.firstcaller

								}

								if this.grabcount+this.notcalllocout == 3 {
									//亮底牌，通知谁是地主
									this.sendLandlordOwer()
								} else {
									//抢地主
									this.grabLandlord(rg.pos, isgrab)
								}

							} else {
								logs.Error("grablandlord call param error")
							}

						//出牌
						case "popcards":

							if cv, ok := rg.msg["cards"].([]interface{}); ok {
								this.PopCards(rg.pos, cv)
							}

						//离开房间
						case "leaveroom":
							this.LeaveRoom(rg.pos)

						//固定消息聊天
						case "chat":
							if id, ok := rg.msg["id"].(float64); ok {
								this.Chat(rg.pos, int(id))
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

//快速匹配后向玩家推送匹配的玩家
func (this *Room) sendMatchPlayer() {
	//向三个玩家推送匹配的用户
	res := make(map[string]interface{})
	res["cmd"] = "matcher"
	res["err"] = ""
	res["room"] = this.roomid

	//向ua推送,左侧是uc,右侧是ub
	wr := fmt.Sprintf("%.1f/%%", float32(this.uc.user.Wn)/float32(this.uc.user.Ct)*100)
	res["u1"] = map[string]interface{}{"name": this.uc.user.Nickname, "up": this.uc.user.Up,
		"ct": this.uc.user.Ct, "wr": wr, "isr": 0, "img": this.uc.user.Img}

	wr = fmt.Sprintf("%.1f/%%", float32(this.ub.user.Wn)/float32(this.ub.user.Ct)*100)
	res["u2"] = map[string]interface{}{"name": this.ub.user.Nickname, "up": this.ub.user.Up,
		"ct": this.ub.user.Ct, "wr": wr, "isr": 0, "img": this.ub.user.Img}

	data, err := json.Marshal(&res)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.ua.send <- data

	//向ub推送,左侧是ua,右侧是uc
	wr = fmt.Sprintf("%.1f/%%", float32(this.ua.user.Wn)/float32(this.ua.user.Ct)*100)
	res["u1"] = map[string]interface{}{"name": this.ua.user.Nickname, "up": this.ua.user.Up,
		"ct": this.ua.user.Ct, "wr": wr, "isr": 0, "img": this.ua.user.Img}

	wr = fmt.Sprintf("%.1f/%%", float32(this.uc.user.Wn)/float32(this.uc.user.Ct)*100)
	res["u2"] = map[string]interface{}{"name": this.uc.user.Nickname, "up": this.uc.user.Up,
		"ct": this.uc.user.Ct, "wr": wr, "isr": 0, "img": this.uc.user.Img}

	data, err = json.Marshal(&res)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.ub.send <- data

	//向uc推送,左侧是ub,右侧是ua
	wr = fmt.Sprintf("%.1f/%%", float32(this.ub.user.Wn)/float32(this.ub.user.Ct)*100)
	res["u1"] = map[string]interface{}{"name": this.ub.user.Nickname, "up": this.ub.user.Up,
		"ct": this.ub.user.Ct, "wr": wr, "isr": 0, "img": this.ub.user.Img}

	wr = fmt.Sprintf("%.1f/%%", float32(this.ua.user.Wn)/float32(this.ua.user.Ct)*100)
	res["u2"] = map[string]interface{}{"name": this.ua.user.Nickname, "up": this.ua.user.Up,
		"ct": this.ua.user.Ct, "wr": wr, "isr": 0, "img": this.ua.user.Img}
	data, err = json.Marshal(&res)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.uc.send <- data

	//开始发牌
	this.dealCards(0)
}

//某个玩家准备好了, 1为左侧用户，2为右侧用户，3为自己
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
	cards := gamerule.GenRandCards()
	//每个人17张，留三张作为底牌
	res := make(map[string]interface{})

	res["cmd"] = "dealcards"
	res["err"] = ""
	res["again"] = again
	res["u1"] = cards[0:0]
	res["u2"] = cards[0:0]

	//向ua发牌

	this.ua.cards = cards[0:17]
	res["u3"] = cards[0:17]
	data, err := json.Marshal(&res)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.ua.send <- data

	//向ub发牌
	this.ub.cards = cards[17:34]
	res["u3"] = cards[17:34]
	data, err = json.Marshal(&res)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.ub.send <- data

	//向uc发牌

	this.uc.cards = cards[34:51]
	res["u3"] = cards[34:51]
	data, err = json.Marshal(&res)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.uc.send <- data

	//设置底牌
	this.backcards[0] = cards[51]
	this.backcards[1] = cards[52]
	this.backcards[2] = cards[53]
}

//服务器随机第一个叫地主的玩家
func (this *Room) sendFirstCaller() {
	rand.Seed(time.Now().Unix())
	n := rand.Intn(3) + 1

	var res string
	res = `{"cmd":"caller","err":"","user":`
	switch n {
	//选定ua
	case 1:
		//向ua发送
		ret := fmt.Sprintf("%s%d}", res, 3)
		this.ua.send <- []byte(ret)
		//向ub发送
		ret = fmt.Sprintf("%s%d}", res, 1)
		this.ub.send <- []byte(ret)
		//向uc发送
		ret = fmt.Sprintf("%s%d}", res, 2)
		this.uc.send <- []byte(ret)
	//选定ub
	case 2:
		//向ua发送
		ret := fmt.Sprintf("%s%d}", res, 2)
		this.ua.send <- []byte(ret)
		//向ub发送
		ret = fmt.Sprintf("%s%d}", res, 3)
		this.ub.send <- []byte(ret)
		//向uc发送
		ret = fmt.Sprintf("%s%d}", res, 1)
		this.uc.send <- []byte(ret)
	//选定uc
	case 3:
		//向ua发送
		ret := fmt.Sprintf("%s%d}", res, 1)
		this.ua.send <- []byte(ret)
		//向ub发送
		ret = fmt.Sprintf("%s%d}", res, 2)
		this.ub.send <- []byte(ret)
		//向uc发送
		ret = fmt.Sprintf("%s%d}", res, 3)
		this.uc.send <- []byte(ret)
	}
}

//玩家叫地主
// pos 具体哪个玩家
// call 该玩家是否叫地主 1 叫 0 不叫
func (this *Room) callLandlord(pos, call int) {
	var res = make(map[string]interface{})
	res["cmd"] = "calllandlord"
	res["err"] = ""
	res["call"] = call

	//按照顺序应该是ub执行下一步操作
	if call == 0 {
		res["ntdo"] = "c"
	} else {
		res["ntdo"] = "g"
	}

	switch pos {
	case 1:
		this.ua.iscall = call

		//向ua发送
		res["user"] = 3 //叫地主的是ua
		res["nt"] = 2   //ub执行下一步操作
		data, err := json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}

		this.ua.send <- data

		//向ub发送
		res["user"] = 1 //叫地主的是ua
		res["nt"] = 3   //ub执行下一步操作
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}

		this.ub.send <- data

		//向uc发送
		res["user"] = 2 //叫地主的是ua
		res["nt"] = 1   //ub执行下一步操作
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}

		this.uc.send <- data

	case 2:
		this.ub.iscall = call

		//向ua发送
		res["user"] = 2 //叫地主的是ub
		res["nt"] = 1   //uc执行下一步操作
		data, err := json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}

		this.ua.send <- data

		//向ub发送
		res["user"] = 3 //叫地主的是ub
		res["nt"] = 2   //uc执行下一步操作
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}

		this.ub.send <- data

		//向uc发送
		res["user"] = 1 //叫地主的是ub
		res["nt"] = 3   //uc执行下一步操作
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}

		this.uc.send <- data

	case 3:
		this.uc.iscall = call

		//向ua发送
		res["user"] = 1 //叫地主的是uc
		res["nt"] = 3   //ua执行下一步操作
		data, err := json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}

		this.ua.send <- data

		//向ub发送
		res["user"] = 2 //叫地主的是uc
		res["nt"] = 1   //ua执行下一步操作
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}

		this.ub.send <- data

		//向uc发送
		res["user"] = 3 //叫地主的是uc
		res["nt"] = 2   //ua执行下一步操作
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}

		this.uc.send <- data

	}
}

//玩家抢地主
// pos 具体哪个玩家
// grab 该玩家是否抢地主 1 抢 0 不抢
func (this *Room) grabLandlord(pos, grab int) {
	var res = make(map[string]interface{})
	res["cmd"] = "grablandlord"
	res["err"] = ""
	res["grab"] = grab

	n := this.getNextOperator(pos)
	if n == 0 {
		//后面没有人抢地主了
		this.sendLandlordOwer()
		return
	}

	switch pos {
	case 1:
		if this.ua.iscall == 1 {
			//向ua发送
			res["user"] = 3 //抢地主的是ua
			res["nt"] = this.getRelativePos(1, n)
			data, err := json.Marshal(&res)
			if err != nil {
				logs.Error("json marshal error:%s", err.Error())
			}
			this.ua.send <- data

			//向ub发送
			res["user"] = 1 //抢地主的是ua
			res["nt"] = this.getRelativePos(2, n)
			data, err = json.Marshal(&res)
			if err != nil {
				logs.Error("json marshal error:%s", err.Error())
			}
			this.ub.send <- data

			//向uc发送
			res["user"] = 2 //抢地主的是ua
			res["nt"] = this.getRelativePos(3, n)
			data, err = json.Marshal(&res)
			if err != nil {
				logs.Error("json marshal error:%s", err.Error())
			}
			this.uc.send <- data
		}

	case 2:
		if this.ub.iscall == 1 {
			//向ua发送
			res["user"] = 2 //抢地主的是ub
			res["nt"] = this.getRelativePos(1, n)
			data, err := json.Marshal(&res)
			if err != nil {
				logs.Error("json marshal error:%s", err.Error())
			}
			this.ua.send <- data

			//向ub发送
			res["user"] = 3 //抢地主的是ub
			res["nt"] = this.getRelativePos(2, n)
			data, err = json.Marshal(&res)
			if err != nil {
				logs.Error("json marshal error:%s", err.Error())
			}
			this.ub.send <- data

			//向uc发送
			res["user"] = 1 //抢地主的是ub
			res["nt"] = this.getRelativePos(3, n)
			data, err = json.Marshal(&res)
			if err != nil {
				logs.Error("json marshal error:%s", err.Error())
			}
			this.uc.send <- data
		}
	case 3:
		if this.uc.iscall == 1 {
			//向ua发送
			res["user"] = 1 //抢地主的是uc
			res["nt"] = this.getRelativePos(1, n)
			data, err := json.Marshal(&res)
			if err != nil {
				logs.Error("json marshal error:%s", err.Error())
			}
			this.ua.send <- data

			//向ub发送
			res["user"] = 2 //抢地主的是uc
			res["nt"] = this.getRelativePos(2, n)
			data, err = json.Marshal(&res)
			if err != nil {
				logs.Error("json marshal error:%s", err.Error())
			}
			this.ub.send <- data

			//向uc发送
			res["user"] = 3 //抢地主的是uc
			res["nt"] = this.getRelativePos(3, n)
			data, err = json.Marshal(&res)
			if err != nil {
				logs.Error("json marshal error:%s", err.Error())
			}
			this.uc.send <- data
		}

	}
}

//判断下一步操作的人
//返回值 1 ua 2 ub 3 uc 0 没有
func (this *Room) getNextOperator(pos int) int {
	var n = 0
	switch pos {
	case 1:
		//判断ua之后该谁
		if this.ub.iscall == 1 {
			n = 2
		} else if this.uc.iscall == 1 {
			n = 3
		} else {
			n = 0
		}
	case 2:
		//判断ub之后该谁
		if this.uc.iscall == 1 {
			n = 3
		} else if this.ua.iscall == 1 {
			n = 1
		} else {
			n = 0
		}
	case 3:
		//判断uc之后该谁
		if this.ua.iscall == 1 {
			n = 1
		} else if this.ub.iscall == 1 {
			n = 2
		} else {
			n = 0
		}
	}
	return n
}

//获取u2相对u1的位置
//返回值 1 左侧 2 右侧 3 自己
func (this *Room) getRelativePos(u1, u2 int) int {
	var n = 3
	switch u1 {
	//ua
	case 1:
		switch u2 {
		//ua
		case 1:
			n = 3
			//ub
		case 2:
			n = 2
			//uc
		case 3:
			n = 1
		}
		//ub
	case 2:
		switch u2 {
		//ua
		case 1:
			n = 1
			//ub
		case 2:
			n = 3
			//uc
		case 3:
			n = 2
		}
		//uc
	case 3:
		switch u2 {
		//ua
		case 1:
			n = 2
			//ub
		case 2:
			n = 1
			//uc
		case 3:
			n = 3
		}
	}
	return n
}

//亮底牌，通知谁是地主
func (this *Room) sendLandlordOwer() {
	var res = make(map[string]interface{})
	res["cmd"] = "landlordower"
	res["err"] = ""
	res["bc"] = this.backcards

	var lordpos = 0
	//没有人抢地主，地主是第一个叫地主的玩家
	if this.firstgraber == 0 {
		lordpos = this.firstcaller
	} else {
		//地主是第一个抢地主的玩家。注意：如果叫地主的玩家也抢了地主，我们将其记录为firstgraber了
		lordpos = this.firstgraber
	}

	switch lordpos {
	case 1:
		this.ua.utype = 1
		this.ub.utype = 2
		this.uc.utype = 2
		this.ua.cards = append(this.ua.cards, this.backcards[:]...)

		//向ua发送
		res["lo"] = 3 //ua是地主
		data, err := json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.ua.send <- data

		//向ub发送
		res["lo"] = 1 //ua是地主
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.ub.send <- data

		//向uc发送
		res["lo"] = 2 //ua是地主
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.uc.send <- data

	case 2:
		this.ua.utype = 2
		this.ub.utype = 1
		this.uc.utype = 2
		this.ub.cards = append(this.ub.cards, this.backcards[:]...)

		//向ua发送
		res["lo"] = 2 //ub是地主
		data, err := json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.ua.send <- data

		//向ub发送
		res["lo"] = 3 //ub是地主
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.ub.send <- data

		//向uc发送
		res["lo"] = 1 //ub是地主
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.uc.send <- data
	case 3:
		this.ua.utype = 2
		this.ub.utype = 2
		this.uc.utype = 1
		this.uc.cards = append(this.uc.cards, this.backcards[:]...)

		//向ua发送
		res["lo"] = 1 //uc是地主
		data, err := json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.ua.send <- data

		//向ub发送
		res["lo"] = 2 //uc是地主
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.ub.send <- data

		//向uc发送
		res["lo"] = 3 //uc是地主
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.uc.send <- data
	}

}

//玩家出牌
//cv 玩家出的牌
func (this *Room) PopCards(pos int, cv []interface{}) {

	//将[]interface{} 转为 []int
	cards := make([]int, len(cv))
	for i, v := range cv {
		cards[i] = int(v.(float64))
	}

	if this.propos == pos {
		this.precards = nil
	}

	if len(cards) != 0 {
		this.propos = pos
	}

	var res = make(map[string]interface{})
	res["cmd"] = "popcards"
	res["err"] = ""

	if len(cards) != 0 {
		//判断牌类型
		myType, myValue, myNv := gamerule.GetCardsType(cards)
		if myType == gamerule.ERROR_CARD {
			res["err"] = "牌类型错误"
			data, err := json.Marshal(&res)
			if err != nil {
				logs.Error("json marshal error:%s", err.Error())
			}
			switch pos {
			case 1:
				this.ua.send <- data
			case 2:
				this.ub.send <- data
			case 3:
				this.uc.send <- data
			}
			return
		} else {
			//判断是否大于上家出的牌
			if len(this.precards) > 0 {
				prevType, prevValue, preNv := gamerule.GetCardsType(this.precards)
				over := gamerule.IsOvercomePrev(myType, myValue, myNv, prevType, prevValue, preNv)

				if !over {
					//没有上家的牌大
					res["err"] = "没有上家的牌大"
					data, err := json.Marshal(&res)
					if err != nil {
						logs.Error("json marshal error:%s", err.Error())
					}
					switch pos {
					case 1:
						this.ua.send <- data
					case 2:
						this.ub.send <- data
					case 3:
						this.uc.send <- data
					}
					return
				}
			}
		}
	}

	//设置上次玩家出的牌
	if len(cards) != 0 {
		this.precards = cards
	}

	this.precards = cards
	//通知其他人出牌情况
	res["cards"] = cards

	switch pos {
	case 1:
		num := len(this.ua.cards) - len(cards)
		this.ua.cards = gamerule.Difference(this.ua.cards, cards)
		res["num"] = num

		//ua
		res["user"] = 3 //ua出的牌
		res["nt"] = 2   //该ub出牌
		data, err := json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.ua.send <- data

		//ub
		res["user"] = 1 //ua出的牌
		res["nt"] = 3   //该ub出牌
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.ub.send <- data

		//uc
		res["user"] = 2 //ua出的牌
		res["nt"] = 1   //该ub出牌
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.uc.send <- data

		//判断是否出完牌
		if num == 0 {
			//结算
			this.Settlement(pos)
		}

	case 2:
		num := len(this.ub.cards) - len(cards)
		this.ub.cards = gamerule.Difference(this.ub.cards, cards)
		res["num"] = num

		//ua
		res["user"] = 2 //ub出的牌
		res["nt"] = 1   //该uc出牌
		data, err := json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.ua.send <- data

		//ub
		res["user"] = 3 //ub出的牌
		res["nt"] = 2   //该uc出牌
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.ub.send <- data

		//uc
		res["user"] = 1 //ub出的牌
		res["nt"] = 3   //该uc出牌
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.uc.send <- data

		//判断是否出完牌
		if num == 0 {
			//结算
			this.Settlement(pos)
		}

	case 3:
		num := len(this.uc.cards) - len(cards)
		this.uc.cards = gamerule.Difference(this.uc.cards, cards)
		res["num"] = num

		//ua
		res["user"] = 1 //uc出的牌
		res["nt"] = 3   //该ua出牌
		data, err := json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.ua.send <- data

		//ub
		res["user"] = 2 //uc出的牌
		res["nt"] = 1   //该ua出牌
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.ub.send <- data

		//uc
		res["user"] = 3 //uc出的牌
		res["nt"] = 2   //该ua出牌
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.uc.send <- data

		//判断是否出完牌
		if num == 0 {
			//结算
			this.Settlement(pos)
		}
	}
}

//结算
// pos 赢的玩家
func (this *Room) Settlement(pos int) {
	var res = make(map[string]interface{})
	res["cmd"] = "settlement"
	res["err"] = ""

	reward := make(map[string]interface{})

	var rua, rub, ruc int

	//计算金币
	if this.ua.utype == 1 {
		//ua是地主
		switch pos {
		case 1:
			rua = 200
			rub = -100
			ruc = -100
		case 2:
			rua = -200
			rub = 100
			ruc = 100
		case 3:
			rua = -200
			rub = 100
			ruc = 100
		}
	} else if this.ub.utype == 1 {
		//ub是地主
		switch pos {
		case 1:
			rua = 100
			rub = -200
			ruc = 100
		case 2:
			rua = 200
			rub = -100
			ruc = -100
		case 3:
			rua = 100
			rub = -200
			ruc = 100
		}
	} else {
		//uc是地主
		switch pos {
		case 1:
			rua = 100
			rub = 100
			ruc = -200
		case 2:
			rua = 100
			rub = 100
			ruc = -200
		case 3:
			rua = -100
			rub = -100
			ruc = 200
		}
	}
	this.ua.user.Up += rua
	this.ub.user.Up += rub
	this.uc.user.Up += ruc

	//胜利次数
	switch pos {
	case 1:
		this.ua.user.Wn += 1
	case 2:
		this.ub.user.Wn += 1
	case 3:
		this.uc.user.Wn += 1
	}
	//对局次数
	this.ua.user.Ct += 1
	this.ub.user.Ct += 1
	this.uc.user.Ct += 1

	//更新数据库
	//更新ua
	bm := make(map[string]interface{})
	bm["up"] = this.ua.user.Up
	bm["ct"] = this.ua.user.Ct
	bm["wn"] = this.ua.user.Wn

	err := model.UpdateUser(this.ua.user.ObjectId, bm)
	if err != nil {
		logs.Error("update db error:%s", err.Error())
	}
	//更新ub
	bm["up"] = this.ub.user.Up
	bm["ct"] = this.ub.user.Ct
	bm["wn"] = this.ub.user.Wn

	err = model.UpdateUser(this.ub.user.ObjectId, bm)
	if err != nil {
		logs.Error("update db error:%s", err.Error())
	}
	//更新uc
	bm["up"] = this.uc.user.Up
	bm["ct"] = this.uc.user.Ct
	bm["wn"] = this.uc.user.Wn

	err = model.UpdateUser(this.uc.user.ObjectId, bm)
	if err != nil {
		logs.Error("update db error:%s", err.Error())
	}

	//发送消息
	var winer int
	//通知ua
	switch pos {
	case 1:
		winer = 3
	case 2:
		winer = 2
	case 3:
		winer = 1
	}
	res["win"] = winer
	reward["u1"] = ruc
	reward["u2"] = rub
	reward["u3"] = rua
	res["reward"] = reward
	res["c1"] = this.uc.cards
	res["c2"] = this.ub.cards
	res["c3"] = this.ua.cards
	res["u1"] = this.uc.user.Up
	res["u2"] = this.ub.user.Up
	res["u3"] = this.ua.user.Up

	data, err := json.Marshal(&res)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.ua.send <- data

	//通知ub
	switch pos {
	case 1:
		winer = 1
	case 2:
		winer = 3
	case 3:
		winer = 2
	}
	res["win"] = winer
	reward["u1"] = rua
	reward["u2"] = ruc
	reward["u3"] = rub
	res["reward"] = reward
	res["c1"] = this.ua.cards
	res["c2"] = this.uc.cards
	res["c3"] = this.ub.cards
	res["u1"] = this.ua.user.Up
	res["u2"] = this.uc.user.Up
	res["u3"] = this.ub.user.Up

	data, err = json.Marshal(&res)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.ub.send <- data

	//通知uc
	switch pos {
	case 1:
		winer = 2
	case 2:
		winer = 1
	case 3:
		winer = 3
	}
	res["win"] = winer
	reward["u1"] = rub
	reward["u2"] = rua
	reward["u3"] = ruc
	res["reward"] = reward
	res["c1"] = this.ub.cards
	res["c2"] = this.ua.cards
	res["c3"] = this.uc.cards
	res["u1"] = this.ub.user.Up
	res["u2"] = this.ua.user.Up
	res["u3"] = this.uc.user.Up

	data, err = json.Marshal(&res)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	this.uc.send <- data

	//房间重置
	//	this.readycout = 0
	this.ResetRoom()
}

//离开房间,目前的策略是其中一个玩家离开，就相当于所有玩家离开
func (this *Room) LeaveRoom(pos int) {
	var res = `{"cmd":"dissolve"}`

	switch pos {
	case 1:
		//发送给ub
		this.ub.send <- []byte(res)
		//发送给uc
		this.uc.send <- []byte(res)

	case 2:
		//发送给ua
		this.ua.send <- []byte(res)
		//发送给uc
		this.uc.send <- []byte(res)

	case 3:
		//发送给ua
		this.ua.send <- []byte(res)
		//发送给ub
		this.ub.send <- []byte(res)
	}

	this.ua.LeaveRoom()
	this.ub.LeaveRoom()
	this.uc.LeaveRoom()
	this.ua = nil
	this.ub = nil
	this.uc = nil

	this.stop <- 1
}

//固定消息聊天
func (this *Room) Chat(pos, id int) {
	var res = make(map[string]interface{})
	res["cmd"] = "chat"
	res["err"] = ""
	res["id"] = id

	switch pos {
	case 1:
		//发送给ub
		res["user"] = 2
		data, err := json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.ub.send <- []byte(data)
		//发送给uc
		res["user"] = 1
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.uc.send <- []byte(data)
	case 2:
		//发送给ua
		res["user"] = 1
		data, err := json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.ua.send <- []byte(data)
		//发送给uc
		res["user"] = 2
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.uc.send <- []byte(data)
	case 3:
		//发送给ua
		res["user"] = 2
		data, err := json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.ua.send <- []byte(data)
		//发送给ub
		res["user"] = 1
		data, err = json.Marshal(&res)
		if err != nil {
			logs.Error("json marshal error:%s", err.Error())
		}
		this.ub.send <- []byte(data)
	}

}

//房间
func (this *Room) ResetRoom() {
	this.verifycount = 0
	this.notcalllocout = 0
	this.grabcount = 0
	this.firstcaller = 0
	this.firstgraber = 0
	this.propos = 0

	this.ua.iscall = 1
	this.ub.iscall = 1
	this.uc.iscall = 1
	this.ua.utype = 0
	this.ub.utype = 0
	this.uc.utype = 0
}
