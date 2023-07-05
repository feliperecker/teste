package core

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"time"

	"github.com/go-bolo/core/configuration"
	"github.com/go-catupiry/catu/http_client"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/gookit/event"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"

	gorm_logger "gorm.io/gorm/logger"

	"gorm.io/gorm"
)

type App interface {
	AddPlugin(pluginName string, p Plugin) error
	GetPlugin(pluginName string) (p Plugin)
	HasPlugin(pluginName string) (has bool)

	GetEvents() *event.Manager

	GetConfiguration() configuration.ConfigurationInterface

	// DB:
	InitDatabase(name, engine string, isDefault bool) error
	GetDB() *gorm.DB
	GetDBByName(dbName string) *gorm.DB
	SetDB(dbName string, db *gorm.DB) error
	SetModel(name string, f interface{})
	GetModel(name string) interface{}

	// Logger:
	GetLogger() *zap.Logger
	SetLogger(logger *zap.Logger) error

	// Router:
	GetRouter() *echo.Echo
	SetRouterGroup(name, path string) *echo.Group
	GetRouterGroup(name string) *echo.Group
	SetResource(name string, httpController Controller, routerGroup *echo.Group) error
	BindRoute(routeName string, r *Route) echo.HandlerFunc

	GetResponseFormatter(accept string) responseFormatter
	SetResponseFormatter(accept string, rf responseFormatter) error

	StartHTTPServer() error

	SetRoute(routeName string, route *Route) error

	// Theme / view methods
	GetTheme() string
	SetTheme(theme string) error
	GetLayout() string
	SetLayout(layout string) error
	GetTemplates() *template.Template
	HasTemplate(name string) bool
	LoadTemplates() error
	SetTemplateFunction(name string, f interface{})
	RenderTemplate(wr io.Writer, theme string, name string, data interface{}) error

	GetTemplateCtx(c echo.Context, r *Route) string

	// ACL:
	GetAcl() Acl
	SetAcl(acl Acl) error

	// Start and close:
	Bootstrap() error
	Close() error
}

type DefaultAppOptions struct {
	// Gorm configurations / options
	GormOptions gorm.Option
}

func NewApp(opts *DefaultAppOptions) App {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := configuration.NewCfg()

	app := &DefaultApp{
		Acl:                NewAcl(),
		Options:            opts,
		Plugins:            make(map[string]Plugin),
		Events:             event.NewManager("app"),
		Logger:             logger,
		Configuration:      cfg,
		DefaultDB:          "default",
		DBs:                make(map[string]*gorm.DB),
		Models:             make(map[string]Model),
		Resources:          make(map[string]*Resource),
		ResponseFormatters: make(map[string]responseFormatter),
		router:             echo.New(),
		Routes:             make(map[string]*Route),
		Theme:              cfg.GetF("THEME", "site"),
		Layout:             "layouts/default",
	}

	return app
}

type DefaultApp struct {
	Acl           Acl
	Options       *DefaultAppOptions
	Plugins       map[string]Plugin
	Models        map[string]Model
	Events        *event.Manager
	Configuration configuration.ConfigurationInterface

	// Default database
	DefaultDB string
	// avaible databases
	DBs    map[string]*gorm.DB `json:"-"`
	Logger *zap.Logger

	router             *echo.Echo
	Routes             map[string]*Route
	Resources          map[string]*Resource
	ResponseFormatters map[string]responseFormatter

	routerGroups map[string]*echo.Group

	RolesList map[string]Role
	// default theme for HTML responses
	Theme string
	// default layout for HTML responses
	Layout            string
	templates         *template.Template
	templateFunctions template.FuncMap
}

func (app *DefaultApp) GetLogger() *zap.Logger {
	return app.Logger
}

func (app *DefaultApp) SetLogger(logger *zap.Logger) error {
	app.Logger = logger
	return nil
}

func (app *DefaultApp) GetRouter() *echo.Echo {
	return app.router
}

func (app *DefaultApp) SetRouterGroup(name, path string) *echo.Group {
	app.routerGroups[name] = app.router.Group(path)
	return app.routerGroups[name]
}

func (app *DefaultApp) GetRouterGroup(name string) *echo.Group {
	return app.routerGroups[name]
}

