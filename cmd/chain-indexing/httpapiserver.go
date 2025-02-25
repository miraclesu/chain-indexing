package main

import (
	"fmt"

	"github.com/lab259/cors"

	"github.com/crypto-com/chain-indexing/appinterface/cosmosapp"
	"github.com/crypto-com/chain-indexing/appinterface/rdb"
	"github.com/crypto-com/chain-indexing/appinterface/tendermint"
	cosmosapp_infrastructure "github.com/crypto-com/chain-indexing/infrastructure/cosmosapp"
	"github.com/crypto-com/chain-indexing/infrastructure/httpapi"
	"github.com/crypto-com/chain-indexing/infrastructure/httpapi/handlers"
	"github.com/crypto-com/chain-indexing/infrastructure/httpapi/routes"
	tendermint_infrastructure "github.com/crypto-com/chain-indexing/infrastructure/tendermint"
	applogger "github.com/crypto-com/chain-indexing/internal/logger"
)

type HTTPAPIServer struct {
	logger           applogger.Logger
	rdbConn          rdb.Conn
	cosmosAppClient  cosmosapp.Client
	tendermintClient tendermint.Client

	validatorAddressPrefix string
	conNodeAddressPrefix   string

	listeningAddress string
	routePrefix      string

	corsAllowedOrigins []string
	corsAllowedMethods []string
	corsAllowedHeaders []string

	pprof DebugConfig
}

// NewIndexService creates a new server instance for polling and indexing
func NewHTTPAPIServer(logger applogger.Logger, rdbConn rdb.Conn, config *Config) *HTTPAPIServer {
	var cosmosClient cosmosapp.Client
	if config.CosmosApp.Insecure {
		cosmosClient = cosmosapp_infrastructure.NewInsecureHTTPClient(
			config.CosmosApp.HTTPRPCUrl,
			config.Blockchain.BondingDenom,
		)
	} else {
		cosmosClient = cosmosapp_infrastructure.NewHTTPClient(
			config.CosmosApp.HTTPRPCUrl,
			config.Blockchain.BondingDenom,
		)
	}

	var tendermintClient tendermint.Client
	if config.Tendermint.Insecure {
		tendermintClient = tendermint_infrastructure.NewInsecureHTTPClient(
			config.Tendermint.HTTPRPCUrl,
			config.Tendermint.StrictGenesisParsing,
		)
	} else {
		tendermintClient = tendermint_infrastructure.NewHTTPClient(
			config.Tendermint.HTTPRPCUrl,
			config.Tendermint.StrictGenesisParsing,
		)
	}
	return &HTTPAPIServer{
		logger:           logger,
		rdbConn:          rdbConn,
		cosmosAppClient:  cosmosClient,
		tendermintClient: tendermintClient,

		validatorAddressPrefix: config.Blockchain.ValidatorAddressPrefix,
		conNodeAddressPrefix:   config.Blockchain.ConNodeAddressPrefix,
		listeningAddress:       config.HTTP.ListeningAddress,
		routePrefix:            config.HTTP.RoutePrefix,

		corsAllowedOrigins: config.HTTP.CorsAllowedOrigins,
		corsAllowedMethods: config.HTTP.CorsAllowedMethods,
		corsAllowedHeaders: config.HTTP.CorsAllowedHeaders,

		pprof: config.Debug,
	}
}

// Run function runs the polling server to index the data from Tendermint
func (server *HTTPAPIServer) Run() error {
	httpServer := httpapi.NewServer(
		server.listeningAddress,
	).WithLogger(
		server.logger,
	)

	if server.pprof.PprofEnable {
		pprofServer := httpapi.NewServer(
			server.pprof.PprofListeningAddress,
		).WithLogger(
			server.logger,
		)
		fixPath := "/debug/pprof"
		pprofServer = pprofServer.WithPprof(fixPath)
		go func() {
			server.logger.Infof("pprof server start listening on: %s%s", server.pprof.PprofListeningAddress, fixPath)
			if err := pprofServer.ListenAndServe(); err != nil {
				panic(fmt.Errorf("error listening and serving HTTP pprof server: %w", err))
			}
		}()
	}

	if len(server.corsAllowedOrigins) != 0 {
		httpServer = httpServer.WithCors(cors.Options{
			AllowedOrigins: server.corsAllowedOrigins,
			AllowedMethods: server.corsAllowedMethods,
			AllowedHeaders: server.corsAllowedHeaders,
			Debug:          true,
		})
	}

	searchHandler := handlers.NewSearch(server.logger, server.rdbConn.ToHandle())
	blocksHandler := handlers.NewBlocks(server.logger, server.rdbConn.ToHandle())
	statusHandler := handlers.NewStatusHandler(server.logger, server.cosmosAppClient, server.rdbConn.ToHandle())
	transactionsHandler := handlers.NewTransactions(server.logger, server.rdbConn.ToHandle())
	blockEventsHandler := handlers.NewBlockEvents(server.logger, server.rdbConn.ToHandle())
	validatorsHandler := handlers.NewValidators(
		server.logger,
		server.validatorAddressPrefix,
		server.conNodeAddressPrefix,
		server.cosmosAppClient,
		server.tendermintClient,
		server.rdbConn.ToHandle(),
	)
	accountTransactionsHandler := handlers.NewAccountTransactions(server.logger, server.rdbConn.ToHandle())
	accountMessagesHandler := handlers.NewAccountMessages(server.logger, server.rdbConn.ToHandle())
	accountsHandler := handlers.NewAccounts(
		server.logger,
		server.rdbConn.ToHandle(),
		server.cosmosAppClient,
		server.validatorAddressPrefix,
	)
	proposalsHandler := handlers.NewProposals(
		server.logger,
		server.rdbConn.ToHandle(),
		server.cosmosAppClient,
	)
	nftsHandler := handlers.NewNFTs(
		server.logger,
		server.rdbConn.ToHandle(),
	)

	routeRegistry := routes.NewRoutesRegistry(
		searchHandler,
		blocksHandler,
		statusHandler,
		transactionsHandler,
		blockEventsHandler,
		validatorsHandler,
		accountTransactionsHandler,
		accountMessagesHandler,
		accountsHandler,
		proposalsHandler,
		nftsHandler,
	)
	routeRegistry.Register(httpServer, server.routePrefix)

	server.logger.Infof("server start listening on: %s", server.listeningAddress)
	if err := httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("error listening and serving HTTP API server: %v", err)
	}

	return nil
}
