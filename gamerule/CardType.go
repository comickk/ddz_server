package gamerule

import (
	"sort"
)

//判断手牌的类型
//cds 给定的手牌

/**
 * 获取牌对应的值
 *
 * @param cds 牌的集合
 * @return ，是 返回true；否则，返回false
 */
func GetCardsValue(cds []int) []int {
	cvs := make([]int, len(cds))
	for i, v := range cds {
		cvs[i] = v % 100
	}
	return cvs
}

/**
 * 判断给定的牌是否是相同的牌
 *
 * @param cvs 牌值的集合
 * @return ，是 返回true；否则，返回false
 */
func isSameCards(cvs []int) bool {
	cardlen := len(cvs)
	if cardlen == 0 {
		return false
	}

	for i := 0; i < cardlen-1; i++ {
		if cvs[i] != cvs[i+1] {
			return false
		}
	}

	return true
}

/**
 * 判断牌是否为单牌
 *
 * @param cds 牌的集合
 * @return bool:true 单牌； false 非单牌
 *@return  int :当为单牌时，返回牌值, 否则返回0
 */
func isSingle(cds []int) (bool, int) {
	cardlen := len(cds)
	v := GetCardsValue(cds)
	if cardlen == 1 {
		return true, v[0]
	}
	return false, 0
}

/**
 * 判断牌是否为对牌
 *
 * @param cds 牌的集合
 * @return bool:true 对牌； false 非对牌
 *@return  int :当为对牌时，返回牌值 否则返回0
 */
func isDouble(cds []int) (bool, int) {
	cardlen := len(cds)

	if cardlen == 2 {
		//判断牌是否相同
		cvs := GetCardsValue(cds)

		if isSameCards(cvs) {
			return true, cvs[0]
		}
	}

	return false, 0
}

/**
 * 判断牌是否 3不带
 *
 * @param cds 牌的集合
 * @return bool 是 返回true；否则，返回false
 *@return  int :是 返回牌值 否则返回0
 */
func isTreble(cds []int) (bool, int) {
	cardlen := len(cds)

	if cardlen == 3 {
		//判断牌是否相同
		cvs := GetCardsValue(cds)

		if isSameCards(cvs) {
			return true, cvs[0]
		}
	}

	return false, 0
}

/**
 * 判断牌是否 3带1张单牌
 *
 * @param cds 牌的集合
 * @return 是 返回true；否则，返回false
 *@return  int :是 返回三张相同牌的牌值 否则返回0
 */
func isTrebleOne(cds []int) (bool, int) {
	cardlen := len(cds)

	if cardlen == 4 {
		cvs := GetCardsValue(cds)
		//递增排序
		sort.Ints(cvs)

		//四张牌相同，不认为是三带一
		if isSameCards(cvs) {
			return false, 0
		}

		//3带1排序后的组合为AAAB或BAAA,所以需要判断前三张或者后三张相同
		if isSameCards(cvs[:3]) {
			return true, cvs[1]
		}
		if isSameCards(cvs[1:4]) {
			return true, cvs[1]
		}
	}

	return false, 0
}

/**
 * 判断牌是否 3带2,带一对
 *
 * @param cds 牌的集合
 * @return 是 返回true；否则，返回false
 *@return  int :是 返回三张相同牌的牌值 否则返回0
 */
func isTrebleTwo(cds []int) (bool, int) {
	cardlen := len(cds)

	if cardlen == 5 {
		cvs := GetCardsValue(cds)
		//递增排序
		sort.Ints(cvs)

		//3带对排序后的组合为BBAAA或AAABB
		if isSameCards(cvs[:3]) {
			if isSameCards(cvs[3:5]) {
				return true, cvs[2]
			}
			return false, 0
		}

		if isSameCards(cvs[2:5]) {
			if isSameCards(cvs[0:2]) {
				return true, cvs[2]
			}
			return false, 0
		}
	}

	return false, 0
}

/**
 * 判断牌是否 顺子
 *
 * @param cds 牌的集合
 * @return 是 返回true；否则，返回false
 * @return  int :是 返回顺子中最小牌的牌值 否则返回0
 * @return  int :是 该值表明是几连 ，否则返回0
 */
func isContinue(cds []int) (bool, int, int) {
	cardlen := len(cds)

	// 顺子牌的个数在5到12之间
	if cardlen < 5 || cardlen > 12 {
		return false, 0, 0
	}

	cvs := GetCardsValue(cds)
	//递增排序
	sort.Ints(cvs)

	// 小王、大王、2不能加入
	if cvs[cardlen-1] == 19 || cvs[cardlen-1] == 18 || cvs[cardlen-1] == 16 {
		return false, 0, 0
	}

	for i := 0; i < cardlen-1; i++ {
		prev := cvs[i]
		next := cvs[i+1]

		if next-prev != 1 {
			return false, 0, 0
		}
	}

	return true, cvs[0], cardlen
}

