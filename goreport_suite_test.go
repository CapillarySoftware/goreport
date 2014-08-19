package goreport

import (
	gi "github.com/onsi/ginkgo"
	gom "github.com/onsi/gomega"

	"testing"
)

func TestGoreporter(t *testing.T) {
	gom.RegisterFailHandler(gi.Fail)
	gi.RunSpecs(t, "Goreporter Suite")
}
