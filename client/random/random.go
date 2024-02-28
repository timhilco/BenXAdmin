package random

import (
	"benefitsDomain/datatypes"
	"benefitsDomain/domain/person"
	"benefitsDomain/domain/person/personRoles"
	"fmt"
	"math/rand"
)

func CreateRandomPerson(options map[string]interface{}) *person.Person {
	person := person.Person{}

	min := 1000
	max := 30000

	id := rand.Intn(max-min+1) + min

	internalId := fmt.Sprintf("%d", id)
	person.InternalId = internalId
	person.LastName = "Sample_" + internalId
	person.FirstName = "John"
	//person.BirthDate = YYYYMMDD_Date("20230101")
	person.BirthDate = datatypes.YYYYMMDD_Date("20230101")
	o := make(map[string]string)
	person.ContactPreferenceHistory = buildContact(o)

	return &person
}
func buildContact(options map[string]string) []person.ContactPreferenceHistory {
	history := make([]person.ContactPreferenceHistory, 0)
	contactType := [3]string{"Home Address", "Home Email", "Work Email"}
	contactClass := [3]string{"Street", "Email", "Email"}
	for i := 0; i < len(contactType); i++ {
		o := make(map[string]string)
		o["Type"] = contactType[i]
		o["Class"] = contactClass[i]
		var contactPoint person.ContactPoint
		switch contactClass[i] {
		case "Street":
			contactPoint = buildRandomStreetAddress(o)
		case "Email":
			contactPoint = buildRandomEmailAddress(o)

		case "Phone":
			contactPoint = buildRandomPhone(o)

		}
		cpID := contactPoint.GetContactId()
		preferenceItem := person.ContactPreferencePeriod{
			EffectiveBeginDate: "01/01/2023",
			EffectiveEndDate:   "01/01/2024",
			ContactPointId:     cpID,
			ContactPoint:       contactPoint,
		}
		preferenceItems := make([]person.ContactPreferencePeriod, 0)
		preferenceItems = append(preferenceItems, preferenceItem)
		preferenceHistory := person.ContactPreferenceHistory{
			ContactPointType:          contactType[i],
			ContactPointClass:         contactClass[i],
			ContractPreferencePeriods: preferenceItems,
		}
		history = append(history, preferenceHistory)
	}
	return history
}
func buildRandomStreetAddress(options map[string]string) person.StreetAddress {
	s := person.StreetAddress{}
	min := 1000
	max := 30000

	id := rand.Intn(max-min+1) + min
	s.ContactPointId = fmt.Sprintf("%d", id)
	s.AddressLine1 = "100 Main St"
	s.City = "Chicago"
	s.State = "IL"
	s.Zipcode = "12345"
	return s
}
func buildRandomPhone(options map[string]string) person.PhoneNumber {
	s := person.PhoneNumber{}
	min := 1000
	max := 30000

	id := rand.Intn(max-min+1) + min
	s.ContactPointId = fmt.Sprintf("%d", id)
	s.PhoneNumber = "(123) 456-7890)"

	return s
}
func buildRandomEmailAddress(options map[string]string) person.EmailAddress {
	s := person.EmailAddress{}
	min := 1000
	max := 30000

	id := rand.Intn(max-min+1) + min
	s.ContactPointId = fmt.Sprintf("%d", id)
	s.EmailAddress = "sample@company.com"
	return s
}
func CreateRandomWorker(options map[string]interface{}) *personRoles.Worker {
	worker := personRoles.Worker{}
	/*
		min := 1000
		max := 30000

		id := rand.Intn(max-min+1) + min

		internalId := fmt.Sprintf("%d", id)
	*/
	return &worker
}
