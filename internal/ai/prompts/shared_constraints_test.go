package prompts

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBannedAIPhrases(t *testing.T) {
	assert.Contains(t, BannedAIPhrases, "leverage")
	assert.Contains(t, BannedAIPhrases, "utilize")
	assert.Contains(t, BannedAIPhrases, "spearheaded")
	assert.Contains(t, BannedAIPhrases, "game-changer")
	assert.Contains(t, BannedAIPhrases, "low-hanging fruit")
}

func TestAntiAILanguageConstraints(t *testing.T) {
	constraints := AntiAILanguageConstraints()

	assert.Greater(t, len(constraints), 5)

	constraintText := strings.Join(constraints, " ")
	assert.Contains(t, constraintText, "WRITE LIKE A HUMAN")
	assert.Contains(t, constraintText, "BANNED PHRASES")
	assert.Contains(t, constraintText, "NATURAL LANGUAGE ONLY")
	assert.Contains(t, constraintText, "NO CORPORATE BUZZWORDS")
	assert.Contains(t, constraintText, "CONVERSATIONAL TONE")
	assert.Contains(t, constraintText, "SPECIFIC > GENERIC")
	assert.Contains(t, constraintText, "HUMAN TEST")
}

func TestCVAntiAIConstraints(t *testing.T) {
	constraints := CVAntiAIConstraints()

	baseConstraints := AntiAILanguageConstraints()
	assert.Greater(t, len(constraints), len(baseConstraints))

	constraintText := strings.Join(constraints, " ")
	assert.Contains(t, constraintText, "WRITE LIKE A HUMAN")
	assert.Contains(t, constraintText, "TRANSFORM and ENHANCE")
	assert.Contains(t, constraintText, "ELEVATE TRUTHFULLY")
	assert.Contains(t, constraintText, "BEST FOOT FORWARD")
}

func TestCoverLetterAntiAIConstraints(t *testing.T) {
	constraints := CoverLetterAntiAIConstraints()

	baseConstraints := AntiAILanguageConstraints()
	assert.Greater(t, len(constraints), len(baseConstraints))

	constraintText := strings.Join(constraints, " ")
	assert.Contains(t, constraintText, "WRITE LIKE A HUMAN")
	assert.Contains(t, constraintText, "professional conversation")
	assert.Contains(t, constraintText, "genuine interest and personality")
	assert.Contains(t, constraintText, "Skip tired openings")
}

func TestJobMatchAntiAIConstraints(t *testing.T) {
	constraints := JobMatchAntiAIConstraints()

	baseConstraints := AntiAILanguageConstraints()
	assert.Greater(t, len(constraints), len(baseConstraints))

	constraintText := strings.Join(constraints, " ")
	assert.Contains(t, constraintText, "WRITE LIKE A HUMAN")
	assert.Contains(t, constraintText, "NATURAL RECRUITER LANGUAGE")
	assert.Contains(t, constraintText, "CONVERSATIONAL BUT DIRECT")
	assert.Contains(t, constraintText, "concrete examples")
}
