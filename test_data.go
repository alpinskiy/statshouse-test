package main

import (
	"io"
	"log"
	"net/http"
	"strings"

	"pgregory.net/rand"
)

type testData struct {
	Metrics []metric
}

type metric struct {
	Timestamp uint32
	Name      string
	Tags      [][2]string
	Kind      int       // 0 counter, 1 value, 2 unique
	Count     float64   // always set
	Values    []float64 // Kind == 1
	Uniques   []int64   // Kind == 2
}

func newTestData(o options) testData {
	log.Printf("generate random data of length %d", o.sampleNum)
	str := make([]byte, o.maxStrLen)
	values := make([]float64, o.maxStrLen)
	uniques := make([]int64, o.maxStrLen)
	for i := range str {
		str[i] = 'U'
		values[i] = 0xAAAAAAAAAAAA
		uniques[i] = 0xAAAAAAAAAAAA
	}
	r := rand.New()
	res := make([]metric, o.sampleNum)
	for i := range res {
		res[i].generateRandom(str, values, uniques, r)
	}
	return testData{res}
}

func (t testData) Include(url string) string {
	log.Printf("download %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var sb strings.Builder
	if _, err = io.Copy(&sb, resp.Body); err != nil {
		panic(err)
	}
	return sb.String()
}

func (m *metric) generateRandom(str []byte, values []float64, uniques []int64, r *rand.Rand) {
	m.Timestamp = r.Uint32()
	m.Name = string(str[:rand.Intn(len(str)-1)+1])
	m.Tags = [][2]string{{"1", string(str[:rand.Intn(len(str)-1)+1])}, {"2", string(str[:rand.Intn(len(str)-1)+1])}}
	m.Kind = r.Intn(3)
	switch m.Kind {
	case 0: // counter
		m.Count = r.Float64() * 1_000_000
	case 1: // values
		m.Values = values[:r.Intn(len(values)-1)+1]
	case 2: // uniques
		m.Uniques = uniques[:r.Intn(len(uniques)-1)+1]
	}
}
