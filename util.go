package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/VictoriaMetrics/metricsql"
)

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
		if (*e.Window).Duration(1) > 0 || e.Step.Duration(1) > 0 {
			// hack: cannot access `DurationExpr.s` directly, but can get it by appending empty bytes to `DurationExpr`
			b.WriteString(fmt.Sprintf("[%s:%s]", e.Window.AppendString([]byte{}), e.Step.AppendString([]byte{})))
		}
		if e.Offset.Duration(1) > 0 {
			b.WriteString(fmt.Sprintf(" offset %s", e.Offset.AppendString([]byte{})))
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

func Prettier(s string) (string, error) {
	expr, err := metricsql.Parse(s)
	if err != nil {
		return "", err
	}

	return string(prettier(expr, 0)), nil
}
