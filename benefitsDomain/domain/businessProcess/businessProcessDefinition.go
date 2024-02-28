package businessProcess

import (
	"benefitsDomain/datatypes"
	"errors"
	"fmt"
	"log/slog"
	"server/message"
	"strings"
)

// Definition:
// -----------
// {IDs,
// Name
// Flow }
type BusinessProcessDefinition struct {
	InternalId string
	Name       string
	Label      string
	Flow       Flow
}

func (d *BusinessProcessDefinition) StartBenefitProcessFlow(ec *MessageProcessingContext, effectiveDate datatypes.YYYYMMDD_Date) {
	slog.Debug("In BusinessProcessDefinition StartPersonBusinessProcess")
	firstSegmentDefinitionDefinition := d.Flow.SegmentDefinitions[0]
	firstSegmentDefinitionDefinition.Start(ec)

}
func (d *BusinessProcessDefinition) GetTaskForSegment(ssId string, tId string) Task {
	var rTask Task
	var segment *SegmentDefinition
	segments := d.Flow.SegmentDefinitions
	for _, item := range segments {
		if item.Id == ssId {
			segment = item
		}
	}
	for _, task := range segment.Tasks {
		if task.GetID() == tId {
			rTask = task
		}
	}
	return rTask

}
func (d *BusinessProcessDefinition) DetermineNextSegment(completedSegment *SegmentDefinition) ([]*SegmentDefinition, string) {

	segments := d.Flow.SegmentDefinitions
	numberOfSegments := len(segments)
	nilSegment := make([]*SegmentDefinition, 0)
	if numberOfSegments == 1 {
		return nilSegment, "Done"
	}
	foundIndex := -1
	for i, item := range segments {
		if item.Id == completedSegment.Id {
			foundIndex = i
		}
	}
	if foundIndex == (numberOfSegments - 1) {
		return nilSegment, "Done"
	}
	rSegments, _ := d.Flow.GetNextSegments(completedSegment)
	return rSegments, ""

}
func (d *BusinessProcessDefinition) GetEventTaskCrossReference() map[message.EventDefinition]Task {
	eventTask := make(map[message.EventDefinition]Task)
	tasks := d.GetAllElementalTasks()
	for _, v := range tasks {
		eId := v.GetExpectedMessageId()
		e := message.CreateMockEventDefinitionObject(eId)
		eventTask[e] = v

	}
	return eventTask
}
func (d *BusinessProcessDefinition) GetAllElementalTasks() map[string]Task {
	taskMap := make(map[string]Task)
	segments := d.Flow.SegmentDefinitions
	for _, segment := range segments {
		for _, task := range segment.Tasks {
			switch task := task.(type) {
			case ElementalTask:
				taskMap[task.GetID()] = task
			case CompositeTask:
				for _, item := range task.ElementalTasks {
					taskMap[task.GetID()] = item

				}
			}
		}
	}
	return taskMap
}

// Graph represents a set of vertices connected by edges.
type SegmentGraph struct {
	Vertices   map[string]*Vertex
	RootVertex *Vertex
}

// Vertex is a node in the graph that stores the int value at that node
// along with a map to the vertices it is connected to via edges.
type Vertex struct {
	Id                string
	SegmentDefinition *SegmentDefinition
	Edges             map[int]*Edge
}

// Edge represents an edge in the graph and the destination vertex.
type Edge struct {
	Id     int
	Vertex *Vertex
}

// Flow
// ------
// {Series of SegmentDefinitions }
type Flow struct {
	Id                 string
	Name               string
	SegmentFlowGraph   SegmentGraph
	SegmentDefinitions []*SegmentDefinition
}

