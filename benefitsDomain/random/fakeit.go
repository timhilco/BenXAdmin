package random

import (
	"benefitsDomain/datatypes"
	"benefitsDomain/domain/person"
	"benefitsDomain/domain/person/personRoles"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

func CreateFakeItPerson(options map[string]interface{}) *person.Person {
	person := person.Person{}
	externalId := options["ExternalId"].(string)

	person.InternalId = gofakeit.UUID()
	person.ExternalId = externalId
	person.LastName = gofakeit.LastName()
	person.FirstName = gofakeit.FirstName()
	//min Date - 01/01/1960
	minDate := time.Date(1960, time.January, 1, 1, 0, 0, 0, time.UTC)
	maxDate := time.Now().AddDate(-21, 0, 0)
	bd := gofakeit.DateRange(minDate, maxDate).Format("20060102")
	person.BirthDate = datatypes.YYYYMMDD_Date(bd)
	person.ContactPreferenceHistory = buildFakeItContact(nil)
	return &person
}
func buildFakeItContact(options map[string]string) []person.ContactPreferenceHistory {
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
			contactPoint = buildFakeItStreetAddress(o)
		case "Email":
			contactPoint = buildFakeItEmailAddress(o)

		case "Phone":
			contactPoint = buildFakeItPhone(o)

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
func buildFakeItStreetAddress(options map[string]string) person.StreetAddress {
	s := person.StreetAddress{}
	min := 1000
	max := 30000

	id := rand.Intn(max-min+1) + min
	s.ContactPointId = fmt.Sprintf("%d", id)
	a := gofakeit.Address()
	s.AddressLine1 = a.Street
	s.City = a.City
	s.State = a.State
	s.Zipcode = a.Zip
	return s
}
func buildFakeItPhone(options map[string]string) person.PhoneNumber {
	s := person.PhoneNumber{}
	min := 1000
	max := 30000

	id := rand.Intn(max-min+1) + min
	s.ContactPointId = fmt.Sprintf("%d", id)
	s.PhoneNumber = gofakeit.PhoneFormatted()

	return s
}
func buildFakeItEmailAddress(options map[string]string) person.EmailAddress {
	s := person.EmailAddress{}
	min := 1000
	max := 30000

	id := rand.Intn(max-min+1) + min
	s.ContactPointId = fmt.Sprintf("%d", id)
	s.EmailAddress = gofakeit.Email()
	return s
}
func CreateFakeItWorker(options map[string]interface{}) *personRoles.Worker {
	worker := personRoles.Worker{}

	externalId := options["WorkerId"].(string)
	personInternalId := options["PersonInternalId"].(string)

	worker.InternalId = gofakeit.UUID()
	worker.WorkerId = externalId
	worker.PersonInternalId = personInternalId
	worker.Employer = "ABC CORP"
	pay := gofakeit.Float64Range(10000, 1000000)
	s := strconv.FormatFloat(pay, 'g', -2, 32)
	worker.Pay, _ = datatypes.NewBigFloat(s)
	worker.EmploymentStatus = "Active"

	return &worker
}
