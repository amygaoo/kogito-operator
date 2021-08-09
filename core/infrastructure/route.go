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
	if len(host) > 0 {
		if err := ValidateHostName(host); len(err) != 0 {
			r.Log.Error(err[0], "Invalid Custom HostName")
			// TODO: What to do after error
		}

	} else {
		TruncateHostName(&service.ObjectMeta)
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

// TruncateHostname truncates service and route name if name and namespace are longer than DNS1123LabelMaxLength for route creation
func TruncateHostName(ObjectMeta *metav1.ObjectMeta) {
	hostname := ObjectMeta.Name + "-" + ObjectMeta.Namespace
	if extra_characters := len(hostname) - validation.DNS1123LabelMaxLength; extra_characters > 0 && len(ObjectMeta.Name) > extra_characters {
		ObjectMeta.Name = ObjectMeta.Name[:len(ObjectMeta.Name)-extra_characters]
	}
	// TODO: Handle case where len(ObjectMeta.Name) <= extra_characters, then name will be truncated to 0 (not allowed)

}

// From https://github.com/openshift/router/blob/2d1e1f4bd413dd283c92638e23fae940ef4c1e54/pkg/router/controller/unique_host.go
func ValidateHostName(Host string) field.ErrorList {
	result := field.ErrorList{}
	if len(Host) < 1 {
		return result
	}

	specPath := field.NewPath("spec")
	hostPath := specPath.Child("host")

	if len(validation.IsDNS1123Subdomain(Host)) != 0 {
		result = append(result, field.Invalid(hostPath, Host, "host must conform to DNS 952 subdomain conventions"))
	}

	segments := strings.Split(Host, ".")
	for _, s := range segments {
		errs := validation.IsDNS1123Label(s)
		for _, e := range errs {
			result = append(result, field.Invalid(hostPath, Host, e))
		}
	}

	return result
}
