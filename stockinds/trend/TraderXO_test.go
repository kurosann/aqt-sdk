package trend

import (
	"fmt"
	"testing"

	"github.com/kurosann/aqt-sdk/stockinds/utils"
)

// Run:
// go test -v ./trend -run TestTraderXO
func TestTraderXO(t *testing.T) {
	t.Parallel()
	list := utils.GetTestKline()

	stock := NewDefaultTraderXO(list)

	var dataList = stock.GetData()

	var side = stock.AnalysisSide()

	fmt.Printf("-- %s --\n", stock.Name)
	for i := len(dataList) - 1; i > 0; i-- {
		if i < len(dataList)-50 {
			break
		}
		var v = dataList[i]
		fmt.Printf("\t[%d]Time:%s\t Fast:%f\tSlow:%f\tSide:%s\n", i, v.Time.Format("2006-01-02 15:04:05"), v.Fast, v.Slow, side.Data[i].String())
	}
}
