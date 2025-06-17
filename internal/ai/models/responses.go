package models

// MatchResult represents the result of a matching process, including the match score,
// identified strengths and weaknesses, key highlights, and overall feedback.
type MatchResult struct {
	MatchScore int      `json:"matchScore"`
	Strengths  []string `json:"strengths"`
	Weaknesses []string `json:"weaknesses"`
	Highlights []string `json:"highlights"`
	Feedback   string   `json:"feedback"`
}

// CoverLetterFormat defines the format type for a cover letter, such as HTML, Markdown, or plain text.
// It is used to specify how the cover letter content is structured and presented.
type CoverLetterFormat string

const (
	CoverLetterTypeHtml      CoverLetterFormat = "html"
	CoverLetterTypeMarkdown  CoverLetterFormat = "markdown"
	CoverLetterTypePlainText CoverLetterFormat = "plain_text"
)

// CoverLetter represents a cover letter with its format and content.
type CoverLetter struct {
	Format  CoverLetterFormat `json:"format"`
	Content string            `json:"content"`
}
