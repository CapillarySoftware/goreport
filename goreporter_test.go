package goreporter

import (
	gi "github.com/onsi/ginkgo"
	gom "github.com/onsi/gomega"
)

var _ = gi.Describe("Goreporter", func() {
	gi.It("Basic creation", func() {
		url := "ipc:///tmp/test"
		rep, err := NewReporter(&url)
		gom.Expect(err).Should(gom.BeNil())
		key := "key"
		rep.AddStat(key, 30)
		rep.Flush()
	})
})
