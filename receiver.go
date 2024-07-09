package main

import (
	"alpinskiy/statshouse-test/gen2"
	"bytes"
	"context"
	"fmt"
	"log"
	"math"
	"net"
	"slices"
	"sort"
	"sync"
)

type series map[tags]map[uint32]*value
type tag [2]string // name, value
type tags [17]tag  // metric name + 16 tags

type value struct {
	counter float64
	values  []float64
	uniques []int64
}

func listenUDP(addr string, fn func([]byte)) (func(), error) {
	pc, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, err
	}
	log.Printf("listen UDP %s\n", addr)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer pc.Close()
		buffer := make([]byte, math.MaxUint16)
		for {
			if n, _, err := pc.ReadFrom(buffer); err == nil {
				fn(buffer[:n])
			} else if ctx.Err() != nil {
				break
			} else {
				log.Fatal(err)
			}
		}
	}()
	return func() {
		cancel()
		pc.Close()
		wg.Wait()
	}, nil
}

var (
	jsonPacketPrefix   = []byte("{")                    // 0x7B
	metricsBatchPrefix = []byte{0x39, 0x02, 0x58, 0x56} // little-endian 0x56580239
)

func (s series) parseAdd(b []byte) {
	var bb gen2.BatchBytes
	switch {
	case bytes.HasPrefix(b, metricsBatchPrefix):
		if _, err := bb.ReadBoxed(b); err != nil {
			log.Fatal(err)
		}
	case bytes.HasPrefix(b, jsonPacketPrefix):
		if err := bb.UnmarshalJSON(b); err != nil {
			log.Fatal(err)
		}
	case msgpackLooksLikeMap(b):
		if _, err := msgpackUnmarshalStatshouseAddMetricBatch(&bb, b); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal(fmt.Errorf("unknown prefix"))
	}
	for _, b := range bb.Metrics {
		tags := tags{{"", string(b.Name)}}
		for i := 0; i < len(b.Tags); i++ {
			if len(b.Tags[i].Value) != 0 {
				tags[i+1] = tag{string(b.Tags[i].Key), string(b.Tags[i].Value)}
			}
		}
		vals := s[tags]
		if vals == nil {
			vals = map[uint32]*value{}
			s[tags] = vals
		}
		val := vals[b.Ts]
		if val == nil {
			val = &value{}
			vals[b.Ts] = val
		}
		val.addMetricBytes(b)
	}
}

func (s series) normalize() {
	tmp := make(series, len(s))
	for tags, seconds := range s {
		delete(s, tags)
		sort.Slice(tags[:], func(i, j int) bool {
			return slices.Compare(tags[i][:], tags[j][:]) < 0
		})
		if secondsT, ok := tmp[tags]; ok {
			for tags, val := range seconds {
				if valT, ok := secondsT[tags]; ok {
					valT.addValue(val)
				} else {
					secondsT[tags] = val
				}
			}
		} else {
			tmp[tags] = seconds
		}
	}
	for tags, seconds := range tmp {
		for _, v := range seconds {
			v.sort()
		}
		delete(tmp, tags)
		s[tags] = seconds
	}
}

func (v *value) addMetricBytes(b gen2.MetricBytes) {
	v.counter += b.Counter
	v.values = append(v.values, b.Value...)
	v.uniques = append(v.uniques, b.Unique...)
}

func (v *value) addValue(b *value) {
	v.counter += b.counter
	v.values = append(v.values, b.values...)
	v.uniques = append(v.uniques, b.uniques...)
}

func (v *value) sort() {
	sort.Float64s(v.values)
	sort.Slice(v.uniques, func(i, j int) bool {
		return v.uniques[i] < v.uniques[j]
	})
}
