package businessProcess

import (
	"benefitsDomain/datatypes"
	"benefitsDomain/domain/db"
)

type ResourceContext struct {
	creator                            string
	personDataStore                    *db.PersonMongoDB
	businessProcessDataStore           *BusinessProcessMongoDB
	businessProcessDefinitionDataStore BusinessProcessDefinitionDataStore
	eventBroker                        MessageBroker
	planDataStore                      *db.PlanMongoDB
	environmentVariables               datatypes.EnvironmentVariables
}

func (c *ResourceContext) GetPersonDataStore() *db.PersonMongoDB {
	return c.personDataStore
}
func (c *ResourceContext) GetPlanDataStore() *db.PlanMongoDB {
	return c.planDataStore
}
func (c *ResourceContext) GetBusinessProcessStore() *BusinessProcessMongoDB {
	return c.businessProcessDataStore
}
func (c *ResourceContext) GetBusinessProcessDefinitionDataStore() BusinessProcessDefinitionDataStore {
	return c.businessProcessDefinitionDataStore
}
func (c *ResourceContext) GetMessageBroker() MessageBroker {
	return c.eventBroker
}
func (c *ResourceContext) GetEnvironmentVariables() datatypes.EnvironmentVariables {
	return c.environmentVariables
}
func (c *ResourceContext) SetMessageBroker(eb MessageBroker) {
	c.eventBroker = eb
}
func (c *ResourceContext) SetEnvironmentVariables(ev datatypes.EnvironmentVariables) {
	c.environmentVariables = ev
}
func (c *ResourceContext) Close() {
	c.eventBroker.Close()
}
func NewResourceContext(creator string, p *db.PersonMongoDB, bp *BusinessProcessMongoDB, bpd BusinessProcessDefinitionDataStore,
	plan *db.PlanMongoDB, eb MessageBroker, ev datatypes.EnvironmentVariables) *ResourceContext {
	r := &ResourceContext{
		creator:                            creator,
		personDataStore:                    p,
		businessProcessDataStore:           bp,
		businessProcessDefinitionDataStore: bpd,
		eventBroker:                        eb,
		planDataStore:                      plan,
		environmentVariables:               ev,
	}
	return r
}
