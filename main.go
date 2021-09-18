// +build !js

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

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
			if err := http.ListenAndServe("localhost:5002", nil); err != nil {
				panic(err)
			}
		}()
	}
}

func main() {
	promql, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	expr, err := metricsql.Parse(string(promql))
	if err != nil {
		panic(err)
	}

	ret := prettier(expr, 0)
	fmt.Printf("%s", ret)

}
