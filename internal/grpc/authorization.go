package grpc

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/open-edge-platform/orch-library/go/pkg/errors"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/open-edge-platform/orch-library/go/pkg/openpolicyagent"
	"google.golang.org/grpc/metadata"
)

func (s *Server) authCheckAllowed(ctx context.Context, request string) error {
	if s.opaClient == nil {
		log.Debugf("ignoring Authorization")
		return nil
	}
	md, ok := metadata.FromIncomingContext(ctx)

	log.Debugf("Checking auth for request %s with metadata: %+v\n", request, md)

	if !ok {
		return errors.NewInvalid("authentication failed") // errors.NewInvalidArgument(errors.WithMessage("authentication failed"))
	}
	opaInputStruct := openpolicyagent.OpaInput{
		Input: map[string]interface{}{
			"request":  emptypb.Empty{},
			"metadata": md,
		},
	}

	// can safely ignore the JSON error - will not happen with OPA data
	completeInputJSON, _ := json.Marshal(opaInputStruct)

	bodyReader := bytes.NewReader(completeInputJSON)

	requestPackage := request[0:strings.LastIndex(request, ".")]
	requestName := request[strings.LastIndex(request, ".")+1:]

	trueBool := true
	resp, err := s.opaClient.PostV1DataPackageRuleWithBodyWithResponse(
		ctx,
		requestPackage,
		requestName,
		&openpolicyagent.PostV1DataPackageRuleParams{
			Pretty:  &trueBool,
			Metrics: &trueBool,
		},
		"application/json",
		bodyReader)
	if err != nil {
		return errors.NewInternal("unable to reach Open Policy Agent")
	}

	resultBool, boolErr := resp.JSON200.Result.AsOpaResponseResult1()
	if boolErr != nil {
		resultObj, objErr := resp.JSON200.Result.AsOpaResponseResult0()
		if objErr != nil {
			log.Debugf("access denied by OPA rule %s %v", requestName, objErr)
			return errors.NewForbidden("access denied by OPA rule %s %v", requestName, objErr)
		}
		log.Debugf("access denied by OPA rule %s %v", requestName, resultObj)
		return errors.NewForbidden("access denied by OPA rule %s %v", requestName, objErr)

	}
	if resultBool {
		log.Infof("%s Authorized", requestName)
		return nil
	}

	log.Debugf("access denied by OPA rule %s. OPA response %d %v", requestName, resp.StatusCode(), resp.HTTPResponse)
	return errors.NewForbidden("access denied by OPA rule %s. OPA response %d %v", requestName, resp.StatusCode(), resp.HTTPResponse)
}
