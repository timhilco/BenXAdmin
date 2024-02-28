package commands

import (
	"benefitsDomain/datatypes"
	"benefitsDomain/domain/benefitPlan"
	"fmt"

	"os"

	"log/slog"

	"github.com/xuri/excelize/v2"
)

func LoadBenefitFromSpreadsheet(fileName string) (string, []*benefitPlan.Benefit, error) {

	err := os.Chdir("../..")
	if err != nil {
		slog.Error("Error: ", err)
	}

	s, _ := os.Getwd()
	fmt.Println(s)
	f, err := excelize.OpenFile(fileName)
	if err != nil {
		slog.Error("Error: ", err)
		return "", nil, err
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			slog.Error("Error: ", err)
			return
		}
	}()

	// Get all the rows in the Sheet1.
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		slog.Error("Error: ", err)
		return "", nil, err
	}
	const (
		C_STATE_BENEFIT = iota
		C_STATE_BENEFIT_PLAN
		C_STATE_RATE_TABLE
		C_STATE_RATE_TIERS_TABLE
		C_STATE_HEADER
		C_STATE_DETAILS
		C_NOT_APPLICABLE
	)
	var effectiveDate string = ""
	benefits := make([]*benefitPlan.Benefit, 0)
	var aCoverageBenefitPlan *benefitPlan.CoverageBenefitPlan
	var aContributionBenefitPlan *benefitPlan.ContributionBenefitPlan
	parentState := C_STATE_HEADER
	gen1State := C_NOT_APPLICABLE
	var b *benefitPlan.Benefit
	var isHeaderNext bool = false
	var coverageType string
	var coverage benefitPlan.CoverageLevel
	var rateTable benefitPlan.SimpleHealthCareRateTable
	var rateTiersTable []benefitPlan.CoverageTier
	var benefitType int = 0
	for _, row := range rows {
		colA := row[0]
		var colB string
		if len(row) > 1 {
			colB = row[1]
		}
		fmt.Printf("%s <> %s\n", colA, colB)
		switch parentState {
		case C_STATE_HEADER:
			switch colA {
			case "<Effective Date>":
				effectiveDate = colB
			case "<Benefit>":
				b = &benefitPlan.Benefit{
					CoverageBenefitPlans: make([]*benefitPlan.CoverageBenefitPlan, 0),
					InternalId:           datatypes.GetGlobalInternalIdentifier(),
				}
				//benefits = append(benefits, b)
				b.BenefitName = colB
				parentState = C_STATE_BENEFIT
			default:
				fmt.Printf("%v", row)
			}
		case C_STATE_BENEFIT:
			switch colA {
			case "<Benefit Id>":
				b.BenefitId = colB
			case "<Benefit Type>":
				b.BenefitType = colB
				benefitType = benefitPlan.DetermineBenefitType(colB)
			case "<Coverage Type>":
				//coverageType = colB
			case "<Benefit Plan>":
				switch benefitType {
				case benefitPlan.C_BENEFIT_TYPE_COVERED_PEOPLE,
					benefitPlan.C_BENEFIT_TYPE_COVERED_AMOUNT:
					aCoverageBenefitPlan = &benefitPlan.CoverageBenefitPlan{
						BenefitPlanName: colB,
						InternalId:      datatypes.GetGlobalInternalIdentifier(),
					}
					b.CoverageBenefitPlans = append(b.CoverageBenefitPlans, aCoverageBenefitPlan)
				case benefitPlan.C_BENEFIT_TYPE_CONTRIBUTION:
					aContributionBenefitPlan = &benefitPlan.ContributionBenefitPlan{}
					b.ContributionPlan = *aContributionBenefitPlan
				case benefitPlan.C_BENEFIT_TYPE_UNKNOWN:
					fmt.Printf("%s", "Error: Unknown Type")
				}
				parentState = C_STATE_BENEFIT_PLAN
				gen1State = C_NOT_APPLICABLE
			case "</Benefit>":
				parentState = C_STATE_HEADER
				gen1State = C_NOT_APPLICABLE
				benefits = append(benefits, b)
			default:
				fmt.Printf("%v", row)
			}

		case C_STATE_BENEFIT_PLAN:
			switch colA {
			case "<Benefit Plan Id>":
				aCoverageBenefitPlan.BenefitPlanId = colB
			case "<Rates>":
				parentState = C_STATE_BENEFIT_PLAN
				gen1State = C_STATE_RATE_TABLE
				coverageType = colB
				aCoverageBenefitPlan.CoverageType = coverageType
				switch coverageType {
				case "CoveredPeople":
					coverage = &benefitPlan.CoveredPeopleCoverageLevel{}
				case "CoveredAmount":
					coverage = &benefitPlan.CoveredAmountCoverageLevel{}
				}
				rateTable = benefitPlan.SimpleHealthCareRateTable{}
				isHeaderNext = true
			case "<Details>":
				gen1State = C_STATE_DETAILS
			case "</Details>":
				parentState = C_STATE_BENEFIT_PLAN
				gen1State = C_NOT_APPLICABLE
			case "</Rates>":
				coverage.SetRateTable(rateTable)
				coverage.SetCoverageTiers(rateTiersTable)
				aCoverageBenefitPlan.Coverage = coverage
			case "</Benefit Plan>":
				parentState = C_STATE_BENEFIT
				gen1State = C_NOT_APPLICABLE
			default: // gen1Switch
				switch gen1State {
				case C_STATE_RATE_TABLE:
					if colA == "<Rates-Tiers>" {
						gen1State = C_STATE_RATE_TIERS_TABLE
						rateTiersTable = make([]benefitPlan.CoverageTier, 0)
						isHeaderNext = true
					} else {
						if isHeaderNext {
							headers := make([]string, 0)
							for i := 0; i < 3; i++ {
								headers = append(headers, row[i])
							}
							rateTable.ColumnHeaders = headers
							isHeaderNext = false
						} else {
							ee, _ := datatypes.NewBigFloat(colB)
							er, _ := datatypes.NewBigFloat(row[2])
							entry := benefitPlan.SimpleHealthCareRateEntry{
								CoverageCategory:    colA,
								EmployeeMonthlyRate: ee,
								EmployerMonthlyRate: er,
							}
							t := rateTable.SimpleHealthCareRateTableRows
							a := append(t, entry)
							rateTable.SimpleHealthCareRateTableRows = a
						}
					}
				case C_STATE_RATE_TIERS_TABLE:
					if isHeaderNext {
						/*
							headers := make([]string, 0)
							for i := 0; i < 3; i++ {
								headers = append(headers, row[i])
							}
							rateTable.ColumnHeaders = headers
						*/
						isHeaderNext = false
					} else {
						ct := benefitPlan.CoverageTier{
							CoverageDescription: colA,
							CoverageFormula:     colB,
						}
						rateTiersTable = append(rateTiersTable, ct)
					}
				case C_STATE_DETAILS:
					switch colA {
					case "Minimum":
						bf, _ := datatypes.NewBigFloat(colB)
						aContributionBenefitPlan.MinimumContribution = bf
					case "Maximum":
						bf, _ := datatypes.NewBigFloat(colB)
						aContributionBenefitPlan.MaximumContribution = bf
					}
				default:
					fmt.Printf("Default: %v\n", row)
				}
			}

		}
	}
	//fmt.Printf("Benefit: %v\n", b)

	return effectiveDate, benefits, nil

}
