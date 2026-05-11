package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"honeygarden/internal/domain"
)

func OK(c *gin.Context, data any)      { c.JSON(http.StatusOK, gin.H{"data": data}) }
func Created(c *gin.Context, data any) { c.JSON(http.StatusCreated, gin.H{"data": data}) }
func NoContent(c *gin.Context)         { c.Status(http.StatusNoContent) }

func Err(c *gin.Context, err error) {
	code := http.StatusInternalServerError
	switch {
	case errors.Is(err, domain.ErrNotFound):
		code = http.StatusNotFound
	case errors.Is(err, domain.ErrAlreadyExists):
		code = http.StatusConflict
	case errors.Is(err, domain.ErrForbidden):
		code = http.StatusForbidden
	case errors.Is(err, domain.ErrInvalidInput):
		code = http.StatusBadRequest
	case errors.Is(err, domain.ErrUnauthorized):
		code = http.StatusUnauthorized
	}
	c.JSON(code, gin.H{"error": err.Error()})
}
