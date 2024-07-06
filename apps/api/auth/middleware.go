package auth

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	"go.dot.industries/brease/worker"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	errors2 "github.com/juju/errors"
	"go.dot.industries/brease/env"
	"go.uber.org/zap"

	unkey "github.com/WilfredAlmeida/unkey-go/features"
)

const (
	ApiKeyHeader     = "X-API-KEY"
	ContextJwtKey    = "jwt"
	ContextUserIDKey = "userId"
	ContextOrgKey    = "orgId"
	bearerAuthPrefix = "Bearer "
)

type validateAuthTokenArgs struct {
	logger     *zap.Logger
	token      string
	rootAPIKey string
	headers    http.Header
}

type validationErr struct {
	Status int
	Error  error
}

type validateAuthTokenResult struct {
	authed        bool
	error         *validationErr
	token         *jwt.Token
	userID        string
	orgID         string
	authenticator string
}

func NewAuthInterceptor(logger *zap.Logger) connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			isClient := req.Spec().IsClient
			if isClient {
				// TODO: client side auth interceptor
				// Send a token with client requests.
				// req.Header().Set(tokenHeader, "sample")
			} else if !strings.Contains(req.Spec().Procedure, "RefreshToken") {
				// server only
				rootAPIKey := env.Getenv("ROOT_API_KEY", "")
				canUseRootAPIKey := rootAPIKey != ""

				if !canUseRootAPIKey {
					logger.Fatal("🔥 ROOT_API_KEY is not specified.")
				}

				authHeader := req.Header().Get(ApiKeyHeader)
				if authHeader == "" {
					authHeader = req.Header().Get("Authorization")
				}
				if authHeader == "" {
					return nil, connect.NewError(
						connect.CodeUnauthenticated,
						fmt.Errorf("API key not set"),
					)
				}

				args := validateAuthTokenArgs{
					logger:     logger,
					rootAPIKey: rootAPIKey,
					token:      authHeader,
					headers:    req.Header(),
				}

				pool := authPool(args)
				pool.Run(ctx)
				authed, valErr, err := getResult(pool, logger)

				if authed == nil {
					logger.Warn("All authenticators failed for request.", zap.Any("validationErr", valErr), zap.Error(err))
					if valErr != nil {
						return nil, connect.NewError(
							connect.CodeUnauthenticated,
							valErr.Error,
						)
					}
					if err != nil {
						return nil, connect.NewError(
							connect.CodeUnauthenticated,
							err,
						)
					}
					// if no errors occurred
					return nil, connect.NewError(
						connect.CodeUnauthenticated,
						fmt.Errorf(""),
					)
				}

				// update context values
				if authed.token != nil {
					ctx = context.WithValue(ctx, ContextJwtKey, authed.token)
				}
				if authed.userID != "" {
					ctx = context.WithValue(ctx, ContextUserIDKey, authed.userID)
				}
				if authed.orgID != "" {
					ctx = context.WithValue(ctx, ContextOrgKey, authed.orgID)
				}
			}

			return next(ctx, req)
		}
	}
	return interceptor
}

func authPool(args validateAuthTokenArgs) worker.WorkerPool {
	pool := worker.New(3)
	pool.GenerateFrom([]worker.Job{
		{
			Descriptor: worker.JobDescriptor{ID: "rootAPIKey"},
			ExecFn:     validateRootAPIKey,
			Args:       args,
		},
		{
			Descriptor: worker.JobDescriptor{ID: "unkey"},
			ExecFn:     validateUnkey,
			Args:       args,
		},
		{
			Descriptor: worker.JobDescriptor{ID: "jwt"},
			ExecFn:     validateJWT,
			Args:       args,
		},
	})
	return pool
}

func getResult(pool worker.WorkerPool, logger *zap.Logger) (authed *validateAuthTokenResult, firstValidationErr *validationErr, firstErr error) {
	for r := range pool.Results() {
		// capture error in case no-authed
		if r.Err != nil && firstErr == nil {
			firstErr = r.Err
		}
		res := r.Value.(validateAuthTokenResult)
		res.authenticator = string(r.Descriptor.ID)

		logger.Debug("Validation result", zap.String("authenticator", string(r.Descriptor.ID)), zap.Bool("success", res.authed))
		if !res.authed && firstValidationErr == nil {
			firstValidationErr = res.error
			continue
		}

		// capture auth success
		if res.authed && authed == nil {
			logger.Debug("Successfully authenticated", zap.String("authenticator", string(r.Descriptor.ID)), zap.String("userId", res.userID), zap.String("orgID", res.orgID))
			authed = &res
		}
	}

	return
}

