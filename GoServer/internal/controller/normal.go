package controller

import (
	v1 "GoServer/api/v1"
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

type Normal struct{}

func (h *Normal) Normal(ctx context.Context, req *v1.NormalReq) (res *v1.NormalRes, err error) {
	name := req.Name
	g.Log().Infof(ctx, "前端传来的 name: %s", name)

	if name == "" {
		name = "unKnown"
	}
	return &v1.NormalRes{
		CODE:    200,
		Message: "Hello " + name,
	}, nil
}
