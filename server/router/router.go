package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/example/qianchuan-saas/auth"
	"github.com/example/qianchuan-saas/handler"
	"github.com/example/qianchuan-saas/qianchuan"
)

func Setup(qc *qianchuan.Client) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	authH := &handler.AuthHandler{}
	accountH := &handler.AccountHandler{QC: qc}
	adH := &handler.AdHandler{QC: qc}
	dashH := &handler.DashboardHandler{}
	reportH := &handler.ReportHandler{}
	ruleH := &handler.RuleHandler{}
	recH := &handler.RecommendationHandler{}
	creativeH := &handler.CreativeHandler{}

	api := r.Group("/api")
	{
		api.POST("/register", authH.Register)
		api.POST("/login", authH.Login)

		authorized := api.Group("/", auth.AuthRequired())
		{
			authorized.GET("/accounts/auth-url", accountH.AuthURL)
			authorized.GET("/accounts", accountH.List)
			authorized.POST("/accounts", accountH.Create)
			authorized.GET("/accounts/:id", accountH.Get)
			authorized.DELETE("/accounts/:id", accountH.Delete)

			authorized.GET("/ads", adH.List)
			authorized.POST("/ads", adH.Create)
			authorized.GET("/ads/:id", adH.Get)
			authorized.PATCH("/ads/:id/status", adH.UpdateStatus)

			authorized.GET("/dashboard/stats", dashH.Stats)
			authorized.GET("/dashboard/trend", dashH.Trend)

			authorized.GET("/recommendations", recH.List)
			authorized.PATCH("/recommendations/:id/status", recH.UpdateStatus)

			authorized.GET("/creatives", creativeH.List)
			authorized.GET("/creatives/:id", creativeH.Get)
			authorized.PUT("/creatives/:id/tags", creativeH.UpdateTags)

			authorized.GET("/reports/ads/:id", reportH.ByAd)
			authorized.GET("/reports/summary", reportH.SummaryByDate)

			authorized.GET("/rules/executions", ruleH.Executions)
			authorized.GET("/rules", ruleH.List)
			authorized.POST("/rules", ruleH.Create)
			authorized.GET("/rules/:id", ruleH.Get)
			authorized.PUT("/rules/:id", ruleH.Update)
			authorized.DELETE("/rules/:id", ruleH.Delete)
		}
	}

	return r
}
