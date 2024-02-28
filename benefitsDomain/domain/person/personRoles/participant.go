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

type CoveragePeriod struct {
	CoverageStartDate        string             `json:"coverageStartDate" bson:"coverageStartDate"`
	CoverageEndDate          string             `json:"coverageEndDate" bson:"coverageEndDate"`
	PayrollReportingState    int                `json:"payrollReportingState" bson:"payrollReportingState"`
	CarrierReportingState    int                `json:"carrierReportingState" bson:"carrierReportingState"`
	ElectedBenefitOfferingId string             `json:"electedBenefitOfferingId" bson:"electedBenefitOfferingId"`
	ActualBenefitOfferingId  string             `json:"actualBenefitOfferingId" bson:"actualBenefitOfferingId"`
	ElectedCoverageLevel     string             `json:"electedCoverageLevel" bson:"electedCoverageLevel"`
	ActualCoverageLevel      string             `json:"actualCoverageLevel" bson:"actualCoverageLevel"`
	OfferedBenefitOfferingId string             `json:"offeredBenefitOfferingId" bson:"offeredBenefitOfferingId"`
	ActualCoverageAmount     datatypes.BigFloat `json:"actualCoverageAmount" bson:"actualCoverageAmount"`
	ElectedCoverageAmount    datatypes.BigFloat `json:"electedCoverageAmount" bson:"electedCoverageAmount"`
	EmployeePreTaxCost       datatypes.BigFloat `json:"employeePreTaxCost" bson:"employeePreTaxCost"`
	EmployerCost             datatypes.BigFloat `json:"employerCost" bson:"employerCost"`
	EmployeeAfterTax         datatypes.BigFloat `json:"employeeAfterTax" bson:"employeeAfterTax"`
	EmployerSubsidy          datatypes.BigFloat `json:"employerSubsidy" bson:"employerSubsidy"`
	LifeImputedIncome        datatypes.BigFloat `json:"lifeImputedIncome" bson:"lifeImputedIncome"`
}
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
func (p *Participant) ApplyEnrollmentElections(effectiveDate datatypes.YYYYMMDD_Date, election commandDataStructures.OpenEnrollmentElection) error {
	// Validate Election
	// Close any current Period
	// Get current
	//ch := p.CoverageHistory
	b, _ := datatypes.NewBigFloat(election.CoverageAmount)
	coveragePeriod := CoveragePeriod{
		CoverageStartDate:        string(effectiveDate),
		PayrollReportingState:    C_STATE_UNREPORTED,
		CarrierReportingState:    C_STATE_UNREPORTED,
		ElectedBenefitOfferingId: election.BenefitPlanId,
		ElectedCoverageLevel:     election.CoverageLevelId,
		ElectedCoverageAmount:    b,
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
