package businessProcess

import (
	"benefitsDomain/datatypes"
	"benefitsDomain/domain/person/personRoles"
	"context"
	"encoding/json"
	"fmt"
	"server/message"
	"server/message/commandDataStructures"

	"log/slog"

	"github.com/google/uuid"
	"github.com/spaolacci/murmur3"
)

type PolicyHandler interface {
	EvaluatePolicy(ec *MessageProcessingContext) bool
}
type PolicyHandlerFunc func(ec *MessageProcessingContext) bool

func (f PolicyHandlerFunc) EvaluatePolicy(ec *MessageProcessingContext) bool {
	b := f(ec)
	return b

}

// Normal Start
func (t *ElementalTask) normalStart(ec *MessageProcessingContext) {
	slog.Debug("In BusinessProcessDefinition normalStart")
	pbp := ec.PersonBusinessProcess
	os := pbp.GetOpenSegments()[0]
	os.BusinessProcessState = t.CompletionState
	pbp.UpdateOpenSegmentForSegmentId(os)
	slog.Info("#############################")
	slog.Info("------- Normal Start --------")
	slog.Info("#############################")

}
func (t *ElementalTask) openEnrollmentStart(ec *MessageProcessingContext) {
	slog.Debug("In BusinessProcessDefinition openEnrollmentStart")
	pbp := ec.PersonBusinessProcess
	os := pbp.GetOpenSegments()[0]
	os.BusinessProcessState = t.CompletionState
	pbp.UpdateOpenSegmentForSegmentId(os)
	options := pbp.buildOpenEnrollmentBusinessProcessData(ec)
	pbp.BusinessProcessData = options
	slog.Info("######################################")
	slog.Info("*------ Open Enrollment Start --------")
	slog.Info("######################################")

}

