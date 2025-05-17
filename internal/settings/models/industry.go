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

// String returns the string representation of an Industry
func (i Industry) String() string {
	switch i {
	case IndustryTechnology:
		return "Technology"
	case IndustrySoftwareDevelopment:
		return "Software Development"
	case IndustryIT:
		return "Information Technology"
	case IndustryDataScience:
		return "Data Science"
	case IndustryArtificialIntelligence:
		return "Artificial Intelligence"
	case IndustryMachineLearning:
		return "Machine Learning"
	case IndustryInternetServices:
		return "Internet Services"
	case IndustryTelecommunications:
		return "Telecommunications"
	case IndustryHealthcare:
		return "Healthcare"
	case IndustryBiotechnology:
		return "Biotechnology"
	case IndustryPharmaceuticals:
		return "Pharmaceuticals"
	case IndustryMedicalDevices:
		return "Medical Devices"
	case IndustryHealthTech:
		return "Health Technology"
	case IndustryFinance:
		return "Finance"
	case IndustryBanking:
		return "Banking"
	case IndustryInvestmentManagement:
		return "Investment Management"
	case IndustryInsurance:
		return "Insurance"
	case IndustryFinTech:
		return "Financial Technology"
	case IndustryRealEstate:
		return "Real Estate"
	case IndustryConstruction:
		return "Construction"
	case IndustryArchitecture:
		return "Architecture"
	case IndustryManufacturing:
		return "Manufacturing"
	case IndustryAutomotive:
		return "Automotive"
	case IndustryAerospace:
		return "Aerospace"
	case IndustryRetail:
		return "Retail"
	case IndustryEcommerce:
		return "E-commerce"
	case IndustryWholesale:
		return "Wholesale"
	case IndustryConsumerGoods:
		return "Consumer Goods"
	case IndustryEducation:
		return "Education"
	case IndustryEdTech:
		return "Education Technology"
	case IndustryHigherEducation:
		return "Higher Education"
	case IndustryMedia:
		return "Media"
	case IndustryEntertainment:
		return "Entertainment"
	case IndustryPublishing:
		return "Publishing"
	case IndustryAdvertising:
		return "Advertising"
	case IndustryMarketing:
		return "Marketing"
	case IndustryPublicRelations:
		return "Public Relations"
	case IndustryHospitality:
		return "Hospitality"
	case IndustryTourism:
		return "Tourism"
	case IndustryFoodBeverage:
		return "Food & Beverage"
	case IndustryEnergyUtilities:
		return "Energy & Utilities"
	case IndustryOilGas:
		return "Oil & Gas"
	case IndustryRenewableEnergy:
		return "Renewable Energy"
	case IndustryEnvironmental:
		return "Environmental"
	case IndustryTransportationLogistics:
		return "Transportation & Logistics"
	case IndustryAgriculture:
		return "Agriculture"
	case IndustryGovernment:
		return "Government"
	case IndustryNonProfit:
		return "Non-Profit"
	case IndustryLegal:
		return "Legal"
	case IndustryConsulting:
		return "Consulting"
	case IndustryHumanResources:
		return "Human Resources"
	case IndustrySecurity:
		return "Security"
	case IndustryCybersecurity:
		return "Cybersecurity"
	default:
		return "Unspecified"
	}
}

// IndustryFromString converts a string to an Industry
func IndustryFromString(s string) Industry {
	switch s {
	case "Technology":
		return IndustryTechnology
	case "Software Development":
		return IndustrySoftwareDevelopment
	case "Information Technology":
		return IndustryIT
	case "Data Science":
		return IndustryDataScience
	case "Artificial Intelligence":
		return IndustryArtificialIntelligence
	case "Machine Learning":
		return IndustryMachineLearning
	case "Internet Services":
		return IndustryInternetServices
	case "Telecommunications":
		return IndustryTelecommunications
	case "Healthcare":
		return IndustryHealthcare
	case "Biotechnology":
		return IndustryBiotechnology
	case "Pharmaceuticals":
		return IndustryPharmaceuticals
	case "Medical Devices":
		return IndustryMedicalDevices
	case "Health Technology":
		return IndustryHealthTech
	case "Finance":
		return IndustryFinance
	case "Banking":
		return IndustryBanking
	case "Investment Management":
		return IndustryInvestmentManagement
	case "Insurance":
		return IndustryInsurance
	case "Financial Technology":
		return IndustryFinTech
	case "Real Estate":
		return IndustryRealEstate
	case "Construction":
		return IndustryConstruction
	case "Architecture":
		return IndustryArchitecture
	case "Manufacturing":
		return IndustryManufacturing
	case "Automotive":
		return IndustryAutomotive
	case "Aerospace":
		return IndustryAerospace
	case "Retail":
		return IndustryRetail
	case "E-commerce":
		return IndustryEcommerce
	case "Wholesale":
		return IndustryWholesale
	case "Consumer Goods":
		return IndustryConsumerGoods
	case "Education":
		return IndustryEducation
	case "Education Technology":
		return IndustryEdTech
	case "Higher Education":
		return IndustryHigherEducation
	case "Media":
		return IndustryMedia
	case "Entertainment":
		return IndustryEntertainment
	case "Publishing":
		return IndustryPublishing
	case "Advertising":
		return IndustryAdvertising
	case "Marketing":
		return IndustryMarketing
	case "Public Relations":
		return IndustryPublicRelations
	case "Hospitality":
		return IndustryHospitality
	case "Tourism":
		return IndustryTourism
	case "Food & Beverage":
		return IndustryFoodBeverage
	case "Energy & Utilities":
		return IndustryEnergyUtilities
	case "Oil & Gas":
		return IndustryOilGas
	case "Renewable Energy":
		return IndustryRenewableEnergy
	case "Environmental":
		return IndustryEnvironmental
	case "Transportation & Logistics":
		return IndustryTransportationLogistics
	case "Agriculture":
		return IndustryAgriculture
	case "Government":
		return IndustryGovernment
	case "Non-Profit":
		return IndustryNonProfit
	case "Legal":
		return IndustryLegal
	case "Consulting":
		return IndustryConsulting
	case "Human Resources":
		return IndustryHumanResources
	case "Security":
		return IndustrySecurity
	case "Cybersecurity":
		return IndustryCybersecurity
	default:
		return IndustryUnspecified
	}
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
