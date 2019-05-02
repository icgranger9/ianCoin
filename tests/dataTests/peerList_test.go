package dataTests

import (
	"../../p3/data"
	"fmt"
	"testing"
)

func TestNewPeerList(t *testing.T){
	pl := data.NewPeerList(9, 32)
	id := pl.GetSelfId()

	if id != 9{
		t.Log("\tFailed with id: ", id)
		t.Fail()
	}
}

//func TestSimpleAdd(t *testing.T){
//	pl := data.NewPeerList(9, 32)
//	pl.Add("192.168.0.0", 0, "00")
//	pl.Add("192.168.1.1", 1, "11")
//	pl.Add("192.168.2.2", 2, "22")
//	pl.Add("192.168.3.3", 3, "33")
//	pl.Add("192.168.4.4", 4, "44")
//	pl.Add("192.168.5.5", 5, "55")
//
//	res := pl.Show()
//
//
//	if res == "" {
//		t.Log(res)
//		t.Fail()
//	}
//}
//
//func TestAdvancedAdd(t *testing.T){
//	pl := data.NewPeerList(9, 32)
//
//	numAdditions := 64
//
//	for val :=0; val < numAdditions; val++{
//		str := "192.168." + fmt.Sprint(val) + "." + fmt.Sprint(val)
//		key := fmt.Sprint(val) + fmt.Sprint(val)
//		pl.Add(str, int32(val), key)
//	}
//
//	for val :=0; val < numAdditions; val++{
//		str := "192.168." + fmt.Sprint(val) + "." + fmt.Sprint(val)
//		key := fmt.Sprint(val) + fmt.Sprint(val)
//		pl.Add(str, int32(val), key)
//	}
//
//	res := pl.Show()
//
//
//	//note: minus one because we can't add our own id
//	if strings.Count(res, "addr") != numAdditions-1 {
//		t.Log(res)
//		t.Fail()
//	}
//}
//
//func TestSimpleDelete(t *testing.T){
//	maxLen := 32
//	pl := data.NewPeerList(9, int32(maxLen))
//
//	numAdditions := 35
//
//	for val :=0; val < numAdditions; val++{
//		str := "192.168." + fmt.Sprint(val) + "." + fmt.Sprint(val)
//		key := fmt.Sprint(val) + fmt.Sprint(val)
//		pl.Add(str, int32(val), key)
//	}
//
//	pl.Delete("192.168.10.10")
//
//	res := pl.Show()
//
//	//mins two because we can't add our own id, and for the one we deleted
//	if strings.Count(res, "addr") != numAdditions-2 {
//		t.Log(res)
//		t.Fail()
//	}
//}
//
//func TestAdvancedDelete(t *testing.T){
//	maxLen := 32
//	pl := data.NewPeerList(9, int32(maxLen))
//
//	numAdditions := 35
//
//	for val :=0; val < numAdditions; val++{
//		str := "192.168." + fmt.Sprint(val) + "." + fmt.Sprint(val)
//		key := fmt.Sprint(val) + fmt.Sprint(val)
//		pl.Add(str, int32(val), key)
//	}
//
//	numDeletions := 35
//
//	for val :=0; val < numDeletions; val++{
//		str := "192.168." + fmt.Sprint(val) + "." + fmt.Sprint(val)
//		pl.Delete(str)
//	}
//
//	res := pl.Show()
//
//	//mins two because we can't add our own id, and for the one we deleted
//	if strings.Count(res, "addr") != numAdditions-numDeletions {
//		t.Log(res)
//		t.Fail()
//	}
//}
//
//func TestSimpleRebalance(t *testing.T){
//	maxLen := 10
//	pl := data.NewPeerList(3, int32(maxLen))
//
//	numAdditions := 35
//
//	for val :=0; val < numAdditions; val++{
//		str := "192.168." + fmt.Sprint(val) + "." + fmt.Sprint(val)
//		key := fmt.Sprint(val) + fmt.Sprint(val)
//		pl.Add(str, int32(val), key)
//	}
//
//	pl.Rebalance()
//
//	res := pl.Show()
//
//
//	if strings.Count(res, "addr") != maxLen {
//		t.Log(res)
//		t.Fail()
//	} else {
//		t.Log(res)
//	}
//}
//
//func TestAdvancedRebalance(t *testing.T){
//
//	data.TestPeerListRebalance()
//}

func TestSimpleJson(t *testing.T){
	maxLen := 32
	pl := data.NewPeerList(3, int32(maxLen))

	numAdditions := 4

	for val :=9999999999; val < numAdditions+9999999999; val++{
		str := "192.168." + fmt.Sprint(val) + "." + fmt.Sprint(val)
		key := fmt.Sprint(val) + ":" + fmt.Sprint(val)
		pl.Add(str, int32(val), key)
	}

	fmt.Println(pl.Show())

	json, _ := pl.PeerMapToJson()
	fmt.Println(json)
}