package businessProcess

import (
	"benefitsDomain/datatypes"
	"benefitsDomain/domain/benefitPlan"
	"benefitsDomain/domain/person"
	"benefitsDomain/domain/person/personRoles"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"server/message"
	"server/message/commandDataStructures"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/spaolacci/murmur3"
)

const (
	C_TASK_ACTION_NORMAL = iota
	C_TASK_ACTION_TIMEOUT
	C_TASK_ACTION_COMPENSATION
	C_TASK_ACTION_UNKNOWN
	C_TASK_ACTION_REENTRANT
)

type PersonBusinessProcess struct {
	InternalId                  string                     `json:"internalId" bson:"internalId"`
	ReferenceNumber             string                     `json:"referenceNumber" bson:"referenceNumber"`
	PersonId                    string                     `json:"personId" bson:"personId"`
	State                       int                        `json:"state" bson:"state"`
	BusinessProcessDefinitionId string                     `json:"businessProcessDefinitionId" bson:"businessProcessDefinitionId"`
	EffectiveDate               datatypes.YYYYMMDD_Date    `json:"effectiveDate" bson:"effectiveDate"`
	CreationDate                datatypes.YYYYMMDD_Date    `json:"creationDate" bson:"creationDate"`
	SourceEventReferenceNumber  string                     `json:"sourceEventReferenceNumber" bson:"sourceEventReferenceNumber"`
	SourceType                  string                     `json:"sourceType" bson:"sourceType"`
	SegmentStates               []SegmentState             `json:"segmentStates" bson:"segmentStates"`
	TriggeredBusinessProcesses  []TriggeredBusinessProcess `json:"triggeredEvents" bson:"triggeredEvents"`
	BusinessProcessData         interface{}                `json:"data" bson:"data"`
}
type SegmentState struct {
	SegmentDefinitionId  string `json:"segmentDefinitionId" bson:"segmentDefinitionId"`
	BusinessProcessState string `json:"businessProcessState" bson:"businessProcessState"`
	SegmentState         int    `json:"segmentState" bson:"segmentState"`
	WaitingTaskId        string `json:"waitingTaskId" bson:"waitingTaskId"`
}
type TriggeredBusinessProcess struct {
	BusinessProcessDefinitionId    string                  `json:"businessProcessDefinitionId" bson:"businessProcessDefinitionId"`
	BusinessProcessReferenceNumber string                  `json:"businessProcessReferenceNumber" bson:"businessProcessReferenceNumber"`
	EffectiveDate                  datatypes.YYYYMMDD_Date `json:"effectiveDate" bson:"effectiveDate"`
}
type BusinessProcessStartContext struct {
	Person                     *person.Person
	Worker                     *personRoles.Worker
	BusinessProcessDefinition  *BusinessProcessDefinition
	EffectiveDate              datatypes.YYYYMMDD_Date
	SourceEventReferenceNumber string
	SourceType                 string
}

