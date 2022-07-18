package group

import (
	"fmt"

	groupLister "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/lister/group"
)

type DynamicLookupResolver struct {
	GroupLister groupLister.Lister
}

func NewDynamicLookupResolver(groupLister groupLister.Lister) *DynamicLookupResolver {
	return &DynamicLookupResolver{
		GroupLister: groupLister,
	}
}

func (r *DynamicLookupResolver) ResolveIDToName(groupID string) (string, error) {
	groups, err := r.GroupLister.List()
	if err != nil {
		return "", fmt.Errorf("listing groups: %w", err)
	}

	for _, group := range groups {
		if group.ID == groupID {
			return group.Name, nil
		}
	}

	// If we get here, there was no group with an ID matching that provided
	return "", fmt.Errorf("no dynamic lookup entry for group ID %q", groupID)
}

func (r *DynamicLookupResolver) ResolveNameToID(groupName string) (string, error) {
	groups, err := r.GroupLister.List()
	if err != nil {
		return "", fmt.Errorf("listing groups: %w", err)
	}

	for _, group := range groups {
		if group.Name == groupName {
			return group.ID, nil
		}
	}

	// If we get here, there was no group with a name matching that provided
	return "", fmt.Errorf("no dynamic lookup entry for group name %q", groupName)
}
