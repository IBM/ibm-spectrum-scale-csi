package scale_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCsiplugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Csiplugin Suite")
}
