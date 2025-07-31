package prompts

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromptTemplate_BuildPrompt(t *testing.T) {
	tests := []struct {
		name               string
		template           *PromptTemplate
		systemInstruction  string
		applicantName      string
		jobDescription     string
		applicantProfile   string
		extraContext       string
		params             map[string]any
		expectedContains   []string
		unexpectedContains []string
	}{
		{
			name: "should_build_basic_prompt_without_optional_fields",
			template: &PromptTemplate{
				Role:       "Test Role",
				Context:    "Test Context",
				Task:       "Test Task",
				OutputSpec: "Test Output",
			},
			applicantName:    "John Doe",
			jobDescription:   "Software Engineer",
			applicantProfile: "Experienced developer",
			expectedContains: []string{
				"# Your Role\nTest Role",
				"# Context\nTest Context",
				"**Applicant Name:** John Doe",
				"**Job Description:**\nSoftware Engineer",
				"**Applicant Profile:**\nExperienced developer",
				"# Your Task\nTest Task",
				"# Output Format\nTest Output",
			},
			unexpectedContains: []string{
				"System Instruction",
				"Examples",
				"Additional Context",
				"Requirements and Constraints",
				"Word Count",
				"Thinking Process",
			},
		},
		{
			name: "should_include_system_instruction_when_provided",
			template: &PromptTemplate{
				Role:       "Assistant",
				Context:    "Context",
				Task:       "Task",
				OutputSpec: "Output",
			},
			systemInstruction: "You are a helpful assistant",
			applicantName:     "Jane Smith",
			jobDescription:    "Data Scientist",
			applicantProfile:  "ML Expert",
			expectedContains: []string{
				"You are a helpful assistant\n\n",
				"# Your Role\nAssistant",
			},
		},
		{
			name: "should_include_examples_when_provided",
			template: &PromptTemplate{
				Role:    "Writer",
				Context: "Writing context",
				Examples: []Example{
					{
						Input:  "Sample input",
						Output: "Sample output",
					},
					{
						Input:  "Another input",
						Output: "Another output",
					},
				},
				Task:       "Write something",
				OutputSpec: "JSON format",
			},
			applicantName:    "Test User",
			jobDescription:   "Writer position",
			applicantProfile: "Writing experience",
			expectedContains: []string{
				"# Examples of Excellent Output",
				"## Example 1",
				"Input: Sample input",
				"Output: Sample output",
				"## Example 2",
				"Input: Another input",
				"Output: Another output",
			},
		},
		{
			name: "should_include_constraints_when_provided",
			template: &PromptTemplate{
				Role:    "Analyzer",
				Context: "Analysis context",
				Task:    "Analyze data",
				Constraints: []string{
					"Be accurate",
					"Be concise",
					"Use proper formatting",
				},
				OutputSpec: "Report format",
			},
			applicantName:    "Analyst",
			jobDescription:   "Data Analyst",
			applicantProfile: "Analysis skills",
			expectedContains: []string{
				"# Requirements and Constraints",
				"- Be accurate",
				"- Be concise",
				"- Use proper formatting",
			},
		},
		{
			name: "should_include_extra_context_when_provided",
			template: &PromptTemplate{
				Role:       "Evaluator",
				Context:    "Evaluation context",
				Task:       "Evaluate candidate",
				OutputSpec: "Score format",
			},
			applicantName:    "Candidate",
			jobDescription:   "Position",
			applicantProfile: "Profile",
			extraContext:     "Additional important information",
			expectedContains: []string{
				"**Additional Context:**\nAdditional important information",
			},
		},
		{
			name: "should_include_current_date_when_in_params",
			template: &PromptTemplate{
				Role:       "Timer",
				Context:    "Time context",
				Task:       "Time task",
				OutputSpec: "Time output",
			},
			applicantName:    "User",
			jobDescription:   "Job",
			applicantProfile: "Profile",
			params: map[string]any{
				"currentDate": "2024-03-15",
			},
			expectedContains: []string{
				"**Current Date:** 2024-03-15",
			},
		},
		{
			name: "should_include_word_range_when_in_params",
			template: &PromptTemplate{
				Role:       "Writer",
				Context:    "Writing",
				Task:       "Write",
				OutputSpec: "Text",
			},
			applicantName:    "Writer",
			jobDescription:   "Writing job",
			applicantProfile: "Writer profile",
			params: map[string]any{
				"wordRange": "300-500",
			},
			expectedContains: []string{
				"**Word Count:** 300-500 words",
			},
		},
		{
			name: "should_include_chain_of_thought_when_in_params",
			template: &PromptTemplate{
				Role:       "Thinker",
				Context:    "Thinking",
				Task:       "Think",
				OutputSpec: "Thoughts",
			},
			applicantName:    "Thinker",
			jobDescription:   "Thinking job",
			applicantProfile: "Thinker profile",
			params: map[string]any{
				"useChainOfThought": true,
			},
			expectedContains: []string{
				"# Thinking Process",
				"Before providing your final answer",
				"1. Key requirements from the job description",
				"2. Matching qualifications from the candidate",
				"3. Gaps or areas of concern",
				"4. Overall fit assessment",
				"Then provide your final JSON response",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.template.BuildPrompt(
				tt.systemInstruction,
				tt.applicantName,
				tt.jobDescription,
				tt.applicantProfile,
				tt.extraContext,
				tt.params,
			)

			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected)
			}

			for _, unexpected := range tt.unexpectedContains {
				assert.NotContains(t, result, unexpected)
			}
		})
	}
}

func TestPromptTemplate_BuildPrompt_CompleteExample(t *testing.T) {
	template := &PromptTemplate{
		Role:    "Professional Career Advisor",
		Context: "You are helping candidates apply for jobs",
		Examples: []Example{
			{
				Input:  "Developer with 5 years experience",
				Output: "Strong technical candidate",
			},
		},
		Task: "Evaluate the candidate",
		Constraints: []string{
			"Be objective",
			"Focus on skills",
		},
		OutputSpec: "JSON with score and feedback",
	}

	result := template.BuildPrompt(
		"You are an AI assistant",
		"John Smith",
		"Senior Software Engineer at Tech Corp",
		"Full-stack developer with React and Go experience",
		"Looking for remote opportunities",
		map[string]any{
			"currentDate":       "2024-03-15",
			"wordRange":         "200-300",
			"useChainOfThought": true,
		},
	)

	assert.True(t, strings.HasPrefix(result, "You are an AI assistant\n\n"))
	assert.Contains(t, result, "# Your Role\nProfessional Career Advisor")
	assert.Contains(t, result, "# Context\nYou are helping candidates apply for jobs")
	assert.Contains(t, result, "# Examples of Excellent Output")
	assert.Contains(t, result, "**Current Date:** 2024-03-15")
	assert.Contains(t, result, "**Applicant Name:** John Smith")
	assert.Contains(t, result, "**Additional Context:**\nLooking for remote opportunities")
	assert.Contains(t, result, "# Requirements and Constraints")
	assert.Contains(t, result, "**Word Count:** 200-300 words")
	assert.Contains(t, result, "# Thinking Process")
}
