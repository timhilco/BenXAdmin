package benefitPlan

import (
	"benefitsDomain/datatypes"
	"fmt"
)

type BenefitPlanCalculationResult struct {
	BenefitPlanId string
	ResultType    string
	Tiers         []TierCalculationResult
}
type TierCalculationResult struct {
	ElectionId          string
	PeopleCovered       string
	CoverageAmount      datatypes.BigFloat
	EmployeeMonthlyRate datatypes.BigFloat
	EmployerMonthlyRate datatypes.BigFloat
}
type SimpleHealthCareRateEntry struct {
	CoverageCategory    string             `json:"coverageCategory" bson:"coverageCategory"`
	EmployeeMonthlyRate datatypes.BigFloat `json:"employeeMonthlyRate" bson:"employeeMonthlyRate"`
	EmployerMonthlyRate datatypes.BigFloat `json:"employerMonthlyRate" bson:"employerMonthlyRate"`
}
type SimpleHealthCareRateTable struct {
	ColumnHeaders                 []string                    `json:"columnHeaders" bson:"columnHeaders"`
	SimpleHealthCareRateTableRows []SimpleHealthCareRateEntry `json:"simpleHealthCareRateTableRows" bson:"simpleHealthCareRateTableRows"`
}

func (b *Benefit) ComputeOptionsPricing(ctx PriceCoverageCalculatorContext) []BenefitPlanCalculationResult {
	results := make([]BenefitPlanCalculationResult, 0)
	plans := b.GetBenefitPlans()
	for _, plan := range plans {
		e := plan.ComputeOptionsPricing(ctx)
		results = append(results, e)
	}
	return results
}
func (bp *CoverageBenefitPlan) ComputeOptionsPricing(ctx PriceCoverageCalculatorContext) BenefitPlanCalculationResult {
	tiers := make([]TierCalculationResult, 0)
	switch bp.CoverageType {
	case C_COVERED_PEOPLE:
		rateTable := bp.Coverage.GetRateTable()
		for i, row := range rateTable.SimpleHealthCareRateTableRows {

			f := row.EmployeeMonthlyRate
			f2 := row.EmployerMonthlyRate
			cov, _ := datatypes.NewBigFloat("0")
			e := TierCalculationResult{
				ElectionId:          fmt.Sprintf("%d", i),
				PeopleCovered:       row.CoverageCategory,
				CoverageAmount:      cov,
				EmployeeMonthlyRate: f,
				EmployerMonthlyRate: f2,
			}
			tiers = append(tiers, e)
		}

	case C_COVERED_AMOUNT:
		rateTable := bp.Coverage.GetRateTable()
		rate := rateTable.SimpleHealthCareRateTableRows[0].EmployeeMonthlyRate
		coverageTiers := bp.Coverage.GetCoverageTiers()
		for i, tier := range coverageTiers {
			formula := tier.CoverageFormula
			times := formula[0:1]
			parms := make(map[string]interface{})
			parms["Times"] = times
			ctx.CalculatorParameters = parms
			a, _ := TimesPayCalculation(ctx)
			var ee, er datatypes.BigFloat
			ee.Mul(a, rate)
			er.Mul(a, rate)
			e := TierCalculationResult{
				ElectionId:          fmt.Sprintf("%d", i+1),
				PeopleCovered:       "",
				CoverageAmount:      a,
				EmployeeMonthlyRate: ee,
				EmployerMonthlyRate: er,
			}
			tiers = append(tiers, e)
		}
	}
	r := BenefitPlanCalculationResult{
		BenefitPlanId: bp.BenefitPlanId,
		ResultType:    bp.CoverageType,
		Tiers:         tiers,
	}
	return r
}
