package trend

import (
	"fmt"
	"math"
	"time"

	"github.com/kurosann/aqt-sdk/stockinds/utils"
	"github.com/kurosann/aqt-sdk/stockinds/utils/commonutils"
)

// BreakoutProbability 超级趋势 struct
type BreakoutProbability struct {
	PercentageStep float64
	NumberOfLines  int
	Name           string
	data           []BreakoutProbabilityData
	kline          utils.Klines
}

type BreakoutProbabilityData struct {
	Time time.Time
	List []BreakoutProbabilityDataItem
	Win  float64
	Loss float64
}

type BreakoutProbabilityDataItem struct {
	// 价格
	Price float64
	Per   float64
	Up    bool
}

// NewBreakoutProbability new Func
//
//	Args:
//		percentageStep 级别之间的间距可以通过百分比步长进行调整。1% 表示每个级别位于前一个级别之上/之下 1%
//		numberOfLines 设置要计算的级别数量，最小1，最大5
func NewBreakoutProbability(list utils.Klines, percentageStep float64, numberOfLines int) *BreakoutProbability {

	if numberOfLines < 1 {
		numberOfLines = 1
	}
	if numberOfLines > 5 {
		numberOfLines = 5
	}
	m := &BreakoutProbability{
		Name:           fmt.Sprintf("BreakoutProbability%.2f-%d", percentageStep, numberOfLines),
		kline:          list,
		PercentageStep: percentageStep,
		NumberOfLines:  numberOfLines,
	}
	return m
}

// NewDefaultBreakoutProbability new Func
func NewDefaultBreakoutProbability(list utils.Klines) *BreakoutProbability {
	return NewBreakoutProbability(list, 1.2, 5)
}

// Calculation Func
func (e *BreakoutProbability) Calculation() *BreakoutProbability {

	var high = e.kline.GetOHLC().High
	var low = e.kline.GetOHLC().Low
	var open = e.kline.GetOHLC().Open
	var close = e.kline.GetOHLC().Close

	//第1、2行是up；
	// 3、4行是down;
	// 5 行是 第1列green汇总；第2列red汇总
	total := make([][]int, 7)
	for i := 0; i < 7; i++ {
		total[i] = make([]int, 4)
	}

	vals := make([][]float64, 5)
	for i := 0; i < 5; i++ {
		vals[i] = make([]float64, 4)
	}

	defer func() {
		high = nil
		low = nil
		open = nil
		close = nil
		total = nil
		vals = nil
	}()

	e.data = make([]BreakoutProbabilityData, len(e.kline))
	for klineIndex, row := range e.kline {
		if klineIndex == 0 {
			continue
		}

		var dataItem = BreakoutProbabilityData{
			Time: row.Time,
		}

		dataItem.List = make([]BreakoutProbabilityDataItem, 10)

		var h = row.High
		var l = row.Low
		var c = row.Close
		var step = c * (e.PercentageStep / 100)

		var green = commonutils.If(close[klineIndex-1] > open[klineIndex-1], true, false).(bool)
		var red = commonutils.If(close[klineIndex-1] < open[klineIndex-1], true, false).(bool)

		// 累加上涨、中跌的次数
		total[5][0] += commonutils.If(green, 1, 0).(int)
		total[5][1] += commonutils.If(red, 1, 0).(int)

		//Run Score Function
		e.score(0, 0, green, red, total, h, high[klineIndex-1], l, low[klineIndex-1], vals)
		e.score(step, 1, green, red, total, h, high[klineIndex-1], l, low[klineIndex-1], vals)
		e.score(step*2, 2, green, red, total, h, high[klineIndex-1], l, low[klineIndex-1], vals)
		e.score(step*3, 3, green, red, total, h, high[klineIndex-1], l, low[klineIndex-1], vals)
		e.score(step*4, 4, green, red, total, h, high[klineIndex-1], l, low[klineIndex-1], vals)

		//Fetch Score Values
		a1 := vals[0][0]
		b1 := vals[0][1]
		a2 := vals[0][2]
		b2 := vals[0][3]

		for i := 0; i < e.NumberOfLines; i++ {

			if (green && math.Min(vals[i][0], vals[i][1]) > 0) || (red && math.Min(vals[i][2], vals[i][3]) > 0) {

				// 用当前值，测试下一根线的概率
				hi := h + (step * float64(i))
				lo := l - (step * float64(i))

				dataItem.List[i] = BreakoutProbabilityDataItem{
					Price: hi,
					Per:   commonutils.If(green, vals[i][0], vals[i][2]).(float64),
					Up:    true,
				}
				dataItem.List[5+i] = BreakoutProbabilityDataItem{
					Price: lo,
					Per:   commonutils.If(green, vals[i][1], vals[i][3]).(float64),
					Up:    false,
				}

			}
		}

		//Run Backtest Function
		if green {
			if math.Max(a1, b1) == a1 {
				e.backtest(total, h, high[klineIndex-1], l, high[klineIndex-1])
			} else {
				e.backtest(total, h, high[klineIndex-1], l, low[klineIndex-1])
			}
		} else {
			if math.Max(a2, b2) == a2 {
				e.backtest(total, h, high[klineIndex-1], l, high[klineIndex-1])
			} else {
				e.backtest(total, h, high[klineIndex-1], l, low[klineIndex-1])
			}
		}

		dataItem.Win = float64(total[6][0])
		dataItem.Loss = float64(total[6][1])

		e.data[klineIndex] = dataItem

	}

	return e
}

// GetData Func
func (e *BreakoutProbability) GetData() []BreakoutProbabilityData {
	if len(e.data) == 0 {
		e = e.Calculation()
	}
	return e.data
}

func (e *BreakoutProbability) backtest(total [][]int, high, highPrev, low float64, v float64) {
	p1 := total[6][0]
	p2 := total[6][1]
	if v == highPrev {
		if high >= v {
			total[6][0] = p1 + 1
		} else {
			total[6][1] = p2 + 1
		}
	} else {
		if low <= v {
			total[6][0] = p1 + 1
		} else {
			total[6][1] = p2 + 1
		}
	}
}

// score Func
func (e *BreakoutProbability) score(x float64, rowIndex int, green, red bool, total [][]int, high, highPrev, low, lowPrev float64, vals [][]float64) {
	ghh := total[rowIndex][0]
	gll := total[rowIndex][1]
	rhh := total[rowIndex][2]
	rll := total[rowIndex][3]
	gtotal := total[5][0]
	rtotal := total[5][1]

	var hh = commonutils.If(high > highPrev+x, true, false).(bool)
	var ll = commonutils.If(low > lowPrev-x, true, false).(bool)

	if green && hh {
		total[rowIndex][0] = ghh + 1
		vals[rowIndex][0] = commonutils.FormatDecimalFloat64(((float64(ghh+1) / float64(gtotal)) * 100), -2)
	}
	if green && ll {
		total[rowIndex][1] = gll + 1
		vals[rowIndex][1] = commonutils.FormatDecimalFloat64(((float64(gll+1) / float64(gtotal)) * 100), -2)
	}
	if red && hh {
		total[rowIndex][2] = rhh + 1
		vals[rowIndex][2] = commonutils.FormatDecimalFloat64(((float64(rhh+1) / float64(rtotal)) * 100), -2)
	}
	if red && ll {
		total[rowIndex][3] = rll + 1
		vals[rowIndex][3] = commonutils.FormatDecimalFloat64(((float64(rll+1) / float64(rtotal)) * 100), -2)
	}
}
