package rpc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/c13n-io/c13n-go/app"
)

func associateStatusCode(err error) error {
	var e app.Error
	if errors.As(err, &e) {
		switch e.Kind {
		case app.Cancelled:
			return status.Errorf(codes.Canceled, "%v", err)
		case app.DeadlineExceeded:
			return status.Errorf(codes.DeadlineExceeded, "%v", err)
		// Missing app.NetworkError
		case app.PermissionError:
			return status.Errorf(codes.PermissionDenied, "%v", err)
		case app.NoRouteFound, app.ContactNotFound, app.DiscussionNotFound:
			return status.Errorf(codes.NotFound, "%v", err)
		case app.InvalidAddress:
			return status.Errorf(codes.InvalidArgument, "%v", err)
		// Missing app.InsufficientBalance
		case app.ContactAlreadyExists, app.DiscussionAlreadyExists:
			return status.Errorf(codes.AlreadyExists, "%v", err)
		case app.UnknownError:
			return status.Errorf(codes.Unknown, "%v", err)
		case app.InternalError:
			return status.Errorf(codes.Internal, "%v", err)
		default:
			return status.Errorf(codes.Internal, "%v", err)
		}
	}
	if _, ok := status.FromError(err); ok {
		return err
	}
	return status.Errorf(codes.Internal, "%v", err)
}