func (f *Flow) GetNextSegments(completedSegment *SegmentDefinition) ([]*SegmentDefinition, string) {
	//Find CompletedVertex
	s := make([]*SegmentDefinition, 0)
	rootVertex := f.SegmentFlowGraph.RootVertex
	verities := f.SegmentFlowGraph.Vertices
	vertex := rootVertex.findVertexForCompletedSegment(completedSegment, verities)
	for _, v := range vertex.Edges {
		s = append(s, v.Vertex.SegmentDefinition)
	}
	return s, ""
}
func (f *Flow) GetSegment(id string) (*SegmentDefinition, error) {

	for _, v := range f.SegmentDefinitions {
		if v.Id == id {
			return v, nil
		}
	}
	return nil, errors.New("not found")
}
func (f *Flow) getFutureSegmentDefinitionsFollowing(sd *SegmentDefinition) []*SegmentDefinition {
	sdsMap := make(map[string]*SegmentDefinition)
	// Find Vertex for SegmentDefinition
	vertexes := f.SegmentFlowGraph.Vertices
	var startingVertex *Vertex
	for _, v := range vertexes {
		if v.SegmentDefinition.Id == sd.Id {
			startingVertex = v
		}

	}
	startingVertex.buildFollowOnVertexes(sdsMap)
	//
	sds := make([]*SegmentDefinition, 0)
	for _, v := range sdsMap {
		sds = append(sds, v)
	}
	return sds

}
func (v *Vertex) buildFollowOnVertexes(sdsMap map[string]*SegmentDefinition) {
	for _, edge := range v.Edges {
		vertex := edge.Vertex
		sdsMap[vertex.SegmentDefinition.Id] = vertex.SegmentDefinition
		vertex.buildFollowOnVertexes(sdsMap)
	}

}
func (v *Vertex) findVertexForCompletedSegment(completedSegment *SegmentDefinition, verities map[string]*Vertex) *Vertex {
	slog.Debug((v.String()))
	if completedSegment.Id == v.SegmentDefinition.Id {
		return v
	}
	count := len(v.Edges)
	if count == 0 {
		return &Vertex{
			Id: "",
		}
	}
	for _, value := range v.Edges {
		vertex := value.Vertex
		r := vertex.findVertexForCompletedSegment(completedSegment, verities)
		if r.Id != "" {
			return r
		}
	}

	return &Vertex{
		Id: "",
	}

}

func (v *Vertex) String() string {
	var s strings.Builder
	s.WriteString("Id: " + v.Id + "-- ")
	s.WriteString("Segment Id: " + v.SegmentDefinition.Id + "-- ")
	for _, v := range v.Edges {
		s.WriteString(v.Vertex.Id + ",")
	}
	return s.String()
}

// SegmentDefinition
// -------
// {Sequence of Tasks}
type SegmentDefinition struct {
	Id             string
	Name           string
	Tasks          []Task
	PreRegSegments []*SegmentDefinition
}

const (
	C_STATE_OPEN = iota
	C_STATE_CLOSED
	C_STATE_REENTRANT
)

func (s *SegmentDefinition) Start(ec *MessageProcessingContext) {
	pbp := ec.PersonBusinessProcess
	ss := SegmentState{
		SegmentDefinitionId: s.Id,
		SegmentState:        C_STATE_OPEN,
	}
	personSegmentStates := pbp.SegmentStates
	if personSegmentStates == nil {
		personSegmentStates = make([]SegmentState, 0)
	}
	personSegmentStates = append(personSegmentStates, ss)
	pbp.SegmentStates = personSegmentStates
	task := s.Tasks[0]
	task.Execute(ec, C_TASK_ACTION_NORMAL)
	pbp.UpdateSegmentState(ec, task, C_TASK_ACTION_NORMAL)
}

func (s *SegmentDefinition) StartSegmentForPersonBusinessProcess(ec *MessageProcessingContext) {
	pbp := ec.PersonBusinessProcess
	pbp.StartSegment(ec, s)

}

/*
	func (s *SegmentDefinition) CloseSegmentForPersonBusinessProcess(ec *MessageProcessingContext) {
		pbp := ec.personBusinessProcess
		pbp.CloseSegment(s,ec.businessProcessDefinition.Flow.SegmentFlowGraph.RootVertex.SegmentDefinition)

}
*/
func (s *SegmentDefinition) DetermineNextTask(ec *MessageProcessingContext, currentTask Task) (Task, string) {
	numberOfTask := len(s.Tasks)
	if numberOfTask == 1 {
		return nil, "Done"
	}
	foundIndex := -1
	for i, item := range s.Tasks {
		if item.GetID() == currentTask.GetID() {
			foundIndex = i
		}
	}
	if foundIndex == (numberOfTask - 1) {
		return nil, "Done"
	}
	task := s.Tasks[foundIndex+1]
	return task, ""
}

