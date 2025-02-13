package routes

import (
	"alime-be/controllers"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	SetupEssentialRoutes(r)
	api := r.Group("/api")
	{
		api.POST("/upload", controllers.HandleGenerateTranscribe)
		api.POST("/translate", controllers.HandleTranslate)
		api.POST("/download-video", controllers.HandleDownloadVideoWithCaption)

		// New route for downloading subtitled videos
		api.GET("/download-subtitled-video", controllers.DownloadSubtitledVideo)
	}
}

func SetupEssentialRoutes(r *gin.Engine) {
	// Set Gin to release mode if not in development
	gin.SetMode(gin.ReleaseMode)

	// Load templates at startup instead of loading them on each request
	r.Delims("{{", "}}") // Optional: set template delimiters
	r.LoadHTMLGlob("./public/html/*")

	// Serve static files with proper caching headers
	r.Static("/public", "./public")

	// Handle root route with proper error handling
	r.GET("/", func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
		c.HTML(http.StatusOK, "index.html", gin.H{
			"ginBoilerplateVersion": "v0.03",
			"goVersion":             runtime.Version(),
		})
	})

	// 404 handler
	r.NoRoute(func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
		c.HTML(404, "404.html", gin.H{})
	})
}
