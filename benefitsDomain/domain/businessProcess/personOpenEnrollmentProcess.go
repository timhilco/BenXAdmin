package businessProcess

import (
	"benefitsDomain/datatypes"
	"benefitsDomain/domain/benefitPlan"
	"benefitsDomain/domain/db"
	"strings"
)

type EnrollmentBenefitOptionRates struct {
	BenefitPlanOptionsCollection []BenefitPlanOptions `json:"benefitPlanOptionsCollection" bson:"benefitPlanOptionsCollection"`
}
type BenefitPlanOptions struct {
	BenefitId string              `json:"benefitId" bson:"benefitId"`
	Options   []BenefitPlanOption `json:"options" bson:"options"`
}
type BenefitPlanOption struct {
	BenefitPlanId      string              `json:"benefitPlanId" bson:"benefitPlanId"`
	CoverageType       string              `json:"coverageType" bson:"coverageType"`
	TierCoverageLevels []TierCoverageLevel `json:"tierCoverageLevels" bson:"tierCoverageLevels"`
}
type TierCoverageLevel struct {
	ElectionId          string             `json:"electionId" bson:"electionId"`
	CoverageLevel       string             `json:"coverageLevel" bson:"coverageLevel"`
	CoverageAmount      datatypes.BigFloat `json:"coverageAmount" bson:"coverageAmount"`
	EmployeeMonthlyRate datatypes.BigFloat `json:"employeeMonthlyRate" bson:"employeeMonthlyRate"`
	EmployerMonthlyRate datatypes.BigFloat `json:"employerMonthlyRate" bson:"employerMonthlyRate"`
}
type OpenEnrollmentElectionRequest struct {
	BenefitPlanElections []OpenEnrollmentElectionCommand `json:"benefitPlanElections"`
}
type OpenEnrollmentElectionCommand struct {
	BenefitId       string `json:"benefitId"`
	BenefitPlanId   string `json:"benefitPlanId"`
	CoverageLevelId string `json:"coverageLevelId"`
	CoverageAmount  string `json:"coverageAmount"`
}

func (e *EnrollmentBenefitOptionRates) Report(rc *ResourceContext) string {
	var sb strings.Builder
	for _, b := range e.BenefitPlanOptionsCollection {
		sb.WriteString("-BenefitPlanOptions\n")
		sb.WriteString(b.Report(rc))
	}
	return sb.String()
}
func (b *BenefitPlanOptions) Report(rc *ResourceContext) string {
	var sb strings.Builder
	sb.WriteString("Benefit Id :" + b.BenefitId + "\n")
	for _, o := range b.Options {
		sb.WriteString("-BenefitPlanOption\n")
		sb.WriteString(o.Report(rc))
	}
	return sb.String()
}
func (b *BenefitPlanOption) Report(rc *ResourceContext) string {
	var sb strings.Builder
	sb.WriteString("Benefit Plan Id :" + b.BenefitPlanId + "\n")
	sb.WriteString("Coverage Type :" + b.CoverageType + "\n")
	sb.WriteString("-TierCoverageLevel\n")
	for _, o := range b.TierCoverageLevels {
		sb.WriteString(o.Report(rc))
	}
	return sb.String()
}
func (b *TierCoverageLevel) Report(rc *ResourceContext) string {
	var sb strings.Builder
	sb.WriteString(b.ElectionId)
	sb.WriteString(" ")
	sb.WriteString(b.CoverageLevel)
	sb.WriteString(" ")
	sb.WriteString(b.CoverageAmount.FormattedString(""))
	sb.WriteString(" ")
	sb.WriteString(b.EmployeeMonthlyRate.FormattedString(""))
	sb.WriteString(" ")
	sb.WriteString(b.EmployerMonthlyRate.FormattedString(""))
	sb.WriteString("\n")

	return sb.String()
}
func (pbp *PersonBusinessProcess) buildOpenEnrollmentBusinessProcessData(mpc *MessageProcessingContext) EnrollmentBenefitOptionRates {
	bpos := make([]BenefitPlanOptions, 0)
	egs := db.GetEligibilityGroups(mpc.ResourceContext.planDataStore)
	for _, eligibilityGroup := range egs {
		if eligibilityGroup.IsEligible() {
			for _, benefit := range eligibilityGroup.Benefits {
				bpo := pbp.buildBenefitPlanOptions(mpc, benefit)
				bpos = append(bpos, bpo)
			}
		}

	}
	eor := EnrollmentBenefitOptionRates{

		BenefitPlanOptionsCollection: bpos,
	}
	return eor
}
func (pbp *PersonBusinessProcess) buildBenefitPlanOptions(mpc *MessageProcessingContext, benefit *benefitPlan.Benefit) BenefitPlanOptions {
	bpos := BenefitPlanOptions{
		BenefitId: benefit.BenefitId,
		Options:   make([]BenefitPlanOption, 0),
	}
	options := make([]BenefitPlanOption, 0)
	ctx := benefitPlan.PriceCoverageCalculatorContext{
		Person:        mpc.Person,
		Worker:        mpc.Worker,
		EffectiveDate: pbp.EffectiveDate,
	}
	op := benefit.ComputeOptionsPricing(ctx)
	for _, plan := range op {
		option := BenefitPlanOption{}
		option.BenefitPlanId = plan.BenefitPlanId
		option.CoverageType = plan.ResultType
		tiers := make([]TierCoverageLevel, 0)
		for _, tier := range plan.Tiers {
			t := TierCoverageLevel{
				ElectionId:          tier.ElectionId,
				CoverageLevel:       tier.PeopleCovered,
				CoverageAmount:      tier.CoverageAmount,
				EmployeeMonthlyRate: tier.EmployeeMonthlyRate,
				EmployerMonthlyRate: tier.EmployerMonthlyRate,
			}
			tiers = append(tiers, t)
		}
		option.TierCoverageLevels = tiers
		options = append(options, option)

	}
	bpos.Options = options
	return bpos

}
