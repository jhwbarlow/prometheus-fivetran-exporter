package group

import (
	"fmt"
)

type Resolver interface {
	ResolveNameToID(groupName string) (string, error)
}

type GroupListerResolver struct {
	GroupLister Lister
}

func NewGroupListerResolver(groupLister Lister) *GroupListerResolver {
	return &GroupListerResolver{
		GroupLister: groupLister,
	}
}

func (r *GroupListerResolver) ResolveNameToID(groupName string) (string, error) {
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
	return "", fmt.Errorf("no entry for group name %q", groupName)
}
