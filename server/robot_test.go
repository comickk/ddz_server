package server

import (
	"UULoServer/gamerule"
	"fmt"
	"testing"
)

func TestSplitCards(t *testing.T) {
	/*var ss []int
	if ss == nil {
		fmt.Println(ss)
	}
	fmt.Println(len(ss))
	ss = make([]int, 10)
	fmt.Println(len(ss))
	var mp map[int][]int
	fmt.Println(len(mp))
	cn := make(map[int]int, 15)
	fmt.Println(len(cn))*/

	robot := &Robot{}
	fmt.Println(robot)
	cds := gamerule.GenRandCards()
	fmt.Println(cds[0:17])
	robot.SplitCards(cds[0:17])
}
