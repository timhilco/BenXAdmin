package businessProcess

import "benefitsDomain/datatypes"

type EnrollmentInstruction interface {
	GetEffectiveDate() datatypes.YYYYMMDD_Date
}
type EnrollBenefitInstruction struct {
	EffectiveDate datatypes.YYYYMMDD_Date
}
type EnrollBenefitPlanInstruction struct {
	EffectiveDate            datatypes.YYYYMMDD_Date
	CoverageEndDate          datatypes.YYYYMMDD_Date
	ElectedBenefitOfferingId string `json:"electedBenefitOfferingId" bson:"electedBenefitOfferingId"`
	ActualBenefitOfferingId  string `json:"actualBenefitOfferingId" bson:"actualBenefitOfferingId"`
	OfferedBenefitOfferingId string `json:"offeredBenefitOfferingId" bson:"offeredBenefitOfferingId"`
	ActualCoverageAmount     string `json:"actualCoverageAmount" bson:"actualCoverageAmount"`
	ElectedCoverageAmount    string `json:"electedCoverageAmount" bson:"electedCoverageAmount"`
	EmployeePreTaxCost       string `json:"employeePreTaxCost" bson:"employeePreTaxCost"`
	EmployerCost             string `json:"employerCost" bson:"employerCost"`
	EmployeeAfterTax         string `json:"employeeAfterTax" bson:"employeeAfterTax"`
	EmployerSubsidy          string `json:"employerSubsidy" bson:"employerSubsidy"`
	LifeImputedIncome        string `json:"lifeImputedIncome" bson:"lifeImputedIncome"`
}
