package message

import (
	"encoding/json"
	"server/message/commandDataStructures"
)

func CreateMockCommandDefinitionObject(id string) EventDefinition {
	var e EventDefinition
	switch id {
	case "PC0001":
		e = EventDefinition{
			ID:   "PC0001",
			Name: "Accept Enrollment Elections",
		}
	case "PC0002":
		e = EventDefinition{
			ID:   "PC0002",
			Name: "Update POI Response",
		}

	}
	return e

}
func BuildCommand(id string, referenceNumber string) Message {
	dfn := CreateMockCommandDefinitionObject(id)
	header := CommandHeader{
		CommandId:   id,
		CommandName: dfn.Name,
	}
	domain := CommandData{
		ReferenceNumber: referenceNumber,
		Target:          "This",
	}

	if id == "PC0001" {
		elections := make([]commandDataStructures.EnrollmentElection, 0)
		medical := commandDataStructures.EnrollmentElection{
			BenefitId:     "B001",
			BenefitPlanId: "BP001-002",
		}
		elections = append(elections, medical)
		dental := commandDataStructures.EnrollmentElection{
			BenefitId:     "B002",
			BenefitPlanId: "BP002-001",
		}
		elections = append(elections, dental)

		data := commandDataStructures.OpenEnrollmentElectionRequest{
			BenefitPlanElections: elections,
		}
		jsonData, _ := json.Marshal(data)
		domain.JsonData = string(jsonData)
	}
	aCommand := Command{
		Header: header,
		Data:   domain,
	}
	return &aCommand

}
