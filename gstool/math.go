package gstool

import (
	"math/rand"
	"time"
)

// MathRandNumber 生成一个随机数 注意是 start <= result < end
func MathRandNumber(start int, end int) int {
	randList := MathRandNumberList(start, end, 1)
	return randList[0]
}

// MathRandNumberList 生成多个随机数
func MathRandNumberList(startNumber int, endNumber int, count int) []int {

	if endNumber < startNumber || (endNumber-startNumber) < count {
		return nil
	}
	nums := make([]int, 0)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		//生成随机数
		num := r.Intn(endNumber-startNumber) + startNumber

		//查重
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}
		if !exist {
			nums = append(nums, num)
		}
	}
	return nums
}

func MathMaxInt(params ...int) int {
	mxInt := 0
	for _, v := range params {
		if v > mxInt {
			mxInt = v
		}
	}
	return mxInt
}