func StartPersonBusinessProcess(ctx context.Context, rc *ResourceContext, startParameters BusinessProcessStartContext) (string, error) {
	person := startParameters.Person
	worker := startParameters.Worker
	businessProcessDefinition := startParameters.BusinessProcessDefinition
	effectiveDate := startParameters.EffectiveDate
	slog.Debug("Starting Person Business Process")
	referenceNumber := fmt.Sprintf("%s_%s_%s", person.ExternalId, businessProcessDefinition.InternalId, effectiveDate)

	pbp := &PersonBusinessProcess{
		InternalId:                  person.InternalId + "_" + referenceNumber,
		PersonId:                    person.InternalId,
		ReferenceNumber:             referenceNumber,
		EffectiveDate:               effectiveDate,
		CreationDate:                datatypes.YYYYMMDD_Date("20231101"),
		BusinessProcessDefinitionId: businessProcessDefinition.InternalId,
		SourceEventReferenceNumber:  startParameters.SourceEventReferenceNumber,
		SourceType:                  startParameters.SourceType,
		State:                       C_STATE_OPEN,
	}
	ds := rc.GetBusinessProcessStore()
	ds.InsertPersonBusinessProcess(pbp)
	ec := &MessageProcessingContext{
		ResourceContext:                   rc,
		Message:                           nil,
		Person:                            person,
		Worker:                            worker,
		PersonBusinessProcess:             pbp,
		BusinessProcessDefinition:         businessProcessDefinition,
		ShouldUpdatePersonBusinessProcess: false,
		ContextDataMap:                    make(map[string]string),
	}
	businessProcessDefinition.StartBenefitProcessFlow(ec, effectiveDate)
	ds.UpdatePersonBusinessProcess(pbp.ReferenceNumber, pbp)
	return pbp.ReferenceNumber, nil
}
func (pbp *PersonBusinessProcess) ProcessMessage(rc *ResourceContext, event message.Message) {
	slog.Info("-------------------------------------------------------------------")
	slog.Info("PersonBusinessProcess::ProcessMessage: Enter " + event.String())
	slog.Info("-------------------------------------------------------------------")
	bpd := rc.GetBusinessProcessDefinitionDataStore().GetBusinessProcessDefinition(pbp.BusinessProcessDefinitionId)
	person, _ := rc.GetPersonDataStore().GetPerson(pbp.PersonId, "Internal")
	ec := MessageProcessingContext{
		ResourceContext:                   rc,
		PersonBusinessProcess:             pbp,
		Message:                           event,
		Person:                            person,
		BusinessProcessDefinition:         bpd,
		ShouldUpdatePersonBusinessProcess: false,
		ContextDataMap:                    make(map[string]string),
	}
	//	fmt.Println(ec.Report())
	oss := pbp.GetAllSegments()
	//isMessageProcessed := false
	for i, ss := range oss {
		shouldExit := false
		data := fmt.Sprintf("Message: %s - Segment State: %s- index: %d", event.String(), ss.String(), i)
		slog.Debug("PersonBusinessProcess::ProcessMessage State Entry" + data)
		segmentDefinition, _ := ec.BusinessProcessDefinition.Flow.GetSegment(ss.SegmentDefinitionId)
		var task Task
		var eventString string
		if ss.SegmentState == C_STATE_OPEN {
			task = bpd.GetTaskForSegment(ss.SegmentDefinitionId, ss.WaitingTaskId)
		} else {
			foundTask := false
			for _, t := range segmentDefinition.Tasks {
				if t.IsReentrant() {
					foundTask = true
					task = t
					eventString = fmt.Sprintf("IncomingMessage: %s -Task Message: %s", event.String(), task.GetExpectedMessageId())
				}
			}
			if foundTask {
				slog.Debug("PersonBusinessProcess::ProcessMessage: Closed Task - Found Reentrant " + eventString)
			} else {
				shouldExit = true
			}
		}
		if shouldExit {
			slog.Debug("PersonBusinessProcess::ProcessMessage: Exit due to Closed Task not Reentrant")

		} else {
			eventString = fmt.Sprintf("IncomingMessage: %s -Task Message: %s", event.String(), task.GetExpectedMessageId())
			isMessageMatch, processingAction := CanProcessMessageForTask(event, ss, task)
			if isMessageMatch {
				switch processingAction {
				case C_TASK_ACTION_TIMEOUT:
					// Check if Timeout Policy is true
					timeoutDefinition := task.GetTimeoutDefinition()
					isTimeout := timeoutDefinition.EvaluatePolicy(ec)
					if isTimeout {
						slog.Debug("PersonBusinessProcess::ProcessMessage: Timeout Policy valid " + eventString)
					} else {
						slog.Info("-------------------------------------------------------------------")
						slog.Debug("PersonBusinessProcess::ProcessMessage: Timeout Policy not valid " + eventString)
						slog.Info("PersonBusinessProcess::ProcessMessage: Exit " + event.String())
						slog.Info("-------------------------------------------------------------------")
						return
					}
				case C_TASK_ACTION_REENTRANT:
					slog.Debug("PersonBusinessProcess::ProcessMessage: Do ReEntrant Processing" + eventString)
					sds := ec.BusinessProcessDefinition.Flow.getFutureSegmentDefinitionsFollowing(segmentDefinition)
					pbp.setSegmentStateToReentrant(sds)
				}
				ss.ProcessMessage(&ec, task, processingAction)
				pbp.UpdateSegmentState(&ec, task, processingAction)
				ec.ShouldUpdatePersonBusinessProcess = true
			} else {
				slog.Debug("PersonBusinessProcess::ProcessMessage: Message does not match any expecting event for Segment: " + ss.SegmentDefinitionId)
			}
		}
	}
	if ec.ShouldUpdatePersonBusinessProcess {
		rc.GetBusinessProcessStore().UpdatePersonBusinessProcess(pbp.ReferenceNumber, pbp)
		slog.Debug("PersonBusinessProcess::ProcessMessage: Updating Person Business Process ")
		pbp.triggerStateChange(rc, event)
		ec.ShouldUpdatePersonBusinessProcess = false

	}
	slog.Info("-------------------------------------------------------------------")
	slog.Info("PersonBusinessProcess::ProcessMessage: Exit " + event.String())
	slog.Info("-------------------------------------------------------------------")
}
func (pbp *PersonBusinessProcess) triggerStateChange(rc *ResourceContext, anMessage message.Message) {
	jsonData, err := json.Marshal(pbp)
	if err != nil {
		slog.Error("Error:", err)
	}
	id, _ := uuid.NewRandom()
	now := datatypes.YYYYMMDD_Date_Now().String()
	header := message.EventHeader{
		EventId:           id.String(),
		Version:           "1.0",
		EventName:         "Person Business Process Data Object Update",
		EventDefinitionId: "PE0001",
		ContextTag:        "PersonBusinessObject",
		Action:            "Update",
		CreationTimestamp: now,
		BusinessDomain:    "Benefits",
		CorrelationId:     id.String(),
		CorrelationIdType: "Session",
		SubjectIdentifier: pbp.PersonId,
	}

	typeHeader := message.BoHeader{
		BusinessObjectResourceType: "PersonBusinessProcess",
		BusinessObjectIdentifier:   pbp.ReferenceNumber,
		DataChangeTimestamp:        now,
	}

	domain := message.EventData{
		ReferenceNumber: pbp.ReferenceNumber,
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
	pid := pbp.PersonId
	hasher.Write([]byte(pid))
	partition, _ := hasher.Sum128()
	err = eb.Publish(ctx, newEvent, int(partition))
	if err != nil {
		slog.Error("Error:", err)

	}
	id, _ = uuid.NewRandom()

	header = message.EventHeader{
		EventId:           id.String(),
		Version:           "1.0",
		EventName:         "POI Requirement Update",
		EventDefinitionId: "PE0006",
		ContextTag:        "PersonBusinessObject_POI",
		Action:            "Update",
		CreationTimestamp: now,
		BusinessDomain:    "Benefits",
		CorrelationId:     id.String(),
		CorrelationIdType: "Session",
		SubjectIdentifier: pbp.PersonId,
	}

	typeHeader = message.BoHeader{
		BusinessObjectResourceType: "PersonBusinessProcess",
		BusinessObjectIdentifier:   pbp.ReferenceNumber,
		DataChangeTimestamp:        now,
	}

	domain = message.EventData{
		ReferenceNumber: pbp.ReferenceNumber,
		JsonData:        string(jsonData),
		Target:          "InterestedParties",
	}
	newEvent = &message.Event{
		Header:     header,
		TypeHeader: typeHeader,
		Data:       domain,
	}
	ctx = context.Background()
	err = eb.Publish(ctx, newEvent, int(partition))
	if err != nil {
		slog.Error("Error:", err)

	}
}

func (pbp *PersonBusinessProcess) setSegmentStateToReentrant(s []*SegmentDefinition) {
	for _, v := range s {
		ss, _ := pbp.GetSegmentState(v)
		ss.SegmentState = C_STATE_REENTRANT
		pbp.SetSegmentState(v, ss)

	}

}
func (pbp *PersonBusinessProcess) GetSegmentState(s *SegmentDefinition) (SegmentState, error) {
	ss := pbp.SegmentStates
	for _, item := range ss {
		if item.SegmentDefinitionId == s.Id {
			return item, nil
		}
	}
	return SegmentState{}, errors.New("not found")
}
func (pbp *PersonBusinessProcess) SetSegmentState(s *SegmentDefinition, state SegmentState) error {
	ss := make([]SegmentState, 0)
	for _, item := range pbp.SegmentStates {
		if item.SegmentDefinitionId == s.Id {
			ss = append(ss, state)
		} else {
			ss = append(ss, item)
		}
	}
	pbp.SegmentStates = ss
	return nil

}
func (pbp *PersonBusinessProcess) UpdateOpenSegmentForSegmentId(state SegmentState) {
	ss := pbp.SegmentStates
	newSS := make([]SegmentState, 0)
	for _, v := range ss {
		if v.SegmentState == C_STATE_OPEN && v.SegmentDefinitionId == state.SegmentDefinitionId {
			newSS = append(newSS, state)
		} else {
			newSS = append(newSS, v)
		}
	}
	pbp.SegmentStates = newSS

}
func (pbp *PersonBusinessProcess) GetOpenSegmentInBusinessProcess() (SegmentState, error) {
	ss := pbp.SegmentStates
	for _, v := range ss {
		if v.SegmentState == C_STATE_OPEN {
			return v, nil
		}
	}
	return SegmentState{}, errors.New("no open segment")
}
func (pbp *PersonBusinessProcess) GetOpenSegments() []SegmentState {
	rss := make([]SegmentState, 0)
	pss := pbp.SegmentStates
	for _, v := range pss {
		if v.SegmentState == C_STATE_OPEN {
			rss = append(rss, v)
		}
	}
	return rss
}
func (pbp *PersonBusinessProcess) GetAllSegments() []SegmentState {
	rss := make([]SegmentState, 0)
	pss := pbp.SegmentStates
	rss = append(rss, pss...)

	return rss
}
func (ss SegmentState) ProcessMessage(ec *MessageProcessingContext, task Task, processingAction int) {
	slog.Debug("In SegmentState::ProcessMessage")
	task.Execute(ec, processingAction)
}
func (pbp *PersonBusinessProcess) UpdateFlowState(ec *MessageProcessingContext, completedSegment *SegmentDefinition) {
	slog.Debug("In PersonBusiness Process::UpdateFlowState:" + completedSegment.Id)
	nextSegmentDefinition, _ := ec.BusinessProcessDefinition.DetermineNextSegment(completedSegment)
	isDone := len(nextSegmentDefinition) == 0
	if isDone {
		pbp.Close()
		ec.ShouldUpdatePersonBusinessProcess = true
	} else {
		for _, s := range nextSegmentDefinition {
			pbp.StartSegment(ec, s)
		}
	}

}
func (ss SegmentState) String() string {
	var sb strings.Builder
	sb.WriteString(ss.SegmentDefinitionId + ":")
	sb.WriteString(ss.WaitingTaskId + ":")
	sb.WriteString(strconv.Itoa(ss.SegmentState) + ":")
	sb.WriteString(ss.BusinessProcessState)

	return sb.String()
}
func (pbp *PersonBusinessProcess) UpdateSegmentState(ec *MessageProcessingContext, currentTask Task, processingAction int) {
	if processingAction == C_TASK_ACTION_COMPENSATION {
		return
	}
	slog.Debug("In PersonBusiness Process::UpdateSegmentState:" + currentTask.GetID())
	s := currentTask.GetParentSegment()
	nextTask, action := s.DetermineNextTask(ec, currentTask)
	switch action {
	case "Done":
		pbp.CloseSegment(ec, s)
	default:
		ss, _ := pbp.GetSegmentState(s)
		ss.WaitingTaskId = nextTask.GetID()
		pbp.UpdateOpenSegmentForSegmentId(ss)
		ec.ShouldUpdatePersonBusinessProcess = true
	}
}
func (pbp *PersonBusinessProcess) isStartValid(ec *MessageProcessingContext, s *SegmentDefinition) bool {
	if s.PreRegSegments == nil {
		return true
	}
	for _, v := range s.PreRegSegments {
		ss, _ := pbp.GetSegmentState(v)
		if ss.SegmentState == C_STATE_OPEN {
			return false
		}
	}
	return true

}
func (pbp *PersonBusinessProcess) StartSegment(ec *MessageProcessingContext, s *SegmentDefinition) {
	slog.Debug("In PersonBusinessProcess::StartSegment:" + s.Id)
	if !pbp.isStartValid(ec, s) {
		slog.Debug("In PersonBusinessProcess::StartSegment:" + s.Id + " --Not Starting Segment")
		return
	}

	ss := SegmentState{
		SegmentDefinitionId: s.Id,
		SegmentState:        C_STATE_OPEN,
	}
	personSegmentStates := pbp.SegmentStates
	if personSegmentStates == nil {
		personSegmentStates = make([]SegmentState, 0)
	}
	task := s.Tasks[0]
	if task.GetExpectedMessageId() != "" {
		ss.WaitingTaskId = task.GetID()
	}
	personSegmentStates = append(personSegmentStates, ss)
	pbp.SegmentStates = personSegmentStates
	if task.GetExpectedMessageId() == "" {
		task.Execute(ec, C_TASK_ACTION_NORMAL)
		pbp.UpdateSegmentState(ec, task, C_TASK_ACTION_NORMAL)
	}
	ec.ShouldUpdatePersonBusinessProcess = true

}
func (pbp *PersonBusinessProcess) CloseSegment(ec *MessageProcessingContext, s *SegmentDefinition) {
	slog.Debug("In PersonBusinessProcess::CloseSegment:" + s.Id)
	personSegmentState, _ := pbp.GetSegmentState(s)
	personSegmentState.SegmentState = C_STATE_CLOSED
	personSegmentState.WaitingTaskId = ""
	pbp.SetSegmentState(s, personSegmentState)
	pbp.UpdateFlowState(ec, s)
	ec.ShouldUpdatePersonBusinessProcess = true
	//ec.resourceContext.GetDataStore().UpdatePersonBusinessProcess(pbp.ReferenceNumber, pbp)
}
func (pbp *PersonBusinessProcess) Close() {
	slog.Debug("In PersonBusinessProcess Close")
	pbp.State = C_STATE_CLOSED
	slog.Info(" ############################################")
	slog.Info(" ############################################")
	slog.Info(" ######    Closing Business Process  ########")
	slog.Info(" ############################################")
	slog.Info(" ############################################")
}

const (
	C_STATE_UNREPORTED = iota
	C_STATE_REPORTED
)

func (pbp *PersonBusinessProcess) CreateNewCoveragePeriodFromElections(election commandDataStructures.EnrollmentElection, benefit *benefitPlan.Benefit) personRoles.CoveragePeriod {

	b, _ := datatypes.NewBigFloat("0.0")
	z, _ := datatypes.NewBigFloat(election.CoverageAmount)
	coverageStartDate := benefit.DetermineCoverageStartDate(pbp.EffectiveDate)
	coverageEndDate := benefit.DetermineCoverageEndDate(pbp.EffectiveDate)

	coveragePeriod := personRoles.CoveragePeriod{
		CoverageStartDate:        coverageStartDate,
		CoverageEndDate:          coverageEndDate,
		PayrollReportingState:    C_STATE_UNREPORTED,
		CarrierReportingState:    C_STATE_UNREPORTED,
		ElectedBenefitOfferingId: election.BenefitPlanId,
		ElectedCoverageLevel:     election.CoverageLevelId,
		ActualBenefitOfferingId:  "",
		ActualCoverageLevel:      "",
		OfferedBenefitOfferingId: "",
		ActualCoverageAmount:     b,
		ElectedCoverageAmount:    b,
		EmployeePreTaxCost:       z,
		EmployerCost:             z,
		EmployeeAfterTaxCost:     z,
		EmployerSubsidy:          z,
		LifeImputedIncome:        z,
	}
	f, _ := datatypes.NewBigFloat("10.0")
	coveragePeriod.EmployeePreTaxCost = f
	f, _ = datatypes.NewBigFloat("20.0")
	coveragePeriod.EmployerCost = f
	f, _ = datatypes.NewBigFloat("30.0")
	coveragePeriod.EmployeeAfterTaxCost = f
	f, _ = datatypes.NewBigFloat("40.0")
	coveragePeriod.EmployerSubsidy = f
	f, _ = datatypes.NewBigFloat("50.0")
	coveragePeriod.LifeImputedIncome = f
	return coveragePeriod
}
