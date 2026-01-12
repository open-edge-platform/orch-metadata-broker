/*
* SPDX-FileCopyrightText: (C) 2023 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
 */

package rest

import (
	"context"
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/open-edge-platform/orch-library/go/dazl"
	_ "github.com/open-edge-platform/orch-library/go/dazl/zap"
	ginlogger "github.com/open-edge-platform/orch-library/go/pkg/logging/gin"
	ginmiddleware "github.com/open-edge-platform/orch-library/go/pkg/middleware/gin"
	openapiutils "github.com/open-edge-platform/orch-library/go/pkg/openapi"
	pb "github.com/open-edge-platform/orch-metadata-broker/pkg/api/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"net/http"
	"strings"
)

var log = dazl.GetPackageLogger()

var allowedHeaders = map[string]struct{}{
	"x-request-id": {},
}

func isHeaderAllowed(s string) (string, bool) {
	// check if allowedHeaders contain the header
	if _, isAllowed := allowedHeaders[s]; isAllowed {
		// send uppercase header
		return strings.ToUpper(s), true
	}
	// if not in the allowed header, don't send the header
	return s, false
}

const ActiveProjectID = "ActiveProjectID"

func NewServer(restPort int, grpcPort int, basePath string, allowedCorsOrigins string, openapiSpecFile string) *http.Server {
	gin.DefaultWriter = ginlogger.NewWriter(log)

	// creating mux for gRPC gateway. This will multiplex or route request different gRPC service
	mux := runtime.NewServeMux(
		// convert header in response(going from gateway) from metadata received.
		runtime.WithOutgoingHeaderMatcher(isHeaderAllowed),
		runtime.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
			authHeader := request.Header.Get("Authorization")
			uaHeader := request.Header.Get("User-Agent")
			projectIDHeader := request.Header.Get(ActiveProjectID)
			// send all the headers received from the client
			md := metadata.Pairs("auth", authHeader, "client", uaHeader, ActiveProjectID, projectIDHeader)
			return md
		}),
		runtime.WithRoutingErrorHandler(ginmiddleware.HandleRoutingError),
	)

	// setting up a dail up for gRPC service by specifying endpoint/target url
	err := pb.RegisterMetadataServiceHandlerFromEndpoint(context.Background(), mux, fmt.Sprintf("localhost:%d", grpcPort),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
	if err != nil {
		log.Fatalw("Failed to register MetadataService handler", dazl.Error(err))
	}

	router := gin.New()
	// check if another method is allowed for the current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed' and HTTP status code 405
	// otherwise will return 'Page Not Found' and HTTP status code 404.
	router.HandleMethodNotAllowed = true
	router.Handle("GET", "/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	var msgSizeLimitBytes int64 = 1 * 1024 * 1024

	router.Use(ginlogger.NewGinLogger(log))
	router.Use(secure.New(secure.Config{ContentTypeNosniff: true}))
	router.Use(ginmiddleware.MessageSizeLimiter(msgSizeLimitBytes))
	router.Use(ginmiddleware.UnicodePrintableCharsChecker())
	router.StaticFile(fmt.Sprintf("%smetadata.orchestrator.apis/api/v1", basePath), openapiSpecFile)

	spec, err := openapiutils.LoadOpenAPISpec(openapiSpecFile)
	if err != nil {
		log.Fatalw("Failed to load open API spec", dazl.Error(err))
	}

	// Restrict GET verb for different endpoints of the API
	allPaths := openapiutils.ExtractAllPaths(spec)

	var allowedMethods []string
	for verb := range allPaths {
		allowedMethods = append(allowedMethods, verb)
	}

	corsOrigins := strings.Split(allowedCorsOrigins, ",")
	if len(corsOrigins) > 1 {
		config := cors.DefaultConfig()
		config.AllowOrigins = corsOrigins
		router.Use(cors.New(config))
	}

	router.Group(fmt.Sprintf("%smetadata.orchestrator.apis/v1/*{grpc_gateway}", basePath)).Match(allowedMethods, "", gin.WrapH(mux))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Ok")
	})
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", restPort),
		Handler: router,
	}
}
