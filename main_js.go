package main

import (
	"syscall/js"

	"github.com/VictoriaMetrics/metricsql"
)

var (
	blockingCh chan struct{}
)

func init() {
	blockingCh = make(chan struct{})
}

func main() {
	g := js.Global()
	doc := g.Get("document")

	input := doc.Call("getElementById", "input")
	output := doc.Call("getElementById", "output")

	prettier := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		promql := input.Get("value").String()
		// fmt.Printf(promql)
		if promql == "" {
			promql = input.Get("placeholder").String()
		}

		expr, err := metricsql.Parse(promql)
		if err != nil {
			g.Call("alert", err.Error())
			return nil
		}

		ret := string(prettier(expr, 0))
		output.Set("innerHTML", js.ValueOf(ret))
		return nil
	})

	doc.Call("getElementById", "prettierBtn").Call("addEventListener", "click", prettier)

	<-blockingCh
}
