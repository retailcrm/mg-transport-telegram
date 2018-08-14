package main

import (
	"os"
	"os/signal"
	"syscall"

	"io/ioutil"

	"github.com/getsentry/raven-go"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"
)

func init() {
	parser.AddCommand("run",
		"Run mg-telegram",
		"Run mg-telegram.",
		&RunCommand{},
	)
}

// RunCommand struct
type RunCommand struct{}

// Execute command
func (x *RunCommand) Execute(args []string) error {
	config = LoadConfig(options.Config)
	orm = NewDb(config)
	logger = newLogger()

	go start()

	c := make(chan os.Signal, 1)
	signal.Notify(c)
	for sig := range c {
		switch sig {
		case os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM:
			orm.DB.Close()
			return nil
		default:
		}
	}

	return nil
}

func start() {
	routing := setup()
	routing.Run(config.HTTPServer.Listen)
}

func setup() *gin.Engine {
	loadTranslateFile()

	binding.Validator = new(defaultValidator)

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("validatecrmurl", validateCrmURL)
	}

	if config.Debug == false {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	if config.Debug {
		r.Use(gin.Logger())
	}

	r.Static("/static", "./static")
	r.HTMLRender = createHTMLRender()

	r.Use(func(c *gin.Context) {
		setLocale(c.GetHeader("Accept-Language"))
	})

	errorHandlers := []ErrorHandlerFunc{
		PanicLogger(),
		ErrorLogger(),
		ErrorResponseHandler(),
	}
	sentry, _ := raven.New(config.SentryDSN)

	if sentry != nil {
		errorHandlers = append(errorHandlers, ErrorCaptureHandler(sentry, false))
	}

	r.Use(ErrorHandler(errorHandlers...))

	r.GET("/", connectHandler)
	r.GET("/settings/:uid", settingsHandler)
	r.POST("/save/", saveHandler)
	r.POST("/create/", createHandler)
	r.POST("/add-bot/", addBotHandler)
	r.POST("/delete-bot/", deleteBotHandler)
	r.POST("/actions/activity", activityHandler)
	r.POST("/telegram/:token", telegramWebhookHandler)
	r.POST("/webhook/", mgWebhookHandler)

	return r
}

func createHTMLRender() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromFiles("home", "templates/layout.html", "templates/home.html")
	r.AddFromFiles("form", "templates/layout.html", "templates/form.html")
	return r
}

func loadTranslateFile() {
	bundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)
	files, err := ioutil.ReadDir("translate")
	if err != nil {
		logger.Error(err)
	}
	for _, f := range files {
		if !f.IsDir() {
			bundle.MustLoadMessageFile("translate/" + f.Name())
		}
	}
}
