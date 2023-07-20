package bolo

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/go-bolo/bolo/acl"
	"github.com/go-bolo/bolo/configuration"
	"github.com/go-bolo/bolo/helpers"
	"github.com/go-bolo/clock"
	"github.com/go-playground/validator/v10"
	"github.com/microcosm-cc/bluemonday"

	"github.com/gookit/event"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"

	gorm_logger "gorm.io/gorm/logger"

	"gorm.io/gorm"
)

type App interface {
	GetEnv() string

	GetClock() clock.Clock
	SetClock(clock clock.Clock) error

	AddPlugin(p Plugin) error
	GetPlugin(pluginName string) (p Plugin)
	HasPlugin(pluginName string) (has bool)

	GetEvents() *event.Manager

	GetConfiguration() configuration.ConfigurationInterface

	// DB:
	InitDatabase(name, engine string, isDefault bool) error
	GetDB() *gorm.DB
	GetDBByName(dbName string) *gorm.DB
	SetDB(dbName string, db *gorm.DB) error
	SetModel(name string, model Model) error
	GetModel(name string) Model
	// Run gorm migrate for each registered model
	SyncDB() error

	// Logger:
	GetLogger() *zap.Logger
	SetLogger(logger *zap.Logger) error

	// Router:
	GetRouter() *echo.Echo
	SetRouterGroup(name, path string) *echo.Group
	GetRouterGroup(name string) *echo.Group
	SetResource(r *Resource) error

	BindRoute(routeName string, r *Route) echo.HandlerFunc
	SetRoute(routeName string, route *Route) error

	GetDefaultContentType() string
	GetContentTypes() []string
	SetContentTypes(contentTypes []string) error
	GetResponseFormatter(accept string) responseFormatter
	SetResponseFormatter(accept string, rf responseFormatter) error

	StartHTTPServer() error

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

	GetTemplate(c echo.Context, r *Route) string

	// ACL:
	GetAcl() acl.Acl
	SetAcl(acl acl.Acl) error
	// HTML / Text sanitizer:
	GetSanitizer() *bluemonday.Policy
	SetSanitizer(policy *bluemonday.Policy) error
	// Start and close:
	Bootstrap() error
	Close() error
}

type DefaultAppOptions struct {
	DefaultContentType string
	// Avaiblable content types for negotiation:
	ContentTypes []string
	// Gorm configurations / options
	GormOptions gorm.Option
}

func NewApp(opts *DefaultAppOptions) App {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := configuration.NewCfg()

	if len(opts.ContentTypes) == 0 {
		opts.ContentTypes = []string{"text/html", "application/json"}
	}

	if opts.DefaultContentType == "" {
		opts.DefaultContentType = "application/json"
	}

	app := &DefaultApp{
		Acl:                acl.NewAcl(&acl.NewAclOpts{Logger: logger}),
		Clock:              clock.New(),
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
		Theme:              cfg.GetF(THEME, "site"),
		Layout:             "layouts/default",
		templateFunctions:  make(template.FuncMap),
	}
	// Default police:
	app.Sanitizer = bluemonday.UGCPolicy()
	app.Sanitizer.AllowDataURIImages()

	app.router.GET("/health", HealthCheckHandler)
	app.router.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			SetDefaultValues(c, app)
			return next(c)
		}
	})

	app.router.Binder = &CustomBinder{}
	app.router.HTTPErrorHandler = CustomHTTPErrorHandler
	app.router.Validator = &helpers.CustomValidator{Validator: validator.New()}
	// add core plugin, is set as plugin to be overriden if needed:
	app.AddPlugin(NewCorePlugin(&CorePluginOpts{}))

	return app
}

type DefaultApp struct {
	Acl           acl.Acl
	Clock         clock.Clock
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
	ResponseFormatters map[string]responseFormatter `json:"-"`

	routerGroups map[string]*echo.Group

	Sanitizer *bluemonday.Policy

	// default theme for HTML responses
	Theme string
	// default layout for HTML responses
	Layout            string
	templates         *template.Template
	templateFunctions template.FuncMap
}

func (app *DefaultApp) GetClock() clock.Clock {
	return app.Clock
}

func (app *DefaultApp) SetClock(clock clock.Clock) error {
	app.Clock = clock
	return nil
}

func (app *DefaultApp) GetEnv() string {
	return os.Getenv(ENV_VARIABLE_NAME)
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

func (app *DefaultApp) SetResource(r *Resource) error {
	app.Resources[r.Name] = r
	return nil
}

func (app *DefaultApp) BindRoute(routeName string, r *Route) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("route", r)

		res, err := r.Action(c)

		if err != nil {
			return err
		}

		return app.GetResponseFormatter(GetAccept(c))(app, c, r, res)
	}
}

func (app *DefaultApp) GetDefaultContentType() string {
	if app.Options.DefaultContentType != "" {
		return app.Options.DefaultContentType
	}

	return "application/json"
}

func (app *DefaultApp) GetContentTypes() []string {
	return app.Options.ContentTypes
}

func (app *DefaultApp) SetContentTypes(contentTypes []string) error {
	app.Options.ContentTypes = contentTypes
	return nil
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
	port := app.Configuration.Get(PORT)
	if port == "" {
		port = "8080"
	}

	app.GetLogger().Info("Server listening on port " + port)
	return http.ListenAndServe(":"+port, app.GetRouter())
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
	return app.templates.Lookup(name) == nil
}

