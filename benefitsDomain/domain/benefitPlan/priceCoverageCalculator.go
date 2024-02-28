package benefitPlan

import (
	"benefitsDomain/datatypes"
	"benefitsDomain/domain/person"
	"benefitsDomain/domain/person/personRoles"
)

type PriceCoverageCalculatorContext struct {
	EffectiveDate        datatypes.YYYYMMDD_Date
	Person               *person.Person
	Worker               *personRoles.Worker
	Participant          *personRoles.Participant
	CalculatorParameters map[string]interface{}
	//PersonEnrollment     *businessProcess.PersonBusinessProcess
}
type PriceCoverageCalculator interface {
	Calculate(ec *BenefitPlanPolicyContext) (string interface{})
}
type PriceCoverageCalculatorFunc func(ec *BenefitPlanPolicyContext) (string interface{})

func (f PriceCoverageCalculatorFunc) Calculate(ec *BenefitPlanPolicyContext) (string interface{}) {
	b := f(ec)
	return b

}
func TimesPayCalculation(ctx PriceCoverageCalculatorContext) (datatypes.BigFloat, error) {
	//worker := ctx.Worker
	//pay, _ := worker.GetCurrentPay()
	pay, _ := datatypes.NewBigFloat("50000.00")
	sTimes := ctx.CalculatorParameters["Times"].(string)
	times, _ := datatypes.NewBigFloat(sTimes)
	var r datatypes.BigFloat
	r = r.Mul(pay, times)
	return r, nil
}
