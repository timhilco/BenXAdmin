package db

import "benefitsDomain/domain/benefitPlan"

func CreateMockEligibilityGroupObjects(id string, planDB *PlanMongoDB) *benefitPlan.EligibilityGroup {
	switch id {
	case "EG001":
		return CreateActiveEligibilityGroup(planDB)

	}
	return nil
}
func CreateActiveEligibilityGroup(planDB *PlanMongoDB) *benefitPlan.EligibilityGroup {
	benefits := make([]*benefitPlan.Benefit, 0)
	b, _ := planDB.GetBenefit("B001")
	benefits = append(benefits, b)
	b, _ = planDB.GetBenefit("B002")
	benefits = append(benefits, b)
	b, _ = planDB.GetBenefit("B003")
	benefits = append(benefits, b)
	b, _ = planDB.GetBenefit("B004")
	benefits = append(benefits, b)
	b, _ = planDB.GetBenefit("B005")
	benefits = append(benefits, b)
	eg := &benefitPlan.EligibilityGroup{
		InternalId:           "EG001",
		EligibilityGroupId:   "EG001",
		EligibilityGroupName: "Active Program",
		Benefits:             benefits,
	}
	return eg
}
func CreateMockBenefitObjects(id string) *benefitPlan.Benefit {
	switch id {
	case "B001":
		return CreateMedicalBenefit()
	case "B002":
		return CreateDentalBenefit()
	}
	return nil
}
func CreateMedicalBenefit() *benefitPlan.Benefit {
	benefit := benefitPlan.Benefit{
		InternalId:  "B001",
		BenefitId:   "B001",
		BenefitName: "Active Medical",
		BenefitType: "Medical",
	}
	plans := make([]*benefitPlan.CoverageBenefitPlan, 0)
	plan1 := &benefitPlan.CoverageBenefitPlan{
		InternalId:      "B001-BP001",
		BenefitPlanId:   "Medical_BP001",
		BenefitPlanName: "Aetna Gold Plan",
		//	Coverage:        "You + 1000 Deductible",
		//Rate:            "101.00",
		Provider: "Aetna",
	}
	plans = append(plans, plan1)
	plan2 := &benefitPlan.CoverageBenefitPlan{
		InternalId:      "B001-BP002",
		BenefitPlanId:   "Medical_BP002",
		BenefitPlanName: "BC/BS Gold Plan",
		//	Coverage:        "You + 1000 Deductible",
		//Rate:            "102.00",
		Provider: "BC/BS NC",
	}
	plans = append(plans, plan2)
	benefit.CoverageBenefitPlans = plans
	return &benefit

}
func CreateDentalBenefit() *benefitPlan.Benefit {
	benefit := benefitPlan.Benefit{
		InternalId:  "B002",
		BenefitId:   "B002",
		BenefitName: "Active Dental",
		BenefitType: "Dental",
	}
	plans := make([]*benefitPlan.CoverageBenefitPlan, 0)
	plan1 := &benefitPlan.CoverageBenefitPlan{
		InternalId:      "B002-BP001",
		BenefitPlanId:   "Dental_BP001",
		BenefitPlanName: "Aetna Gold Plan",
		//Coverage:        "You + 1000 Deductible",
		//Rate:            "201.00",
		Provider: "Aetna",
	}
	plans = append(plans, plan1)
	plan2 := &benefitPlan.CoverageBenefitPlan{
		InternalId:      "B002-BP002",
		BenefitPlanId:   "Dental_BP002",
		BenefitPlanName: "BC/BS Gold Plan",
		//Coverage:        "You + 1000 Deductible",
		//Rate:            "202.00",
		Provider: "BC/BS NC",
	}
	plans = append(plans, plan2)
	benefit.CoverageBenefitPlans = plans
	return &benefit

}
func GetEligibilityGroups(planDB *PlanMongoDB) []*benefitPlan.EligibilityGroup {
	egs := make([]*benefitPlan.EligibilityGroup, 0)
	eg := CreateMockEligibilityGroupObjects("EG001", planDB)
	egs = append(egs, eg)
	return egs
}