func (app *DefaultApp) SetResource(name string, httpController Controller, routerGroup *echo.Group) error {
	app.Resources[name] = &Resource{
		Name: name,
		// Group:    routerGroup,
		// Handlers: httpController,
	}
	return nil
}

func (app *DefaultApp) BindRoute(routeName string, r *Route) echo.HandlerFunc {
	return func(c echo.Context) error {
		res, err := r.Action(c)

		// TODO! formatter
		log.Println("TODO!>>>>>>", res, err)

		availableTypes := []string{"application/json", "text/plain", "text/html", "text/*"}
		ctype := NegotiateContentType(c.Request(), availableTypes, "application/json")

		log.Println("ctype>>>>>>>>>>>>>>>>>>>>>", ctype)

		return app.GetResponseFormatter(ctype)(app, c, r, res)
	}
}

func (app *DefaultApp) GetResponseFormatter(accept string) responseFormatter {
	return app.ResponseFormatters[accept]
}

func (app *DefaultApp) SetResponseFormatter(accept string, rf responseFormatter) error {
	app.ResponseFormatters[accept] = rf
	return nil
}

func (app *DefaultApp) SetRoute(routeName string, route *Route) error {
	app.Routes[routeName] = route
	return nil
}

func (app *DefaultApp) StartHTTPServer() error {
	panic("not implemented") // TODO: Implement
}

func (app *DefaultApp) GetTheme() string {
	return app.Theme
}

func (app *DefaultApp) SetTheme(theme string) error {
	app.Theme = theme
	return nil
}

func (app *DefaultApp) GetLayout() string {
	return app.Layout
}

func (app *DefaultApp) SetLayout(layout string) error {
	if layout == "" {
		return fmt.Errorf("SetLayout: layout cannot be empty")
	}

	app.Layout = layout
	return nil
}

func (app *DefaultApp) GetTemplates() *template.Template {
	return app.templates
}

func (app *DefaultApp) HasTemplate(name string) bool {
	if app.templates.Lookup(name) == nil {
		return false

	}
	return true
}

func (app *DefaultApp) LoadTemplates() error {
	panic("not implemented") // TODO: Implement
}

func (app *DefaultApp) SetTemplateFunction(name string, f interface{}) {
	app.templateFunctions[name] = f
}

func (app *DefaultApp) RenderTemplate(wr io.Writer, theme string, name string, data interface{}) error {
	panic("not implemented") // TODO: Implement
}

func (app *DefaultApp) GetTemplateCtx(c echo.Context, r *Route) string {
	templateCtx := c.Get("template")
	if templateCtx != nil {
		return templateCtx.(string)
	}

	if r.Template != "" {
		return r.Template
	}

	return "template-not-set"
}

func (app *DefaultApp) GetAcl() Acl {
	return app.Acl
}

func (app *DefaultApp) SetAcl(acl Acl) error {
	app.Acl = acl
	return nil
}

func (app *DefaultApp) AddPlugin(pluginName string, p Plugin) error {
	app.Plugins[pluginName] = p
	return nil
}

func (app *DefaultApp) GetPlugin(pluginName string) (p Plugin) {
	return app.Plugins[pluginName]
}

func (app *DefaultApp) HasPlugin(pluginName string) (has bool) {
	_, has = app.Plugins[pluginName]
	return has
}

func (app *DefaultApp) GetEvents() *event.Manager {
	return app.Events
}

func (app *DefaultApp) GetConfiguration() configuration.ConfigurationInterface {
	return app.Configuration
}

// GetDB - returns the default connection:
func (app *DefaultApp) GetDB() *gorm.DB {
	return app.DBs[app.DefaultDB]
}

func (app *DefaultApp) GetDBByName(dbName string) *gorm.DB {
	return app.DBs[dbName]
}

func (app *DefaultApp) SetDB(dbName string, db *gorm.DB) error {
	app.DBs[dbName] = db
	return nil
}

func (app *DefaultApp) Migrate() error {
	panic("not implemented") // TODO: Implement
}

