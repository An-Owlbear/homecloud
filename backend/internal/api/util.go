package api

import (
	"fmt"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

func render(c echo.Context, code int, t templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := t.Render(c.Request().Context(), buf); err != nil {
		return err
	}

	return c.HTML(code, buf.String())
}

func staticFilter(dir string, filter string) echo.HandlerFunc {
	return func(c echo.Context) error {
		filePath := c.Param("*")
		filePath = filepath.Clean(filePath)
		fullPath := filepath.Join(dir, filePath)

		// Checks if the file exists, and returns error if it doesn't
		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				return echo.ErrNotFound
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Error accessing file")
		}
		if fileInfo.IsDir() {
			return echo.ErrNotFound
		}

		// Checks the file matches a regex filter, and returns a not found error if it doesn't
		re, err := regexp.Compile(filter)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error in file filter")
		}

		if !re.MatchString(fullPath) {
			fmt.Printf("File filter does not match regexp: %s", filter)
			return echo.ErrNotFound
		}

		return c.File(fullPath)
	}
}
