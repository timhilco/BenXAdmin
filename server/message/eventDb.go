package message

import (
	"encoding/json"
	"server/message/commandDataStructures"

	"github.com/google/uuid"
)

func CreateMockEventDefinitionObject(id string) EventDefinition {
	var e EventDefinition
	switch id {
	case "PE0001":
		e = EventDefinition{
			ID:   "PE0001",
			Name: "Person Business Process Data Object Update",
		}
	case "PE0002":
		e = EventDefinition{
			ID:   "PE0002",
			Name: "Carrier Report Date Timestamp Clock Tick",
		}
	case "PE0003":
		e = EventDefinition{
			ID:   "PE0003",
			Name: "Confirmation Statement Date Timestamp Clock Tick",
		}
	case "PE0004":
		e = EventDefinition{
			ID:   "PE0004",
			Name: "Payroll Report Date Timestamp Clock Tick",
		}
	case "PE0005":
		e = EventDefinition{
			ID:   "PE0005",
			Name: "New Day Clock Tick",
		}
	case "PE0006":
		e = EventDefinition{
			ID:   "PE0006",
			Name: "POI Requirement Update",
		}

	}
	return e

}
func BuildEvent(id string, referenceNumber string) Message {
	dfn := CreateMockEventDefinitionObject(id)
	uid, _ := uuid.NewRandom()
	header := EventHeader{
		EventId:           uid.String(),
		EventName:         id + "-" + dfn.Name,
		EventDefinitionId: id,
	}
	domain := EventData{
		ReferenceNumber: referenceNumber,
		Target:          "This",
	}
	if id == "PE0004" {
		data := commandDataStructures.PayrollRelease{
			BenefitId: "B001",
			PayrollId: "Payroll-Active-SemiMonthly",
		}
		jsonData, _ := json.Marshal(data)
		domain.JsonData = string(jsonData)
	}
	anEvent := Event{
		Header: header,
		Data:   domain,
	}
	return &anEvent

}
func BuildClosePartitionEvent() Message {
	header := EventHeader{
		EventId:   "CloseEventPartition",
		EventName: "CloseEventPartition",
	}
	domain := EventData{}

	anEvent := Event{
		Header: header,
		Data:   domain,
	}
	return &anEvent

}
