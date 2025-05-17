package llm_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/your-org/llm-fast-wrapper/internal/llm"
)

var _ = Describe("OpenAIStreamer", func() {
	It("streams 10 tokens", func() {
		streamer := llm.NewOpenAIStreamer()
		ch, err := streamer.Stream("hi")
		Expect(err).NotTo(HaveOccurred())
		count := 0
		for {
			select {
			case _, ok := <-ch:
				if !ok {
					Expect(count).To(Equal(10))
					return
				}
				count++
			case <-time.After(2 * time.Second):
				Fail("timeout")
			}
		}
	})
})
