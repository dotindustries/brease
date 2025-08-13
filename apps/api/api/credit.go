package api

import (
	"context"

	authv1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/auth/v1"
	"connectrpc.com/connect"
	unkey "github.com/unkeyed/sdks/api/go/v2"
	"github.com/unkeyed/sdks/api/go/v2/models/components"
)

func (b *BreaseHandler) UpdateCredit(ctx context.Context, c *connect.Request[authv1.UpdateCreditRequest]) (*connect.Response[authv1.UpdateCreditResponse], error) {
	uk := unkey.New(
		unkey.WithSecurity(c.Msg.RootKey),
	)
	_, err := uk.Keys.UpdateCredits(ctx, components.V2KeysUpdateCreditsRequestBody{
		KeyID:     c.Msg.TargetKey,
		Value:     &c.Msg.Value,
		Operation: components.Operation(c.Msg.Operation),
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&authv1.UpdateCreditResponse{}), nil
}
