package commandDataStructures

type OpenEnrollmentElectionRequest struct {
	BenefitPlanElections []EnrollmentElection `json:"benefitPlanElections"`
}
type EnrollmentElection struct {
	BenefitId          string `json:"benefitId"`
	BenefitPlanId      string `json:"benefitPlanId"`
	CoverageLevelId    string `json:"coverageLevelId"`
	CoverageAmount     string `json:"coverageAmount"`
	ContributionAmount string `json:"contributionAmount"`
}

type PayrollRelease struct {
	BenefitId string `json:"benefitId" bson:"benefitId"`
	PayrollId string `json:"payrollId" bson:"payrollId"`
}
type CarrierRelease struct {
	BenefitId string `json:"benefitId" bson:"benefitId"`
	CarrierId string `json:"carrierId" bson:"carrierId"`
}