func (t *ElementalTask) sendEnrollmentCommunication(ec *MessageProcessingContext) {
	slog.Debug(" In BusinessProcessDefinition sendEnrollmentCommunication")
	pbp := ec.PersonBusinessProcess
	os := pbp.GetOpenSegments()[0]
	os.BusinessProcessState = t.CompletionState
	pbp.UpdateOpenSegmentForSegmentId(os)
	slog.Info("################################################################")
	slog.Info("------Sending out Enrollment Communication ******** " + os.SegmentDefinitionId)
	slog.Info("################################################################")

}
func (t *ElementalTask) acceptEnrollmentElections(ec *MessageProcessingContext) {
	slog.Debug("In BusinessProcessDefinition acceptEnrollmentElections")
	m := ec.Message
	command, _ := m.(*message.Command)
	jsonData := command.Data.JsonData
	var elections commandDataStructures.OpenEnrollmentElectionRequest
	err := json.Unmarshal([]byte(jsonData), &elections)
	if err != nil {
		fmt.Println(err)
	}

	pbp := ec.PersonBusinessProcess
	effectiveDate := pbp.EffectiveDate
	for _, election := range elections.BenefitPlanElections {
		personId := pbp.PersonId
		benefitId := election.BenefitId
		dataStore := ec.ResourceContext.GetPersonDataStore()
		config := make(map[string]string)
		config["keyType"] = "PersonId/BenefitId"
		config["PersonId"] = personId
		config["BenefitId"] = benefitId
		participant, err := dataStore.GetParticipant(config)

		fmt.Printf("%v", participant)
		shouldCreateNewParticipant := false
		if err != nil {
			shouldCreateNewParticipant = true
			participant, _ = personRoles.NewParticipant(personId, benefitId)
		}
		planDB := ec.ResourceContext.GetPlanDataStore()
		benefit, _ := planDB.GetBenefit(benefitId)
		newCoveragePeriod := pbp.CreateNewCoveragePeriodFromElections(election, benefit)
		participant.ApplyNewCoveragePeriod(effectiveDate, newCoveragePeriod)
		//participant.ApplyEnrollmentElections(effectiveDate, election)
		if shouldCreateNewParticipant {
			dataStore.InsertParticipant(participant)
		} else {
			dataStore.UpdateParticipant(participant.InternalId, participant)
		}
	}
	os := pbp.GetOpenSegments()[0]
	os.BusinessProcessState = t.CompletionState
	pbp.UpdateOpenSegmentForSegmentId(os)
	slog.Info("####################################################")
	slog.Info("------- Updating Business Process Elections --------")
	slog.Info("####################################################")

}
func (t *ElementalTask) closeBusinessProcess(ec *MessageProcessingContext) {
	slog.Debug(" In BusinessProcessDefinition closeBusinessProcess")
	pbp := ec.PersonBusinessProcess
	os := pbp.GetOpenSegments()[0]
	os.BusinessProcessState = t.CompletionState
	pbp.UpdateOpenSegmentForSegmentId(os)
	slog.Info("#######################################")
	slog.Info("------- Close Business Process --------")
	slog.Info("#######################################")

}
func (t *ElementalTask) releaseElections(ec *MessageProcessingContext) {
	slog.Debug(" In BusinessProcessDefinition releaseElections")
	m := ec.Message
	event, _ := m.(*message.Event)
	jsonData := event.Data.JsonData
	data := &commandDataStructures.PayrollRelease{}
	_ = json.Unmarshal([]byte(jsonData), &data)
	benefitId := data.BenefitId
	if benefitId == "" {
		slog.Info("#############################################################")
		slog.Info("------- Release Election - Bad Benefit Id  Exiting --------- ")
		slog.Info("#############################################################")
		return

	}
	pbp := ec.PersonBusinessProcess
	personId := pbp.PersonId
	dataStore := ec.ResourceContext.GetPersonDataStore()
	config := make(map[string]string)
	config["keyType"] = "PersonId/BenefitId"
	config["PersonId"] = personId
	config["BenefitId"] = benefitId
	participant, err := dataStore.GetParticipant(config)
	if err != nil {
		slog.Info("#############################################################")
		slog.Info("------- Release Election Error: " + err.Error() + "--------- ")
		slog.Info("#############################################################")
		return

	}
	coveragePeriods := &participant.CoverageHistory
	cp := &coveragePeriods.CoveragePeriods[0]
	cp.PayrollReportingState = personRoles.C_STATE_REPORTED
	dataStore.UpdateParticipant(participant.InternalId, participant)
	os := pbp.GetOpenSegments()[0]
	os.BusinessProcessState = t.CompletionState
	pbp.UpdateOpenSegmentForSegmentId(os)
	slog.Info("******* Publish Payroll Reporting Event: " + personId + " - " + benefitId)
	triggerPayrollReleaseEvent(participant, ec.ResourceContext, m)
	slog.Info("##################################################################")
	slog.Info("------- Release Election  --------- " + personId + " - " + benefitId)
	slog.Info("##################################################################")

}
func triggerPayrollReleaseEvent(participant *personRoles.Participant, rc *ResourceContext, anMessage message.Message) {

	jsonData, err := json.Marshal(participant)
	if err != nil {
		slog.Error("Error:", err)
	}
	id, _ := uuid.NewRandom()
	now := datatypes.YYYYMMDD_Date_Now().String()
	header := message.EventHeader{
		EventId:           id.String(),
		Version:           "1.0",
		EventName:         "Participant Benefit Election Reporting Update",
		EventDefinitionId: "PE0007",
		ContextTag:        "Participant_BenefitElectionReportingState",
		Action:            "Updated",
		CreationTimestamp: now,
		BusinessDomain:    "Benefits",
		CorrelationId:     id.String(),
		CorrelationIdType: "Session",
		SubjectIdentifier: participant.PersonId,
	}

	typeHeader := message.BoHeader{
		BusinessObjectResourceType: "Participant",
		BusinessObjectIdentifier:   participant.InternalId,
		DataChangeTimestamp:        now,
	}

	domain := message.EventData{
		ReferenceNumber: "UNKNOWN",
		JsonData:        string(jsonData),
		Target:          "InterestedParties",
	}
	newEvent := &message.Event{
		Header:     header,
		TypeHeader: typeHeader,
		Data:       domain,
	}
	eb := rc.GetMessageBroker()
	ctx := context.Background()
	hasher := murmur3.New128()
	pid := participant.PersonId
	hasher.Write([]byte(pid))
	partition, _ := hasher.Sum128()
	err = eb.Publish(ctx, newEvent, int(partition))
	if err != nil {
		slog.Error("Error:", err)

	}
	slog.Info("##################################################################")
	slog.Info("----- Trigger Payroll Release Event ------------------------------")
	slog.Info("##################################################################")
}
func (t *ElementalTask) defaultElections(ec *MessageProcessingContext) {
	slog.Debug(" In BusinessProcessDefinition releaseElections")
	pbp := ec.PersonBusinessProcess
	os := pbp.GetOpenSegments()[0]
	os.BusinessProcessState = t.CompletionState
	pbp.UpdateOpenSegmentForSegmentId(os)
	slog.Info("##################################################################")
	slog.Info("-------- Timeout - Default Elections  -------- " + ec.Message.GetMessageName())
	slog.Info("##################################################################")

}

