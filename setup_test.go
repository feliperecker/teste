package core_test

import (
	"os"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/approvals/go-approval-tests/reporters"
	"github.com/go-bolo/core"
)

func TestMain(m *testing.M) {
	r := approvals.UseReporter(reporters.NewVSCodeReporter())
	defer r.Close()
	approvals.UseFolder("testData")

	os.Exit(m.Run())

	m.Run()
}

func GetTestApp() core.App {
	os.Setenv("TEMPLATE_FOLDER", "./_mocks/themes")

	app := core.NewApp(&core.DefaultAppOptions{})
	app.SetTheme("site")

	return app
}
