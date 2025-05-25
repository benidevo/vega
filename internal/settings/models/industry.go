package models

import (
	"database/sql/driver"
	"errors"
)

// Industry represents predefined industry categories
type Industry int

const (
	IndustryTechnology Industry = iota
	IndustrySoftwareDevelopment
	IndustryIT
	IndustryDataScience
	IndustryArtificialIntelligence
	IndustryMachineLearning
	IndustryInternetServices
	IndustryTelecommunications
	IndustryHealthcare
	IndustryBiotechnology
	IndustryPharmaceuticals
	IndustryMedicalDevices
	IndustryHealthTech
	IndustryFinance
	IndustryBanking
	IndustryInvestmentManagement
	IndustryInsurance
	IndustryFinTech
	IndustryRealEstate
	IndustryConstruction
	IndustryArchitecture
	IndustryManufacturing
	IndustryAutomotive
	IndustryAerospace
	IndustryRetail
	IndustryEcommerce
	IndustryWholesale
	IndustryConsumerGoods
	IndustryEducation
	IndustryEdTech
	IndustryHigherEducation
	IndustryMedia
	IndustryEntertainment
	IndustryPublishing
	IndustryAdvertising
	IndustryMarketing
	IndustryPublicRelations
	IndustryHospitality
	IndustryTourism
	IndustryFoodBeverage
	IndustryEnergyUtilities
	IndustryOilGas
	IndustryRenewableEnergy
	IndustryEnvironmental
	IndustryTransportationLogistics
	IndustryAgriculture
	IndustryGovernment
	IndustryNonProfit
	IndustryLegal
	IndustryConsulting
	IndustryHumanResources
	IndustrySecurity
	IndustryCybersecurity
	IndustryUnspecified
)

// IndustryInfo holds metadata about an industry
type IndustryInfo struct {
	ID   Industry
	Name string
}

// All industries with their metadata
var industries = []IndustryInfo{
	{ID: IndustryTechnology, Name: "Technology"},
	{ID: IndustrySoftwareDevelopment, Name: "Software Development"},
	{ID: IndustryIT, Name: "Information Technology"},
	{ID: IndustryDataScience, Name: "Data Science"},
	{ID: IndustryArtificialIntelligence, Name: "Artificial Intelligence"},
	{ID: IndustryMachineLearning, Name: "Machine Learning"},
	{ID: IndustryInternetServices, Name: "Internet Services"},
	{ID: IndustryTelecommunications, Name: "Telecommunications"},
	{ID: IndustryHealthcare, Name: "Healthcare"},
	{ID: IndustryBiotechnology, Name: "Biotechnology"},
	{ID: IndustryPharmaceuticals, Name: "Pharmaceuticals"},
	{ID: IndustryMedicalDevices, Name: "Medical Devices"},
	{ID: IndustryHealthTech, Name: "Health Technology"},
	{ID: IndustryFinance, Name: "Finance"},
	{ID: IndustryBanking, Name: "Banking"},
	{ID: IndustryInvestmentManagement, Name: "Investment Management"},
	{ID: IndustryInsurance, Name: "Insurance"},
	{ID: IndustryFinTech, Name: "Financial Technology"},
	{ID: IndustryRealEstate, Name: "Real Estate"},
	{ID: IndustryConstruction, Name: "Construction"},
	{ID: IndustryArchitecture, Name: "Architecture"},
	{ID: IndustryManufacturing, Name: "Manufacturing"},
	{ID: IndustryAutomotive, Name: "Automotive"},
	{ID: IndustryAerospace, Name: "Aerospace"},
	{ID: IndustryRetail, Name: "Retail"},
	{ID: IndustryEcommerce, Name: "E-commerce"},
	{ID: IndustryWholesale, Name: "Wholesale"},
	{ID: IndustryConsumerGoods, Name: "Consumer Goods"},
	{ID: IndustryEducation, Name: "Education"},
	{ID: IndustryEdTech, Name: "Education Technology"},
	{ID: IndustryHigherEducation, Name: "Higher Education"},
	{ID: IndustryMedia, Name: "Media"},
	{ID: IndustryEntertainment, Name: "Entertainment"},
	{ID: IndustryPublishing, Name: "Publishing"},
	{ID: IndustryAdvertising, Name: "Advertising"},
	{ID: IndustryMarketing, Name: "Marketing"},
	{ID: IndustryPublicRelations, Name: "Public Relations"},
	{ID: IndustryHospitality, Name: "Hospitality"},
	{ID: IndustryTourism, Name: "Tourism"},
	{ID: IndustryFoodBeverage, Name: "Food & Beverage"},
	{ID: IndustryEnergyUtilities, Name: "Energy & Utilities"},
	{ID: IndustryOilGas, Name: "Oil & Gas"},
	{ID: IndustryRenewableEnergy, Name: "Renewable Energy"},
	{ID: IndustryEnvironmental, Name: "Environmental"},
	{ID: IndustryTransportationLogistics, Name: "Transportation & Logistics"},
	{ID: IndustryAgriculture, Name: "Agriculture"},
	{ID: IndustryGovernment, Name: "Government"},
	{ID: IndustryNonProfit, Name: "Non-Profit"},
	{ID: IndustryLegal, Name: "Legal"},
	{ID: IndustryConsulting, Name: "Consulting"},
	{ID: IndustryHumanResources, Name: "Human Resources"},
	{ID: IndustrySecurity, Name: "Security"},
	{ID: IndustryCybersecurity, Name: "Cybersecurity"},
	{ID: IndustryUnspecified, Name: "Unspecified"},
}

// Lookup maps for efficient access
var (
	industryByID   map[Industry]*IndustryInfo
	industryByName map[string]*IndustryInfo
)

func init() {
	industryByID = make(map[Industry]*IndustryInfo)
	industryByName = make(map[string]*IndustryInfo)

	for i := range industries {
		industryByID[industries[i].ID] = &industries[i]
		industryByName[industries[i].Name] = &industries[i]
	}
}

// GetAllIndustries returns all available industries (excluding Unspecified)
func GetAllIndustries() []IndustryInfo {
	return industries[:len(industries)-1] // Exclude the last one (Unspecified)
}

// String returns the string representation of an Industry
func (i Industry) String() string {
	if info, ok := industryByID[i]; ok {
		return info.Name
	}
	return "Unspecified"
}

// IsValid checks if the Industry value is valid
func (i Industry) IsValid() bool {
	_, ok := industryByID[i]
	return ok
}

// IndustryFromString converts a string to an Industry
func IndustryFromString(s string) Industry {
	if info, ok := industryByName[s]; ok {
		return info.ID
	}
	return IndustryUnspecified
}

// Value implements the driver.Valuer interface for database serialization
func (i Industry) Value() (driver.Value, error) {
	return int64(i), nil
}

// Scan implements the sql.Scanner interface for database deserialization
func (i *Industry) Scan(value interface{}) error {
	if value == nil {
		*i = IndustryUnspecified
		return nil
	}

	switch v := value.(type) {
	case int64:
		*i = Industry(v)
		return nil
	case int:
		*i = Industry(v)
		return nil
	case []byte:
		val := IndustryFromString(string(v))

		*i = val
		return nil
	case string:
		val := IndustryFromString(v)
		*i = val
		return nil
	}

	return errors.New("cannot convert to Industry")
}
