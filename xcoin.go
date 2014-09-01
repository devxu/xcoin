package main

import (
	"encoding/json"
	"fmt"
	"github.com/nsf/termbox-go"
	"io/ioutil"
	"net/http"
	"os"
	// "reflect"
	"sort"
	"time"
)

type Coin struct {
	Name string
	High float64
	Low  float64
	Last float64
	Vol  float64
	Buy  float64
	Sell float64
}

type CoinSorter []*Coin

func (a CoinSorter) Len() int           { return len(a) }
func (a CoinSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a CoinSorter) Less(i, j int) bool { return a[j].Vol*a[j].Last < a[i].Vol*a[i].Last }

var hasNewData = make(chan bool)
var dataUpdateTime time.Time
var coinList = make([]*Coin, 0)

func main() {

	err := termbox.Init()
	checkError(err)
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc)

	go func() {
		for {
			queryCoinData()
			time.Sleep(10 * time.Second)
		}
	}()

	go func() {
		for {
			<-hasNewData
			redraw()
		}
	}()

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				return
			default:
				if ev.Ch == 'q' {
					return
				}
				continue
			}
		case termbox.EventError:
			checkError(ev.Err)
		}

	}
}

func queryCoinData() {
	url := "http://api.btc38.com/v1/ticker.php?c=all&mk_type=cny"
	resp, err := http.Get(url)
	checkError(err)
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(data))
	if data == nil {
		fmt.Println("data is empty")
		return
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	checkError(err)

	//清空
	coinList = coinList[len(coinList):]

	for k, v := range result {
		var record = v.(map[string]interface{})
		if record["ticker"] == nil || record["ticker"] == "" {
			continue
		}

		var ticker = record["ticker"].(map[string]interface{})
		high := ticker["high"].(float64)
		low := ticker["low"].(float64)
		last := ticker["last"].(float64)
		vol := ticker["vol"].(float64)
		buy := ticker["buy"].(float64)
		sell := ticker["sell"].(float64)

		coin := &Coin{}
		coin.Name = k
		coin.Last = last
		coin.High = high
		coin.Low = low
		coin.Vol = vol
		coin.Buy = buy
		coin.Sell = sell

		coinList = append(coinList, coin)
	}

	sort.Sort(CoinSorter(coinList))
	dataUpdateTime = time.Now()
	hasNewData <- true
}

func redraw() {
	const colorDef = termbox.ColorDefault
	termbox.Clear(colorDef, colorDef)
	w, h := termbox.Size()

	top_margin := 2
	left_margin := 5

	rect_width := w - 2*left_margin
	rect_height := h - 2*top_margin

	printRune(left_margin, top_margin, '┌')
	printRune(left_margin, top_margin+rect_height-1, '└')
	printRune(left_margin+rect_width-1, top_margin, '┐')
	printRune(left_margin+rect_width-1, top_margin+rect_height-1, '┘')
	fill(left_margin+1, top_margin, rect_width-2, 1, '─')
	fill(left_margin+1, top_margin+rect_height-1, rect_width-2, 1, '─')
	fill(left_margin, top_margin+1, 1, rect_height-2, '│')
	fill(left_margin+rect_width-1, top_margin+1, 1, rect_height-2, '│')

	headX := left_margin + 5
	headY := top_margin + 1
	printStr(headX, headY, "%s", "Seq")
	headX += 8
	printStr(headX, headY, "%s", "Name")
	headX += 10
	printStr(headX, headY, "%s", "Last")
	headX += 15
	printStr(headX, headY, "%s", "High")
	headX += 15
	printStr(headX, headY, "%s", "Low")
	headX += 15
	printStr(headX, headY, "%s", "Vol")
	headX += 20
	printStr(headX, headY, "%s", "Buy")
	headX += 15
	printStr(headX, headY, "%s", "Sell")
	headX += 10
	printStr(headX, headY, "Update time: %s", dataUpdateTime.Format("15:04:05"))

	if len(coinList) > 0 {
		for i, coin := range coinList {
			x := left_margin + 5
			y := top_margin + 3 + i

			printStr(x, y, "%d", i+1)
			x += 8
			printStr(x, y, "%s", coin.Name)
			x += 10
			printStr(x, y, "%.4f", coin.Last)
			x += 15
			printStr(x, y, "%.4f", coin.High)
			x += 15
			printStr(x, y, "%.4f", coin.Low)
			x += 15
			printStr(x, y, "%.4f", coin.Vol)
			x += 20
			printStr(x, y, "%.4f", coin.Buy)
			x += 15
			printStr(x, y, "%.4f", coin.Sell)
		}
	}

	termbox.Flush()
}

func fill(x, y, w, h int, ch rune) {
	for ly := 0; ly < h; ly++ {
		for lx := 0; lx < w; lx++ {
			printRune(x+lx, y+ly, ch)
		}
	}
}

func printRune(x, y int, ch rune) {
	termbox.SetCell(x, y, ch, termbox.ColorDefault, termbox.ColorDefault)
}

func printStr(x, y int, format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	for _, ch := range str {
		printRune(x, y, ch)
		x++
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("error = ", err)
		os.Exit(0)
	}
}
