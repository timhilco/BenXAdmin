package businessProcess

/*
func PersonEnrollInBenefit(ctx context.Context, processingContext *domain.ResourceContext, person *domain.Person, benefit *domain.Benefit, instruction *EnrollBenefitInstruction) (string, error) {
	// Create Participant((
	slog.Info("Person Enroll In Benefit :", person.LastName)
	participant := domain.Participant{}
	//participant.InternalId = domain.GetGlobalInternalIdentifier()
	participant.InternalId = person.InternalId + "-" + benefit.BenefitId
	participant.PersonId = person.InternalId
	participant.BenefitId = benefit.InternalId
	// Update Participant based on Instructions
	participant.InitialEnrollmentDate = instruction.EffectiveDate
	coverageHistory := domain.CoverageHistory{
		CoveragePeriods:     make([]domain.CoveragePeriod, 0),
		ContributionPeriods: nil,
	}
	participant.CoverageHistory = coverageHistory

	// Insert Participant into Database
	ds := processingContext.GetDataStore()
	err := ds.InsertParticipant(&participant)

	return "OK", err
}
func PersonEnrollInBenefitPlan(ctx context.Context, processingContext *domain.ResourceContext, participant *domain.Participant, instructions EnrollBenefitPlanInstruction) (string, error) {
	slog.Info("Person Enroll In Benefit Plan:", participant.PersonId)
	coveragePeriod := domain.CoveragePeriod{
		CoverageStartDate:        string(instructions.EffectiveDate),
		CoverageEndDate:          string(instructions.CoverageEndDate),
		ElectedBenefitOfferingId: instructions.ElectedBenefitOfferingId,
		ActualBenefitOfferingId:  instructions.ActualBenefitOfferingId,
		ActualCoverageAmount:     instructions.ActualCoverageAmount,
		ElectedCoverageAmount:    instructions.ElectedCoverageAmount,
		EmployeePreTaxCost:       instructions.EmployeePreTaxCost,
		LifeImputedIncome:        instructions.LifeImputedIncome,
	}
	participant.AddCoveragePeriodToCoverageHistory(coveragePeriod, "Insert")
	err := processingContext.GetDataStore().UpdateParticipant(participant.InternalId, participant)

	return "OK", err
}
*/
