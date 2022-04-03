package rpc

import (
	"context"
	"encoding/base64"
	"net"
	"strings"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"github.com/c13n-io/c13n-go/app"
	pb "github.com/c13n-io/c13n-go/rpc/services"
	"github.com/c13n-io/c13n-go/slog"
)

// Server is the RPC server struct.
type Server struct {
	Log *slog.Logger

	App *app.App

	Listener net.Listener

	grpcCreds credentials.TransportCredentials
	authFunc  grpc_auth.AuthFunc

	// Embedded field
	*grpc.Server
}

// New creates a new instance of Server with all services registered.
func New(address string, app *app.App, options ...func(*Server) error) (*Server, error) {
	var err error

	server := &Server{
		Log: slog.NewLogger("server"),
		App: app,
	}

	for _, option := range options {
		if err := option(server); err != nil {
			return nil, err
		}
	}

	// Create grpc server and register services
	grpcOpts := server.grpcServerOpts(server.authFunc)
	server.Server = grpc.NewServer(grpcOpts...)
	server.registerAllServices()

	// Announce on the specified address
	server.Listener, err = net.Listen("tcp", address)
	if err != nil {
		server.Log.WithError(err).Error("Could not establish listener")
		return nil, err
	}

	return server, nil
}

func (s *Server) grpcServerOpts(authFunc grpc_auth.AuthFunc) []grpc.ServerOption {
	var opts []grpc.ServerOption

	fieldExtractorFunc := grpc_ctxtags.CodeGenRequestFieldExtractor

	// Disable payload logging for levels below debug
	payloadLogServerDecider := func(ctx context.Context, fullMethodName string,
		servingObject interface{}) bool {

		return (slog.LogLevel >= logrus.DebugLevel)
	}

	logrusOpts := []grpc_logrus.Option{
		grpc_logrus.WithLevels(codeToLogLevel),
	}

	authValidator := authFunc
	if authValidator == nil {
		s.Log.Warnf("Server requirement for client authorization disabled")
		authValidator = func(ctx context.Context) (context.Context, error) {
			return ctx, nil
		}
	}

	opts = append(opts,
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(
				grpc_ctxtags.WithFieldExtractor(fieldExtractorFunc),
			),
			grpc_logrus.UnaryServerInterceptor(s.Log, logrusOpts...),
			grpc_auth.UnaryServerInterceptor(authValidator),
			grpc_logrus.PayloadUnaryServerInterceptor(s.Log, payloadLogServerDecider),
			grpc_validator.UnaryServerInterceptor(),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(
				grpc_ctxtags.WithFieldExtractor(fieldExtractorFunc),
			),
			grpc_logrus.StreamServerInterceptor(s.Log, logrusOpts...),
			grpc_auth.StreamServerInterceptor(authValidator),
			grpc_logrus.PayloadStreamServerInterceptor(s.Log, payloadLogServerDecider),
			grpc_validator.StreamServerInterceptor(),
		),
	)

	switch s.grpcCreds {
	case nil:
		s.Log.Warnf("Server certificate requirement for connection disabled")
	default:
		opts = append(opts, grpc.Creds(s.grpcCreds))
	}

	return opts
}

func codeToLogLevel(c codes.Code) logrus.Level {
	// Override default log level for code subset.
	switch c {
	case codes.InvalidArgument, codes.AlreadyExists, codes.NotFound, codes.Internal:
		return logrus.ErrorLevel
	case codes.Unknown:
		return logrus.WarnLevel
	}

	return grpc_logrus.DefaultCodeToLevel(c)
}

func (s *Server) registerAllServices() {
	// Create services
	contacter := NewContactServiceServer(s.App)
	messenger := NewMessageServiceServer(s.App)
	discusser := NewDiscussionServiceServer(s.App)
	channeler := NewChannelServiceServer(s.App)
	nodeInformant := NewNodeInfoServiceServer(s.App)
	financier := NewPaymentServiceServer(s.App)

	// Register services
	pb.RegisterContactServiceServer(s.Server, contacter)
	pb.RegisterMessageServiceServer(s.Server, messenger)
	pb.RegisterDiscussionServiceServer(s.Server, discusser)
	pb.RegisterChannelServiceServer(s.Server, channeler)
	pb.RegisterNodeInfoServiceServer(s.Server, nodeInformant)
	pb.RegisterPaymentServiceServer(s.Server, financier)
}

// WithBasicAuth creates an authorization interceptor with the provided basic auth credentials.
func WithBasicAuth(username, hashedPassword string) func(*Server) error {
	hashedPwd := []byte(hashedPassword)
	return func(server *Server) error {
		f := func(ctx context.Context) (context.Context, error) {
			authError := status.Errorf(codes.Unauthenticated, "Invalid credentials")

			encCreds, err := grpc_auth.AuthFromMD(ctx, "Basic")
			if err != nil {
				return nil, authError
			}
			creds, err := base64.StdEncoding.DecodeString(encCreds)
			if err != nil {
				return nil, status.Errorf(codes.Unauthenticated,
					"Invalid base64 encoded data")
			}

			// Ensure that BasicAuth decoded creds are of the following format
			// username:password
			credsStr := string(creds)
			if !strings.HasPrefix(credsStr, username+":") {
				return nil, authError
			}

			// Check password hash match
			pass := strings.TrimPrefix(credsStr, username+":")
			if err := bcrypt.CompareHashAndPassword(hashedPwd, []byte(pass)); err != nil {
				return nil, authError
			}

			return ctx, nil
		}

		server.authFunc = f
		return nil
	}
}

// WithTLS enables TLS for the Server instance.
func WithTLS(tlsCertPath string, tlsKeyPath string) func(*Server) error {
	return func(server *Server) error {
		creds, err := getCredentials(tlsCertPath, tlsKeyPath)
		if err != nil {
			return errors.Wrap(err, "could not read server certificate")
		}

		server.grpcCreds = creds
		return nil
	}
}

// Cleanup cleans up and terminates the rpc server.
func (s *Server) Cleanup() error {
	return s.App.Cleanup()
}
