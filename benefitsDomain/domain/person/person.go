package person

import (
	"benefitsDomain/datatypes"
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
)

type Person struct {
	InternalId               string                     `json:"internalId" bson:"internalId"`
	ExternalId               string                     `json:"externalId" bson:"externalId"`
	FirstName                string                     `json:"firstName" bson:"firstName"`
	LastName                 string                     `json:"lastName" bson:"lastName"`
	BirthDate                datatypes.YYYYMMDD_Date    `json:"birthDate" bson:"birthDate"`
	ContactPreferenceHistory []ContactPreferenceHistory `json:"contactPreferenceHistory" bson:"contactPreferenceHistory"`
}
type ContactPreferencePeriod struct {
	EffectiveBeginDate string      `json:"effectiveBeginDate" bson:"effectiveBeginDate"`
	EffectiveEndDate   string      `json:"effectiveEndDate" bson:"effectiveEndDate"`
	ContactPoint       interface{} `json:"contactPoint" bson:"contactPoint"`
	ContactPointId     string      `json:"contactPointId" bson:"contactPointId"`
}
type ContactPreferenceHistory struct {
	ContactPointType          string                    `json:"contactPointType" bson:"contactPointType"`
	ContactPointClass         string                    `json:"contactPointClass" bson:"contactPointClass"`
	ContractPreferencePeriods []ContactPreferencePeriod `json:"contractPreferencePeriods" bson:"contractPreferencePeriods"`
}
type ContactPoint interface {
	GetContactId() string
}
type EmailAddress struct {
	ContactPointId string `json:"contactPointId" bson:"contactPointId"`
	EmailAddress   string `json:"emailAddress" bson:"emailAddress"`
}

func (c EmailAddress) GetContactId() string {
	return c.ContactPointId
}

type StreetAddress struct {
	ContactPointId string `json:"contactPointId" bson:"contactPointId"`
	AddressLine1   string `json:"addressLine1" bson:"addressLine1"`
	City           string `json:"city" bson:"city"`
	State          string `json:"state" bson:"state"`
	Zipcode        string `json:"zipcode" bson:"zipcode"`
}

func (c StreetAddress) GetContactId() string {
	return c.ContactPointId
}

type PhoneNumber struct {
	ContactPointId string `json:"contactPointId" bson:"contactPointId"`
	PhoneNumber    string `json:"phoneNumber" bson:"phoneNumber"`
}

func (c PhoneNumber) GetContactId() string {
	return c.ContactPointId
}

func (p *Person) ConvertContactPointsToStructs() error {
	newHistories := make([]ContactPreferenceHistory, 0, len(p.ContactPreferenceHistory))
	for _, v := range p.ContactPreferenceHistory {
		h := ContactPreferenceHistory{}
		cpc := v.ContactPointClass
		cpt := v.ContactPointType
		cpi := make([]ContactPreferencePeriod, 0, len(v.ContractPreferencePeriods))
		h.ContactPointClass = cpc
		h.ContactPointType = cpt
		for _, item := range v.ContractPreferencePeriods {
			i := ContactPreferencePeriod{}
			i.EffectiveBeginDate = item.EffectiveBeginDate
			i.EffectiveEndDate = item.EffectiveEndDate
			i.ContactPointId = item.ContactPointId
			b, err := bson.Marshal(item.ContactPoint)
			if err != nil {
				return err
			}
			switch cpc {
			case "Street":
				var res StreetAddress
				err := bson.Unmarshal(b, &res)
				if err != nil {
					return err
				}
				i.ContactPoint = res
			case "Phone":
				var res PhoneNumber
				err := bson.Unmarshal(b, &res)
				if err != nil {
					return err
				}
				i.ContactPoint = res
			case "Email":
				var res EmailAddress
				err := bson.Unmarshal(b, &res)
				if err != nil {
					return err
				}
				i.ContactPoint = res

			}
			cpi = append(cpi, i)
		}
		h.ContractPreferencePeriods = cpi
		newHistories = append(newHistories, h)
	}

	p.ContactPreferenceHistory = newHistories
	return nil

}
func (p *Person) Equal(component string, otherPerson *Person) bool {
	if p.InternalId != otherPerson.InternalId {
		return false
	}
	if p.FirstName != otherPerson.FirstName {
		return false
	}
	if p.LastName != otherPerson.LastName {
		return false
	}
	if p.BirthDate != otherPerson.BirthDate {
		return false
	}
	return true
}
func PersonJSON_EQUAL(firstPerson string, secondPerson string) bool {
	first, _ := CreatePersonFromJson([]byte(firstPerson))
	second, _ := CreatePersonFromJson([]byte(secondPerson))

	return first.Equal("ALL", second)
}

func CreatePersonFromJson(j []byte) (*Person, error) {

	person := Person{}

	err := json.Unmarshal(j, &person)

	return &person, err
}
func CreateJsonFromPerson(person *Person) ([]byte, error) {
	// Convert struct to JSON
	jsonData, err := json.Marshal(person)
	if err != nil {
		log.Fatal(err)
	}
	return jsonData, err
}
func CreatePersonFromJsonFile(filename string) (*Person, error) {

	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	person := Person{}

	err = json.Unmarshal(data, &person) // here!
	if err != nil {
		panic(err)
	}
	return &person, nil
}
func (p *Person) Report(ev datatypes.EnvironmentVariables) string {
	dir := ev.TemplateDirectory
	templateFile := dir + "person.tmpl"
	buf := new(bytes.Buffer)
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(buf, p)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
