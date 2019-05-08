package model

import (
	"UULoServer/logs"
	"fmt"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

//商品类
type Goods struct {
	Id       int `xorm:"pk 'id'"`  //主键，该条记录的唯一标识
	Gold     int `xorm:"gold"`     //金豆数
	Price    int `xorm:"price"`    //商品价值,单位为分
	Hot      int `xorm:"hot"`      //是否热卖
	Discount int `xorm:"discount"` //优惠信息,多送百分比
}

//购买记录
type Record struct {
	OrderId string    `xorm:"pk 'orderid'"` //订单号
	User    string    `xorm:"user"`
	Good    int       `xorm:"good"`
	Num     int       `xorm:"num"`
	Price   int       `xorm:"price"`
	Ct      time.Time `xomr:"created 'ct'"`
}

var list []Goods     //金豆列表
var zuanlist []Goods //钻列表

func (this *Goods) TableName() string {
	return "shop"
}

func (this *Record) TableName() string {
	return "exrecord"
}

//新订单
func NewRecord(goodid int, user string) (*Record, error) {

	t := fmt.Sprintf("%v", time.Now().UnixNano()/1e6) //订单号

	var pg *Goods
	pg = nil
	if goodid > 20 {
		if zuanlist != nil {
			pg = &zuanlist[goodid%10]
		}
	} else {
		if list != nil {
			pg = &list[goodid%10]
		}
	}

	if pg != nil {

		if goodid > 20 { //币换钻石
			//得到当前汇率
			price := GetCurrPrice()
			//查询用户余额
			wallet := GetUserShare(user)
			//计算可用余额是否足够支付商品价值
			value := float64(pg.Price/100) / price //商品单价
			if wallet >= value {
				//添加基础用户订单表
				_, err := engine.Exec("INSERT INTO coin_bill (create_time,user_id,coin_number,states,user_state,user_sum,user_type) values(?, ?, ?, ?,?,?,?)", time.Now(), user, -value, "兑换为游戏钻石", 2, value, 23)
				if err != nil {
					logs.Error("INSERT INTO coin_bill error:%s", err.Error())
					return nil, nil
				}
				//更新用户share
				_, err = engine.Exec("UPDATE option_s SET share = ? WHERE user_id = ?", wallet-float64(pg.Price/100), user)
				if err != nil {
					logs.Error("update option_s error:%s", err.Error())
					return nil, nil
				}

				r := &Record{OrderId: t, User: user, Good: goodid, Price: pg.Price, Num: pg.Gold, Ct: time.Now()}
				//添加游戏内订单表
				_, err = engine.InsertOne(r)
				if err != nil {
					logs.Error("new record  error:%s", err.Error())
				}
				return r, err

			} else {
				logs.Error("user( %s ) money ( %d ) is not enough", user, wallet)
				return nil, nil
			}
		} else { //钻石换豆
			r := &Record{OrderId: t, User: user, Good: goodid, Price: pg.Price, Num: pg.Gold, Ct: time.Now()}
			//添加游戏内订单表
			_, err := engine.InsertOne(r)
			if err != nil {
				logs.Error("new record  error:%s", err.Error())
			}
			return r, err
		}
	} else {

		return nil, nil
	}
	return nil, nil
}

func GetGoodList(goodstype int) ([]string, error) {
	var err error

	if goodstype == 1 {
		if list == nil {
			list = make([]Goods, 0)
			err = engine.Sql("SELECT * from shop where id <20").Find(&list) //取得豆列表
		}
	} else {
		if zuanlist == nil {
			zuanlist = make([]Goods, 0)
			err = engine.Sql("SELECT * from shop where id >20").Find(&zuanlist) //取得钻列表
		}
	}

	// has, err := engine.Where("id>?", 20).Get(list)
	// if has {

	// } else {
	// 	return nil, err
	// }

	// err = engine.Find(&list)
	if err != nil {
		logs.Error("open room cfg error:%s", err.Error())
		return nil, err
	}

	var liststr []string

	if goodstype == 1 {
		liststr = make([]string, len(list))
		for i, v := range list {
			liststr[i] = fmt.Sprintf("%d|%d|%d|%d|%d", v.Id, v.Gold, v.Price, v.Hot, v.Discount)
		}
	} else {
		liststr = make([]string, len(zuanlist))
		for i, v := range zuanlist {
			liststr[i] = fmt.Sprintf("%d|%d|%d|%d|%d", v.Id, v.Gold, v.Price, v.Hot, v.Discount)
		}
	}

	return liststr, err
}

//取得当前最新价格
func GetCurrPrice() float64 {

	results, err := engine.QueryString("select * from ltb")

	if err != nil {
		logs.Error("Get Curr Price  error:%s", err.Error())
		return 0
	}

	// if p, ok := results[0]["price"].(float64); ok {
	// 	return p
	// }
	p, _ := strconv.ParseFloat(results[0]["price"], 64)
	return p
	// bits := binary.LittleEndian.Uint64(results[0]["price"])
	// return math.Float64frombits(bits)
}

//取得用户可用余额
func GetUserShare(id string) float64 {

	// sqlstr := fmt.Sprintf(`select * from user where id="%s"`, id)
	// results, err := engine.Query(sqlstr)

	strsql := fmt.Sprintf(`select * from option_s where user_id="%s"`, id)

	results, err := engine.QueryString(strsql)

	if err != nil {
		logs.Error("Get User Share error:%s", err.Error())
		return 0
	}

	if len(results) != 1 {
		return 0
	}

	w, _ := strconv.ParseFloat(results[0]["share"], 64)
	return w
	// if w, ok := results[0]["share"].(float64); ok {
	// 	return w
	// }
	// return 0.3
}
