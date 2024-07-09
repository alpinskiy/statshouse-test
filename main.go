package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"
)

type options struct {
	sampleNum int
	maxStrLen int
	viewCode  bool
	keepTemp  bool
	fail      bool
}

func parseCommandLine() options {
	var res options
	flag.IntVar(&res.sampleNum, "n", 200, "number of samples")
	flag.IntVar(&res.maxStrLen, "max-str-len", 32, "maximum string length")
	flag.BoolVar(&res.viewCode, "view-code", false, "open generated source files in Visual Studio Code")
	flag.BoolVar(&res.keepTemp, "keep-temp", false, "do not remove generated temporary files")
	flag.BoolVar(&res.fail, "fail", false, "run tests which are expected to fail")
	flag.Parse()
	return res
}

func main() {
	opt := parseCommandLine()
	data := newTestData(opt)
	compareLibs(data, opt, &cppTransport{}, &cppRegistry{}, &rust{}, &java{})
	if opt.fail {
		compareLibs(data, opt, &cppTransport{}, &golang{}) // go library does not send timestamp
		compareLibs(data, opt, &python{})                  // TODO: write template
	}
}

func compareLibs(data any, opt options, libs ...lib) {
	s := make([]string, len(libs))
	for i, v := range libs {
		s[i] = typeName(v)
	}
	fmt.Printf("compare [%s]", strings.Join(s, ", "))
	var err error
	defer func() {
		if err == nil {
			fmt.Println("test passed!")
		} else {
			log.Println(err)
			fmt.Println("test failed!")
		}
	}()
	out := make([]series, len(libs))
	for i := range out {
		out[i] = series{}
	}
	var i int
	var cancel func()
	cancel, err = listenUDP(":13337", func(b []byte) {
		out[i].parseAdd(b)
	})
	if err != nil {
		return
	}
	defer cancel()
	// run
	var l lib
	for i, l = range libs {
		if err = runClient(l, data, opt); err != nil {
			return
		}
		log.Println("wait a second for data")
		time.Sleep(time.Second)
	}
	// compare output
	out[0].normalize()
	for i := 1; i < len(out); i++ {
		out[i].normalize()
		if diff := compareSeries(out[0], out[i]); !diff.empty() {
			err = fmt.Errorf("%s != %s: %s", typeName(libs[0]), typeName(libs[i]), diff.String())
			return
		}
	}
}
