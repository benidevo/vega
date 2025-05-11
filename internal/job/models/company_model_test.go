package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCompany(t *testing.T) {
	name := "Test Company"
	company := NewCompany(name)

	assert.Equal(t, name, company.Name)
	assert.NotZero(t, company.CreatedAt)
	assert.NotZero(t, company.UpdatedAt)
	assert.Equal(t, company.CreatedAt, company.UpdatedAt)
}
