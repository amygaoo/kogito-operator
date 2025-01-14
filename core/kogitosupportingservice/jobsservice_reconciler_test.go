// Copyright 2020 Red Hat, Inc. and/or its affiliates
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

package kogitosupportingservice

import (
	"github.com/kiegroup/kogito-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-operator/core/operator"
	"github.com/kiegroup/kogito-operator/core/test"
	"github.com/kiegroup/kogito-operator/internal/app"
	"github.com/kiegroup/kogito-operator/meta"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReconcileKogitoJobsService_Reconcile(t *testing.T) {
	ns := t.Name()
	jobsService := test.CreateFakeJobsService(ns)
	cli := test.NewFakeClientBuilder().AddK8sObjects(jobsService).Build()
	context := operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: meta.GetRegisteredSchema(),
	}
	r := &jobsServiceSupportingServiceResource{
		supportingServiceContext: supportingServiceContext{
			Context:                  context,
			instance:                 jobsService,
			supportingServiceHandler: app.NewKogitoSupportingServiceHandler(context),
			infraHandler:             app.NewKogitoInfraHandler(context),
			runtimeHandler:           app.NewKogitoRuntimeHandler(context),
		},
	}
	// first reconciliation

	err := r.Reconcile()
	assert.NoError(t, err)
	// second time
	err = r.Reconcile()
	assert.NoError(t, err)

	_, err = kubernetes.ResourceC(cli).Fetch(jobsService)
	assert.NoError(t, err)
	assert.NotNil(t, jobsService.GetStatus())
	assert.Len(t, *jobsService.GetStatus().GetConditions(), 2)
}
