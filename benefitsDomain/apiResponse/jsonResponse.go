package apiResponse

import "benefitsDomain/datatypes"

type PersonProfileViewResponse struct {
	InternalId        string            `json:"internalId" bson:"internalId"`
	ExternalId        string            `json:"externalId" bson:"externalId"`
	FirstName         string            `json:"firstName" bson:"firstName"`
	LastName          string            `json:"lastName" bson:"lastName"`
	BirthDate         string            `json:"birthDate" bson:"birthDate"`
	Workers           []Worker          `json:"workers" bson:"workers"`
	Participants      []Participant     `json:"participants" bson:"participants"`
	BusinessProcesses []BusinessProcess `json:"BusinessProcesses" bson:"businessProcesses"`
}
type Participant struct {
	InternalId                 string             `json:"internalId" bson:"internalId"`
	PersonId                   string             `json:"personId" bson:"personId"`
	BenefitId                  string             `json:"benefitId" bson:"benefitId"`
	CoverageStartDate          string             `json:"coverageStartDate" bson:"coverageStartDate"`
	CoverageEndDate            string             `json:"coverageEndDate" bson:"coverageEndDate"`
	PayrollReportingState      int                `json:"payrollReportingState" bson:"payrollReportingState"`
	CarrierReportingState      int                `json:"carrierReportingState" bson:"carrierReportingState"`
	ElectedBenefitPlanId       string             `json:"electedBenefitPlanId" bson:"electedBenefitPlanId"`
	ActualBenefitPlanId        string             `json:"actualBenefitPlanId" bson:"actualBenefitPlanId"`
	ElectedTierCoverageLevelId string             `json:"electedTierCoverageLevelId" bson:"electedTierCoverageLevelId"`
	ActualTierCoverageLevelId  string             `json:"actualTierCoverageLevelId" bson:"actualTierCoverageLevelId"`
	ActualCoverageAmount       datatypes.BigFloat `json:"actualCoverageAmount" bson:"actualCoverageAmount"`
	ElectedCoverageAmount      datatypes.BigFloat `json:"electedCoverageAmount" bson:"electedCoverageAmount"`
	EmployeePreTaxCost         datatypes.BigFloat `json:"employeePreTaxCost" bson:"employeePreTaxCost"`
	EmployerCost               datatypes.BigFloat `json:"employerCost" bson:"employerCost"`
	EmployeeAfterTaxCost       datatypes.BigFloat `json:"employeeAfterTaxCost" bson:"employeeAfterTaxCost"`
	EmployerSubsidy            datatypes.BigFloat `json:"employerSubsidy" bson:"employerSubsidy"`
	LifeImputedIncome          datatypes.BigFloat `json:"lifeImputedIncome" bson:"lifeImputedIncome"`
}
type BusinessProcess struct {
	InternalId                  string `json:"internalId" bson:"internalId"`
	ReferenceNumber             string `json:"referenceNumber" bson:"referenceNumber"`
	BusinessProcessDefinitionId string `json:"businessProcessDefinitionId" bson:"businessProcessDefinitionId"`
	EffectiveDate               string `json:"effectiveDate" bson:"effectiveDate"`
	State                       string `json:"state" bson:"state"`
}
type Worker struct {
	InternalId       string `json:"internalId" bson:"internalId"`
	WorkerId         string `json:"workerId" bson:"workerId"`
	Employer         string `json:"employer" bson:"employer"`
	Pay              string `json:"pay" bson:"pay"`
	EmploymentStatus string `json:"employmentStatus" bson:"employmentStatus"`
}
