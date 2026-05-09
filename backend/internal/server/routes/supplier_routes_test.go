//go:build unit

package routes

import (
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRegisterSupplierRoutesRemovesApplyAndKeepsDashboardProfileScoped(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	jwt := middleware.JWTAuthMiddleware(func(c *gin.Context) { c.Next() })

	RegisterSupplierRoutes(v1, &handler.Handlers{
		Admin:    &handler.AdminHandlers{},
		Supplier: handler.NewSupplierHandler(nil),
	}, jwt)

	routes := map[string]bool{}
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	require.False(t, routes[http.MethodPost+" /api/v1/supplier/apply"])
	require.True(t, routes[http.MethodPut+" /api/v1/supplier/profile"])
	require.True(t, routes[http.MethodGet+" /api/v1/supplier/dashboard/stats"])
	require.True(t, routes[http.MethodGet+" /api/v1/supplier/usage/summary"])
	require.True(t, routes[http.MethodPost+" /api/v1/supplier/accounts"])
}
