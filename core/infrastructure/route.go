// Copyright 2019 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package infrastructure

import (
	"strings"

	"github.com/kiegroup/kogito-operator/api"
	"github.com/kiegroup/kogito-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-operator/core/client/openshift"
	"github.com/kiegroup/kogito-operator/core/operator"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// RouteHandler ...
type RouteHandler interface {
	FetchRoute(key types.NamespacedName) (*routev1.Route, error)
	GetHostFromRoute(routeKey types.NamespacedName) (string, error)
	CreateRoute(instance api.KogitoService, service *corev1.Service) (route *routev1.Route)
}

type routeHandler struct {
	operator.Context
}

// NewRouteHandler ...
func NewRouteHandler(context operator.Context) RouteHandler {
	return &routeHandler{
		context,
	}
}

func (r *routeHandler) FetchRoute(key types.NamespacedName) (*routev1.Route, error) {
	route := &routev1.Route{}
	exists, err := kubernetes.ResourceC(r.Client).FetchWithKey(key, route)
	if err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}
	return route, nil
}

func (r *routeHandler) GetHostFromRoute(routeKey types.NamespacedName) (string, error) {
	route, err := r.FetchRoute(routeKey)
	if err != nil || route == nil {
		return "", err
	}
	return route.Spec.Host, nil
}

// createRequiredRoute creates a new Route resource based on the given Service
func (r *routeHandler) CreateRoute(instance api.KogitoService, service *corev1.Service) (route *routev1.Route) {
	if service == nil || len(service.Spec.Ports) == 0 {
		r.Log.Warn("Impossible to create a Route without a target service")
		return route
	}

	host := instance.GetSpec().GetHost()
	r.Log.Debug("THIS IS HOST", "route", host)

	if err := ValidateHostName(host); len(err) > 0 {
		// host invalid
		r.Log.Warn("Unable to use custom hostname", "error", err.ToAggregate().Error())
		host = ""

	}
	if len(host) == 0 {
		// host empty
		if err := TruncateName(&service.ObjectMeta); err != nil {
			r.Log.Warn("Unable to truncate name for route creation", "error", err)
		}
	}

	route = &routev1.Route{
		ObjectMeta: service.ObjectMeta,
		Spec: routev1.RouteSpec{
			Host: host,
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString(service.Spec.Ports[0].Name),
			},
			To: routev1.RouteTargetReference{
				Kind: openshift.KindService.Name,
				Name: service.Name,
			},
		},
	}

	route.ResourceVersion = ""
	return route
}

// TruncateName truncates service and route name if name and namespace are longer than DNS1123LabelMaxLength
func TruncateName(ObjectMeta *metav1.ObjectMeta) error {
	hostname := ObjectMeta.Name + "-" + ObjectMeta.Namespace
	extra_characters := len(hostname) - validation.DNS1123LabelMaxLength

	if extra_characters > 0 && len(ObjectMeta.Name) > extra_characters {
		ObjectMeta.Name = ObjectMeta.Name[:len(ObjectMeta.Name)-extra_characters]

	} else if extra_characters > 0 {
		// Handle case where len(ObjectMeta.Name) <= extra_characters, then name will be truncated to 0 (not allowed)
		namePath := field.NewPath("name")
		return field.Invalid(namePath, hostname, "length of name and namespace combined must be under 62 characters or provide custom hostname")
	}
	return nil

}

// ValidateRouteHostname validates the hostname provided by the user
// see: https://github.com/openshift/router/blob/release-4.6/pkg/router/controller/unique_host.go#L231
func ValidateHostName(Host string) field.ErrorList {
	result := field.ErrorList{}
	specPath := field.NewPath("spec")
	hostPath := specPath.Child("host")

	if len(Host) < 1 {
		return result
	}

	if len(validation.IsDNS1123Subdomain(Host)) != 0 {
		result = append(result, field.Invalid(hostPath, Host, "host must conform to DNS 952 subdomain conventions"))
	}

	segments := strings.Split(Host, ".")
	for _, s := range segments {
		errs := validation.IsDNS1123Label(s)
		for _, e := range errs {
			result = append(result, field.Invalid(hostPath, s, e))
		}
	}

	return result
}
