package web

//https://studygolang.com/articles/9467
//https://studygolang.com/articles/12977?fr=sidebar
// https://blog.csdn.net/books1958/article/details/41748719
import (
	"UULoServer/logs"
	"UULoServer/model"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func timeHandler(w http.ResponseWriter, r *http.Request) {

	tm := time.Now().Format("time.RFC1123")
	w.Write([]byte("The time is: " + tm))
}

func loginHandle(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	//w.Header().Set("content-type", "application/json")             //返回数据格式是json

	//获取客户端通过GET/POST方式传递的参数
	r.ParseForm()
	param_userid, found1 := r.Form["userid"]
	//param_password, found2 := r.Form["password"]

	if !(found1) {
		fmt.Fprint(w, "滚")
		return
	}

	unionid := param_userid[0]
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

	u, has = model.GetUser(unionid) //1 从游戏用户表中查找 userid

	res := make(map[string]interface{})
	if err != nil {
		res["err"] = err.Error()
	} else {
		res["err"] = "0"
	}

	if has { //2 找到后返回数据
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

	} else { //3 未找到,从 app  user 表中查询
		var bu *model.User
		bu, has = model.GetBaseUser(unionid)

		if !has { //4 未找到返回错误
			res["err"] = "1" //用户不存在

		} else { //5 找到 则将数据添加进游戏表中并返回
			provice := ""
			city := "郑州"
			sex := 1
			country := "中国"
			headimgurl := ""

			u, err = model.NewUser(bu.Username, provice, city, country,
				headimgurl, bu.Id, sex)

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
		}
	}

	data, err := json.Marshal(&res)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	w.Write(data)
}

func RoomCfgHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	//goldRoom: [
	// 场次ID + | + 底分 + | + 房间初始倍数 + | + 入场最小值 + | + 入场最大值 + | + 门票 + | + 输赢封顶
	// {id: 0, underPoint: 0, initMul: 0, minEnterPoint: 0, maxEnterPoint: 0, ticket: 0, maxEarn: 0},

	// U钻场数据结构
	//gemRoom: [
	// 场次ID + | + 底分 + | + 房间初始倍数 + | + 可赢钻石 + | + 入场金币 + | + 金币购买的积分数量
	// {id: 0, underPoint: 0, initMul: 0, gem: 0, ticket: 0, gamePoint: 0},
	//w.Write(data)

	res := make(map[string][]string)

	res["cp"] = []string{"1|10|5|100|10000|1|1000000", "2|10|5|100|10000|1|1000000"}
	res["sp"] = []string{"10|10|5|100|10000|100000", "11|20|5|100|10000|100000"}

	data, err := json.Marshal(&res)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	w.Write(data)

}

func RankHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型

	////   名字 |  胜局  | 连胜  |  总局  |  4  | 头像 | goldnum
	res := make(map[string][]string)

	res["gold"] = []string{"t1|10|5|100|11|1000000|222", "t1|10|5|100|11|1000000|222"}
	res["ud"] = []string{"t1|10|5|100|11|1000000|222", "t1|10|5|100|11|1000000|222"}

	data, err := json.Marshal(&res)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	w.Write(data)
}

//请求显示商品列表
func GoodsListHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型

	//// 商品ID + | + 金币 + | + 对应人民币 + | + 是否热卖 + | + 多送百分比
	res := make(map[string][]string)

	res["list"] = []string{"10|100|100|1|10",
		"11|100|100|0|10",
		"12|100|100|0|10",
		"50|100|100|1|11",
		"51|100|100|1|11",
		"52|100|100|1|11"}
	//res["ud"] = []string{"t1|10|5|100|11|1000000|222", "t1|10|5|100|11|1000000|222"}

	data, err := json.Marshal(&res)
	if err != nil {
		logs.Error("json marshal error:%s", err.Error())
	}
	w.Write(data)
}

func Start() {

	//var format string = time.RFC1123
	//th := timeHandler(format)

	http.HandleFunc("/time", timeHandler)
	http.HandleFunc("/login", loginHandle)
	http.HandleFunc("/roomcfg", RoomCfgHandle)
	http.HandleFunc("/rank", RankHandle)
	http.HandleFunc("/goodslist", GoodsListHandle)

	fmt.Println("http server start ...")
	http.ListenAndServe(":8080", nil)
}
