package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
)

// Event
// -----
// Standard Event Message (Event or Command) Follow industry Event Standard
// { ID
// Name
// Context Tag
// Action
// Timestamp
// Data}
type EventHeader struct {
	EventId                       string `json:"eventId" bson:"eventId"`
	Version                       string `json:"version" bson:"version"`
	Topic                         string `json:"topic" bson:"topic"`
	EventCategoryHeaderDataSchema string `json:"eventCategoryHeaderDataSchema" bson:"eventCategoryHeaderDataSchema"`
	EventBodyDataSchema           string `json:"eventBodyDataSchema" bson:"eventBodyDataSchema"`
	EventName                     string `json:"eventName" bson:"eventName"`
	// Add to spec
	EventDefinitionId              string                   `json:"eventDefinitionId" bson:"eventDefinitionId"`
	ContextTag                     string                   `json:"contextTag" bson:"contextTag"`
	Action                         string                   `json:"action" bson:"action"`
	CreationTimestamp              string                   `json:"creationTimestamp" bson:"creationTimestamp"`
	BusinessDomain                 string                   `json:"businessDomain" bson:"businessDomain"`
	CorrelationId                  string                   `json:"correlationId" bson:"correlationId"`
	CorrelationIdType              string                   `json:"correlationIdType" bson:"correlationIdType"`
	SubjectIdentifier              string                   `json:"subjectIdentifier" bson:"subjectIdentifier"`
	PublisherId                    string                   `json:"publisherId" bson:"publisherId"`
	PublisherApplicationName       string                   `json:"publisherApplicationName" bson:"publisherApplicationName"`
	PublisherApplicationInstanceId string                   `json:"publisherApplicationInstanceId" bson:"publisherApplicationInstanceId"`
	PublishingPlatformsHistory     []PublishingPlatformItem `json:"publishingPlatformsHistory" bson:"publishingPlatformsHistory"`
	SystemOfRecord                 SystemOfRecordItem       `json:"systemOfRecord" bson:"systemOfRecord"`
	CorrelatedResource             []CorrelatedResourceItem `json:"correlatedResourceItem" bson:"correlatedResourceItem"`
}

type EventData struct {
	Date            string `json:"date" bson:"date"`
	ReferenceNumber string `json:"referenceNumber" bson:"referenceNumber"`
	Target          string `json:"target" bson:"target"`
	JsonData        string `json:"data" bson:"data"`
}
type BoHeader struct {
	BusinessObjectResourceType        string                                 `json:"businessObjectResourceType" bson:"businessObjectResourceType"`
	BusinessObjectIdentifier          string                                 `json:"businessObjectIdentifier" bson:"businessObjectIdentifier"`
	AdditionalBusinessObjectResources []AdditionalBusinessObjectResourceItem `json:"additionalBusinessObjectResources" bson:"additionalBusinessObjectResources"`
	DataChangeTimestamp               string                                 `json:"dataChangeTimestamp" bson:"dataChangeTimestamp"`
}
type AdditionalBusinessObjectResourceItem struct {
	AdditionalBusinessObjectResourceType string `json:"additionalBusinessObjectResourceType" bson:"additionalBusinessObjectResourceType"`
	AdditionalBusinessObjectResourceId   string `json:"additionalBusinessObjectResourceId" bson:"additionalBusinessObjectResourceId"`
}
type Event struct {
	// The following fields need to be removed
	//ID              string      `json:"id" bson:"id"`
	//Name            string      `json:"name" bson:"name"`
	//ReferenceNumber string      `json:"referenceNumber" bson:"referenceNumber"`
	Header     EventHeader `json:"eventHeader" bson:"eventHeader"`
	TypeHeader BoHeader    `json:"boEventHeader" bson:"boEventHeader"`
	Data       EventData   `json:"eventData" bson:"eventData"`
}

func (e Event) String() string {
	return fmt.Sprintf("Event: %s-%s", e.Header.EventId, e.Header.EventName)

}
func (e Event) SetReferenceNumber(s string) {
	e.Data.ReferenceNumber = s

}
func (e Event) GetReferenceNumber() string {
	return e.Data.ReferenceNumber

}
func (e Event) GetTarget() string {
	return e.Data.Target

}
func (e Event) GetMessageId() string {
	return e.Header.EventId

}
func (e Event) GetMessageDfnId() string {
	return e.Header.EventDefinitionId

}
func (e Event) GetMessageName() string {
	return e.Header.EventName

}
func (e Event) GetDataDate() string {
	return e.Data.Date

}
func (e Event) MarshallJson() ([]byte, error) {
	jsonData, err := json.Marshal(e)
	if err != nil {
		log.Fatal(err)
	}
	return jsonData, err

}

type EventDefinition struct {
	ID        string
	Name      string
	EventDate string
}

func (e EventDefinition) String() string {
	return fmt.Sprintf("ED: %s %s", e.ID, e.Name)
}
func (e Event) Report(templateDirectory string) string {

	templateFile := templateDirectory + "/eventTemplate.tmpl"
	buf := new(bytes.Buffer)
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(buf, e)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
