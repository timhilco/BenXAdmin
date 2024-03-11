package benefitPlan

import (
	"benefitsDomain/datatypes"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
)

type Benefit struct {
	InternalId           string                  `json:"internalId" bson:"internalId"`
	BenefitId            string                  `json:"benefitId" bson:"benefitId"`
	BenefitName          string                  `json:"benefitName" bson:"benefitName"`
	BenefitType          string                  `json:"benefitType" bson:"benefitType"`
	CoverageBenefitPlans []*CoverageBenefitPlan  `json:"coverageBenefitPlans" bson:"coverageBenefitPlans"`
	ContributionPlan     ContributionBenefitPlan `json:"contributionBenefitPlan" bson:"contributionBenefitPlan"`
}

const (
	C_BENEFIT_TYPE_UNKNOWN = iota
	C_BENEFIT_TYPE_COVERED_PEOPLE
	C_BENEFIT_TYPE_COVERED_AMOUNT
	C_BENEFIT_TYPE_CONTRIBUTION
)
const C_COVERED_AMOUNT string = "CoveredAmount"
const C_COVERED_PEOPLE string = "CoveredPeople"

type BenefitPlan interface {
}
type CoverageBenefitPlan struct {
	InternalId      string `json:"internalId" bson:"internalId"`
	BenefitPlanId   string `json:"benefitPlanId" bson:"benefitPlanId"`
	BenefitPlanName string `json:"benefitPlanName" bson:"benefitPlanName"`
	CoverageType    string `json:"coverageType" bson:"coverageType"`
	Coverage        CoverageLevel
	Provider        string `json:"provider" bson:"provider"`
}
type ContributionBenefitPlan struct {
	InternalId          string             `json:"internalId" bson:"internalId"`
	BenefitPlanId       string             `json:"benefitPlanId" bson:"benefitPlanId"`
	BenefitPlanName     string             `json:"benefitPlanName" bson:"benefitPlanName"`
	MinimumContribution datatypes.BigFloat `json:"minimumContribution" bson:"minimumContribution"`
	MaximumContribution datatypes.BigFloat `json:"maximumContribution" bson:"maximumContribution"`
	Administrator       string             `json:"administrator" bson:"administrator"`
}

type BenefitPlanBSON struct {
	InternalId      string `json:"internalId" bson:"internalId"`
	BenefitPlanId   string `json:"benefitPlanId" bson:"benefitPlanId"`
	BenefitPlanName string `json:"benefitPlanName" bson:"benefitPlanName"`
	CoverageType    string `json:"coverageType" bson:"coverageType"`
	Coverage        bson.Raw
	Provider        string `json:"provider" bson:"provider"`
}
type CoverageLevel interface {
	SetRateTable(rt SimpleHealthCareRateTable)
	SetCoverageTiers(ct []CoverageTier)
	GetRateTable() SimpleHealthCareRateTable
	GetCoverageTiers() []CoverageTier
	//json.Marshaler
	//bson.Marshaler
	//json.Unmarshaler
	//bson.Unmarshaler

}

type CoveredPeopleCoverageLevel struct {
	RateTable SimpleHealthCareRateTable `json:"rateTable" bson:"rateTable"`
}

func (c *CoveredPeopleCoverageLevel) SetRateTable(rt SimpleHealthCareRateTable) {
	c.RateTable = rt
}

func (c *CoveredPeopleCoverageLevel) SetCoverageTiers(ct []CoverageTier) {
	// Do nothing
}
func (c *CoveredPeopleCoverageLevel) GetRateTable() SimpleHealthCareRateTable {
	return c.RateTable
}

func (c CoveredPeopleCoverageLevel) GetCoverageTiers() []CoverageTier {
	return nil
}

type CoveredAmountCoverageLevel struct {
	CoverageTiers []CoverageTier            `json:"coverageTiers" bson:"coverageTiers"`
	RateTable     SimpleHealthCareRateTable `json:"rateTable" bson:"rateTable"`
}

func (c *CoveredAmountCoverageLevel) SetRateTable(rt SimpleHealthCareRateTable) {
	c.RateTable = rt
}
func (c *CoveredAmountCoverageLevel) SetCoverageTiers(ct []CoverageTier) {
	c.CoverageTiers = ct
}
func (c *CoveredAmountCoverageLevel) GetRateTable() SimpleHealthCareRateTable {
	return c.RateTable
}

