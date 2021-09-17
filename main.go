package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/VictoriaMetrics/metricsql"
)

var (
	flVersion = flag.Bool("version", false, "print version")
	flDebug   = flag.Bool("debug", false, "enable debug")

	BuildTime    string
	BuildBranch  string
	BuildVersion string
)

const version = "0.1.0"

func init() {
	flag.Parse()
	if *flVersion {
		log.Printf("PromQL Prettier %s\nGit branch: %s\nGit commit: %s\nBuild: %s\n",
			version, BuildBranch, BuildVersion, BuildTime)

		os.Exit(0)
	}
	if *flDebug {
		go func() {
			// http://localhost:5002/debug/pprof/
			http.ListenAndServe("localhost:5002", nil)
		}()
	}
}

func main() {
	promql, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	expr, err := metricsql.Parse(string(promql))
	if err != nil {
		log.Fatalf("invalid promql, err: %v", err)
	}

	ret := prettier(expr, 0)
	fmt.Printf("%s", ret)

}

func genPadding(ident int) string {
	return strings.Repeat("  ", ident)
}

func needParens(expr metricsql.Expr) bool {
	switch e := expr.(type) {
	case *metricsql.MetricExpr, *metricsql.NumberExpr, *metricsql.StringExpr:
		return false
	case *metricsql.FuncExpr:
		return e.Name != "time"
	default:
		return true
	}
}

func wrapParensWhenNecesary(expr metricsql.Expr, b *bytes.Buffer, ident int) {
	paddings := genPadding(ident)
	if needParens(expr) {
		b.WriteString(paddings + "(\n")
		b.Write(prettier(expr, ident+1))
		b.WriteString("\n" + paddings + ")")
	} else {
		b.Write(prettier(expr, ident))
	}
}

func prettier(expr metricsql.Expr, ident int) []byte {
	paddings := genPadding(ident)
	var buf []byte

	switch e := expr.(type) {
	case *metricsql.MetricExpr, *metricsql.NumberExpr, *metricsql.StringExpr:
		buf = append(buf, paddings...)
		buf = e.AppendString(buf)
	case *metricsql.RollupExpr:
		var b bytes.Buffer

		wrapParensWhenNecesary(e.Expr, &b, ident)
		if len(e.Window) > 0 || len(e.Step) > 0 {
			b.WriteString(fmt.Sprintf("[%s:%s]", e.Window, e.Step))
		}
		if len(e.Offset) > 0 {
			b.WriteString(fmt.Sprintf(" offset %s", e.Offset))
		}

		buf = append(buf, b.Bytes()...)
	case *metricsql.BinaryOpExpr:
		var b bytes.Buffer

		wrapParensWhenNecesary(e.Left, &b, ident+1)
		b.WriteString(fmt.Sprintf("\n%s%s", paddings, e.Op))
		if e.Bool {
			b.WriteString(" bool")
		}
		if e.GroupModifier.Op != "" {
			b.WriteString(" ")
			b.Write(e.GroupModifier.AppendString(nil))
		}
		if e.JoinModifier.Op != "" {
			b.WriteString(" ")
			b.Write(e.JoinModifier.AppendString(nil))
		}
		b.WriteString("\n")
		wrapParensWhenNecesary(e.Right, &b, ident+1)

		buf = append(buf, b.Bytes()...)
	case *metricsql.AggrFuncExpr:
		var b bytes.Buffer

		b.WriteString(paddings + e.Name)
		if e.Modifier.Op != "" {
			b.WriteString(" ")
			b.Write(e.Modifier.AppendString(nil))
		}
		b.WriteString(" (\n")
		for i, a := range e.Args {
			b.Write(prettier(a, ident+1))
			if i < len(e.Args)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString(paddings + ")")

		buf = append(buf, b.Bytes()...)
	case *metricsql.FuncExpr:
		if e.Name == "time" {
			buf = append(buf, []byte(paddings+"time ()")...)
		} else {
			var b bytes.Buffer

			b.WriteString(paddings + e.Name + " (\n")
			for i, a := range e.Args {
				b.Write(prettier(a, ident+1))
				if i < len(e.Args)-1 {
					b.WriteString(",")
				}
				b.WriteString("\n")
			}
			b.WriteString(paddings + ")")

			buf = append(buf, b.Bytes()...)
		}
	}

	return buf
}
