package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/arthurcgc/waf-api/internal/pkg/manager"
	echo "github.com/labstack/echo/v4"
)

type CreateOpts struct {
	Name      string
	Replicas  int
	Namespace string
	Plan      string
	Bind      manager.Bind
}

func (a *Api) createInstance(c echo.Context) error {
	var opts CreateOpts
	err := json.NewDecoder(c.Request().Body).Decode(&opts)
	if err != nil {
		return err
	}

	args := manager.CreateArgs{
		Name:      opts.Name,
		Namespace: opts.Namespace,
		Replicas:  opts.Replicas,
		PlanName:  opts.Plan,
		Bind:      opts.Bind,
	}

	if err := a.manager.CreateInstance(c.Request().Context(), args); err != nil {
		return fmt.Errorf("error during deploy: %s", err.Error())
	}

	return c.String(http.StatusCreated, "Created nginx resource!\n")
}
