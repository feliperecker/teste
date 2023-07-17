package core

import (
	"github.com/gookit/event"
	"go.uber.org/zap"
)

type CorePlugin struct {
	Plugin
	Name string
}

func (p *CorePlugin) Init(app App) error {
	app.GetLogger().Debug("CorePlugin.Init Running init", zap.String("PluginName", p.Name))

	app.GetEvents().On("bindMiddlewares", event.ListenerFunc(func(e event.Event) error {
		return p.BindMiddlewares(app)
	}), event.High)

	return nil
}

func (p *CorePlugin) GetName() string {
	return p.Name
}

func (p *CorePlugin) SetName(name string) error {
	p.Name = name
	return nil
}

func (p *CorePlugin) BindMiddlewares(a App) error {
	BindMiddlewares(a, p)
	return nil
}

type CorePluginOpts struct{}

func NewCorePlugin(opts *CorePluginOpts) *CorePlugin {
	return &CorePlugin{
		Name: "core",
	}
}
