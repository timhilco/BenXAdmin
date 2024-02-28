package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
)

// Command
// -----
// Standard Command Message (Command or Command) Follow industry Command Standard
// { ID
// Name
// Context Tag
// Action
// Timestamp
// Data}
type CommandHeader struct {
	CommandId                       string `json:"commandId" bson:"commandId"`
	Version                         string `json:"version" bson:"version"`
	Topic                           string `json:"topic" bson:"topic"`
	CommandCategoryHeaderDataSchema string `json:"commandCategoryHeaderDataSchema" bson:"commandCategoryHeaderDataSchema"`
	CommandBodyDataSchema           string `json:"commandBodyDataSchema" bson:"commandBodyDataSchema"`
	CommandName                     string `json:"commandName" bson:"commandName"`
	// Add to spec
	CommandDefinitionId            string                   `json:"commandDefinitionId" bson:"commandDefinitionId"`
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

type CommandData struct {
	Date            string `json:"date" bson:"date"`
	ReferenceNumber string `json:"referenceNumber" bson:"referenceNumber"`
	Target          string `json:"target" bson:"target"`
	JsonData        string `json:"data" bson:"data"`
}
type Command struct {
	Header CommandHeader `json:"commandHeader" bson:"commandHeader"`
	Data   CommandData   `json:"commandData" bson:"commandData"`
}

func (c Command) String() string {
	return fmt.Sprintf("Command: %s-%s", c.Header.CommandId, c.Header.CommandName)

}

func (c *Command) SetReferenceNumber(s string) {
	c.Data.ReferenceNumber = s

}
func (c Command) GetReferenceNumber() string {
	return c.Data.ReferenceNumber

}
func (c Command) GetTarget() string {
	return c.Data.Target

}
func (c Command) GetMessageId() string {
	return c.Header.CommandId

}
func (c Command) GetMessageDfnId() string {
	return c.Header.CommandDefinitionId

}
func (c Command) GetMessageName() string {
	return c.Header.CommandName

}
func (c Command) GetDataDate() string {
	return c.Data.Date

}
func (c Command) MarshallJson() ([]byte, error) {
	jsonData, err := json.Marshal(c)
	if err != nil {
		log.Fatal(err)
	}
	return jsonData, err

}

type CommandDefinition struct {
	ID          string
	Name        string
	CommandDate string
}

func (c CommandDefinition) String() string {
	return fmt.Sprintf("ED: %s %s", c.ID, c.Name)
}
func (c Command) Report(templateDirectory string) string {

	templateFile := templateDirectory + "commandTemplate.tmpl"
	buf := new(bytes.Buffer)
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(buf, c)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
