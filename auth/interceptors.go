package auth

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

const ClientIDKey = "ClientID"

type authStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (a *authStream) Context() context.Context {
	return a.ctx
}

func UnaryAuthInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	newCtx, err := updateContextWithAuthInfo(ctx)

	if err != nil {
		return nil, err
	}

	return handler(newCtx, req)
}

func StreamAuthInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	newCtx, err := updateContextWithAuthInfo(stream.Context())

	if err != nil {
		return err
	}

	return handler(srv, &authStream{
		ServerStream: stream,
		ctx:          newCtx,
	})
}

func updateContextWithAuthInfo(ctx context.Context) (context.Context, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "unable to get peer from context")
	}

	tlsAuth, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "tls auth info not available")
	}

	if len(tlsAuth.State.VerifiedChains) == 0 || len(tlsAuth.State.VerifiedChains[0]) == 0 {
		return nil, status.Error(codes.PermissionDenied, "could not verify peer certificate")
	}

	crt := tlsAuth.State.VerifiedChains[0][0]

	// TODO: additionally verify CA cert via comparing Issue name of crt
	// and Subject name of a loaded CA cert
	serial := crt.SerialNumber

	ctxWithClientID := context.WithValue(ctx, ClientIDKey, serial)
	return ctxWithClientID, nil
}
