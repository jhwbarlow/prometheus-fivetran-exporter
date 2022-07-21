package group

import (
	"fmt"

	"go.uber.org/zap"
)

type Resolver interface {
	ResolveNameToID(groupName string) (string, error)
}

type GroupListerResolver struct {
	GroupLister Lister
	logger      *zap.SugaredLogger
}

func NewGroupListerResolver(logger *zap.SugaredLogger, groupLister Lister) *GroupListerResolver {
	logger = getComponentLogger(logger, "group_lister_resolver")

	return &GroupListerResolver{
		GroupLister: groupLister,
		logger:      logger,
	}
}

func (r *GroupListerResolver) ResolveNameToID(groupName string) (string, error) {
	groups, err := r.GroupLister.List()
	if err != nil {
		r.logger.Errorw("listing groups", "group_name", groupName, "error", err)
		return "", fmt.Errorf("listing groups for group name %q: %w", groupName, err)
	}

	for _, group := range groups {
		if group.Name == groupName {
			id := group.ID
			r.logger.Infow("resolved group name to ID", "group_name", groupName, "id", id)
			return id, nil
		}
	}

	// If we get here, there was no group with a name matching that provided
	r.logger.Errorw("no entry for group name", "group_name", groupName)
	return "", fmt.Errorf("no entry for group name %q", groupName)
}
