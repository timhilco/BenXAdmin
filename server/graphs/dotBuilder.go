package graphs

import (
	"benefitsDomain/domain/businessProcess"
	"fmt"

	"strings"
)

func BuildBusinessProcessDefinitionDotString(f *businessProcess.Flow) string {

	//lines := make([]string, 0)
	segmentGraph := f.SegmentFlowGraph
	vertices := segmentGraph.Vertices
	root := segmentGraph.RootVertex
	var sb strings.Builder
	sb.WriteString(" digraph BusinessProcessDefinition { \n")
	sb.WriteString("	fontname=\"Helvetica,Arial,sans-serif\"\n ")
	sb.WriteString("	node [fontname=\"Helvetica,Arial,sans-serif\", height=1.5]\n")
	sb.WriteString("	edge [fontname=\"Helvetica,Arial,sans-serif\",len=2]\n")
	text := "	label = \"Open Enrollment\";\n"
	sb.WriteString(text)
	sb.WriteString("	fontsize=20;\n")
	sb.WriteString("	overlap=scale;\n")
	sb.WriteString("	rankdir=\"LR\";\n")
	// Entity
	sb.WriteString(" 	// Terminal Nodes \n")
	sb.WriteString("	node [shape=circle;style=filled;color=lightgray;height=1;fixedsize=true]; \n")

	sb.WriteString(" 	// Merge Nodes \n")
	sb.WriteString("	node [shape=circle;style=filled;color=yellow;height=.5;fixedsize=true];  \n")

	sb.WriteString(" 	//Task Nodes \n")
	sb.WriteString(" 	node [shape=box;style=rounded;color=blue;fixedsize=false];  \n")
	var tsb strings.Builder
	buildTaskNodes(root, vertices, &tsb)
	sb.WriteString(tsb.String())
	sb.WriteString(" 	//Events \n")
	sb.WriteString(" 	node [shape=box;style=rounded;color=blue;fixedsize=false];  \n")

	sb.WriteString(" 	//Edges \n")
	sb.WriteString(" 	edge [color=black;penwidth=3.0];  \n")
	var tsb2 strings.Builder
	buildEdges(f, vertices, &tsb2)
	sb.WriteString(tsb2.String())

	sb.WriteString("}")
	return sb.String()

}
func buildTaskNodes(root *businessProcess.Vertex, vertices map[string]*businessProcess.Vertex, sb *strings.Builder) {
	segmentTasks := root.SegmentDefinition.Tasks
	for _, v := range segmentTasks {
		switch v := v.(type) {
		case businessProcess.ElementalTask:
			s := buildTaskString(&v)
			sb.WriteString(s)
		case businessProcess.CompositeTask:
			for _, item := range v.ElementalTasks {
				t := item.(businessProcess.ElementalTask)
				s := buildTaskString(&t)
				sb.WriteString(s)
			}

		}
		for _, edge := range root.Edges {
			buildTaskNodes(edge.Vertex, vertices, sb)

		}
	}

}
func buildTaskString(t *businessProcess.ElementalTask) string {

	name := t.GetID()
	name = strings.ReplaceAll(name, "-", "_")
	label := t.GetName()
	var s strings.Builder
	s.WriteString("     " + name)
	s.WriteString("[label = <\n")
	s.WriteString("          <table cellborder=\"0\" style=\"rounded\">\n")
	s.WriteString("             <tr><td>")
	s.WriteString(label)
	s.WriteString("</td></tr>\n")
	s.WriteString("             <hr/>\n")
	s.WriteString("             <tr><td></td></tr>\n")
	s.WriteString("          </table>\n")
	s.WriteString("       > margin=0 shape=none]\n")
	return s.String()

}
func buildEdges(flow *businessProcess.Flow, vertices map[string]*businessProcess.Vertex, sb *strings.Builder) {
	segments := flow.SegmentDefinitions
	for _, segment := range segments {
		buildSegmentTaskEdges(segment, sb)
	}
	buildInnerSegmentTaskEdges(flow, sb)
	/*
		segmentTasks := root.SegmentDefinition.Tasks
		for _, v := range segmentTasks {
			switch v := v.(type) {
			case businessProcess.ElementalTask:
				s := buildTaskString(&v)
				sb.WriteString(s)
			case businessProcess.CompositeTask:
				for _, item := range v.ElementalTasks {
					t := item.(businessProcess.ElementalTask)
					s := buildTaskString(&t)
					sb.WriteString(s)
				}

			}
			for _, edge := range root.Edges {
				buildTaskNodes(edge.Vertex, vertices, sb)

			}
		}


		for _, edge := range root.Edges {
			buildEdges(edge.Vertex, vertices, sb)

		}
	*/
}
func buildSegmentTaskEdges(segment *businessProcess.SegmentDefinition, sb *strings.Builder) {
	tasks := segment.Tasks
	lastIndex := len(tasks) - 1
	for i, leftTask := range tasks {
		if i != lastIndex {
			taskList := make([]businessProcess.ElementalTask, 0)
			switch leftTask := leftTask.(type) {
			case businessProcess.ElementalTask:
				taskList = append(taskList, leftTask)
			case businessProcess.CompositeTask:
				for _, item := range leftTask.ElementalTasks {
					taskList = append(taskList, item.(businessProcess.ElementalTask))
				}
			}
			rightTask := tasks[lastIndex]
			switch rightTask := rightTask.(type) {
			case businessProcess.ElementalTask:
				taskList = append(taskList, rightTask)
			case businessProcess.CompositeTask:
				for _, item := range rightTask.ElementalTasks {
					taskList = append(taskList, item.(businessProcess.ElementalTask))
				}
			}
			lastIndex := len(taskList) - 1
			for j, leftTask := range taskList {
				if j != lastIndex {
					rightTask := taskList[j+1]
					s := buildElementalTaskEdge(leftTask, rightTask)
					sb.WriteString(s)
				}
			}
		}
	}
}
func buildElementalTaskEdge(leftTask businessProcess.ElementalTask, rightTask businessProcess.ElementalTask) string {
	left := leftTask.GetID()
	left = strings.ReplaceAll(left, "-", "_")
	right := rightTask.GetID()
	right = strings.ReplaceAll(right, "-", "_")
	s := fmt.Sprintf("     %s -> %s\n", left, right)
	return s

}
func buildInnerSegmentTaskEdges(flow *businessProcess.Flow, sb *strings.Builder) {

	segmentGraph := flow.SegmentFlowGraph
	vertices := segmentGraph.Vertices
	root := segmentGraph.RootVertex
	vertexPairMap := make(map[string]vertexPair)
	buildEdgePairs(root, vertices, vertexPairMap)
	for _, v := range vertexPairMap {
		leftTasks := v.leftSegment.Tasks
		leftTask := leftTasks[len(leftTasks)-1]
		left := leftTask.GetID()
		left = strings.ReplaceAll(left, "-", "_")
		rightTasks := v.rightSegment.Tasks
		rightTask := rightTasks[len(rightTasks)-1]
		right := rightTask.GetID()
		right = strings.ReplaceAll(right, "-", "_")
		s := fmt.Sprintf("     %s -> %s\n", left, right)
		sb.WriteString(s)
	}
	/*
		for _, v := range segmentTasks {
			switch v := v.(type) {
			case businessProcess.ElementalTask:
				s := buildTaskString(&v)
				sb.WriteString(s)
			case businessProcess.CompositeTask:
				for _, item := range v.ElementalTasks {
					t := item.(businessProcess.ElementalTask)
					s := buildTaskString(&t)
					sb.WriteString(s)
				}

			}
		}
			segments := flow.SegmentDefinitions
			lastIndex := len(segments) - 1
			for i, segment := range segments {
				if i != lastIndex {
				}
			}
	*/
}

