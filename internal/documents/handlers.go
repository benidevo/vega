package documents

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/benidevo/vega/internal/common/alerts"
	"github.com/benidevo/vega/internal/common/logger"
	"github.com/benidevo/vega/internal/common/render"
	"github.com/benidevo/vega/internal/config"
	"github.com/benidevo/vega/internal/documents/models"
	"github.com/gin-gonic/gin"
)

type DocumentHandler struct {
	service  Service
	cfg      *config.Settings
	log      *logger.PrivacyLogger
	renderer *render.HTMLRenderer
}

func NewDocumentHandler(service Service, cfg *config.Settings, renderer *render.HTMLRenderer) *DocumentHandler {
	return &DocumentHandler{
		service:  service,
		cfg:      cfg,
		log:      logger.GetPrivacyLogger("documents_handler"),
		renderer: renderer,
	}
}

func (h *DocumentHandler) GetDocumentsHub(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		alerts.RenderError(c, http.StatusUnauthorized, "Authentication required", alerts.ContextGeneral)
		return
	}
	userID := userIDValue.(int)

	tab := c.DefaultQuery("tab", "cover-letters")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize := 20

	metrics, err := h.service.GetDocumentMetrics(c.Request.Context(), userID)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to get document metrics")
		metrics = &models.DocumentMetrics{}
	}

	var documents []*models.DocumentSummary
	var total int

	if tab == "resumes" {
		documents, total, err = h.service.GetDocumentsByType(
			c.Request.Context(), userID, models.DocumentTypeResume, page, pageSize,
		)
	} else {
		tab = "cover-letters"
		documents, total, err = h.service.GetDocumentsByType(
			c.Request.Context(), userID, models.DocumentTypeCoverLetter, page, pageSize,
		)
	}

	if err != nil {
		h.log.Error().Err(err).Str("tab", tab).Msg("Failed to get documents")
		alerts.RenderError(c, http.StatusInternalServerError, "Failed to load documents", alerts.ContextGeneral)
		return
	}

	totalPages := (total + pageSize - 1) / pageSize

	data := gin.H{
		"PageTitle":        "My Documents",
		"ActiveTab":        tab,
		"Documents":        documents,
		"TotalDocuments":   total,
		"CoverLetterCount": metrics.CoverLetterCount,
		"ResumeCount":      metrics.ResumeCount,
		"CurrentPage":      page,
		"TotalPages":       totalPages,
		"HasPrevPage":      page > 1,
		"HasNextPage":      page < totalPages,
		"PrevPage":         page - 1,
		"NextPage":         page + 1,
	}

	if c.GetHeader("HX-Request") == "true" && c.GetHeader("HX-Target") == "documents-content" {
		h.renderer.HTML(c, http.StatusOK, "documents/partials/document_list.html", data)
		return
	}
	h.renderer.HTML(c, http.StatusOK, "documents/hub.html", data)
}

func (h *DocumentHandler) GetDocument(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		alerts.RenderError(c, http.StatusUnauthorized, "Authentication required", alerts.ContextGeneral)
		return
	}
	userID := userIDValue.(int)

	docIDStr := c.Param("id")
	docID, err := strconv.Atoi(docIDStr)
	if err != nil {
		alerts.RenderError(c, http.StatusBadRequest, "Invalid document ID", alerts.ContextGeneral)
		return
	}

	doc, err := h.service.GetDocument(c.Request.Context(), docID, userID)
	if err != nil {
		if err == models.ErrDocumentNotFound {
			alerts.RenderError(c, http.StatusNotFound, "Document not found", alerts.ContextGeneral)
		} else {
			h.log.Error().Err(err).Int("doc_id", docID).Msg("Failed to get document")
			alerts.RenderError(c, http.StatusInternalServerError, "Failed to load document", alerts.ContextGeneral)
		}
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(doc.Content))
}

func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		alerts.RenderError(c, http.StatusUnauthorized, "Authentication required", alerts.ContextGeneral)
		return
	}
	userID := userIDValue.(int)

	docIDStr := c.Param("id")
	docID, err := strconv.Atoi(docIDStr)
	if err != nil {
		alerts.RenderError(c, http.StatusBadRequest, "Invalid document ID", alerts.ContextGeneral)
		return
	}

	err = h.service.DeleteDocument(c.Request.Context(), docID, userID)
	if err != nil {
		if err == models.ErrDocumentNotFound {
			alerts.RenderError(c, http.StatusNotFound, "Document not found", alerts.ContextGeneral)
		} else {
			h.log.Error().Err(err).Int("doc_id", docID).Msg("Failed to delete document")
			alerts.RenderError(c, http.StatusInternalServerError, "Failed to delete document", alerts.ContextGeneral)
		}
		return
	}

	c.Header("HX-Trigger", "document-deleted")
	alerts.RenderSuccess(c, "Document deleted successfully", alerts.ContextGeneral)
}

