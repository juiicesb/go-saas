package saas

import (
	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/goxiaoy/go-saas/common"
	"github.com/goxiaoy/go-saas/common/http"
	"github.com/goxiaoy/go-saas/data"
)

type ErrorFormatter func(context *gin.Context, err error)

var (
	DefaultErrorFormatter ErrorFormatter = func(context *gin.Context, err error) {
		if errors.Is(err, common.ErrTenantNotFound) {
			context.AbortWithError(404, err)
		} else {
			context.AbortWithError(500, err)
		}
	}
)

func MultiTenancy(hmtOpt *http.WebMultiTenancyOption, ts common.TenantStore, ef ErrorFormatter, options ...common.PatchTenantResolveOption) gin.HandlerFunc {
	if ef == nil {
		ef = DefaultErrorFormatter
	}
	return func(context *gin.Context) {
		df := []common.TenantResolveContributor{
			http.NewCookieTenantResolveContributor(hmtOpt.TenantKey, context.Request),
			http.NewFormTenantResolveContributor(hmtOpt.TenantKey, context.Request),
			http.NewHeaderTenantResolveContributor(hmtOpt.TenantKey, context.Request),
			http.NewQueryTenantResolveContributor(hmtOpt.TenantKey, context.Request),
		}
		if hmtOpt.DomainFormat != "" {
			df := append(df[:1], df[0:]...)
			df[0] = http.NewDomainTenantResolveContributor(hmtOpt.DomainFormat, context.Request)
		}
		trOpt := common.NewTenantResolveOption(df...)
		for _, option := range options {
			option(trOpt)
		}
		//get tenant config
		tenantConfigProvider := common.NewDefaultTenantConfigProvider(common.NewDefaultTenantResolver(*trOpt), ts)
		tenantConfig, trCtx, err := tenantConfigProvider.Get(context, true)
		if err != nil {
			ef(context, err)
		}
		previousTenant, _ := common.FromCurrentTenant(trCtx)
		//set current tenant
		newContext := common.NewCurrentTenant(trCtx, tenantConfig.ID, tenantConfig.Name)
		//data filter
		dataFilterCtx := data.NewEnableMultiTenancyDataFilter(newContext)

		//with newContext
		context.Request = context.Request.WithContext(dataFilterCtx)
		//next
		context.Next()
		//cancel
		context.Request = context.Request.WithContext(common.NewCurrentTenantInfo(dataFilterCtx, previousTenant))
	}
}