type vertexPair struct {
	leftSegment  *businessProcess.SegmentDefinition
	rightSegment *businessProcess.SegmentDefinition
}

func buildEdgePairs(vertex *businessProcess.Vertex, vertices map[string]*businessProcess.Vertex, pm map[string]vertexPair) {

	leftSegment := vertex.SegmentDefinition
	for _, edge := range vertex.Edges {
		completeEdgeNodes(leftSegment, edge.Vertex, vertices, pm)

	}

}
func completeEdgeNodes(leftSegment *businessProcess.SegmentDefinition, vertex *businessProcess.Vertex, vertices map[string]*businessProcess.Vertex, pm map[string]vertexPair) {
	rightSegment := vertex.SegmentDefinition
	key := leftSegment.Id + rightSegment.Id
	vp := vertexPair{
		leftSegment:  leftSegment,
		rightSegment: rightSegment,
	}
	pm[key] = vp
	for _, edge := range vertex.Edges {
		completeEdgeNodes(rightSegment, edge.Vertex, vertices, pm)
	}

}

/*
func getNodeByType(allNodes map[string]GraphNode, nodeType string) []GraphNode {
	returnSlice := make([]GraphNode, 0)
	for _, v := range allNodes {
		if v.GetType() == nodeType {
			returnSlice = append(returnSlice, v)
		}
	}
	return returnSlice

}
func getNodeNames(allNodes map[string]GraphNode, nodeType string) string {
	var sb strings.Builder

	nodes := getNodeByType(allNodes, nodeType)
	for _, node := range nodes {
		text := fmt.Sprintf(" 		%s;\n", node.GetName())
		sb.WriteString(text)
	}

	return sb.String()

}

func getEdgeNames(allEdges map[string]GraphEdge) string {

	var sb strings.Builder

		for _, edge := range allEdges {
			text := fmt.Sprintf("   %s -- %s;\n", edge.fromName, edge.toName)
			sb.WriteString(text)
		}

	for _, edge := range allEdges {
		text := fmt.Sprintf("   %s -- %s;\n", edge.fromName, edge.toName)
		fmt.Println(text)
	}

	sb.WriteString("SPONSOR -- HRIS	;\n")
	sb.WriteString("SPONSOR -- SPONSOR_PRODUCT_GRP ;\n")
	sb.WriteString("SPONSOR -- SPONSOR_CARRIER_CFG ;\n")
	sb.WriteString("HRIS -- PERSON ;\n")
	sb.WriteString("SPONSOR_PRODUCT_GRP -- PRODUCT ;\n")
	sb.WriteString("SPONSOR_CARRIER_CFG -- CARRIER ;\n")
	sb.WriteString("SPONSOR_CARRIER_CFG -- SPONSOR_CARRIER_GROUP;\n")
	sb.WriteString("SPONSOR_CARRIER_GROUP -- SPONSOR_CARRIER_PROFILE ;\n")
	return sb.String()

}
*/
/*
func buildBoundedContextGraph(g *Graph, logger zerolog.Logger, parentChildTable map[string]TableNode, tableMap map[string]Table, leaf TableLeaf, parent string, context *graphProcessingContext) {

	table := leaf.GetTable()
	fullTableName := leaf.GetName()
	fmt.Println(fullTableName)
	var tableName string
	if fullTableName == "ROOT" {
		fullTableName = leaf.GetTable().name
	}

	parts := strings.Split(fullTableName, ".")
	tableName = parts[1]
	context.usedNodeNames = append(context.usedNodeNames, fullTableName)
	ignoreChildren := false
	tableType := table.GetTableType()
	node, err := newGraphNodeFromLeaf(leaf, tableName, tableType)
	if err != nil {
		return
	}
	g.addNode(tableName, node)
	allLeafs := parentChildTable[fullTableName].Leafs
	pks := selectRelevantTables(allLeafs, tableMap, context.usedNodeNames)
	l := len(pks)

	if l == 0 || ignoreChildren {
		return
	}
	for key, leaf := range pks {
		childTable := tableMap[key]
		childTableName := childTable.GetTableNameWithoutPrefix()
		childLeafs := parentChildTable[key].Leafs
		selectedLeafs := selectRelevantTables(childLeafs, tableMap, context.usedNodeNames)
		switch leaf.(type) {
		case referenceLeaf:
		case associationLeaf:
			for _, v := range selectedLeafs {
				associationName := v.GetName()
				childNode, err := newGraphNodeFromLeaf(leaf, associationName, "Primary")
				if err != nil {
					return
				}
				g.addNode(childTableName, childNode)
				addAssociationEdges(g, childTableName, tableMap, childTable)
			}
		case dependentLeaf:
		case rootLeaf:
			logger.Debug().Msg("Error: Hit rootleaf")
		default:
			logger.Debug().Msg("Error: Hit default leaf")
		}

		fmt.Println(selectedLeafs)
		fmt.Println(k)
		fmt.Println(v)
		//cparts := strings.Split(childTableName, ".")
		//cName := cparts[1]
		//edgeName := tableName + ":" + cName
		//g.addEdge(edgeName, tableName, cName)
		childLeafs := parentChildTable[key].leafs
		switch childNode.(type) {
		case AssociationGraphNode:
			addAssociationEdges(g, childTableName, tableMap, childTable)
		case DependentGraphNode:
		case RootGraphNode:
			buildBoundedContextGraph(g, logger, parentChildTable, tableMap, leaf, childTableName, context)
		default:
			// FundamentalGraphNode:
		}
	}

}


func selectRelevantTables(leafs map[string]TableLeaf, tableMap map[string]Table, usedNodeNames []string) map[string]TableLeaf {
	l := make(map[string]TableLeaf)
	for key, leaf := range leafs {
		childTable, ok := tableMap[key]
		if ok {
			childTableName := childTable.GetTableNameWithoutPrefix()
			if !slices.Contains(usedNodeNames, childTableName) {
				l[key] = leaf
			}
		}
	}
	return l
}
func newGraphNodeFromLeaf(leaf TableLeaf, tableName string, tableType string) (GraphNode, error) {
	var node GraphNode
	switch leaf.(type) {
	case rootLeaf:
		node = RootGraphNode{
			name: tableName,
		}

	default:
		switch tableType {
		case "Primary":

			node = FundamentalGraphNode{
				name: tableName,
			}

		case "Dependent_To":

			node = DependentGraphNode{
				name: tableName,
			}

		case "Association_To":
			node = AssociationGraphNode{
				name: tableName,
			}

		default:

			return nil, errors.New("no node created")
		}
	}
	fmt.Println("Node: " + node.String())
	return node, nil
}
func addAssociationEdges(g *Graph, tableName string, allTables map[string]Table, table Table) {
	tables := table.GetCorrespondingAssociationsTables(allTables)
	for _, v := range tables {
		cName := v.GetTableNameWithoutPrefix()
		edgeName := cName + ":" + tableName
		g.addEdge(edgeName, cName, tableName)
	}

}
*/
