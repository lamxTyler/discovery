package random

type RandomList interface {
	Len() int                // 列表的长度
	Index(i int) interface{} // 根据索引返回项目
	Weight(i int) int        // 每个项目的权重
}

type RandListWithWeight interface {
	RandomList
	TotalWeight() int
}

func RandomAward(r RandomList) interface{} {
	listLen := r.Len()
	if listLen == 0 {
		return nil
	}
	totalWeight := 0
	if rs, ok := r.(RandListWithWeight); ok {
		totalWeight = rs.TotalWeight()
	} else {
		for i := 0; i < listLen; i++ {
			totalWeight += r.Weight(i)
		}
	}
	if totalWeight == 0 {
		return nil
	}
	randNum := Between(0, totalWeight)
	for i := 0; i < listLen; i++ {
		itemWeight := r.Weight(i)
		if randNum < itemWeight {
			return r.Index(i)
		} else {
			randNum -= itemWeight
		}
	}
	return nil
}

func RandomAwardEx(r RandomList) (int, interface{}) {
	listLen := r.Len()
	if listLen == 0 {
		return -1, nil
	}
	totalWeight := 0
	if rs, ok := r.(RandListWithWeight); ok {
		totalWeight = rs.TotalWeight()
	} else {
		for i := 0; i < listLen; i++ {
			totalWeight += r.Weight(i)
		}
	}
	if totalWeight == 0 {
		return -1, nil
	}
	randNum := Between(0, totalWeight)
	for i := 0; i < listLen; i++ {
		itemWeight := r.Weight(i)
		if randNum < itemWeight {
			return i, r.Index(i)
		} else {
			randNum -= itemWeight
		}
	}
	return -1, nil
}
