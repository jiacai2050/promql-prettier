package main

import (
	"fmt"
	"syscall/js"

	"github.com/VictoriaMetrics/metricsql"
)

var c chan struct{}

func init() {
	c = make(chan struct{})
}

var cb js.Func

func main() {
	g := js.Global()
	doc := g.Get("document")

	input := doc.Call("getElementById", "input")
	output := doc.Call("getElementById", "output")
	cb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		promql := input.Get("value").String()
		fmt.Println("input is %s", promql)
		expr, err := metricsql.Parse(promql)
		if err != nil {
			g.Call("alert", err.Error())
			return nil
		}

		ret := string(prettier(expr, 0))
		output.Set("innerHTML", js.ValueOf(ret))
		return nil
	})

	doc.Call("getElementById", "prettierBtn").Call("addEventListener", "click", cb)

	<-c
	fmt.Println("WebAssembly App Finish!")
}
