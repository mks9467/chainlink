package loader

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/graph-gophers/dataloader"
	"github.com/smartcontractkit/chainlink/core/services/chainlink"
)

type loadersKey struct{}

type Dataloader struct {
	app chainlink.Application

	NodesByChainIDLoader *dataloader.Loader
	ChainsByIDLoader     *dataloader.Loader
}

func New(app chainlink.Application) *Dataloader {
	nodes := &nodeBatcher{app: app}
	chains := &chainBatcher{app: app}

	return &Dataloader{
		app: app,

		NodesByChainIDLoader: dataloader.NewBatchedLoader(nodes.getByChainID),
		ChainsByIDLoader:     dataloader.NewBatchedLoader(chains.getByID),
	}
}

// GetUser wraps the User dataloader for efficient retrieval by user ID
// func (dl *Dataloader) GetNodesByChainID(chainID string) (*types.Node, error) {
// 	thunk := dl.NodesByChainIDLoader.Load(dl.ctx, dataloader.StringKey(chainID))
// 	result, err := thunk()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return result.(*model.User), nil
// }

// Middleware inserts the dataloader into the context
func Middleware(app chainlink.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), loadersKey{}, New(app))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// For returns the dataloader for a given context
func For(ctx context.Context) *Dataloader {
	return ctx.Value(loadersKey{}).(*Dataloader)
}
