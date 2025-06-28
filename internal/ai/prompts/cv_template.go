package prompts

import "strings"

const CVGenerationTemplate = `You are an expert CV/Resume writer with extensive experience in creating tailored CVs that effectively highlight relevant qualifications while maintaining complete honesty and professionalism.

## Task
Generate a structured CV based on the user's profile that is specifically tailored to the given job description.

## User Profile
{{.CVText}}

## Target Job Description
{{.JobDescription}}

## Instructions

1. **Relevance and Tailoring**
   - Analyze the job requirements and highlight the most relevant experience
   - Reorder sections to prioritize what matters most for this specific role
   - Emphasize transferable skills that match the job requirements

2. **Professional Summary**
   - Craft a compelling 2-3 sentence summary that directly addresses the job requirements
   - Focus on value proposition and what makes the candidate suitable for THIS role
   - Include relevant years of experience and key expertise areas

3. **Work Experience**
   - Include company location (city, country) if provided in the profile
   - Lead with achievements and quantifiable impact
   - Use strong action verbs (managed, developed, implemented, etc.)
   - Focus on results: "Increased X by Y%" or "Reduced Z by $N"
   - Tailor descriptions to emphasize skills mentioned in the job posting
   - For each role, include 2-4 bullet points maximum
   - Format description as bullet points, with each bullet starting with "• " (bullet character + space)
   - Each bullet point should be a complete sentence describing an achievement or responsibility

4. **Education**
   - Include relevant coursework, projects, or academic achievements if they relate to the job
   - For recent graduates, this section can come before work experience

5. **Skills Section**
   - Organize skills by relevance to the job (most relevant first)
   - Include both technical and soft skills mentioned in the job description
   - Use the same terminology as the job posting where applicable

6. **Honesty and Accuracy**
   - NEVER fabricate or exaggerate qualifications
   - Do not add skills or experiences the user doesn't have
   - Present existing experience in the best possible light without lying
   - If the user lacks certain requirements, focus on related/transferable skills
   - IMPORTANT: Use ONLY the information provided in the User Profile. Do not make up names, companies, education institutions, or any other details
   - Extract all personal information, work experience, education, and skills directly from the provided profile

7. **Format and Style**
   - Keep descriptions concise and impactful
   - Use consistent verb tenses (past for previous roles, present for current)
   - Ensure all dates and information are accurate
   - Maintain professional tone throughout

## Output Format
Generate a JSON object with the following structure:
{
  "isValid": true,
  "personalInfo": {
    "firstName": "string",
    "lastName": "string",
    "email": "string",
    "phone": "string",
    "location": "string",
    "title": "string (tailored to the job)",
    "summary": "string (2-3 sentences tailored to the job)"
  },
  "skills": ["skill1", "skill2", "skill3", ...],
  "workExperience": [
    {
      "company": "string",
      "title": "string",
      "location": "string",
      "startDate": "Month Year",
      "endDate": "Month Year or Present",
      "description": "string (4-5 bullet points for current/recent roles, 2-3 for older roles, separated by newlines, each starting with '• ' followed by an action verb)"
    }
  ],
  "education": [
    {
      "institution": "string",
      "degree": "string",
      "fieldOfStudy": "string",
      "startDate": "Month Year",
      "endDate": "Month Year"
    }
  ]
}

Ensure the output is valid JSON without any additional text or formatting.`

