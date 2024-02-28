package benefitPlan

type BenefitPlanPolicyContext struct {
	//PersonRoles *personRoles.PersonRoles
}
type PolicyHandler interface {
	EvaluatePolicy(ec *BenefitPlanPolicyContext) bool
}
type PolicyHandlerFunc func(ec *BenefitPlanPolicyContext) bool

func (f PolicyHandlerFunc) EvaluatePolicy(ec *BenefitPlanPolicyContext) bool {
	b := f(ec)
	return b

}
