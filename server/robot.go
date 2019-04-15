package server

import (
	"fmt"
	//"UULoServer/gamerule"
	"sort"
)

type SCards []int //用于给牌根据牌值递增排序

func (p SCards) Len() int           { return len(p) }
func (p SCards) Less(i, j int) bool { return p[i]%100 < p[j]%100 }
func (p SCards) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

//机器人
type Robot struct {
	c_single []int //单牌  递增排序
	c_double []int //对牌 递增排序
	c_treble []int //3张 递增排序
	c_bomb   []int //炸弹 递增排序
	c_rocket []int //火箭

	//m_single map[int]int //单牌 key 牌值 value 牌
	//m_double map[int]

	cards []int //拥有的总牌
}

//拆牌
func (this *Robot) SplitCards(cds []int) {
	this.c_single = nil
	this.c_double = nil
	this.c_treble = nil
	this.c_bomb = nil
	this.c_rocket = nil

	this.cards = cds
	sort.Sort(SCards(this.cards))
	fmt.Println(this.cards)

	cn := make(map[int]int, 15) // key 牌值 value 牌出现的次数
	for _, v := range this.cards {
		cv := v % 100
		cn[cv] += 1
	}

	//fmt.Println(cn)

	//先判断是否有大小王（火箭）
	if cn[19] == 1 && cn[18] == 1 {
		this.c_rocket = make([]int, 2)
		this.c_rocket[0] = 518
		this.c_rocket[1] = 519
		delete(cn, 19)
		delete(cn, 18)
	}

	//判断其余牌
	for _, v := range cn {

		switch v {
		case 1:

		case 2:

		case 3:

		case 4:
		}
	}
}
