/*
Copyright 2021 NDDO.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package configgnmihandler

import (
	"context"
	"time"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yndd/cache/pkg/validator"
	"github.com/yndd/ndd-runtime/pkg/meta"
	"github.com/yndd/ndd-yang/pkg/yparser"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// errors
	errTargetNotFoundInCache = "could not find target in cache"
)

func (s *subServer) Set(ctx context.Context, pf *gnmi.Path, upd *gnmi.Update) (*gnmi.SetResponse, error) {
	if pf == nil {
		pf = &gnmi.Path{}
	}
	path := &gnmi.Path{}
	if upd.GetPath() != nil {
		path = upd.GetPath()
	}
	origin := getOrigin(pf, path)

	cacheNsTargetName := meta.NamespacedName(pf.GetTarget()).GetPrefixNamespacedName(origin)
	log := s.log.WithValues("origin", origin, "target", pf.GetTarget(), "cacheNsTargetName", cacheNsTargetName)

	log.Debug("Set update.", "upd", upd)

	ce, err := s.cache.GetEntry(cacheNsTargetName)
	if err != nil {
		log.Debug("Set Update/Replace cache entry node found", "error", err)
		return nil, status.Errorf(codes.NotFound, errTargetNotFoundInCache)
	}

	// validates and updates the running config
	if err := validator.ValidateUpdate(ce, []*gnmi.Update{upd}, true, false, validator.Origin_GnmiServer); err != nil {
		log.Debug("Set Update/Replace validate update failed", "error", err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	if err := ce.SetSystemCacheStatus(true); err != nil {
		log.Debug("Set system cache status failed", "error", err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &gnmi.SetResponse{
		Response: []*gnmi.UpdateResult{
			{
				Timestamp: time.Now().UnixNano(),
				Path:      path,
				Op:        gnmi.UpdateResult_UPDATE,
			},
		}}, nil
}

func (s *subServer) Delete(ctx context.Context, pf *gnmi.Path, del *gnmi.Path) (*gnmi.SetResponse, error) {
	if pf == nil {
		pf = &gnmi.Path{}
	}
	path := &gnmi.Path{}
	if del != nil {
		path = del
	}
	origin := getOrigin(pf, path)

	cacheNsTargetName := meta.NamespacedName(pf.GetTarget()).GetPrefixNamespacedName(origin)
	log := s.log.WithValues("origin", origin, "target", pf.GetTarget(), "cacheNsTargetName", cacheNsTargetName)

	log.Debug("Set Delete...", "path", yparser.GnmiPath2XPath(del, true))

	return nil, status.Errorf(codes.Unimplemented, "not implemented")

	/*
		ce, err := s.cache.GetEntry(cacheNsTargetName)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, errTargetNotFoundInCache)
		}

		if err := ce.SetSystemCacheStatus(true); err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}

		return &gnmi.SetResponse{
			Response: []*gnmi.UpdateResult{
				{
					Timestamp: time.Now().UnixNano(),
				},
			}}, nil
	*/
}