/**
 * 判断牌是否 双顺,连队
 *
 * @param cds 牌的集合
 * @return 是 返回true；否则，返回false
 * @return  int :是 返回连队中最小牌的牌值 否则返回0
 * @return  int :是 该值表明是几连 ，否则返回0
 */
func isDbContinue(cds []int) (bool, int, int) {
	cardlen := len(cds)

	//判断牌数量
	if cardlen < 6 || cardlen%2 != 0 {
		return false, 0, 0
	}

	cvs := GetCardsValue(cds)
	//递增排序
	sort.Ints(cvs)

	// 小王、大王、2不能加入
	if cvs[cardlen-1] == 19 || cvs[cardlen-1] == 18 || cvs[cardlen-1] == 16 {
		return false, 0, 0
	}

	//判断前面n-1对
	for i := 0; i < cardlen/2-1; i++ {
		if cvs[i*2] != cvs[i*2+1] {
			return false, 0, 0
		}

		if cvs[i*2+2]-cvs[i*2] != 1 {
			return false, 0, 0
		}
	}

	//判断最后1对是否相等
	if cvs[cardlen-1] != cvs[cardlen-2] {
		return false, 0, 0
	}

	return true, cvs[0], cardlen / 2
}

/**
 * 判断牌是否 三顺,飞机不带
 *
 * @param cds 牌的集合
 * @return 是 返回true；否则，返回false
 * @return  int :是 返回三顺中最小牌的牌值 否则返回0
 * @return  int :是 该值表明是几连 ，否则返回0
 */
func isTbContinue(cds []int) (bool, int, int) {
	cardlen := len(cds)

	//判断牌数量
	if cardlen < 6 || cardlen%3 != 0 {
		return false, 0, 0
	}

	cvs := GetCardsValue(cds)
	sort.Ints(cvs)

	// 小王、大王、2不能加入
	if cvs[cardlen-1] == 19 || cvs[cardlen-1] == 18 || cvs[cardlen-1] == 16 {
		return false, 0, 0
	}

	//判断前n-1对
	for i := 0; i < cardlen/3-1; i++ {
		if !isSameCards(cvs[i*3 : i*3+3]) {
			return false, 0, 0
		}
		if cvs[i*3+3]-cvs[i*3] != 1 {
			return false, 0, 0
		}
	}

	//判断最后三个是否相同
	if !isSameCards(cvs[cardlen-3 : cardlen]) {
		return false, 0, 0
	}

	return true, cvs[0], cardlen / 3
}

/**
 * 判断牌是否 飞机带翅膀--飞机带单
 *
 * @param cds 牌的集合
 * @return 是 返回true；否则，返回false
 * @return  int :是 返回三顺中最小牌的牌值 否则返回0
 * @return  int :是 该值表明是几连 ，否则返回0
 */
func isAirplaneSingle(cds []int) (bool, int, int) {
	cardlen := len(cds)

	//判断牌数量
	if cardlen < 8 || cardlen > 20 || cardlen%4 != 0 {
		return false, 0, 0
	}
	cvs := GetCardsValue(cds)
	sort.Ints(cvs)

	//tbc存放三张相同的牌的其中一张,eg:AAA只需存放A
	var tbc []int
	//tw存放带的单牌
	var tw []int

	var index int
	var branch bool

	//从排好序的牌中挑出重复的牌和单牌
	for i := 3; i <= cardlen; {
		if isSameCards(cvs[i-3 : i]) {
			tbc = append(tbc, cvs[i-3])
			index = i
			branch = true
			i = i + 3
		} else {
			tw = append(tw, cvs[i-3])
			index = i
			branch = false
			i++
		}
	}
	if branch {
		tw = append(tw, cvs[index:cardlen]...)
	} else {
		tw = append(tw, cvs[cardlen-2:cardlen]...)
	}

	//判断重复的牌和单牌数量是否相等
	if len(tbc) != len(tw) {
		return false, 0, 0
	}

	//判断重复的牌是否是为连续的
	for i := 0; i < len(tbc)-1; i++ {
		prev := tbc[i]
		next := tbc[i+1]

		if next-prev != 1 {
			return false, 0, 0
		}
	}

	return true, tbc[0], len(tbc)
}

/**
 * 判断牌是否 飞机带翅膀--飞机带对
 *
 * @param cds 牌的集合
 * @return 是 返回true；否则，返回false
 * @return  int :是 返回三顺中最小牌的牌值 否则返回0
 * @return  int :是 该值表明是几连 ，否则返回0
 */
