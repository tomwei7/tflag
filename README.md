# tflag

simple golang flag bind tool

### install

```bash
go get github.com/tomwei7/tflag
```

### usage

```go
package main

import (
    "time"
    "flag"
    "log"

    "github.com/tomwei7/tflag"
)

type Config struct {
	Addr    string        `flag:"addr" usage:"listen address" default:"127.0.0.1:2233"`
	Timeout time.Duration `flag:"timeout" usage:"http timeout" default:"1s"`
}

func main() {
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
```

### TODO

- [x] support slice, 
- [ ] unitest
- [ ] support array