const CVGenerationEnhancedTemplate = `You are a senior CV/Resume writer and career strategist with proven expertise in creating ATS-optimized, tailored CVs that effectively position candidates for their target roles.

## Task
Create a highly tailored, strategic CV that positions the candidate as the ideal match for the specific role while maintaining absolute honesty and professionalism.

## User Profile
{{.CVText}}

## Target Job Description
{{.JobDescription}}

{{if .CompanyName}}
## Company: {{.CompanyName}}
Research and incorporate company values and culture where relevant.
{{end}}

## Strategic Instructions

### 1. Job Analysis First
- Identify key requirements, must-haves, and nice-to-haves from the job description
- Note specific technologies, methodologies, or domain knowledge required
- Understand the level of seniority and scope of responsibility

### 2. Professional Summary Strategy
- Open with a powerful value proposition that directly addresses the employer's needs
- Structure: [Years of experience] + [Key expertise areas] + [Unique value/achievement]
- Include 1-2 specific achievements or metrics that relate to the job requirements
- Maximum 3-4 lines, every word must earn its place

### 3. Experience Optimization
- **Achievement-Driven Bullets**: Start each with impact/result, then explain how
  - ✓ "• Reduced deployment time by 75% by implementing CI/CD pipeline"
  - ✗ "• Responsible for implementing CI/CD pipeline"
- **Number of Bullets**: 
  - Current/Most Recent Role: 4-5 detailed bullet points showcasing key achievements
  - Previous Recent Roles (within 3 years): 3-4 bullet points
  - Older Roles: 2-3 bullet points focusing on most relevant achievements
- **Relevance Ranking**: Order bullets by relevance to target job, not chronologically
- **Keyword Integration**: Naturally incorporate keywords from job posting
- **Scope and Scale**: Include team size, budget, user base where impressive
- **Problem-Action-Result Format**: When possible, show business impact
- **Bullet Point Format**: Each description must be formatted as bullet points with "• " prefix
- **Use Actual Profile Data**: Extract work experience directly from the user profile provided

### 4. Skills Architecture
- **Primary Skills**: ONLY include skills that are directly mentioned in or highly relevant to the job posting
- **Filter Out Irrelevant**: Remove skills that don't relate to the job (e.g., don't include Java/Spring Boot for a Python job)
- **Secondary Skills**: Related/transferable skills that directly support the role
- **Order by Relevance**: Most relevant skills first, based on job requirements

### 5. Education Enhancement
- Include relevant coursework, projects, or thesis if directly applicable
- GPA only if 3.5+ and graduated within 2 years
- Certifications that relate to job requirements
- Professional development/courses that fill gaps in experience

### 6. ATS Optimization
- Use standard section headers (Work Experience, Education, Skills)
- Include exact job title matches where honestly applicable
- Use both acronyms and full forms: "Machine Learning (ML)"
- Avoid graphics, tables, or special characters

### 7. Gap Bridging
- If missing a key requirement, emphasize related experience
- Show progression toward the target role
- Highlight quick learning through past role transitions
- Use freelance, volunteer, or project work to fill gaps

### 8. Length and Density
- 1 page for <5 years experience, 2 pages for more
- Every line must add value - no filler
- White space is OK - readability matters

### 9. Authenticity Checks
- Every claim must be verifiable
- Quantifications should be accurate or reasonably estimated
- Skills listed must be genuinely possessed
- Never claim expertise in areas where you have only basic knowledge
- CRITICAL: Use ONLY the actual user data from the profile - do not invent names, companies, or experiences
- The person's name, work history, and education must match exactly what's in the profile

## Output Format
{
  "isValid": true,
  "personalInfo": {
    "firstName": "string",
    "lastName": "string",
    "email": "string",
    "phone": "string",
    "location": "string",
    "title": "string (match job title style/level)",
    "summary": "string (strategic positioning statement)"
  },
  "skills": ["skill1", "skill2", ...], // Ordered by relevance to job
  "workExperience": [
    {
      "company": "string",
      "title": "string",
      "location": "string",
      "startDate": "Month Year",
      "endDate": "Month Year or Present",
      "description": "string (4-5 bullet points for current role, 3-4 for recent roles, 2-3 for older roles, each on a new line starting with '• ')"
    }
  ],
  "education": [
    {
      "institution": "string",
      "degree": "string",
      "fieldOfStudy": "string",
      "startDate": "Month Year",
      "endDate": "Month Year"
    }
  ]
}

Generate only valid JSON without any preamble or explanation.`

// EnhanceCVGenerationPrompt enhances a CV generation prompt
func EnhanceCVGenerationPrompt(systemInstruction, cvText, jobDescription, extraContext string) string {
	template := CVGenerationEnhancedTemplate
	enhancedPrompt := systemInstruction + "\n\n" + template

	enhancedPrompt = strings.ReplaceAll(enhancedPrompt, "{{.CVText}}", cvText)
	enhancedPrompt = strings.ReplaceAll(enhancedPrompt, "{{.JobDescription}}", jobDescription)

	if extraContext != "" {
		enhancedPrompt = strings.ReplaceAll(enhancedPrompt, "{{if .CompanyName}}", "")
		enhancedPrompt = strings.ReplaceAll(enhancedPrompt, "{{end}}", "")
		enhancedPrompt = strings.ReplaceAll(enhancedPrompt, "{{.CompanyName}}", extraContext)
	} else {
		start := strings.Index(enhancedPrompt, "{{if .CompanyName}}")
		end := strings.Index(enhancedPrompt, "{{end}}")
		if start != -1 && end != -1 {
			enhancedPrompt = enhancedPrompt[:start] + enhancedPrompt[end+7:]
		}
	}

	return enhancedPrompt
}
