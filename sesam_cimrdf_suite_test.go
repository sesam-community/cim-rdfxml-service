package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSesamShaid(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SesamShaid Suite")
}
