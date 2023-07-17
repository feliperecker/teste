package core_test

import (
	"os"
	"testing"
	"time"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/approvals/go-approval-tests/reporters"
	"github.com/go-bolo/clock"
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

	c := clock.NewMock()
	t, _ := time.Parse("2006-01-02", "2023-07-16")
	c.Set(t)

	app.SetClock(c)

	return app
}