// DB:
func (app *DefaultApp) InitDatabase(name string, engine string, isDefault bool) error {
	var err error
	var db *gorm.DB

	dbURI := app.Configuration.GetF("DB_URI", "test.sqlite?charset=utf8mb4")
	dbSlowThreshold := app.Configuration.GetInt64F("DB_SLOW_THRESHOLD", 400)
	logQuery := app.Configuration.GetF("LOG_QUERY", "")

	logrus.WithFields(logrus.Fields{
		"dbURI":           dbURI,
		"dbSlowThreshold": dbSlowThreshold,
		"logQuery":        logQuery,
	}).Debug("catu.App.InitDatabase starting db with configs")

	if dbURI == "" {
		return errors.New("catu.App.InitDatabase DB_URI environment variable is required")
	}

	dsn := dbURI + "?charset=utf8mb4&parseTime=True&loc=Local"

	dbLogger := gorm_logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), gorm_logger.Config{
		SlowThreshold:             time.Duration(dbSlowThreshold) * time.Millisecond,
		LogLevel:                  gorm_logger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	})

	logg := dbLogger.LogMode(gorm_logger.Warn)

	if logQuery != "" {
		logg = dbLogger.LogMode(gorm_logger.Info)
	}

	var gormCFG gorm.Option

	if app.Options.GormOptions != nil {
		o := app.Options.GormOptions.(*gorm.Config)
		o.Logger = logg
		gormCFG = o
	} else {
		gormCFG = &gorm.Config{
			Logger: logg,
		}
	}

	switch engine {
	case "mysql":
		db, err = gorm.Open(mysql.Open(dsn), gormCFG)
	case "sqlite":
		log.Println(">>>>>>>>>>>>>>>>>>>>>>>")
		db, err = gorm.Open(sqlite.Open(dbURI), gormCFG)
		log.Println(">>>>>>>>>>>>>>>>>>>>>>><<<<<<<<<")
	default:
		return errors.New("catu.App.InitDatabase invalid database engine. Options available: mysql or sqlite")
	}

	if err != nil {
		return errors.Wrap(err, "catu.App.InitDatabase error on database connection")
	}

	app.SetDB(app.DefaultDB, db)

	return nil
}

func (app *DefaultApp) SetModel(name string, f interface{}) {
	panic("not implemented") // TODO: Implement
}

func (app *DefaultApp) GetModel(name string) interface{} {
	panic("not implemented") // TODO: Implement
}

func (app *DefaultApp) Bootstrap() error {
	var err error

	l := app.GetLogger().With(zap.String("on", "Bootstrap"))
	l.Debug("DefaultApp.Bootstrap: running")

	err = app.GetAcl().LoadRoles()
	if err != nil {
		return fmt.Errorf("DefaultApp.Bootstrap: Error loading roles: %w", err)
	}

	for _, p := range app.Plugins {
		err = p.Init(app)
		if err != nil {
			return fmt.Errorf("DefaultApp.Bootstrap: Error on run plugin init: %s: %w", p.GetName(), err)
		}
	}

	app.Events.MustTrigger("configuration", event.M{"app": app})

	err = app.InitDatabase(app.DefaultDB, configuration.GetEnv("DB_ENGINE", "sqlite"), true)
	if err != nil {
		return err
	}

	SetDefaultResponseFormatters(app)

	http_client.Init()

	app.Events.MustTrigger("bindMiddlewares", event.M{"app": app})
	app.Events.MustTrigger("bindRoutes", event.M{"app": app})
	app.Events.MustTrigger("setResponseFormats", event.M{"app": app})
	app.Events.MustTrigger("setTemplateFunctions", event.M{"app": app})

	// app.router.Renderer = &TemplateRenderer{
	// 	templates: app.GetTemplates(),
	// }

	for routeName, r := range app.Routes {
		l.Debug("DefaultApp.Bootstrap: registering route", zap.String("route", routeName))

		router := app.GetRouter()

		switch r.Method {
		case "POST":
		case "GET":
			router.GET(r.Path, app.BindRoute(routeName, r))
		default:
			return fmt.Errorf("DefaultApp.Bootstrap: invalid route method: %s", r.Method)
		}
	}

	app.Events.MustTrigger("bootstrap", event.M{"app": app})

	return nil
}

func (app *DefaultApp) Close() error {
	panic("not implemented") // TODO: Implement
}