func isAirplaneDouble(cds []int) (bool, int, int) {
	cardlen := len(cds)

	//判断牌数量
	if cardlen < 10 || cardlen > 20 || cardlen%5 != 0 {
		return false, 0, 0
	}
	cvs := GetCardsValue(cds)
	sort.Ints(cvs)

	//tbc存放三张相同的牌，eg：AAABBB
	var tbc []int

	//tw存放带的对牌,eg：CCDD
	var tw []int

	var index int
	var branch bool

	//从排好序的牌中挑出三张重复的牌和对牌
	for i := 3; i <= cardlen; {
		if isSameCards(cvs[i-3 : i]) {
			tbc = append(tbc, cvs[i-3:i]...)
			index = i
			branch = true
			i = i + 3
		} else {
			tw = append(tw, cvs[i-3])
			index = i
			branch = false
			i++
		}

	}
	if branch {
		tw = append(tw, cvs[index:cardlen]...)
	} else {
		tw = append(tw, cvs[cardlen-2:cardlen]...)
	}

	//判断重复的牌和对牌数量是否相等
	if len(tbc)/3 != len(tw)/2 {
		return false, 0, 0
	}

	//判断重复的牌是否是为连续的
	for i := 0; i < len(tbc)/3-1; i++ {
		if tbc[i*3+3]-tbc[i*3] != 1 {
			return false, 0, 0
		}
	}

	//判断对牌
	for i := 0; i < len(tw)/2-1; i++ {
		if tw[i*2+1] != tw[i*2] {
			return false, 0, 0
		}
	}
	if tw[len(tw)-1] != tw[len(tw)-2] {
		return false, 0, 0
	}

	return true, tbc[0], len(tbc) / 3

}

/**
 * 判断牌是否 四带两单
 *
 * @param cds 牌的集合
 * @return 是 返回true；否则，返回false
 * @return  int :是 返回四张相同牌的牌值 否则返回0
 */
func isFourTwo(cds []int) (bool, int) {
	cardlen := len(cds)

	//判断牌数量
	if cardlen != 6 {
		return false, 0
	}
	cvs := GetCardsValue(cds)
	sort.Ints(cvs)

	//查看是否有连续四个相等的
	for i := 0; i < 3; i++ {
		if isSameCards(cvs[i : i+4]) {
			return true, cvs[i]
		}
	}
	return false, 0
}

/**
 * 判断牌是否 四带两队
 *
 * @param cds 牌的集合
 * @return 是 返回true；否则，返回false
 * @return  int :是 返回四张相同牌的牌值 否则返回0
 */
func isFourTwo4(cds []int) (bool, int) {
	cardlen := len(cds)

	//判断牌数量
	if cardlen != 8 {
		return false, 0
	}
	cvs := GetCardsValue(cds)
	sort.Ints(cvs)

	//查看是否有连续四个相等的
	var i int
	for i = 0; i < 5; i++ {
		if isSameCards(cvs[i : i+4]) {
			break
		}
	}
	if i >= 5 {
		return false, 0
	}

	//获取其余四个的值
	pre := cvs[0:i]
	aft := cvs[i+4 : cardlen]
	prelen := len(pre)

	if prelen == 0 {
		if aft[0] == aft[1] && aft[2] == aft[3] {
			return true, cvs[i]
		}
	} else if prelen == 2 {
		if pre[0] == pre[1] && aft[0] == aft[1] {
			return true, cvs[i]
		}
	} else if prelen == 4 {
		if pre[0] == pre[1] && pre[2] == pre[3] {
			return true, cvs[i]
		}
	}

	return false, 0
}

/**
 * 判断牌是否 炸弹
 *
 * @param cds 牌的集合
 * @return 是 返回true；否则，返回false
 * @return  int :是 返回四张相同牌的牌值 否则返回0
 */
func isBomb(cds []int) (bool, int) {
	cardlen := len(cds)

	//判断牌数量
	if cardlen != 4 {
		return false, 0
	}
	cvs := GetCardsValue(cds)
	if isSameCards(cvs) {
		return true, cvs[0]
	}
	return false, 0

}

/**
 * 判断牌是否 火箭
 *
 * @param cds 牌的集合
 * @return 是 返回true；否则，返回false
 */
func isRocket(cds []int) bool {
	cardlen := len(cds)

	//判断牌数量
	if cardlen != 2 {
		return false
	}
	sort.Ints(cds)
	if cds[0] == 518 && cds[1] == 519 {
		return true
	}
	return false

}
