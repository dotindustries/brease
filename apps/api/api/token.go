package api

import (
	authv1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/auth/v1"
	"connectrpc.com/connect"
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.dot.industries/brease/auth"
)

func (b *BreaseHandler) GetToken(ctx context.Context, c *connect.Request[emptypb.Empty]) (*connect.Response[authv1.TokenPair], error) {
	ownerID := auth.CtxString(ctx, auth.ContextOrgKey)
	userID := auth.CtxString(ctx, auth.ContextUserIDKey)
	permissions := auth.CtxStringArr(ctx, auth.ContextPermissionsKey)
	if !auth.HasPermission(ctx, auth.PermissionCreateRule) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("permission denied"))
	}
	tp, err := b.generateTokenPair(ownerID, userID, auth.ClaimPair[any]{
		Key:   auth.ContextPermissionsKey,
		Value: permissions,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to generate tokens: %w", err))
	}

	if err = b.db.SaveAccessToken(ctx, ownerID, tp); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to save tokens to database: %w", err))
	}

	return connect.NewResponse(tp), nil
}

func (b *BreaseHandler) RefreshToken(ctx context.Context, c *connect.Request[authv1.RefreshTokenRequest]) (*connect.Response[authv1.TokenPair], error) {
	token, err := b.token.Parse(c.Msg.RefreshToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid refreshToken: %w", err))
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid refreshToken"))
	}

	orgID, ok := claims["sub"].(string)
	if !ok {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid refreshToken sub"))
	}

	oldTokenPairs, err := b.db.GetAccessTokens(ctx, orgID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("refreshToken not found"))
	}

	userId, ok := claims[auth.ContextUserIDKey].(string)
	if !ok {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid refreshToken userId"))
	}

	permissions, ok := claims[auth.ContextPermissionsKey].([]string)
	if !ok {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid refreshToken permissions"))
	}

	unknown := true
	for _, tp := range oldTokenPairs {
		if tp.RefreshToken == c.Msg.RefreshToken {
			unknown = false
			break
		}
	}
	if unknown {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown refreshToken"))
	}

	tp, err := b.generateTokenPair(orgID, userId, auth.ClaimPair[any]{
		Key:   auth.ContextPermissionsKey,
		Value: permissions,
	})
	if err != nil {
		return nil, err
	}

	if err = b.db.SaveAccessToken(ctx, orgID, tp); err != nil {
		return nil, fmt.Errorf("failed to save tokens to database: %w", err)
	}

	return connect.NewResponse(tp), nil
}

func (b *BreaseHandler) generateTokenPair(ownerID string, userID string, addClaims ...auth.ClaimPair[any]) (tp *authv1.TokenPair, err error) {
	t, err := b.token.Sign(ownerID, userID, 0, addClaims...)
	if err != nil {
		return tp, fmt.Errorf("failed to generate accessToken: %w", err)
	}

	rt, err := b.token.Sign(ownerID, userID, time.Hour*24, addClaims...)
	if err != nil {
		return tp, fmt.Errorf("failed to generate refreshToken: %w", err)
	}

	tp = &authv1.TokenPair{
		AccessToken:  t,
		RefreshToken: rt,
	}

	return tp, nil
}