/*
	func (t *ElementalTask) triggerEventWithoutParameters(ec *MessageProcessingContext) {
		slog.Debug(" In BusinessProcessDefinition triggerEvent")
		//pbp := ec.PersonBusinessProcess
		//os := pbp.GetOpenSegments()[0]
		//os.BusinessProcessState = t.CompletionState
		//pbp.UpdateOpenSegmentForSegmentId(os)
		slog.Info("******* Timeout - Trigger Event  ******** " + ec.Message.GetMessageName())
	}
*/
func (t *ElementalTask) triggerBusinessProcess(ec *MessageProcessingContext, parameters map[string]string) {
	slog.Debug(" In BusinessProcessDefinition triggerBusinessProcess")
	pbp := ec.PersonBusinessProcess
	i := ec.ProcessingArrayIndex
	th := t.ExitControlHandler[i].(TaskHandlerParameterizedFunc)
	parms := th.parameters
	//p["effectiveDate"] = "today"
	effectiveDate := ec.PersonBusinessProcess.EffectiveDate
	businessProcessDefinitionId := parms["businessProcessDefinitionId"]
	ctx := context.Background()

	ds := ec.ResourceContext.GetBusinessProcessDefinitionDataStore()
	businessProcessDefinition := ds.GetBusinessProcessDefinition(businessProcessDefinitionId)
	startParameters := BusinessProcessStartContext{
		Person:                     ec.Person,
		BusinessProcessDefinition:  businessProcessDefinition,
		EffectiveDate:              effectiveDate,
		SourceEventReferenceNumber: pbp.ReferenceNumber,
		SourceType:                 "Event",
	}
	refnum, _ := StartPersonBusinessProcess(ctx, ec.ResourceContext, startParameters)
	triggeredEvent := TriggeredBusinessProcess{
		BusinessProcessDefinitionId:    businessProcessDefinitionId,
		BusinessProcessReferenceNumber: refnum,
		EffectiveDate:                  effectiveDate,
	}
	te := pbp.TriggeredBusinessProcesses
	te = append(te, triggeredEvent)
	pbp.TriggeredBusinessProcesses = te
	slog.Info("##################################################################")
	slog.Info("--------- Trigger triggerBusinessProcess  -------- " + businessProcessDefinitionId + ":" + businessProcessDefinition.Name)
	slog.Info("##################################################################")

}

// Timeout Policy Handlers
func (t *TimeoutDefinition) matchDate(ec *MessageProcessingContext) bool {
	var b bool
	timeoutDate := t.TimeoutPolicyDate
	eventDate := ec.Message.GetDataDate()
	if timeoutDate.Equal(eventDate) {
		b = false
	} else {
		b = true
	}
	slog.Info("******* Match Date   ******** " + ec.Message.GetMessageName())

	return b
}
func (t *ElementalTask) processPOIResponse(ec *MessageProcessingContext) {
	slog.Debug("In BusinessProcessDefinition processPOIResponse")
	//pbp := ec.PersonBusinessProcess
	//os := pbp.GetOpenSegments()[0]
	//os.BusinessProcessState = t.CompletionState
	//pbp.UpdateOpenSegmentForSegmentId(os)
	slog.Info("#####################################")
	slog.Info("------- ProcessPOIResponse  -------- ")
	slog.Info("#####################################")
}
func (t *ElementalTask) compensatePOIProcess(ec *MessageProcessingContext) {
	slog.Debug("In BusinessProcessDefinition compensatePOIProcess")
	//pbp := ec.PersonBusinessProcess
	//os := pbp.GetOpenSegments()[0]
	//os.BusinessProcessState = t.CompletionState
	//pbp.UpdateOpenSegmentForSegmentId(os)
	slog.Info("######################################")
	slog.Info("------ CompensatePOIProcess  ---------")
	slog.Info("######################################")
}
