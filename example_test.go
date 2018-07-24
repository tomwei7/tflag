package tflag_test

import (
	"flag"
	"log"
	"time"

	"github.com/tomwei7/tflag"
)

func ExampleVar() {
	type Config struct {
		Addr    string        `flag:"addr" usage:"listen address" default:"127.0.0.1:2233"`
		Timeout time.Duration `flag:"timeout" usage:"http timeout" default:"1s"`
	}
	cfg := &Config{}
	parseFunc, err := tflag.Var("http", cfg)
	if err != nil {
		log.Fatalf("flagenv error: %s", err)
	}
	if !flag.Parsed() {
		flag.Parse()
	}

	// call parse function after flag parsed
	parseFunc()
}
