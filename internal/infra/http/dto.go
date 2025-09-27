package http

type CreateDocumentRequest struct {
	ID          string   `json:"id" binding:"required"`
	Title       string   `json:"title" binding:"required"`
	Sort        int32    `json:"sort" binding:"required"`
	ChildrenIDs []string `json:"children_ids,omitempty"`
}

type UpdateDocumentRequest struct {
	Title       string   `json:"title" binding:"required"`
	Sort        int32    `json:"sort" binding:"required"`
	ChildrenIDs []string `json:"children_ids,omitempty"`
	ParentID    *string  `json:"parent_id"`
}