// Task
// ----
// Is Unit of Work
// {Name
// Collection  of TaskHandlers}
type Task interface {
	Execute(ec *MessageProcessingContext, processingAction int)
	GetID() string
	GetName() string
	GetParentSegment() *SegmentDefinition
	GetTimeoutDefinition() *TimeoutDefinition
	GetCompensationDefinition() *CompensationDefinition
	GetExpectedMessageId() string
	IsReentrant() bool
	String() string
	Report() string
}
type ElementalTask struct {
	Id                     string
	Name                   string
	ParentSegment          *SegmentDefinition
	NormalHandler          TaskHandler
	ExitControlHandler     []TaskHandler
	CompletionState        string
	ExpectedMessageId      string
	reentrant              bool
	TimeoutDefinition      *TimeoutDefinition
	CompensationDefinition *CompensationDefinition
}
type TimeoutDefinition struct {
	TimeoutHandler    TaskHandler
	TimeOutPolicy     PolicyHandler
	TimeoutPolicyDate datatypes.YYYYMMDD_Date
	TimeoutEvent      message.EventDefinition
}
type CompensationDefinition struct {
	CompensationHandler TaskHandler
	CompensationEvent   message.EventDefinition
}

func (t *ElementalTask) SetNormalHandler(h TaskHandler) {
	t.NormalHandler = h
}
func (t *ElementalTask) SetExitControlHandler(h []TaskHandler) {
	t.ExitControlHandler = h
}
func (t ElementalTask) Execute(ec *MessageProcessingContext, processingAction int) {
	switch processingAction {
	case C_TASK_ACTION_NORMAL,
		C_TASK_ACTION_REENTRANT:
		t.NormalHandler.HandleEvent(ec)
	case C_TASK_ACTION_TIMEOUT:
		t.TimeoutDefinition.TimeoutHandler.HandleEvent(ec)
	case C_TASK_ACTION_COMPENSATION:
		t.CompensationDefinition.CompensationHandler.HandleEvent(ec)
	default:
	}
	//Process Exit Handler
	if t.ExitControlHandler != nil {
		for i, h := range t.ExitControlHandler {
			ec.ProcessingArrayIndex = i
			h.HandleEvent(ec)
		}
	}
}
func (t ElementalTask) GetID() string {
	return t.Id
}
func (t ElementalTask) GetName() string {
	return t.Name
}
func (t ElementalTask) GetParentSegment() *SegmentDefinition {
	return t.ParentSegment
}
func (t ElementalTask) GetExpectedMessageId() string {
	return t.ExpectedMessageId
}
func (t ElementalTask) GetTimeoutDefinition() *TimeoutDefinition {
	return t.TimeoutDefinition
}
func (t ElementalTask) GetCompensationDefinition() *CompensationDefinition {
	return t.CompensationDefinition
}
func (t ElementalTask) IsReentrant() bool {
	return t.reentrant
}
func (t ElementalTask) String() string {
	var sb strings.Builder
	sb.WriteString(t.Id + ":")
	sb.WriteString(t.Name + ":")
	sb.WriteString(t.ExpectedMessageId + ":")
	if t.reentrant {
		sb.WriteString("ReEntrant:")
	}
	if t.TimeoutDefinition != nil {
		sb.WriteString("Timeout Event ID -" + t.TimeoutDefinition.TimeoutEvent.ID + ":")
	}

	return sb.String()

}
func (t ElementalTask) Report() string {
	var sb strings.Builder
	sb.WriteString("Id: " + t.Id + "\n")
	sb.WriteString("Name: " + t.Name + "\n")
	sb.WriteString("Expected Message Id: " + t.ExpectedMessageId + "\n")
	//s := fmt.Sprintf("%v", t.NormalHandler)
	//sb.WriteString("NormalHandler: " + s + "\n")
	sb.WriteString("Parent Segment: " + t.ParentSegment.Id + "-" + t.ParentSegment.Name + "\n")
	if t.reentrant {
		sb.WriteString("ReEntrant\n")
	}
	if t.TimeoutDefinition != nil {
		sb.WriteString("Timeout Event ID -" + t.TimeoutDefinition.TimeoutEvent.ID + "\n")
	}

	return sb.String()

}

type CompositeTask struct {
	Id                     string
	Name                   string
	ParentSegment          *SegmentDefinition
	ElementalTasks         []Task
	ExpectedMessageId      string
	reentrant              bool
	TimeoutDefinition      *TimeoutDefinition
	CompensationDefinition *CompensationDefinition
}

