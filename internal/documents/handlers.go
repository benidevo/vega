package documents

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/benidevo/vega/internal/common/alerts"
	"github.com/benidevo/vega/internal/common/logger"
	"github.com/benidevo/vega/internal/common/render"
	"github.com/benidevo/vega/internal/config"
	"github.com/benidevo/vega/internal/documents/models"
	jobmodels "github.com/benidevo/vega/internal/job/models"
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
	pageSize := 9

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
		"page":             "documents",
		"activeNav":        "documents",
		"title":            "My Documents",
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
	h.renderer.HTML(c, http.StatusOK, "layouts/base.html", data)
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

	redirectURL := fmt.Sprintf("/jobs/%d/details", doc.JobID)
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userID := userIDValue.(int)

	docIDStr := c.Param("id")
	docID, err := strconv.Atoi(docIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	doc, err := h.service.GetDocument(c.Request.Context(), docID, userID)
	if err != nil {
		if err == models.ErrDocumentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		} else {
			h.log.Error().Err(err).Int("doc_id", docID).Msg("Failed to get document for export")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export document"})
		}
		return
	}

	var jobTitle, companyName string
	if doc.JobID > 0 {
		documents, _, err := h.service.GetDocumentsByType(
			c.Request.Context(),
			userID,
			doc.DocumentType,
			1, // page
			1, // just need one to get job info
		)
		if err == nil && len(documents) > 0 {
			for _, summary := range documents {
				if summary.JobID == doc.JobID {
					jobTitle = summary.JobTitle
					companyName = summary.CompanyName
					break
				}
			}
		}
	}

	// Prepare response data
	responseData := gin.H{
		"content":     doc.Content,
		"type":        string(doc.DocumentType),
		"jobId":       doc.JobID,
		"jobTitle":    jobTitle,
		"companyName": companyName,
	}

	// For cover letters, try to fetch personal info
	if doc.DocumentType == models.DocumentTypeCoverLetter {
		var coverLetterData struct {
			Content      string                  `json:"content"`
			PersonalInfo *jobmodels.PersonalInfo `json:"personalInfo,omitempty"`
		}

		if err := json.Unmarshal([]byte(doc.Content), &coverLetterData); err == nil && coverLetterData.PersonalInfo != nil {
			responseData["content"] = coverLetterData.Content
			responseData["personalInfo"] = gin.H{
				"firstName": coverLetterData.PersonalInfo.FirstName,
				"lastName":  coverLetterData.PersonalInfo.LastName,
				"title":     coverLetterData.PersonalInfo.Title,
				"email":     coverLetterData.PersonalInfo.Email,
				"phone":     coverLetterData.PersonalInfo.Phone,
				"location":  coverLetterData.PersonalInfo.Location,
				"linkedin":  coverLetterData.PersonalInfo.LinkedIn,
			}
		} else if doc.JobID > 0 {
			resumeDoc, err := h.service.GetDocumentByJobAndType(
				c.Request.Context(),
				userID,
				doc.JobID,
				models.DocumentTypeResume,
			)

			if err == nil && resumeDoc != nil {
				var cv jobmodels.GeneratedCV
				if err := json.Unmarshal([]byte(resumeDoc.Content), &cv); err == nil {
					responseData["personalInfo"] = gin.H{
						"firstName": cv.PersonalInfo.FirstName,
						"lastName":  cv.PersonalInfo.LastName,
						"title":     cv.PersonalInfo.Title,
						"email":     cv.PersonalInfo.Email,
						"phone":     cv.PersonalInfo.Phone,
						"location":  cv.PersonalInfo.Location,
						"linkedin":  cv.PersonalInfo.LinkedIn,
					}
				} else {
					h.log.Warn().Err(err).Msg("Failed to parse resume JSON for personal info")
				}
			} else if err != nil && err != models.ErrDocumentNotFound {
				h.log.Warn().Err(err).Int("job_id", doc.JobID).Msg("Failed to fetch resume for personal info")
			}
		}
	}

	c.JSON(http.StatusOK, responseData)
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
	pageSize := 9

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

