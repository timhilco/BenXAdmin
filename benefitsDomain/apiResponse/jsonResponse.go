package apiResponse

type PersonProfileViewResponse struct {
	InternalId        string            `json:"internalId" bson:"internalId"`
	ExternalId        string            `json:"externalId" bson:"externalId"`
	FirstName         string            `json:"firstName" bson:"firstName"`
	LastName          string            `json:"lastName" bson:"lastName"`
	BirthDate         string            `json:"birthDate" bson:"birthDate"`
	Workers           []Worker          `json:"workers" bson:"workers"`
	Participants      []Participant     `json:"participants" bson:"participants"`
	BusinessProcesses []BusinessProcess `json:"BusinessProcesses" bson:"businessProcesses"`
}
type Participant struct {
	InternalId string `json:"internalId" bson:"internalId"`
	PersonId   string `json:"personId" bson:"personId"`
	BenefitId  string `json:"benefitId" bson:"benefitId"`
}
type BusinessProcess struct {
	InternalId                  string `json:"internalId" bson:"internalId"`
	ReferenceNumber             string `json:"referenceNumber" bson:"referenceNumber"`
	BusinessProcessDefinitionId string `json:"businessProcessDefinitionId" bson:"businessProcessDefinitionId"`
	EffectiveDate               string `json:"effectiveDate" bson:"effectiveDate"`
	State                       string `json:"state" bson:"state"`
}
type Worker struct {
	InternalId       string `json:"internalId" bson:"internalId"`
	WorkerId         string `json:"workerId" bson:"workerId"`
	Employer         string `json:"employer" bson:"employer"`
	Pay              string `json:"pay" bson:"pay"`
	EmploymentStatus string `json:"employmentStatus" bson:"employmentStatus"`
}
