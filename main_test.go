package main_test

import (
	"testing"

	main "promql-prettier"
)

func TestFormatting(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: `go_goroutines{instance!="localhost:9090", job!~"prometheus.*"}`, want: `go_goroutines{instance!="localhost:9090", job!~"prometheus.*"}`},
		{input: `{__name__="go_goroutines",instance="localhost:9090",job="prometheus"}`, want: `go_goroutines{instance="localhost:9090", job="prometheus"}`},
		{input: "(((metric_name_long)))", want: "metric_name_long"},
		{input: `histogram_quantile(0.9, rate(instance_cpu_time_seconds{app="lion", proc="web",job="cluster-manager"}[5m]))`, want: `histogram_quantile (
  0.9,
  rate (
    instance_cpu_time_seconds{app="lion", proc="web", job="cluster-manager"}[5m:]
  )
)`},
		{input: `topk(5, (sum without(env) (instance_cpu_time_ns{app="lion", proc="web", rev="34d0f99", env="prod", job="cluster-manager"})))`, want: `topk (
  5,
  sum without (env) (
    instance_cpu_time_ns{app="lion", proc="web", rev="34d0f99", env="prod", job="cluster-manager"}
  )
)`},
		{input: `sum (instance_cpu_time_ns{app="lion", proc="web", rev="34d0f99", env="prod", job="cluster-manager"}) without (label)`, want: `sum without (label) (
  instance_cpu_time_ns{app="lion", proc="web", rev="34d0f99", env="prod", job="cluster-manager"}
)`},
		{input: `sum (instance_cpu_time_ns{app="lion", proc="web", rev="34d0f99", env="prod", job="cluster-manager"} + http_request_total{job="apiserver", handler="/api/comments"}) without (label)`, want: `sum without (label) (
    instance_cpu_time_ns{app="lion", proc="web", rev="34d0f99", env="prod", job="cluster-manager"}
  +
    http_request_total{job="apiserver", handler="/api/comments"}
)`},
		{input: `first_long{foo="bar", hello="world"} + second{foo="bar"} + third{foo="bar", localhost="9090"} + forth`, want: `  (
      (
          first_long{foo="bar", hello="world"}
        +
          second{foo="bar"}
      )
    +
      third{foo="bar", localhost="9090"}
  )
+
  forth`},
	}

	for _, test := range tests {
		got, err := main.Prettier(test.input)
		want := test.want

		if err != nil {
			t.Errorf("got %q, want %q", err, want)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

}
