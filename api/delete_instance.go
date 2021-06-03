package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/arthurcgc/waf-api/internal/pkg/manager"
	echo "github.com/labstack/echo/v4"
)

type DeleteOpts struct {
	Name      string
	Namespace string
}

func (a *Api) deleteInstance(c echo.Context) error {
	var opts DeleteOpts
	err := json.NewDecoder(c.Request().Body).Decode(&opts)
	if err != nil {
		return err
	}

	args := manager.DeleteArgs{
		Name:      opts.Name,
		Namespace: opts.Namespace,
	}

	if err := a.manager.DeleteInstance(c.Request().Context(), args); err != nil {
		a.logger.Error(err)
		return fmt.Errorf("error during deploy: %s", err.Error())
	}

	return c.String(http.StatusOK, "Successfully deleted waf resource!\n")
}
