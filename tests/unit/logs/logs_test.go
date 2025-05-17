package logs_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/your-org/llm-fast-wrapper/internal/logs"
)

var _ = Describe("PostgresLogger", func() {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		Skip("TEST_DATABASE_URL not set")
	}
	logger, err := logs.NewPostgresLogger(dsn)
	Expect(err).NotTo(HaveOccurred())
	It("stores a log entry", func() {
		err := logger.LogPrompt("p", "t", time.Now())
		Expect(err).NotTo(HaveOccurred())
	})
})
