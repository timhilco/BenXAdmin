package benefitPlan

import (
	"benefitsDomain/datatypes"
)

type BenefitPlanCalculator interface {
	CalculateCoverage(BenefitPlanCalculationContext) BenefitPlanCoverageCalculationResults
}
type BenefitPlanCalculationContext struct {
	EffectiveDate datatypes.YYYYMMDD_Date
	//PersonRoles   *personRoles.PersonRoles
}

type BenefitPlanCoverageCalculationResults interface {
}

func (b *CoverageBenefitPlan) CalculateCoverage(BenefitPlanCalculationContext) BenefitPlanCoverageCalculationResults {
	return nil

}
