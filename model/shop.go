package model

import (
	"UULoServer/logs"
	"fmt"

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

var list []Goods

func (this *Goods) TableName() string {
	return "shop"
}

func GetGoodList() ([]string, error) {
	var err error
	if list == nil {
		list = make([]Goods, 0)
		err = engine.Find(&list)

		if err != nil {
			logs.Error("open room cfg error:%s", err.Error())
			return nil, err
		}
	}

	liststr := make([]string, len(list))
	for i, v := range list {
		liststr[i] = fmt.Sprintf("%d|%d|%d|%d|%d", v.Id, v.Gold, v.Price, v.Hot, v.Discount)
	}
	return liststr, err
}
