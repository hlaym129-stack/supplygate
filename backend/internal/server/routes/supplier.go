package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterSupplierRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	jwtAuth middleware.JWTAuthMiddleware,
) {
	authenticated := v1.Group("")
	authenticated.Use(gin.HandlerFunc(jwtAuth))

	authenticated.GET("/supplier/model-pricing", h.Supplier.ModelPricing)
	authenticated.GET("/supplier/profile", h.Supplier.GetProfile)
	authenticated.PUT("/supplier/profile", h.Supplier.ApplyProfile)
	authenticated.GET("/supplier/usage/summary", h.Supplier.UsageSummary)
	authenticated.GET("/supplier/dashboard/stats", h.Supplier.DashboardStats)
	authenticated.GET("/supplier/dashboard/trend", h.Supplier.DashboardTrend)
	authenticated.GET("/supplier/dashboard/models", h.Supplier.DashboardModels)
	authenticated.GET("/supplier/dashboard/recent", h.Supplier.DashboardRecent)
	authenticated.GET("/supplier/settlement-statements", h.Supplier.ListSettlementStatements)

	supplier := authenticated.Group("/supplier")
	supplier.Use(middleware.SupplierOnly(h.SupplierService))
	{
		supplier.GET("/groups", h.Supplier.ListGroups)
		supplier.GET("/proxies", h.Supplier.ListProxies)
		supplier.GET("/models", h.Supplier.ListModels)
		supplier.GET("/accounts", h.Supplier.ListAccounts)
		supplier.POST("/accounts/test", h.Supplier.TestAccount)
		supplier.POST("/accounts/:id/test", h.Supplier.TestAccountUpdate)
		supplier.POST("/accounts", h.Supplier.CreateAccount)
		supplier.PUT("/accounts/:id", h.Supplier.UpdateAccount)
		supplier.POST("/accounts/:id/edit-request", h.Supplier.RequestAccountEdit)
		supplier.GET("/accounts/:id/pricing-revisions", h.Supplier.ListPricingRevisions)
		supplier.POST("/accounts/:id/pricing-change", h.Supplier.SubmitPricingChange)

		registerSupplierAccountOAuthRoutes(supplier, h)
	}
}

func registerSupplierAccountOAuthRoutes(supplier *gin.RouterGroup, h *handler.Handlers) {
	accounts := supplier.Group("/accounts")
	{
		accounts.POST("/generate-auth-url", h.Admin.OAuth.GenerateAuthURL)
		accounts.POST("/generate-setup-token-url", h.Admin.OAuth.GenerateSetupTokenURL)
		accounts.POST("/exchange-code", h.Admin.OAuth.ExchangeCode)
		accounts.POST("/exchange-setup-token-code", h.Admin.OAuth.ExchangeSetupTokenCode)
		accounts.POST("/cookie-auth", h.Admin.OAuth.CookieAuth)
		accounts.POST("/setup-token-cookie-auth", h.Admin.OAuth.SetupTokenCookieAuth)
	}

	openai := supplier.Group("/openai")
	{
		openai.POST("/generate-auth-url", h.Admin.OpenAIOAuth.GenerateAuthURL)
		openai.POST("/exchange-code", h.Admin.OpenAIOAuth.ExchangeCode)
		openai.POST("/refresh-token", h.Admin.OpenAIOAuth.RefreshToken)
	}

	gemini := supplier.Group("/gemini")
	{
		gemini.POST("/oauth/auth-url", h.Admin.GeminiOAuth.GenerateAuthURL)
		gemini.POST("/oauth/exchange-code", h.Admin.GeminiOAuth.ExchangeCode)
		gemini.GET("/oauth/capabilities", h.Admin.GeminiOAuth.GetCapabilities)
	}

	antigravity := supplier.Group("/antigravity")
	{
		antigravity.POST("/oauth/auth-url", h.Admin.AntigravityOAuth.GenerateAuthURL)
		antigravity.POST("/oauth/exchange-code", h.Admin.AntigravityOAuth.ExchangeCode)
		antigravity.POST("/oauth/refresh-token", h.Admin.AntigravityOAuth.RefreshToken)
	}
}
