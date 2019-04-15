// Card
package gamerule

import (
	"math/rand"
	"time"

	"github.com/deckarep/golang-set"
)

//牌类型
/*
单牌: 一张单牌
对牌：数值相同的两张牌，eg:梅花4 + 方块4
三张牌：数值相同的三张牌，eg:三个J
三带一：数值相同的三张牌 + 一张单牌或者一对牌，eg:333+6 或者 444+99
顺子： 五张或更多的连续单牌，eg:34567或者10JQKA。不包括2和双王
双顺： 三队或更多的连续对牌，eg:334455或者JJQQKKAA。不包括2和双王
三顺： 两个或更多的连续三张牌，eg:333444或者JJJQQQKKK。不包括2和双王
飞机带翅膀：三顺+同数量的单牌（或同数量的对牌），eg:444555+7+9 或 333444555+77+99+JJ
四带二：四张相同数值的牌 + 两单牌（或两对牌），四带二不是炸弹，eg:5555+3+8 或 4444 + 55 + 33
炸弹：四张相同数值的牌，eg:7777
火箭：双王，最大的牌.
*/
const (
	ERROR_CARD           = iota //错误,牌不符合规则
	SINGLE_CARD                 //单牌
	DOUBLE_CARD                 //对牌
	TREBLE_CARD                 //3张,3不带
	TREBLE_ONE_CARD             //3带1张单牌
	TREBLE_TWO_CARD             //3带2,3带一对
	CONTINUE_CARD               //顺子,连牌,最低5连
	DB_CONTINUE_CARD            //双顺,连队,最低3对
	TB_CONTINUE_CARD            //三顺,飞机不带
	AIRPLANE_SINGLE_CARD        //飞机带翅膀--飞机带单
	AIRPLANE_DOUBLE_CARD        //飞机带翅膀--飞机带对
	FOUR_TWO_CARD               //四带两单
	FOUR_TWO4_CARD              //四带两对
	BOMB_CARD                   //炸弹
	ROCKET_CARD                 //火箭
)

//牌值定义
var cards = []int{
	//大王，小王
	519, 518,
	//黑桃2~3
	416, 414, 413, 412, 411, 410, 409, 408, 407, 406, 405, 404, 403,
	//红桃2~3
	316, 314, 313, 312, 311, 310, 309, 308, 307, 306, 305, 304, 303,
	//梅花2~3
	216, 214, 213, 212, 211, 210, 209, 208, 207, 206, 205, 204, 203,
	//方块2~3
	116, 114, 113, 112, 111, 110, 109, 108, 107, 106, 105, 104, 103,
}

/**
 * 生成随机的牌
 * @param cds 手牌
 * @return 随机牌
 */
func GenRandCards() []int {
	length := len(cards)
	rand.Seed(time.Now().Unix())
	index := rand.Perm(length)

	randcds := make([]int, length)
	for i := 0; i < length; i++ {
		randcds[i] = cards[index[i]]
	}
	return randcds
}

/**
 * 获取牌类型
 * @param cds 手牌
 * @return cardtype 牌型, v 牌值 n 当牌类型为顺子、连队、飞机时表明是几连
 */
func GetCardsType(cds []int) (cardtype, v, n int) {
	var b bool
	cardtype = ERROR_CARD
	v = 0
	n = 0
	if b, v = isSingle(cds); b {
		cardtype = SINGLE_CARD //单牌
	} else if b, v = isDouble(cds); b {
		cardtype = DOUBLE_CARD //对牌
	} else if b, v = isTreble(cds); b {
		cardtype = TREBLE_CARD //3不带
	} else if b, v = isTrebleOne(cds); b {
		cardtype = TREBLE_ONE_CARD //3带1张单牌
	} else if b, v = isTrebleTwo(cds); b {
		cardtype = TREBLE_TWO_CARD //3带2,3带一对
	} else if b, v, n = isContinue(cds); b {
		cardtype = CONTINUE_CARD //顺子,连牌,最低5连
	} else if b, v, n = isDbContinue(cds); b {
		cardtype = DB_CONTINUE_CARD //双顺,连队,最低3对
	} else if b, v, n = isTbContinue(cds); b {
		cardtype = TB_CONTINUE_CARD //三顺,飞机不带
	} else if b, v, n = isAirplaneSingle(cds); b {
		cardtype = AIRPLANE_SINGLE_CARD //飞机带翅膀--飞机带单
	} else if b, v, n = isAirplaneDouble(cds); b {
		cardtype = AIRPLANE_DOUBLE_CARD //飞机带翅膀--飞机带对
	} else if b, v = isFourTwo(cds); b {
		cardtype = FOUR_TWO_CARD //四带两单
	} else if b, v = isFourTwo4(cds); b {
		cardtype = FOUR_TWO4_CARD //四带两对
	} else if b, v = isBomb(cds); b {
		cardtype = BOMB_CARD //炸弹
	} else if b = isRocket(cds); b {
		cardtype = ROCKET_CARD //火箭
	} else {
		cardtype = ERROR_CARD
		v = 0
	}

	return
}

/**
 * 判断我出的牌和上家的牌的大小，决定是否可以出牌
 * @param myType 我的牌的类型
 * @param myValue 我的牌值 GetCardsType返回的值
 * @param myNv  当牌类型为顺子、连队、飞机时表明是几连，其余情况设置为0
 * @param prevType 上家的牌型
 * @param prevValue 上家的牌值，GetCardsType返回的值
 * @param preNv  当牌类型为顺子、连队、飞机时表明是几连，其余情况设置为0
 * @return 可以出牌，返回true；否则，返回false。
 */
func IsOvercomePrev(myType, myValue, myNv, prevType, prevValue, preNv int) bool {
	//集中判断是否王炸，免得多次判断王炸
	if prevType == ROCKET_CARD {
		return false
	} else if myType == ROCKET_CARD {
		return true
	}

	// 集中判断对方不是炸弹，我出炸弹的情况
	if prevType != BOMB_CARD && myType == BOMB_CARD {
		return true
	}

	// 根据规则 主要有2种情况，1.我出和上家一种类型的牌,此时需判断牌值大小；
	// 2.我出炸弹，此时，和上家的牌的类型可能不同，炸弹的情况在上面已做判断。

	//牌类型不一样，不能出
	if myType != prevType {
		return false
	}

	//牌类型一样

	//当牌类型为顺子、连队、飞机时,要判断连数是否一致,即都为n连
	if CONTINUE_CARD <= myType && myType <= AIRPLANE_DOUBLE_CARD {
		if myNv == 0 || preNv == 0 || myNv != preNv {
			return false
		}
	}

	//判断牌值大小
	if myValue > prevValue {
		return true
	}

	return false
}

//两组牌之间的差集
func Difference(src, des []int) []int {
	srcset := mapset.NewSet()
	for _, v := range src {
		srcset.Add(v)
	}
	desset := mapset.NewSet()
	for _, v := range des {
		desset.Add(v)
	}
	ds := srcset.Difference(desset).ToSlice()

	//将[]interface{} 转为 []int
	dc := make([]int, len(ds))
	for i, v := range ds {
		dc[i] = v.(int)
	}

	return dc
}