func (h *DocumentHandler) RegenerateDocument(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		alerts.RenderError(c, http.StatusUnauthorized, "Authentication required", alerts.ContextGeneral)
		return
	}
	userID := userIDValue.(int)

	docIDStr := c.Param("id")
	docID, err := strconv.Atoi(docIDStr)
	if err != nil {
		alerts.RenderError(c, http.StatusBadRequest, "Invalid document ID", alerts.ContextGeneral)
		return
	}

	doc, err := h.service.GetDocument(c.Request.Context(), docID, userID)
	if err != nil {
		if err == models.ErrDocumentNotFound {
			alerts.RenderError(c, http.StatusNotFound, "Document not found", alerts.ContextGeneral)
		} else {
			h.log.Error().Err(err).Int("doc_id", docID).Msg("Failed to get document for regeneration")
			alerts.RenderError(c, http.StatusInternalServerError, "Failed to regenerate document", alerts.ContextGeneral)
		}
		return
	}

	redirectURL := fmt.Sprintf("/jobs/%d", doc.JobID)
	if doc.DocumentType == models.DocumentTypeCoverLetter {
		redirectURL = fmt.Sprintf("/jobs/%d/cover-letter", doc.JobID)
	} else if doc.DocumentType == models.DocumentTypeResume {
		redirectURL = fmt.Sprintf("/jobs/%d/cv", doc.JobID)
	}

	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Redirect", redirectURL)
		c.Status(http.StatusOK)
	} else {
		c.Redirect(http.StatusSeeOther, redirectURL)
	}
}

func (h *DocumentHandler) ExportDocument(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		alerts.RenderError(c, http.StatusUnauthorized, "Authentication required", alerts.ContextGeneral)
		return
	}
	userID := userIDValue.(int)

	docIDStr := c.Param("id")
	docID, err := strconv.Atoi(docIDStr)
	if err != nil {
		alerts.RenderError(c, http.StatusBadRequest, "Invalid document ID", alerts.ContextGeneral)
		return
	}

	format := c.DefaultQuery("format", "html")

	doc, err := h.service.GetDocument(c.Request.Context(), docID, userID)
	if err != nil {
		if err == models.ErrDocumentNotFound {
			alerts.RenderError(c, http.StatusNotFound, "Document not found", alerts.ContextGeneral)
		} else {
			h.log.Error().Err(err).Int("doc_id", docID).Msg("Failed to get document for export")
			alerts.RenderError(c, http.StatusInternalServerError, "Failed to export document", alerts.ContextGeneral)
		}
		return
	}

	if format != "html" {
		alerts.RenderError(c, http.StatusBadRequest, "Unsupported export format", alerts.ContextGeneral)
		return
	}

	filename := fmt.Sprintf("document_%d.html", docID)
	if doc.DocumentType == models.DocumentTypeCoverLetter {
		filename = fmt.Sprintf("cover_letter_%d.html", doc.JobID)
	} else if doc.DocumentType == models.DocumentTypeResume {
		filename = fmt.Sprintf("resume_%d.html", doc.JobID)
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(doc.Content))
}

func (h *DocumentHandler) GetDocumentPartial(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		alerts.RenderError(c, http.StatusUnauthorized, "Authentication required", alerts.ContextGeneral)
		return
	}
	userID := userIDValue.(int)

	tab := c.DefaultQuery("tab", "cover-letters")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize := 20

	var documents []*models.DocumentSummary
	var total int
	var err error

	if tab == "resumes" {
		documents, total, err = h.service.GetDocumentsByType(
			c.Request.Context(), userID, models.DocumentTypeResume, page, pageSize,
		)
	} else {
		tab = "cover-letters"
		documents, total, err = h.service.GetDocumentsByType(
			c.Request.Context(), userID, models.DocumentTypeCoverLetter, page, pageSize,
		)
	}

	if err != nil {
		h.log.Error().Err(err).Str("tab", tab).Msg("Failed to get documents")
		alerts.RenderError(c, http.StatusInternalServerError, "Failed to load documents", alerts.ContextGeneral)
		return
	}

	totalPages := (total + pageSize - 1) / pageSize

	data := gin.H{
		"ActiveTab":      tab,
		"Documents":      documents,
		"TotalDocuments": total,
		"CurrentPage":    page,
		"TotalPages":     totalPages,
		"HasPrevPage":    page > 1,
		"HasNextPage":    page < totalPages,
		"PrevPage":       page - 1,
		"NextPage":       page + 1,
	}

	h.renderer.HTML(c, http.StatusOK, "documents/partials/document_list.html", data)
}
