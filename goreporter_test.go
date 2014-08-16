package goreporter

import (
	gi "github.com/onsi/ginkgo"
	gom "github.com/onsi/gomega"
)

var _ = gi.Describe("Goreporter", func() {
	gi.It("Basic creation with invalid queue name", func() {
		url := "fail"
		r, err := NewReporter(&url)
		r.SetTimeout(1)
		gom.Expect(err).ShouldNot(gom.Equal(gom.BeNil()))
	})

	gi.It("Basic new stat with failed flush", func() {
		url := "ipc:///tmp/test"
		r, err := NewReporter(&url)
		r.SetTimeout(1)
		gom.Expect(err).Should(gom.BeNil())
		key := "key"
		r.AddStat(key, 30)
		err = r.Flush()
		gom.Expect(err).ShouldNot(gom.Equal(gom.BeNil()))
	})

})