func (c *CoveredAmountCoverageLevel) GetCoverageTiers() []CoverageTier {
	return c.CoverageTiers
}

type CoverageTier struct {
	CoverageDescription string `json:"coverageDescription" bson:"coverageDescription"`
	CoverageFormula     string `json:"coverageFormula" bson:"coverageFormula"`
}

func (b *Benefit) GetBenefitPlans() []*CoverageBenefitPlan {
	return b.CoverageBenefitPlans
}
func (b *Benefit) DetermineCoverageStartDate(enrollmentEffectiveDate datatypes.YYYYMMDD_Date) datatypes.YYYYMMDD_Date {
	return enrollmentEffectiveDate
}
func (b *Benefit) DetermineCoverageEndDate(enrollmentEffectiveDate datatypes.YYYYMMDD_Date) datatypes.YYYYMMDD_Date {
	s := "20250101"
	return datatypes.YYYYMMDD_Date(s)
}

func (bp *CoverageBenefitPlan) GetRateTable() SimpleHealthCareRateTable {
	return bp.Coverage.GetRateTable()
}

func (bp *CoverageBenefitPlan) UnmarshalBSON(data []byte) error {
	// Unmarshall only the type
	bpTemp := BenefitPlanBSON{}
	if err := bson.Unmarshal(data, &bpTemp); err != nil {
		return err
	}

	// Set the type to the prop
	switch bpTemp.CoverageType {
	case C_COVERED_PEOPLE:
		p := &CoveredPeopleCoverageLevel{}
		err := bson.Unmarshal(bpTemp.Coverage, p)
		if err != nil {
			return err
		}
		bp.Coverage = p
	case C_COVERED_AMOUNT:
		a := &CoveredAmountCoverageLevel{}
		err := bson.Unmarshal(bpTemp.Coverage, a)
		if err != nil {
			return err
		}
		bp.Coverage = a
	default:
		slog.Info("DEFAULT")
		return fmt.Errorf("unknown type: %v", bpTemp.CoverageType)
	}
	bp.InternalId = bpTemp.InternalId
	bp.BenefitPlanId = bpTemp.BenefitPlanId
	bp.BenefitPlanName = bpTemp.BenefitPlanName
	bp.CoverageType = bpTemp.CoverageType
	bp.Provider = bpTemp.Provider
	return nil
}
func (bp *CoverageBenefitPlan) MarshalBSON() ([]byte, error) {
	// Unmarshall only the type
	type BenefitPlanBSONAlias CoverageBenefitPlan
	bpTemp := BenefitPlanBSON{}

	fmt.Println(bp.CoverageType)
	var b []byte
	var err error

	coverage := bp.Coverage
	switch bp.CoverageType {
	case C_COVERED_AMOUNT:
		x := coverage.(*CoveredAmountCoverageLevel)
		b, err = bson.Marshal(x)
	case C_COVERED_PEOPLE:
		x := coverage.(*CoveredPeopleCoverageLevel)
		b, err = bson.Marshal(x)
	}
	if err != nil {
		return nil, err
	}
	bpTemp.InternalId = bp.InternalId
	bpTemp.BenefitPlanId = bp.BenefitPlanId
	bpTemp.BenefitPlanName = bp.BenefitPlanName
	bpTemp.CoverageType = bp.CoverageType
	bpTemp.Provider = bp.Provider
	bpTemp.Coverage = b
	b2, err := bson.Marshal((*BenefitPlanBSONAlias)(bp))
	//b2, err := bson.Marshal(bp)

	return b2, err
}
func DetermineBenefitType(name string) int {
	switch name {
	case "Medical",
		"Dental":
		return C_BENEFIT_TYPE_COVERED_PEOPLE
	case "EELife":
		return C_BENEFIT_TYPE_COVERED_AMOUNT
	case "FSA Health Care",
		"FSA Dependent":
		return C_BENEFIT_TYPE_CONTRIBUTION
	}
	return C_BENEFIT_TYPE_UNKNOWN
}
