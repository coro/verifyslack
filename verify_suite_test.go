package verifyslack_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVerify(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Slack Verification Handler Suite")
}

type validationTimeGetter struct {
	validationTime time.Time
}

func (v validationTimeGetter) Now() time.Time {
	return v.validationTime
}