func (t CompositeTask) Execute(ec *MessageProcessingContext, processingAction int) {
	for _, t := range t.ElementalTasks {
		t.Execute(ec, processingAction)
	}
}
func (t CompositeTask) GetID() string {
	return t.Id
}
func (t CompositeTask) GetName() string {
	return t.Name
}
func (t CompositeTask) GetParentSegment() *SegmentDefinition {
	return t.ParentSegment
}
func (t CompositeTask) GetExpectedMessageId() string {
	return t.ExpectedMessageId
}
func (t CompositeTask) GetTimeoutDefinition() *TimeoutDefinition {
	return t.TimeoutDefinition
}
func (t CompositeTask) GetCompensationDefinition() *CompensationDefinition {
	return t.CompensationDefinition
}
func (t CompositeTask) IsReentrant() bool {
	return t.reentrant
}
func (t CompositeTask) String() string {
	var sb strings.Builder
	sb.WriteString("Composite Task :")

	for _, t := range t.ElementalTasks {
		sb.WriteString(":" + t.String())

	}
	return sb.String()
}
func (t CompositeTask) Report() string {
	var sb strings.Builder
	sb.WriteString("Composite Task :\n")

	for _, t := range t.ElementalTasks {
		sb.WriteString(t.Report() + "\n")

	}
	return sb.String()

}

// TaskWorker
// ----------
// Type: Normal, Timeout, Compensation
// { Event
// BusinessLogic NormalHandler (Single or Composite)
// Exit Control Action Handler ( Trigger, Conditional, Break) }

//TaskHandler
//-----------
// Is Business Logic or Exit Control Action
// script or go Function

type TaskHandler interface {
	HandleEvent(ec *MessageProcessingContext)
}
type TaskHandlerFunc func(ec *MessageProcessingContext)

func (f TaskHandlerFunc) HandleEvent(ec *MessageProcessingContext) {
	f(ec)

}

type TaskHandlerParameterizedFunc struct {
	handler    func(ec *MessageProcessingContext, parameters map[string]string)
	parameters map[string]string
}

func (f TaskHandlerParameterizedFunc) HandleEvent(ec *MessageProcessingContext) {
	f.handler(ec, f.parameters)

}

//PolicyHandler
//-----------

func (to *TimeoutDefinition) EvaluatePolicy(ec MessageProcessingContext) bool {
	b := to.TimeOutPolicy.EvaluatePolicy(&ec)
	return b
}
func CanProcessMessageForTask(e message.Message, ss SegmentState, task Task) (bool, int) {
	segmentState := ss.SegmentState
	data := fmt.Sprintf("Message: %s - Task: %s - SegmentState: %d", e.String(), task.String(), segmentState)
	slog.Debug("Message::CanProcessForTask - " + data)

	b := false
	a := C_TASK_ACTION_UNKNOWN
	if segmentState == C_STATE_OPEN {
		b, a = checkForNormalMessage(e, task)
		if b {
			return b, a
		}
	}
	b, a = checkForTimeoutMessage(e, task)
	if b {
		return b, a
	}
	b, a = checkForReentrantTask(e, task)
	if b {
		return b, a
	}
	b, a = checkForCompensationTask(e, task)
	if b {
		return b, a
	}

	data = fmt.Sprintf("Exit No Match - Event: %s - Task Expected: %s", e.String(), task.GetExpectedMessageId())
	slog.Debug("Message::CanProcessForTask - " + data)
	return false, C_TASK_ACTION_UNKNOWN
}
func checkForNormalMessage(e message.Message, task Task) (bool, int) {
	waitingMessageId := task.GetExpectedMessageId()
	if e.GetMessageDfnId() != waitingMessageId {
		return false, C_TASK_ACTION_NORMAL
	}
	return true, C_TASK_ACTION_NORMAL

}
func checkForTimeoutMessage(e message.Message, task Task) (bool, int) {
	timeoutDefinition := task.GetTimeoutDefinition()
	if timeoutDefinition == nil {
		return false, C_TASK_ACTION_TIMEOUT
	} else {
		if e.GetMessageId() != timeoutDefinition.TimeoutEvent.ID {
			return false, C_TASK_ACTION_TIMEOUT
		} else {
			return true, C_TASK_ACTION_TIMEOUT
		}
	}

}
func checkForCompensationTask(e message.Message, task Task) (bool, int) {
	compensationDefinition := task.GetCompensationDefinition()
	if compensationDefinition == nil {
		return false, C_TASK_ACTION_COMPENSATION
	} else {
		eid := e.GetMessageDfnId()
		if eid != compensationDefinition.CompensationEvent.ID {
			return false, C_TASK_ACTION_COMPENSATION
		} else {
			return true, C_TASK_ACTION_COMPENSATION
		}
	}

}
func checkForReentrantTask(e message.Message, task Task) (bool, int) {
	if !task.IsReentrant() {
		return false, C_TASK_ACTION_REENTRANT
	}
	waitingMessageId := task.GetExpectedMessageId()
	if e.GetMessageId() != waitingMessageId {
		return false, C_TASK_ACTION_REENTRANT
	}
	return true, C_TASK_ACTION_REENTRANT

}
