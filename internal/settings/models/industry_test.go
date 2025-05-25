package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndustryString(t *testing.T) {
	tests := []struct {
		industry Industry
		want     string
	}{
		{IndustryTechnology, "Technology"},
		{IndustrySoftwareDevelopment, "Software Development"},
		{IndustryHealthcare, "Healthcare"},
		{IndustryFinance, "Finance"},
		{IndustryEducation, "Education"},
		{IndustryUnspecified, "Unspecified"},
		{Industry(999), "Unspecified"}, // Invalid industry
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.industry.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIndustryIsValid(t *testing.T) {
	validIndustries := GetAllIndustries()
	for _, info := range validIndustries {
		assert.True(t, info.ID.IsValid(), "Industry %s should be valid", info.Name)
	}

	assert.True(t, IndustryUnspecified.IsValid())

	invalidIndustry := Industry(999)
	assert.False(t, invalidIndustry.IsValid())
}

func TestIndustryFromString(t *testing.T) {
	tests := []struct {
		name string
		want Industry
	}{
		{"Technology", IndustryTechnology},
		{"Software Development", IndustrySoftwareDevelopment},
		{"Healthcare", IndustryHealthcare},
		{"Finance", IndustryFinance},
		{"Education", IndustryEducation},
		{"Unspecified", IndustryUnspecified},
		{"Invalid Industry", IndustryUnspecified}, // Unknown returns Unspecified
		{"", IndustryUnspecified},                 // Empty returns Unspecified
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IndustryFromString(tt.name)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetAllIndustries(t *testing.T) {
	industries := GetAllIndustries()

	require.NotEmpty(t, industries)

	for _, info := range industries {
		assert.NotEqual(t, IndustryUnspecified, info.ID, "GetAllIndustries should not include Unspecified")
	}

	expectedNames := []string{
		"Technology",
		"Software Development",
		"Healthcare",
		"Finance",
		"Education",
	}

	industryNames := make(map[string]bool)
	for _, info := range industries {
		industryNames[info.Name] = true
	}

	for _, expected := range expectedNames {
		assert.True(t, industryNames[expected], "Expected industry %s not found", expected)
	}
}

func TestIndustryValue(t *testing.T) {
	tests := []struct {
		industry Industry
		want     int64
	}{
		{IndustryTechnology, 0},
		{IndustrySoftwareDevelopment, 1},
		{IndustryHealthcare, 8},
		{IndustryFinance, 13},
		{IndustryEducation, 28},
		{IndustryUnspecified, 53},
	}

	for _, tt := range tests {
		t.Run(tt.industry.String(), func(t *testing.T) {
			value, err := tt.industry.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.want, value)
		})
	}
}

func TestIndustryScan(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    Industry
		wantErr bool
	}{
		{
			name:  "scan int64",
			value: int64(0),
			want:  IndustryTechnology,
		},
		{
			name:  "scan int",
			value: 13,
			want:  IndustryFinance,
		},
		{
			name:  "scan string name",
			value: "Healthcare",
			want:  IndustryHealthcare,
		},
		{
			name:  "scan byte slice",
			value: []byte("Education"),
			want:  IndustryEducation,
		},
		{
			name:  "scan nil",
			value: nil,
			want:  IndustryUnspecified,
		},
		{
			name:  "scan unknown string",
			value: "Unknown Industry",
			want:  IndustryUnspecified,
		},
		{
			name:    "scan invalid type",
			value:   123.45,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var industry Industry
			err := industry.Scan(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, industry)
			}
		})
	}
}

func TestIndustryLookupMaps(t *testing.T) {
	for _, info := range industries {
		found, ok := industryByID[info.ID]
		assert.True(t, ok, "Industry ID %d not found in industryByID map", info.ID)
		assert.Equal(t, info.Name, found.Name)

		found, ok = industryByName[info.Name]
		assert.True(t, ok, "Industry name %s not found in industryByName map", info.Name)
		assert.Equal(t, info.ID, found.ID)
	}

	assert.Equal(t, len(industries), len(industryByID))
	assert.Equal(t, len(industries), len(industryByName))
}
