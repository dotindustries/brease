package api

import (
	contextv1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/context/v1"
	"connectrpc.com/connect"
	"context"
	"fmt"
	"go.dot.industries/brease/auth"
)

func (b *BreaseHandler) GetRuleVersions(ctx context.Context, c *connect.Request[contextv1.ListRuleVersionsRequest]) (*connect.Response[contextv1.ListRuleVersionsResponse], error) {
	orgID := CtxString(ctx, auth.ContextOrgKey)

	ctxID := c.Msg.ContextId
	ruleID := c.Msg.RuleId
	pageSize := c.Msg.PageSize
	pageToken := c.Msg.PageToken
	rules, err := b.db.RuleVersions(ctx, orgID, ctxID, ruleID, int(pageSize), pageToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to fetch rule versions: %v", err))
	}
	return &connect.Response[contextv1.ListRuleVersionsResponse]{
		Msg: &contextv1.ListRuleVersionsResponse{
			Rules:         rules,
			NextPageToken: "", // TODO: there's no pagination yet
		},
	}, nil
}
