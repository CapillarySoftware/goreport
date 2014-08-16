package goreporter

import (
	gi "github.com/onsi/ginkgo"
	gom "github.com/onsi/gomega"
)

var _ = gi.Describe("Goreporter", func() {
	gi.It("Basic creation with invalid queue name", func() {
		url := "fail"
		_, err := NewReporter(&url)
		gom.Expect(err).ShouldNot(gom.Equal(gom.BeNil()))
	})

	gi.It("Basic new stat with failed flush", func() {
		url := "ipc:///tmp/test"
		r, err := NewReporter(&url)
		gom.Expect(err).Should(gom.BeNil())
		key := "key"
		r.AddStat(key, 30)
		r.AddStatWIndex(key, 30, "index")
		err = r.Flush()
		gom.Expect(err).ShouldNot(gom.Equal(gom.BeNil()))
	})

})
