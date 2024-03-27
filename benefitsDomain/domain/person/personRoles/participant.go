package personRoles

import (
	"benefitsDomain/datatypes"
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"os"
	"server/message/commandDataStructures"
)

type Participant struct {
	InternalId            string                  `json:"internalId" bson:"internalId"`
	PersonId              string                  `json:"personId" bson:"personId"`
	BenefitId             string                  `json:"benefitId" bson:"benefitId"`
	InitialEnrollmentDate datatypes.YYYYMMDD_Date `json:"initialEnrollmentDate" bson:"initialEnrollmentDate"`
	CoverageHistory       CoverageHistory         `json:"coverageHistory" bson:"coverageHistory"`
}
type CoverageHistory struct {
	CoveragePeriods     []CoveragePeriod     `json:"coveragePeriods" bson:"coveragePeriods,omitempty"`
	ContributionPeriods []ContributionPeriod `json:"contributionPeriods" bson:"contributionPeriods,omitempty"`
}

const (
	C_STATE_UNREPORTED = iota
	C_STATE_REPORTED
)

type BenefitPeriod interface {
}
type CoveragePeriod struct {
	CoverageStartDate          datatypes.YYYYMMDD_Date `json:"coverageStartDate" bson:"coverageStartDate"`
	CoverageEndDate            datatypes.YYYYMMDD_Date `json:"coverageEndDate" bson:"coverageEndDate"`
	PayrollReportingState      int                     `json:"payrollReportingState" bson:"payrollReportingState"`
	CarrierReportingState      int                     `json:"carrierReportingState" bson:"carrierReportingState"`
	ElectedBenefitPlanId       string                  `json:"electedBenefitPlanId" bson:"electedBenefitPlanId"`
	ActualBenefitPlanId        string                  `json:"actualBenefitPlanId" bson:"actualBenefitPlanId"`
	ElectedTierCoverageLevelId string                  `json:"electedTierCoverageLevelId" bson:"electedTierCoverageLevelId"`
	ActualTierCoverageLevelId  string                  `json:"actualTierCoverageLevelId" bson:"actualCoverageLevelId"`
	ActualCoverageAmount       datatypes.BigFloat      `json:"actualCoverageAmount" bson:"actualCoverageAmount"`
	ElectedCoverageAmount      datatypes.BigFloat      `json:"electedCoverageAmount" bson:"electedCoverageAmount"`
	//Both Types of Plans
	EmployeePreTaxCost   datatypes.BigFloat `json:"employeePreTaxCost" bson:"employeePreTaxCost"`
	EmployerCost         datatypes.BigFloat `json:"employerCost" bson:"employerCost"`
	EmployeeAfterTaxCost datatypes.BigFloat `json:"employeeAfterTaxCost" bson:"employeeAfterTaxCost"`
	EmployerSubsidy      datatypes.BigFloat `json:"employerSubsidy" bson:"employerSubsidy"`
	LifeImputedIncome    datatypes.BigFloat `json:"lifeImputedIncome" bson:"lifeImputedIncome"`
}

// Contribution Type Plans
type ContributionPeriod struct {
	CoverageStartDate  string `json:"coverageStartDate" bson:"coverageStartDate"`
	CoverageEndDate    string `json:"coverageEndDate" bson:"coverageEndDate"`
	ContributionAmount string `json:"contributionAmount" bson:"contributionAmount"`
}

func CreateJsonFromParticipant(participant *Participant) ([]byte, error) {
	// Convert struct to JSON
	jsonData, err := json.Marshal(participant)
	if err != nil {
		log.Fatal(err)
	}
	return jsonData, err
}
func CreateParticipantFromJsonFile(filename string) (*Participant, error) {

	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	participant := Participant{}

	err = json.Unmarshal(data, &participant) // here!
	if err != nil {
		panic(err)
	}
	return &participant, nil
}
func NewParticipant(personId string, benefitId string) (*Participant, error) {
	participant := Participant{
		InternalId: personId + "_" + benefitId,
		PersonId:   personId,
		BenefitId:  benefitId,
	}
	return &participant, nil
}

func (p *Participant) AddCoveragePeriodToCoverageHistory(cp CoveragePeriod, instruction string) {
	ch := p.CoverageHistory
	if ch.CoveragePeriods == nil {
		ch = CoverageHistory{
			CoveragePeriods: make([]CoveragePeriod, 0),
		}
	}
	ch.AddCoveragePeriod(cp, instruction)
	p.CoverageHistory = ch

}
func (ch *CoverageHistory) AddCoveragePeriod(cp CoveragePeriod, instruction string) {
	// Insert adds new entry in History
	cps := ch.CoveragePeriods
	if cps == nil {
		cps = make([]CoveragePeriod, 0)
	}
	cps = append(cps, cp)
	ch.CoveragePeriods = cps
}
func (p *Participant) ApplyNewCoveragePeriod(effectiveDate datatypes.YYYYMMDD_Date, newCoveragePeriod CoveragePeriod) error {
	p.AddCoveragePeriodToCoverageHistory(newCoveragePeriod, "")
	return nil
}
func (p *Participant) ApplyEnrollmentElections(effectiveDate datatypes.YYYYMMDD_Date, election commandDataStructures.EnrollmentElection) error {
	// Validate Election
	// Close any current Period
	// Get current
	//ch := p.CoverageHistory
	b, _ := datatypes.NewBigFloat(election.CoverageAmount)
	coveragePeriod := CoveragePeriod{
		CoverageStartDate:          effectiveDate,
		PayrollReportingState:      C_STATE_UNREPORTED,
		CarrierReportingState:      C_STATE_UNREPORTED,
		ElectedTierCoverageLevelId: election.TierCoverageLevelId,
		ElectedCoverageAmount:      b,
	}
	p.AddCoveragePeriodToCoverageHistory(coveragePeriod, "")
	return nil
}
func (p *Participant) Report(ev datatypes.EnvironmentVariables) string {

	dir := ev.TemplateDirectory
	templateFile := dir + "participantTemplate.tmpl"
	buf := new(bytes.Buffer)
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(buf, p)
	if err != nil {
		panic(err)
	}
	return buf.String()

}
