// (c) 2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"encoding/json"
	"fmt"

	"github.com/gocraft/web"

	"github.com/ava-labs/ortelius/cfg"
)

type index struct {
	Chains cfg.ChainsConfig
}

type RootRequestContext struct{}

func newRootRouter(chainsConf cfg.ChainsConfig) (*web.Router, error) {
	indexBytes, err := json.Marshal(&index{Chains: chainsConf})
	if err != nil {
		return nil, err
	}

	router := web.New(RootRequestContext{}).
		Middleware((*RootRequestContext).setHeaders).
		NotFound((*RootRequestContext).notFoundHandler).
		Get("/", func(resp web.ResponseWriter, _ *web.Request) {
			if _, err := resp.Write(indexBytes); err != nil {
				// TODO: Write to log
				fmt.Println("Err:", err.Error())
			}
		})

	return router, nil
}

func (*RootRequestContext) setHeaders(w web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	h := w.Header()
	h.Add("access-control-allow-headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
	h.Add("access-control-allow-methods", "GET")
	h.Add("access-control-allow-origin", "*")

	h.Add("Content-Type", "application/json")

	next(w, r)
}

func (*RootRequestContext) notFoundHandler(w web.ResponseWriter, r *web.Request) {
	writeErr(w, 404, "Not Found")
}
