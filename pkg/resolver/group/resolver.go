package group

type Resolver interface {
	ResolveIDToName(groupID string) (string, error)
	ResolveNameToID(groupName string) (string, error)
}
