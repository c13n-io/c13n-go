package rpc

import (
	"github.com/c13n-io/c13n-go/model"
	pb "github.com/c13n-io/c13n-go/rpc/services"
)

func pageOptionsFromKeySet(rpcOpts *pb.KeySetPageOptions) model.PageOptions {
	if rpcOpts == nil {
		return model.PageOptions{}
	}

	return model.PageOptions{
		LastID:   rpcOpts.GetLastId(),
		PageSize: uint64(rpcOpts.GetPageSize()),
		Reverse:  rpcOpts.GetReverse(),
	}
}
