package http

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter(docHandler *DocumentHandler) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api/v1")
	{
		api.GET("/documents", docHandler.ListDocuments)
		api.POST("/documents", docHandler.CreateDocument)
		api.GET("/documents/:id", docHandler.GetDocument)
		api.PATCH("/documents/:id", docHandler.UpdateDocument)
		api.DELETE("/documents/:id", docHandler.DeleteDocument)
	}

	return r
}
