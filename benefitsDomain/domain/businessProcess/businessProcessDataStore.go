package businessProcess

type BusinessProcessDefinitionDataStore interface {
	GetBusinessProcessDefinition(string) *BusinessProcessDefinition
}
