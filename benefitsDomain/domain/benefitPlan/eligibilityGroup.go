package benefitPlan

type EligibilityGroup struct {
	InternalId           string        `json:"internalId" bson:"internalId"`
	EligibilityGroupId   string        `json:"eligibilityGroupId" bson:"eligibilityGroupId"`
	EligibilityGroupName string        `json:"eligibilityGroupName" bson:"eligibilityGroupName"`
	EligibilityPolicy    PolicyHandler `json:"eligibilityPolicy" bson:"eligibilityPolicy"`
	Benefits             []*Benefit    `json:"benefits" bson:"benefits"`
}

func (eg *EligibilityGroup) IsEligible() bool {
	// Execute Eligibility Policy
	return true
}
