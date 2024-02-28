package application

type PersonProfileViewResponse struct {
	InternalId        string            `json:"internalId" bson:"internalId"`
	ExternalId        string            `json:"externalId" bson:"externalId"`
	FirstName         string            `json:"firstName" bson:"firstName"`
	LastName          string            `json:"lastName" bson:"lastName"`
	BirthDate         string            `json:"birthDate" bson:"birthDate"`
	Participants      []Participant     `json:"participants" bson:"participants"`
	BusinessProcesses []BusinessProcess `json:"BusinessProcesses" bson:"businessProcesses"`
}
type Participant struct {
	InternalId string `json:"internalId" bson:"internalId"`
	PersonId   string `json:"personId" bson:"personId"`
	BenefitId  string `json:"benefitId" bson:"benefitId"`
}
type BusinessProcess struct {
	ReferenceNumber string `json:"referenceNumber" bson:"referenceNumber"`
}