func validateUnkey(ctx context.Context, args interface{}) (interface{}, error) {
	a := args.(validateAuthTokenArgs)

	key := a.token
	if key == "" {
		key = a.headers.Get(ApiKeyHeader)
	} else {
		key = strings.TrimPrefix(key, bearerAuthPrefix)
	}
	if key == "" {
		return validateAuthTokenResult{
			error: &validationErr{
				Status: http.StatusUnauthorized,
				Error:  errors2.Unauthorizedf("missing API key"),
			},
		}, nil
	}

	resp, err := unkey.KeyVerify(key)
	if err != nil {
		return validateAuthTokenResult{
			error: &validationErr{
				Status: http.StatusUnauthorized,
				Error:  errors2.NewUnauthorized(err, "could not verify API key"),
			},
		}, nil
	}

	if !resp.Valid {
		return validateAuthTokenResult{
			error: &validationErr{
				Status: http.StatusUnauthorized,
				Error:  errors2.Unauthorizedf("invalid API key"),
			},
		}, nil
	}

	userID := ""
	if uid, ok := resp.Meta[ContextUserIDKey]; ok && uid != nil {
		userID, ok = uid.(string)
		if !ok || userID == "" {
			return validateAuthTokenResult{
				error: &validationErr{
					Status: http.StatusUnauthorized,
					Error:  errors2.BadRequestf("invalid API key: missing userID"),
				},
			}, nil
		}
	}

	orgID := ""
	if oid, ok := resp.Meta[ContextOrgKey]; ok && oid != nil {
		orgID, ok = oid.(string)
		if !ok || orgID == "" {
			return validateAuthTokenResult{
				error: &validationErr{
					Status: http.StatusUnauthorized,
					Error:  errors2.BadRequestf("invalid API key: missing orgID"),
				},
			}, nil
		}
	}

	return validateAuthTokenResult{authed: true, userID: userID, orgID: orgID}, nil
}

func validateRootAPIKey(_ context.Context, args interface{}) (interface{}, error) {
	a := args.(validateAuthTokenArgs)

	if a.rootAPIKey == "" || strings.HasPrefix(a.token, "JWT ") {
		// not configured to authenticate, but no errors
		return validateAuthTokenResult{}, nil
	}

	if a.token != a.rootAPIKey {
		return validateAuthTokenResult{
			error: &validationErr{
				Status: http.StatusUnauthorized,
				Error:  errors2.Unauthorizedf("Invalid API key"),
			},
		}, nil
	}

	orgIDHeader := a.headers.Get("x-org-id")
	// org-id header is mandatory for root key access
	if orgIDHeader == "" {
		return validateAuthTokenResult{
			authed: false,
			error: &validationErr{
				Status: http.StatusUnauthorized,
				Error:  errors2.BadRequestf("x-org-id header not set"),
			},
		}, nil
	}

	return validateAuthTokenResult{authed: true, orgID: orgIDHeader}, nil
}

func validateJWT(_ context.Context, args interface{}) (interface{}, error) {
	a := args.(validateAuthTokenArgs)

	if !strings.HasPrefix(a.token, "JWT ") {
		// cannot authenticate, but no errors -- must be root API key
		return validateAuthTokenResult{}, nil
	}

	tokenHandler := NewToken(a.logger)
	apiKey, hasPrefix := strings.CutPrefix(a.token, "JWT ")
	if !hasPrefix {
		// cannot authenticate
		return validateAuthTokenResult{
			error: &validationErr{
				Status: http.StatusUnauthorized,
				Error:  errors2.Unauthorizedf("API key must be a JWT token"),
			},
		}, nil
	}

	token, err := tokenHandler.Parse(apiKey)
	if err != nil {
		return validateAuthTokenResult{
			error: &validationErr{
				Status: http.StatusUnauthorized,
				Error:  errors2.BadRequestf("Invalid JWT: %w", err),
			},
		}, nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return validateAuthTokenResult{
			error: &validationErr{
				Status: http.StatusUnauthorized,
				Error:  errors2.BadRequestf("Invalid JWT"),
			},
		}, nil
	}

	userID, orgID := "", ""

	oid, ok := claims["sub"]
	if !ok {
		return validateAuthTokenResult{
			error: &validationErr{
				Status: http.StatusUnauthorized,
				Error:  errors2.BadRequestf("Invalid JWT: sub missing"),
			},
		}, nil
	}
	orgID = oid.(string)

	uid, ok := claims[ContextUserIDKey]
	if !ok {
		return validateAuthTokenResult{
			error: &validationErr{
				Status: http.StatusUnauthorized,
				Error:  errors2.BadRequestf("Invalid JWT: '%s' missing", ContextUserIDKey),
			},
		}, nil
	}
	userID = uid.(string)

	// FIXME: do we have to look up the token under the orgID to be sure it's valid?

	return validateAuthTokenResult{authed: true, token: token, userID: userID, orgID: orgID}, nil
}
