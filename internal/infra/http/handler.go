package http

import (
	"net/http"
	"strconv"

	"doc_service/domain"

	"github.com/gin-gonic/gin"
)

type DocumentHandler struct {
	svc domain.Service
}

func NewDocumentHandler(svc domain.Service) *DocumentHandler {
	return &DocumentHandler{svc: svc}
}

func (h *DocumentHandler) CreateDocument(c *gin.Context) {
	var request CreateDocumentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "request is invalid"})
		return
	}

	doc := &domain.Doc{
		ID:          request.ID,
		Title:       request.Title,
		Sort:        request.Sort,
		ChildrenIDs: request.ChildrenIDs,
	}

	createdDoc, err := h.svc.CreateDocument(doc)
	if err != nil {
		status, text := convertErrorToHTTPStatusWithText(err)
		c.AbortWithStatusJSON(status, gin.H{"error": text})
		return
	}

	c.JSON(http.StatusCreated, createdDoc)
}

func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	var limit, offset *int32

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			val := int32(l)
			limit = &val
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			val := int32(o)
			offset = &val
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
			return
		}
	}

	docs, err := h.svc.ListDocuments(limit, offset)
	if err != nil {
		status, text := convertErrorToHTTPStatusWithText(err)
		c.AbortWithStatusJSON(status, gin.H{"error": text})
		return
	}

	c.JSON(http.StatusOK, docs)
}

func (h *DocumentHandler) GetDocument(c *gin.Context) {
	ID := c.Param("id")
	if ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	doc, err := h.svc.GetDocument(ID)
	if err != nil {
		status, text := convertErrorToHTTPStatusWithText(err)
		c.AbortWithStatusJSON(status, gin.H{"error": text})
		return
	}

	c.JSON(http.StatusOK, doc)
}

func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	ID := c.Param("id")
	if ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var request UpdateDocumentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "request is invalid"})
		return
	}

	doc := &domain.Doc{
		ID:          ID,
		Title:       request.Title,
		Sort:        request.Sort,
		ChildrenIDs: request.ChildrenIDs,
	}

	updatedDoc, err := h.svc.UpdateDocument(doc, request.ParentID)
	if err != nil {
		status, text := convertErrorToHTTPStatusWithText(err)
		c.AbortWithStatusJSON(status, gin.H{"error": text})
		return
	}

	c.JSON(http.StatusOK, updatedDoc)
}

func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	ID := c.Param("id")
	if ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	err := h.svc.DeleteDocument(ID)
	if err != nil {
		status, text := convertErrorToHTTPStatusWithText(err)
		c.AbortWithStatusJSON(status, gin.H{"error": text})
		return
	}

	c.Status(http.StatusNoContent)
}

func convertErrorToHTTPStatusWithText(err error) (int, string) {
	switch err {
	case domain.ErrNotFound:
		return http.StatusNotFound, "not found"
	case domain.ErrAlreadyExists:
		return http.StatusConflict, "already exists"
	case domain.ErrInvalidArguments:
		return http.StatusBadRequest, "invalid arguments"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
