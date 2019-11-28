package main_test

import (
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "sesam-cimrdf"
)

var _ = Describe("Microservice options", func() {

	var (
		opt Options
	)

	Describe("when configured", func() {

		Context("with log write", func() {
			BeforeEach(func() {
				opt = Options{"log": ioutil.Discard}
			})
			It("holds writer", func() {
				Expect(opt).To(HaveKeyWithValue("log", ioutil.Discard))
			})
		})
	})

})
