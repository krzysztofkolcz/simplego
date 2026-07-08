package app

import (
	"github.com/example/my-supertokens/internal/config"
	httphandler "github.com/example/my-supertokens/internal/http"
	"github.com/example/my-supertokens/internal/repository"
	"github.com/supertokens/supertokens-golang/recipe/emailpassword"
	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/supertokens"
	"net/http"
)

// App wires all dependencies together.
type App struct {
	Router http.Handler
}

func New(cfg config.Config) (*App, error) {
	if err := initSuperTokens(cfg.SuperTokens); err != nil {
		return nil, err
	}

	authSvc := repository.NewSuperTokensAuthService()
	handler := httphandler.NewHandler(authSvc)
	router := httphandler.NewRouter(handler)

	return &App{Router: router}, nil
}

func initSuperTokens(cfg config.SuperTokensConfig) error {
	return supertokens.Init(supertokens.TypeInput{
		Supertokens: &supertokens.ConnectionInfo{
			ConnectionURI: cfg.ConnectionURI,
			APIKey:        cfg.APIKey,
		},
		AppInfo: supertokens.AppInfo{
			AppName:       "my-supertokens",
			APIDomain:     cfg.APIDomain,
			WebsiteDomain: cfg.WebsiteDomain,
		},
		RecipeList: []supertokens.Recipe{
			emailpassword.Init(nil),
			session.Init(nil),
		},
	})
}
