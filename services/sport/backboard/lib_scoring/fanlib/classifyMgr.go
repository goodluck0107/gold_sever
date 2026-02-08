package fanlib

import (
	"sort"
)

/*
游戏番表管理单元
20190122 特殊番型除了7对、将一色、风一色。。。可能还有别的不用3n+2
*/
type BaseClass struct {
	fanShuArray        sort.IntSlice //根据番数的大小排序
	CustomerScoringMap map[int][]int //每个class层里支持多少个番型
	//20190201 苏大强 添加hukindmask
	HuKindMask uint16
}

//这个单元的初始化在每个游戏里面
func NewBaseClass() *BaseClass {
	return &BaseClass{
		CustomerScoringMap: make(map[int][]int),
	}
}

func (this *BaseClass) recordering(fanID int, checkIndex int) {
	bFind := false
	if this.fanShuArray.Len() != 0 {
		for _, index := range this.fanShuArray {
			if index == checkIndex {
				bFind = true
				break
			}
		}
	}
	//没有发现就添加
	if !bFind {
		//排序添加
		this.fanShuArray = append(this.fanShuArray, checkIndex)
		this.fanShuArray.Sort()
		//新创建番数序列
		this.CustomerScoringMap[checkIndex] = append(this.CustomerScoringMap[checkIndex], fanID)
		return
	}
	//可能没加到查表中去，所以还要检查一次
	//看看里面有没有指定番，没有新加
	bFind = false
	items := this.CustomerScoringMap[checkIndex]
	for _, item := range items {
		if item == fanID {
			bFind = true
			break
		}
	}
	if !bFind {
		this.CustomerScoringMap[checkIndex] = append(this.CustomerScoringMap[checkIndex], fanID)
	}
}
