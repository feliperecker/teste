package core_test

import (
	"testing"

	"github.com/go-bolo/core"
	"github.com/stretchr/testify/assert"
)

func TestPluginWork(t *testing.T) {
	app := core.NewApp(&core.DefaultAppOptions{})

	err := app.Bootstrap()
	assert.Nil(t, err)

	err = app.AddPlugin("example", &URLShortenerPlugin{})
	assert.Nil(t, err)

}
