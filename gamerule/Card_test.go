// Card_test.go
package gamerule

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGenRandCards(t *testing.T) {
	cds := GenRandCards()
	fmt.Println(cds)

	//每个人17张，留三张作为底牌
	res := make(map[string]interface{})

	res["cmd"] = "dealcards"
	res["again"] = 0
	//向ua发牌
	res["u3"] = cds[0:17]
	res["u1"] = cds[0:0]
	res["u2"] = cds[0:0]
	data, err := json.Marshal(&res)
	if err != nil {
		fmt.Errorf("json marshal error:%s", err.Error())
	}
	fmt.Println(string(data))
}

func TestGetCardsValue(t *testing.T) {
	cds := GenRandCards()
	fmt.Print(GetCardsValue(cds))
}

func TestGetCardsType(t *testing.T) {
	//对牌
	tp, v, n := GetCardsType([]int{416, 316})
	t.Log("type:", tp, " value:", v)

	//三不带
	tp, v, n = GetCardsType([]int{411, 311, 211})
	t.Log("type:", tp, " value:", v)

	//三带一
	tp, v, n = GetCardsType([]int{403, 303, 203, 111})
	t.Log("type:", tp, " value:", v)
	//三带一对
	tp, v, n = GetCardsType([]int{404, 403, 104, 303, 204})
	t.Log("type:", tp, " value:", v)

	//顺子
	tp, v, n = GetCardsType([]int{308, 111, 204, 403, 309, 107, 305, 106, 410})
	t.Log("type:", tp, " value:", v, n, "lian")

	//连队
	tp, v, n = GetCardsType([]int{303, 104, 204, 403, 105, 405})
	t.Log("type:", tp, " value:", v, n, "lian")

	//飞机不带
	tp, v, n = GetCardsType([]int{304, 104, 204, 405, 105, 405})
	t.Log("type:", tp, " value:", v, n, "lian")

	//飞机带单
	tp, v, n = GetCardsType([]int{304, 104, 204, 405, 105, 405, 209, 111})
	t.Log("type:", tp, " value:", v, n, "lian")

	//飞机带单
	tp, v, n = GetCardsType([]int{304, 104, 204, 405, 105, 205, 106, 206, 306, 211, 111, 109})
	t.Log("type:", tp, " value:", v, n, "lian")

	//飞机带对
	tp, v, n = GetCardsType([]int{304, 104, 204, 405, 105, 405, 211, 111, 413, 113})
	t.Log("type:", tp, " value:", v, n, "lian")

	//四带两单
	tp, v, n = GetCardsType([]int{304, 104, 204, 404, 413, 113})
	t.Log("type:", tp, " value:", v)

	//四带两对
	tp, v, n = GetCardsType([]int{304, 104, 204, 404, 413, 113, 111, 311})
	t.Log("type:", tp, " value:", v)

	//炸弹
	tp, v, n = GetCardsType([]int{413, 113, 213, 313})
	t.Log("type:", tp, " value:", v)

	//火箭
	tp, v, n = GetCardsType([]int{519, 518})
	t.Log("type:", tp)

}

func TestDifference(t *testing.T) {
	src := []int{304, 104, 204, 405, 105, 205, 106, 206, 306, 211, 111, 109}
	des := []int{306, 211, 111, 109}

	fmt.Println(Difference(src, des))
}
