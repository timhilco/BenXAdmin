package businessProcess

import (
	"benefitsDomain/datatypes"
	"server/message"
)

type BusinessProcessDefinitionMockObjects struct {
}

func NewBusinessProcessDefinitionMock() BusinessProcessDefinitionDataStore {
	return BusinessProcessDefinitionMockObjects{}
}
func (d BusinessProcessDefinitionMockObjects) GetBusinessProcessDefinition(id string) *BusinessProcessDefinition {
	if id == "OE" || id == "BP001" {
		return createOpenEnrollmentProcess()
	}
	if id == "BP002" {
		return createProofOfInsurabilityProcess()
	}
	b := BusinessProcessDefinition{
		InternalId: "B001",
		Name:       "Open Enrollment",
		Label:      "OpenEnrollment",
	}
	return &b

}
func createOpenEnrollmentProcess() *BusinessProcessDefinition {
	b := BusinessProcessDefinition{
		InternalId: "BP001",
		Name:       "Open Enrollment",
		Label:      "OpenEnrollment",
	}
	// Flow
	flow := Flow{
		Id:   "F001",
		Name: "Open Enrollment Flow",
	}
	// Build Segments
	segments := make([]*SegmentDefinition, 0)
	//SegmentDefinition 1
	seg1 := buildSegment_S1()
	segments = append(segments, seg1)
	seg2 := buildSegment_S2()
	segments = append(segments, seg2)
	seg3 := buildSegment_S3()
	segments = append(segments, seg3)
	seg4 := buildSegment_S4()
	segments = append(segments, seg4)
	seg5 := buildSegment_S5()
	preregs := []*SegmentDefinition{seg2, seg3, seg4}
	seg5.PreRegSegments = preregs
	segments = append(segments, seg5)
	flow.SegmentDefinitions = segments
	// Build Graph
	g := SegmentGraph{}
	vMap := make(map[string]*Vertex)
	// Build Edges
	e1 := &Edge{
		Id: 1,
	}
	e2 := &Edge{
		Id: 2,
	}
	e3 := &Edge{
		Id: 3,
	}
	e4 := &Edge{
		Id: 4,
	}
	e5 := &Edge{
		Id: 5,
	}
	e6 := &Edge{
		Id: 6,
	}
	v1 := &Vertex{
		Id:                "V1",
		SegmentDefinition: seg1,
	}
	vMap["V1"] = v1
	v2 := &Vertex{
		Id:                "V2",
		SegmentDefinition: seg2,
	}
	vMap["V2"] = v2
	v3 := &Vertex{
		Id:                "V3",
		SegmentDefinition: seg3,
	}
	vMap["V3"] = v3
	v4 := &Vertex{
		Id:                "V4",
		SegmentDefinition: seg4,
	}
	vMap["V4"] = v4
	v5 := &Vertex{
		Id:                "V5",
		SegmentDefinition: seg5,
	}
	vMap["V5"] = v5
	edges := make(map[int]*Edge)
	edges[1] = e1
	edges[2] = e2
	edges[3] = e3
	v1.Edges = edges
	edges2 := make(map[int]*Edge)
	edges2[1] = e4
	v2.Edges = edges2
	edges3 := make(map[int]*Edge)
	edges3[1] = e5
	v3.Edges = edges3
	edges4 := make(map[int]*Edge)
	edges4[1] = e6
	v4.Edges = edges4
	e1.Vertex = v2
	e2.Vertex = v3
	e3.Vertex = v4
	e4.Vertex = v5
	e5.Vertex = v5
	e6.Vertex = v5

	g.Vertices = vMap
	g.RootVertex = v1
	flow.SegmentFlowGraph = g
	b.Flow = flow
	return &b
}
func buildSegment_S1() *SegmentDefinition {

	electionSegmentDefinition := SegmentDefinition{
		Id:   "S1",
		Name: "Enroll In Benefits",
	}
	segmentTasks := make([]Task, 0)
	compositeTasks := make([]Task, 0)

	startTask := ElementalTask{
		Id:              "S1-T1",
		Name:            "Enrollment Start Task",
		CompletionState: "Started",
		ParentSegment:   &electionSegmentDefinition,
	}
	startTask.SetNormalHandler(TaskHandlerFunc(startTask.openEnrollmentStart))
	compositeTasks = append(compositeTasks, startTask)

	com1Task := ElementalTask{
		Id:              "S1-T2",
		Name:            "Pre Enrollment Communication Task",
		CompletionState: "CommunicationSent",
		ParentSegment:   &electionSegmentDefinition,
	}
	com1Task.SetNormalHandler(TaskHandlerFunc(com1Task.sendEnrollmentCommunication))
	compositeTasks = append(compositeTasks, com1Task)
	compositeTask := CompositeTask{
		Id:             "S1-CT1",
		Name:           "Start Composite Task 1",
		ElementalTasks: compositeTasks,
		ParentSegment:  &electionSegmentDefinition,
	}
	segmentTasks = append(segmentTasks, compositeTask)
	command := message.CreateMockCommandDefinitionObject("PC0001")
	enrollTask := ElementalTask{
		Id:                "S1-T3",
		Name:              "Accept Elections Task",
		ExpectedMessageId: command.ID,
		ParentSegment:     &electionSegmentDefinition,
		reentrant:         true,
	}
	enrollTask.SetNormalHandler(TaskHandlerFunc(enrollTask.acceptEnrollmentElections))
	p := make(map[string]string)
	p["effectiveDate"] = "today"
	p["businessProcessDefinitionId"] = "BP002"
	parameterizedHandle := TaskHandlerParameterizedFunc{
		handler:    enrollTask.triggerBusinessProcess,
		parameters: p,
	}
	exitHandlers := make([]TaskHandler, 0)
	exitHandlers = append(exitHandlers, parameterizedHandle)
	enrollTask.SetExitControlHandler(exitHandlers)
	segmentTasks = append(segmentTasks, enrollTask)
	electionSegmentDefinition.Tasks = segmentTasks
	return &electionSegmentDefinition
}
func buildSegment_S2() *SegmentDefinition {
	seg := &SegmentDefinition{
		Id:   "S2",
		Name: "Release to Payroll",
	}
	event := message.CreateMockEventDefinitionObject("PE0004")
	segmentTasks := make([]Task, 0)
	com1Task := ElementalTask{
		Id:                "S2-T1",
		Name:              "Payroll Report",
		CompletionState:   "ElectionsReleased",
		ParentSegment:     seg,
		ExpectedMessageId: event.ID,
	}
	com1Task.SetNormalHandler(TaskHandlerFunc(com1Task.releaseElections))
	segmentTasks = append(segmentTasks, com1Task)
	seg.Tasks = segmentTasks
	return seg

}
func buildSegment_S3() *SegmentDefinition {
	seg := &SegmentDefinition{
		Id:   "S3",
		Name: "Release to Carrier",
	}
	event := message.CreateMockEventDefinitionObject("PE0002")
	segmentTasks := make([]Task, 0)
	com1Task := ElementalTask{
		Id:                "S3-T1",
		Name:              "Carrier Report",
		CompletionState:   "ElectionsReleased",
		ParentSegment:     seg,
		ExpectedMessageId: event.ID,
	}
	com1Task.SetNormalHandler(TaskHandlerFunc(com1Task.releaseElections))
	segmentTasks = append(segmentTasks, com1Task)
	seg.Tasks = segmentTasks
	return seg

}
func buildSegment_S4() *SegmentDefinition {
	seg := &SegmentDefinition{
		Id:   "S4",
		Name: "Send Post Enrollment Communication",
	}
	theEvent := message.CreateMockEventDefinitionObject("PE0003")
	toEvent := message.CreateMockEventDefinitionObject("PE0003")
	//toEvent := message.CreateMockEventDefinitionObjects("NewDay")

	segmentTasks := make([]Task, 0)
	to := TimeoutDefinition{
		TimeoutPolicyDate: datatypes.YYYYMMDD_Date("2023-09-25"),
		TimeoutEvent:      toEvent,
	}
	com1Task := ElementalTask{
		Id:                "S4-T1",
		Name:              "Post Enrollment Communication Task",
		CompletionState:   "CommunicationSent",
		ParentSegment:     seg,
		ExpectedMessageId: theEvent.ID,
	}
	to.TimeoutHandler = TaskHandlerFunc(com1Task.defaultElections)
	to.TimeOutPolicy = PolicyHandlerFunc(to.matchDate)
	com1Task.SetNormalHandler(TaskHandlerFunc(com1Task.sendEnrollmentCommunication))
	com1Task.TimeoutDefinition = &to
	segmentTasks = append(segmentTasks, com1Task)
	seg.Tasks = segmentTasks
	return seg

}
func buildSegment_S5() *SegmentDefinition {
	closeSegmentDefinition := SegmentDefinition{
		Id:   "S5",
		Name: "SegmentDefinition 5",
	}
	/*
		event2 := message.EventDefinition{
			ID:   "E002",
			Name: "CloseEnrollment",
		}
	*/
	segmentTasks := make([]Task, 0)
	closeEnrollmentTask := ElementalTask{
		Id:            "S5-T1",
		Name:          "Enrollment Finish Task",
		ParentSegment: &closeSegmentDefinition,
	}
	closeEnrollmentTask.SetNormalHandler(TaskHandlerFunc(closeEnrollmentTask.closeBusinessProcess))
	segmentTasks = append(segmentTasks, closeEnrollmentTask)
	/*
		finishTask := ElementalTask{
			Id:   "S2-T2",
			Name: "Enrollment Finish Task",
		}
		tasks = append(tasks, finishTask)
	*/
	closeSegmentDefinition.Tasks = segmentTasks
	return &closeSegmentDefinition
}
func createProofOfInsurabilityProcess() *BusinessProcessDefinition {
	b := BusinessProcessDefinition{
		InternalId: "BP002",
		Name:       "Proof Of Insurability Process",
		Label:      "POI",
	}
	// Flow
	flow := Flow{
		Id:   "F002",
		Name: "POI Flow",
	}
	poiDefinition := SegmentDefinition{
		Id:   "BP-002-S1",
		Name: "BP002-SegmentDefinition 1",
	}
	segmentTasks := make([]Task, 0)

	startTask := ElementalTask{
		Id:              "BP002-S1-T1",
		Name:            "POI Start Task",
		CompletionState: "Started",
		ParentSegment:   &poiDefinition,
	}
	startTask.SetNormalHandler(TaskHandlerFunc(startTask.normalStart))
	segmentTasks = append(segmentTasks, startTask)
	command := message.CreateMockCommandDefinitionObject("PC0002")
	caEvent := message.CreateMockEventDefinitionObject("PE0006")

	cd := CompensationDefinition{
		CompensationEvent: caEvent,
	}

	responseTask := ElementalTask{
		Id:                "BP002-S1-T2",
		Name:              "POI Process Response Task",
		CompletionState:   "Finished",
		ParentSegment:     &poiDefinition,
		ExpectedMessageId: command.ID,
	}
	responseTask.SetNormalHandler(TaskHandlerFunc(responseTask.processPOIResponse))
	cd.CompensationHandler = TaskHandlerFunc(startTask.compensatePOIProcess)
	responseTask.CompensationDefinition = &cd
	segmentTasks = append(segmentTasks, responseTask)

	poiDefinition.Tasks = segmentTasks
	segments := make([]*SegmentDefinition, 0)
	segments = append(segments, &poiDefinition)
	flow.SegmentDefinitions = segments
	// Build Graph
	g := SegmentGraph{}
	vMap := make(map[string]*Vertex)
	// Build Edges
	/*
		e1 := &Edge{
			Id: 1,
		}
	*/

	v1 := &Vertex{
		Id:                "V1",
		SegmentDefinition: &poiDefinition,
	}
	vMap["V1"] = v1

	edges := make(map[int]*Edge)
	//edges[1] = e1

	v1.Edges = edges

	g.Vertices = vMap
	g.RootVertex = v1
	flow.SegmentFlowGraph = g
	b.Flow = flow
	return &b
}
