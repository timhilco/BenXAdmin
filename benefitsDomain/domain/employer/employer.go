package worker

type Employer struct {
	InternalId   string `json:"internalId" bson:"internalId"`
	EmployerId   string `json:"employerId" bson:"employerId"`
	EmployerName string `json:"employerName" bson:"employerName"`
}
