package bolo_test

import (
	"testing"

	bolo "github.com/go-bolo/bolo"
	"github.com/stretchr/testify/assert"
)

func TestPluginWork(t *testing.T) {
	app := bolo.NewApp(&bolo.DefaultAppOptions{})

	err := app.Bootstrap()
	assert.Nil(t, err)

	err = app.AddPlugin("example", &URLShortenerPlugin{})
	assert.Nil(t, err)

}
