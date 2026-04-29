package factory

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/config"
	"github.com/zeroSal/went-web/controller"
	"github.com/zeroSal/went-web/security"
	"github.com/zeroSal/went-web/session"
)

func IrisFactory(
	embedFS embed.FS,
	sec *security.Security,
	sessionProvider session.ProviderInterface,
	configuration *config.Iris,
) (*iris.Application, error) {
	app := iris.New()

	templatesFS, err := fs.Sub(embedFS, "templates")
	if err != nil {
		return nil, fmt.Errorf("failed to get templates subdirectory: %w", err)
	}

	engine := iris.Django(templatesFS, ".html.django")
	if configuration.IsTemplateAutoReload() {
		engine.Reload(true)
	}

	app.RegisterView(engine)

	app.Get("/static/{file:path}", func(ctx iris.Context) {
		file := ctx.Params().Get("file")
		content, err := embedFS.ReadFile("static/" + file)
		if err != nil {
			ctx.StatusCode(404)
			return
		}
		ctx.Write(content)
	})

	sec.SetSessionProvider(sessionProvider)

	handlers := configuration.GetRegistry().Handlers()
	if len(handlers) == 0 {
		return nil, fmt.Errorf("no handlers found in controllers")
	}

	registerHandlers(configuration.GetRegistry(), sec)
	app.Use(sec.Middleware())
	sec.RegisterRoutes(app)

	return app, nil
}

func registerHandlers(
	ctrlRegistry *controller.Registry,
	sec *security.Security,
) {
	if sec == nil {
		return
	}

	handlers := ctrlRegistry.Handlers()
	for _, h := range handlers {
		sec.RegisterHandler(h.Controller, h.Method, h.Handler)
	}
}