func (app *DefaultApp) LoadTemplates() error {
	l := app.GetLogger().With(zap.String("func", "LoadTemplates"))

	rootDir := app.Configuration.GetF(TEMPLATE_FOLDER, "./themes")
	disableTemplating := app.Configuration.GetBool(TEMPLATE_DISABLE)

	if disableTemplating {
		return nil
	}

	tpls, err := findAndParseTemplates(rootDir, app.templateFunctions)
	if err != nil {
		l.Error("error on parse templates", zap.Error(err), zap.String("rootDir", rootDir))
		app.templates = tpls
		return err
	}

	app.templates = tpls

	l.Debug("templates loaded", zap.Int("count", len(app.templates.Templates())))

	return nil
}

func (app *DefaultApp) SetTemplateFunction(name string, f interface{}) {
	app.templateFunctions[name] = f
}

func (app *DefaultApp) RenderTemplate(wr io.Writer, theme string, name string, data interface{}) error {
	return app.GetTemplates().ExecuteTemplate(wr, path.Join(theme, name), data)
}

func (app *DefaultApp) GetTemplate(c echo.Context, r *Route) string {
	templateCtx := c.Get("template")
	if templateCtx != nil {
		return templateCtx.(string)
	}

	if r.Template != "" {
		return r.Template
	}

	return "template-not-set"
}

func (app *DefaultApp) GetAcl() acl.Acl {
	return app.Acl
}

func (app *DefaultApp) SetAcl(acl acl.Acl) error {
	app.Acl = acl
	return nil
}

func (app *DefaultApp) GetSanitizer() *bluemonday.Policy {
	return app.Sanitizer
}

func (app *DefaultApp) SetSanitizer(sanitizer *bluemonday.Policy) error {
	app.Sanitizer = sanitizer
	return nil
}

func (app *DefaultApp) AddPlugin(p Plugin) error {
	app.Plugins[p.GetName()] = p
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

func (app *DefaultApp) SyncDB() error {
	for _, m := range app.Models {
		err := app.GetDB().AutoMigrate(m)
		if err != nil {
			return fmt.Errorf("app.SyncDB: %w", err)
		}
	}

	return nil
}

func (app *DefaultApp) InitDatabase(name string, engine string, isDefault bool) error {
	var err error
	var db *gorm.DB

	dbURI := app.Configuration.GetF(DB_URI, "file::memory:?charset=utf8mb4")
	dbSlowThreshold := app.Configuration.GetInt64F(DB_SLOW_THRESHOLD, 400)
	logQuery := app.Configuration.GetF(LOG_QUERY, "")

	l := app.GetLogger().With(zap.String("on", "InitDatabase"))
	l.Debug("starting db with configs", zap.String("dbURI", dbURI), zap.Int64("dbSlowThreshold", dbSlowThreshold), zap.String("logQuery", logQuery))

	if dbURI == "" {
		return ErrDbUrlIsRequired
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
		db, err = gorm.Open(sqlite.Open(dbURI), gormCFG)
	default:
		return ErrDbInvalidDatabaseEngine
	}

	if err != nil {
		return fmt.Errorf("InitDatabase error on database connectio: %w", err)
	}

	return app.SetDB(app.DefaultDB, db)
}

func (app *DefaultApp) SetModel(name string, m Model) error {
	app.Models[name] = m
	return nil
}

func (app *DefaultApp) GetModel(name string) Model {
	return app.Models[name]
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
		return fmt.Errorf("DefaultApp.Bootstrap: Error on init database connection: %w", err)
	}

	err = SetDefaultResponseFormatters(app)
	if err != nil {
		return err
	}

	HttpClientInit()

	app.Events.MustTrigger("bindMiddlewares", event.M{"app": app})
	app.Events.MustTrigger("bindRoutes", event.M{"app": app})
	app.Events.MustTrigger("setResponseFormats", event.M{"app": app})

	err = SetCoreTemplateFunctions(app)
	if err != nil {
		return err
	}
	app.Events.MustTrigger("setTemplateFunctions", event.M{"app": app})

	err = app.LoadTemplates()
	if err != nil {
		return fmt.Errorf("DefaultApp.Bootstrap Error on LoadTemplates: %w", err)
	}

	app.router.Renderer = &TemplateRenderer{
		templates: app.GetTemplates(),
	}

	for routeName, r := range app.Resources {
		l.Debug("DefaultApp.Bootstrap: registering route", zap.String("route", routeName))
		err := r.BindRoutes(app)
		if err != nil {
			return fmt.Errorf("DefaultApp.Bootstrap: Error on bind resource route: %w", err)
		}
	}

	for routeName, r := range app.Routes {
		l.Debug("DefaultApp.Bootstrap: registering route", zap.String("route", routeName))

		if r.Path == "" {
			return fmt.Errorf("route path is required: %s", r.Path)
		}

		router := app.GetRouter()

		switch r.Method {
		case "POST":
			router.POST(r.Path, app.BindRoute(routeName, r))
		case "GET":
			router.GET(r.Path, app.BindRoute(routeName, r))
		case "PUT":
			router.PUT(r.Path, app.BindRoute(routeName, r))
		case "DELETE":
			router.DELETE(r.Path, app.BindRoute(routeName, r))
		default:
			return fmt.Errorf("DefaultApp.Bootstrap: invalid route method: %s", r.Method)
		}
	}

	app.Events.MustTrigger("bootstrap", event.M{"app": app})

	return nil
}

func (app *DefaultApp) Close() error {
	err, _ := app.Events.Fire("close", event.M{"app": app})
	if err != nil {
		app.GetLogger().Debug("Close error", zap.Error(err))
		return err
	}

	return nil
}
