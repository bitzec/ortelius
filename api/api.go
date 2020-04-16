// (c) 2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"context"
	"net/http"
	"time"

	"github.com/ava-labs/gecko/utils/logging"
	"github.com/gocraft/web"

	"github.com/ava-labs/ortelius/cfg"
)

const (
	XChainAlias = "x"
	PChainAlias = "p"
	CChainAlias = "c"
)

type Server struct {
	log    logging.Logger
	server *http.Server
}

func NewServer(conf cfg.APIConfig) (*Server, error) {
	log, err := logging.New(conf.Logging)
	if err != nil {
		return nil, err
	}

	router, err := newRouter(conf)
	if err != nil {
		return nil, err
	}

	return &Server{
		log: log,
		server: &http.Server{
			Addr:         conf.ListenAddr,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  15 * time.Second,
			Handler:      router,
		},
	}, err
}

func (s *Server) Listen() error {
	s.log.Info("Server listening on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown() error {
	s.log.Info("Server shutting down")
	ctx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFn()
	return s.server.Shutdown(ctx)
}

func newRouter(conf cfg.APIConfig) (*web.Router, error) {
	// Create a root router that does the work common to all requests and provides
	// chain-agnostic endpoints
	router, err := newRootRouter(conf.ChainAliasConfig)
	if err != nil {
		return nil, err
	}

	// Create routers for the main chains
	if xChainID, ok := conf.ChainAliasConfig[XChainAlias]; ok {
		err = NewAVMRouter(router, conf.ServiceConfig, xChainID, XChainAlias, conf.NetworkID)
		if err != nil {
			return nil, err
		}
	}

	if pChainID, ok := conf.ChainAliasConfig[PChainAlias]; ok {
		err = NewPVMRouter(router, conf.ServiceConfig, pChainID, PChainAlias, conf.NetworkID)
		if err != nil {
			return nil, err
		}
	}

	if cChainID, ok := conf.ChainAliasConfig[CChainAlias]; ok {
		err = NewCVMRouter(router, conf.ServiceConfig, cChainID, CChainAlias, conf.NetworkID)
		if err != nil {
			return nil, err
		}
	}

	return router, nil
}
