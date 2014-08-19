package goreport

import (
	// "fmt"
	"github.com/CapillarySoftware/gostat/protoStat"
	gi "github.com/onsi/ginkgo"
	gom "github.com/onsi/gomega"
	nano "github.com/op/go-nanomsg"
	"strings"
	"time"
)

var _ = gi.Describe("Goreport", func() {
	var rep Reporter
	gi.BeforeEach(func() {
		ReporterConfig("ipc:///tmp/goreportertest.ipc", 1)
		rep = NewReporter()
		gom.Expect(rep).ShouldNot(gom.Equal(gom.BeNil()))
	})

	gi.It("End to End integration test with stats", func() {
		pull, err := nano.NewPullSocket()
		gom.Expect(err).Should(gom.BeNil())
		pull.SetRecvTimeout(6 * time.Second)
		pull.SetRecvBuffer(1000)
		pull.Bind("ipc:///tmp/goreportertest.ipc")
		key := "key"
		rep.AddRepeatedStat("repeat")
		rep.AddRepeatedStatWIndex("repeateWIndex", "Index")
		rep.AddStat(key, 2)
		rep.AddStat(key, 2)
		rep.AddStatWIndex(key, 2, "index")
		rep.AddStatWIndex(key, 2, "index")
		msg, err := pull.Recv(0)
		gom.Expect(err).Should(gom.BeNil())
		stats := new(protoStat.ProtoStats)
		stats.Unmarshal(msg)
		gom.Expect(len(stats.Stats)).Should(gom.Equal(4))
		for _, stat := range stats.Stats {
			if strings.Contains(stat.GetKey(), "repeat") {
				gom.Expect(stat.GetValue()).Should(gom.Equal(float64(0)))
			} else {
				gom.Expect(stat.GetValue()).Should(gom.Equal(float64(4)))
			}
		}
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
