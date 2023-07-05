package core_test

import (
	"os"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/approvals/go-approval-tests/reporters"
)

func TestMain(m *testing.M) {
	r := approvals.UseReporter(reporters.NewVSCodeReporter())
	defer r.Close()
	approvals.UseFolder("testData")

	os.Exit(m.Run())

	m.Run()
}
