package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVerify(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Slack Verification Handler Suite")
}
