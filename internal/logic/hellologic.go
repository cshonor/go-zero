package logic

import (
	"context"

	"gozero/internal/svc"
	"gozero/internal/types"
)

type HelloLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHelloLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HelloLogic {
	return &HelloLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HelloLogic) Hello(req *types.HelloReq) (resp *types.HelloResp, err error) {
	resp = &types.HelloResp{
		Message: "Hello from go-zero!",
	}
	return
}

