package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/arthurcgc/waf-api/internal/pkg/manager"
	echo "github.com/labstack/echo/v4"
)

type UpdateOpts struct {
	Name      string
	Replicas  int
	Namespace string
	Plan      string
	Bind      manager.Bind
	Rules     manager.Rules
}

func (a *Api) updateInstance(c echo.Context) error {
	var opts UpdateOpts
	err := json.NewDecoder(c.Request().Body).Decode(&opts)
	if err != nil {
		return err
	}

	args := manager.UpdateArgs{
		Name:      opts.Name,
		Namespace: opts.Namespace,
		Replicas:  opts.Replicas,
		PlanName:  opts.Plan,
		Bind:      opts.Bind,
		Rules:     opts.Rules,
	}

	if err := a.manager.UpdateInstance(c.Request().Context(), args); err != nil {
		return fmt.Errorf("error during update: %s", err.Error())
	}

	return c.String(http.StatusCreated, "Updated waf resource!\n")
}
