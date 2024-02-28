package commandDataStructures

type OpenEnrollmentElectionRequest struct {
	BenefitPlanElections []OpenEnrollmentElection `json:"benefitPlanElections"`
}
type OpenEnrollmentElection struct {
	BenefitId       string `json:"benefitId"`
	BenefitPlanId   string `json:"benefitPlanId"`
	CoverageLevelId string `json:"coverageLevelId"`
	CoverageAmount  string `json:"coverageAmount"`
}

type PayrollRelease struct {
	BenefitId string `json:"benefitId" bson:"benefitId"`
	PayrollId string `json:"payrollId" bson:"payrollId"`
}
type CarrierRelease struct {
	BenefitId string `json:"benefitId" bson:"benefitId"`
	CarrierId string `json:"carrierId" bson:"carrierId"`
}
