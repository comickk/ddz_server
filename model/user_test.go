package model

import (
	"fmt"
	"testing"
)

/*
func TestNewUser(t *testing.T) {
	u, err := NewUser("张三", "河南", "郑州", "CN", "http://ksdfuewr.com/dwerd./0", "wesdfrew", 1)
	if err != nil {
		t.Error(err.Error())
	}
	fmt.Println(u)
}

func TestGetUser(t *testing.T) {
	u, has := GetUser("wesdfrew")
	if has {
		fmt.Println(u)
	} else {
		fmt.Println("wesdfrew not exist")
	}

}
*/
func TestUpdateUser(t *testing.T) {
	bm := make(map[string]interface{})
	bm["img"] = 1
	bm["up"] = 10200
	bm["ct"] = 10
	bm["wn"] = 5
	err := UpdateUser("59647e7f1863ad1778000001", bm)
	if err != nil {
		t.Error(err.Error())
	}
	fmt.Println("update success")
}
