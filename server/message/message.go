package message

const (
	C_MESSAGE_TYPE_PERSON_EVENT = "Person_Event"
	C_MESSAGE_TYPE_COMMAND      = "Command"
)

type Message interface {
	SetReferenceNumber(string)
	GetReferenceNumber() string
	GetMessageId() string
	GetMessageDfnId() string
	GetMessageName() string
	GetTarget() string
	GetDataDate() string
	String() string
	Report(string) string
	MarshallJson() ([]byte, error)
}
type PublishingPlatformItem struct {
	PublisherId                    string `json:"publisherId" bson:"publisherId"`
	PublisherApplicationName       string `json:"publisherApplicationName" bson:"publisherApplicationName"`
	PublisherApplicationInstanceId string `json:"publisherApplicationInstanceId" bson:"publisherApplicationInstanceId"`
	EventId                        string `json:"eventId" bson:"eventId"`
	Topic                          string `json:"topic" bson:"topic"`
	EventName                      string `json:"eventName" bson:"eventName"`
	CreationTimestamp              string `json:"creationTimestamp" bson:"creationTimestamp"`
}
type SystemOfRecordItem struct {
	SystemOfRecordId                  string `json:"systemOfRecordId" bson:"systemOfRecordId"`
	SystemOfRecordApplicationName     string `json:"systemOfRecordApplicationName" bson:"systemOfRecordApplicationName"`
	SystemOfRecordApplicationInstance string `json:"systemOfRecordApplicationInstance" bson:"systemOfRecordApplicationInstance"`
	SystemOfRecordIdDatabaseSchema    string `json:"systemOfRecordIdDatabaseSchema" bson:"systemOfRecordIdDatabaseSchema"`
	PlatformInternalId                string `json:"platformInternalId" bson:"platformInternalId"`
	PlatformExternalId                string `json:"platformExternalId" bson:"platformExternalId"`
}
type CorrelatedResourceItem struct {
	CorrelatedResourceType        string `json:"correlatedResourceType" bson:"correlatedResourceType"`
	CorrelatedResourceId          string `json:"correlatedResourceId" bson:"correlatedResourceId"`
	CorrelatedResourceState       string `json:"correlatedResourceState" bson:"correlatedResourceState"`
	CorrelatedResourceDescription string `json:"correlatedResourceDescription" bson:"correlatedResourceDescription"`
}
