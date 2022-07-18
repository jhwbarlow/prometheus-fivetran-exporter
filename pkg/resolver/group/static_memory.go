package group

import (
	"fmt"

	groupLister "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/lister/group"
)

type StaticMemoryLookupResolver struct {
	lookupTableIDtoName map[string]string
	lookupTableNametoID map[string]string
}

func NewStaticMemoryLookupResolver(groupLister groupLister.Lister) (*StaticMemoryLookupResolver, error) {
	groups, err := groupLister.List()
	if err != nil {
		return nil, fmt.Errorf("listing groups: %w", err)
	}

	lookupTableIDToName := make(map[string]string, len(groups))
	lookupTableNameToID := make(map[string]string, len(groups))
	for _, group := range groups {
		lookupTableIDToName[group.ID] = group.Name
		lookupTableNameToID[group.Name] = group.ID
	}

	return &StaticMemoryLookupResolver{
		lookupTableIDtoName: lookupTableIDToName,
		lookupTableNametoID: lookupTableNameToID,
	}, nil
}

func (r *StaticMemoryLookupResolver) ResolveIDToName(groupID string) (string, error) {
	name, ok := r.lookupTableIDtoName[groupID]
	if !ok {
		return "", fmt.Errorf("no static lookup table entry for group ID %q", groupID)
	}

	return name, nil
}

func (r *StaticMemoryLookupResolver) ResolveNameToID(groupName string) (string, error) {
	name, ok := r.lookupTableNametoID[groupName]
	if !ok {
		return "", fmt.Errorf("no static lookup table entry for group name %q", groupName)
	}

	return name, nil
}