func (h *DocumentHandler) formatCVAsHTML(cv *jobmodels.GeneratedCV) string {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Resume - ` + cv.PersonalInfo.FirstName + ` ` + cv.PersonalInfo.LastName + `</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Helvetica', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 800px;
            margin: 0 auto;
            padding: 40px 20px;
            background: white;
        }
        h1 { font-size: 28px; margin-bottom: 8px; color: #111; }
        h2 { 
            font-size: 16px; 
            margin-top: 24px; 
            margin-bottom: 12px; 
            padding-bottom: 4px;
            border-bottom: 2px solid #333;
            color: #111;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        h3 { font-size: 14px; margin-top: 16px; margin-bottom: 8px; color: #333; }
        .header-section { margin-bottom: 24px; }
        .contact-info { color: #666; font-size: 12px; margin-bottom: 16px; }
        .title { font-size: 14px; color: #444; margin-bottom: 8px; }
        .section { margin-bottom: 24px; }
        .experience-item, .education-item { margin-bottom: 16px; page-break-inside: avoid; }
        .date { color: #666; font-size: 11px; margin-bottom: 4px; }
        .description { font-size: 12px; line-height: 1.5; color: #444; }
        .skills { font-size: 12px; line-height: 1.8; }
        @media print {
            body { padding: 0; }
            h2 { page-break-after: avoid; }
            .experience-item, .education-item { page-break-inside: avoid; }
        }
    </style>
</head>
<body>`

	html += `<div class="header-section">`
	fullName := fmt.Sprintf("%s %s", cv.PersonalInfo.FirstName, cv.PersonalInfo.LastName)
	html += fmt.Sprintf("<h1>%s</h1>", fullName)

	if cv.PersonalInfo.Title != "" {
		html += fmt.Sprintf(`<div class="title">%s</div>`, cv.PersonalInfo.Title)
	}

	var contactParts []string
	if cv.PersonalInfo.Location != "" {
		contactParts = append(contactParts, cv.PersonalInfo.Location)
	}
	if cv.PersonalInfo.Email != "" {
		contactParts = append(contactParts, cv.PersonalInfo.Email)
	}
	if cv.PersonalInfo.Phone != "" {
		contactParts = append(contactParts, cv.PersonalInfo.Phone)
	}
	if cv.PersonalInfo.LinkedIn != "" {
		contactParts = append(contactParts, cv.PersonalInfo.LinkedIn)
	}

	if len(contactParts) > 0 {
		html += fmt.Sprintf(`<div class="contact-info">%s</div>`, strings.Join(contactParts, " | "))
	}
	html += `</div>`

	if cv.PersonalInfo.Summary != "" {
		html += `<div class="section">`
		html += `<h2>Professional Summary</h2>`
		html += fmt.Sprintf(`<p class="description">%s</p>`, cv.PersonalInfo.Summary)
		html += `</div>`
	}

	if len(cv.Skills) > 0 {
		html += `<div class="section">`
		html += `<h2>Skills</h2>`
		html += fmt.Sprintf(`<div class="skills">%s</div>`, strings.Join(cv.Skills, " • "))
		html += `</div>`
	}

	if len(cv.WorkExperience) > 0 {
		html += `<div class="section">`
		html += `<h2>Work Experience</h2>`
		for _, exp := range cv.WorkExperience {
			html += `<div class="experience-item">`
			html += fmt.Sprintf("<h3>%s at %s</h3>", exp.Title, exp.Company)
			dateStr := fmt.Sprintf("%s - %s", exp.StartDate, exp.EndDate)
			if exp.Location != "" {
				dateStr += fmt.Sprintf(" | %s", exp.Location)
			}
			html += fmt.Sprintf(`<div class="date">%s</div>`, dateStr)
			html += fmt.Sprintf(`<div class="description">%s</div>`, h.formatDescription(exp.Description))
			html += `</div>`
		}
		html += `</div>`
	}

	if len(cv.Education) > 0 {
		html += `<div class="section">`
		html += `<h2>Education</h2>`
		for _, edu := range cv.Education {
			html += `<div class="education-item">`
			degreeStr := edu.Degree
			if edu.FieldOfStudy != "" {
				degreeStr += fmt.Sprintf(" in %s", edu.FieldOfStudy)
			}
			html += fmt.Sprintf("<h3>%s</h3>", degreeStr)
			html += fmt.Sprintf("<div>%s</div>", edu.Institution)
			html += fmt.Sprintf(`<div class="date">%s - %s</div>`, edu.StartDate, edu.EndDate)
			html += `</div>`
		}
		html += `</div>`
	}

	if len(cv.Certifications) > 0 {
		html += `<div class="section">`
		html += `<h2>Certifications</h2>`
		for _, cert := range cv.Certifications {
			html += `<div style="margin-bottom: 8px;">`
			html += fmt.Sprintf("<strong>%s</strong>", cert.Name)
			if cert.IssuingOrg != "" {
				html += fmt.Sprintf(" - %s", cert.IssuingOrg)
			}
			if cert.IssueDate != "" {
				html += fmt.Sprintf(" (%s)", cert.IssueDate)
			}
			html += `</div>`
		}
		html += `</div>`
	}

	html += `</body></html>`
	return html
}

func (h *DocumentHandler) formatDescription(text string) string {
	if text == "" {
		return ""
	}

	lines := strings.Split(text, "\n")
	var result strings.Builder
	inList := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "•") || strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") {
			if !inList {
				result.WriteString("<ul>")
				inList = true
			}
			content := strings.TrimSpace(trimmed[1:])
			result.WriteString(fmt.Sprintf("<li>%s</li>", content))
		} else {
			if inList {
				result.WriteString("</ul>")
				inList = false
			}
			if trimmed != "" {
				result.WriteString(fmt.Sprintf("<p>%s</p>", trimmed))
			}
		}
	}

	if inList {
		result.WriteString("</ul>")
	}

	formatted := result.String()
	if formatted == "" {
		return text
	}
	return formatted
}

// SaveDocumentRequest represents the request body for saving a document
type SaveDocumentRequest struct {
	JobID        int    `json:"jobId" binding:"required"`
	DocumentType string `json:"documentType" binding:"required,oneof=resume cover_letter"`
	Content      string `json:"content" binding:"required"`
}

// SaveDocument handles saving a document from the job details page
func (h *DocumentHandler) SaveDocument(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userID := userIDValue.(int)

	var req SaveDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn().Err(err).Msg("Invalid save document request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	var docType models.DocumentType
	if req.DocumentType == "resume" {
		docType = models.DocumentTypeResume
	} else if req.DocumentType == "cover_letter" {
		docType = models.DocumentTypeCoverLetter
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document type"})
		return
	}

	content := req.Content

	if docType == models.DocumentTypeResume {
		var cv jobmodels.GeneratedCV
		if err := json.Unmarshal([]byte(req.Content), &cv); err != nil {
			h.log.Error().Err(err).Msg("Failed to parse resume JSON")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resume data"})
			return
		}
	}

	doc, err := h.service.SaveGeneratedDocument(
		c.Request.Context(),
		userID,
		req.JobID,
		docType,
		content,
	)

	if err != nil {
		h.log.Error().
			Err(err).
			Int("user_id", userID).
			Int("job_id", req.JobID).
			Str("document_type", string(docType)).
			Msg("Failed to save document")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save document"})
		return
	}

	h.log.Info().
		Int("user_id", userID).
		Int("job_id", req.JobID).
		Int("document_id", doc.ID).
		Str("document_type", string(docType)).
		Msg("Document saved successfully")

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Document saved successfully",
		"documentId": doc.ID,
	})
}
