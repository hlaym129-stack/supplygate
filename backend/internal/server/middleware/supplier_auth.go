package middleware

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type supplierAccessChecker interface {
	GetApprovedProfileByUserID(ctx context.Context, userID int64) (*service.SupplierProfile, bool, error)
}

func SupplierOnly(suppliers supplierAccessChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		subject, ok := GetAuthSubjectFromContext(c)
		if !ok {
			AbortWithError(c, 401, "UNAUTHORIZED", "User not found in context")
			return
		}
		if suppliers == nil {
			AbortWithError(c, 503, "SUPPLIER_CHECK_UNAVAILABLE", "Supplier access check unavailable")
			return
		}
		_, allowed, err := suppliers.GetApprovedProfileByUserID(c.Request.Context(), subject.UserID)
		if err != nil {
			AbortWithError(c, 500, "SUPPLIER_CHECK_FAILED", err.Error())
			return
		}
		if !allowed {
			AbortWithError(c, 403, "FORBIDDEN", "Supplier access required")
			return
		}
		c.Next()
	}
}
