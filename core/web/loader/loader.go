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

		NodesByChainIDLoader: dataloader.NewBatchedLoader(nodes.loadByChainIDs),
		ChainsByIDLoader:     dataloader.NewBatchedLoader(chains.loadByIDs),
	}
}

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
