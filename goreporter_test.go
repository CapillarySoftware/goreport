package goreporter

import (
	// "fmt"
	"github.com/CapillarySoftware/gostat/protoStat"
	gi "github.com/onsi/ginkgo"
	gom "github.com/onsi/gomega"
)

var _ = gi.Describe("Goreporter", func() {
	var rep Reporter
	gi.BeforeEach(func() {
		ReporterConfig("ipc:///tmp/goreportertest.ipc", 0)
		rep = NewReporter()
		gom.Expect(rep).ShouldNot(gom.Equal(gom.BeNil()))
	})

	gi.It("Basic new stat with failed flush", func() {
		key := "key"
		rep.AddStat(key, 30)
		rep.AddStatWIndex(key, 30, "index")
	})

	gi.It("Validate update map increments correctly with indexKeys", func() {
		stats := make(map[string]*protoStat.ProtoStat)
		key := "key"
		indexKey := "index"
		value := float64(200)
		for i := 0; i < 2; i++ {
			stat := protoStat.ProtoStat{Key: &key, Value: &value, IndexKey: &indexKey}
			updateMap(stats, &stat)
		}
		gom.Expect(len(stats)).Should(gom.Equal(1))
		for _, s := range stats {
			gom.Expect(s.GetValue()).Should(gom.Equal(float64(400)))
		}
	})

	gi.It("Validate update map increments correctly with standard key value pairs", func() {
		stats := make(map[string]*protoStat.ProtoStat)
		key := "key"
		value := float64(200)
		for i := 0; i < 2; i++ {
			stat := protoStat.ProtoStat{Key: &key, Value: &value}
			updateMap(stats, &stat)
		}
		gom.Expect(len(stats)).Should(gom.Equal(1))
		for _, s := range stats {
			gom.Expect(s.GetValue()).Should(gom.Equal(float64(400)))
		}
	})

})
