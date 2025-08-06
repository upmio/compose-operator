/*
Copyright 2025 The Compose Operator Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

SPDX-License-Identifier: Apache-2.0
*/

package k8sutil

import (
	"context"

	unitv1alpha2 "github.com/upmio/unit-operator/api/v1alpha2"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IUnitControl defines the interface that uses to update, and get Units.
type IUnitControl interface {

	// UpdateUnit updates a Unit.
	UpdateUnit(*unitv1alpha2.Unit) error

	// GetUnit get Unit.
	GetUnit(namespace, name string) (*unitv1alpha2.Unit, error)
}

type UnitController struct {
	client client.Client
}

// NewUnitController creates a concrete implementation of the
// IUnitControl.
func NewUnitController(cfg *rest.Config) (IUnitControl, error) {
	scheme, err := unitv1alpha2.SchemeBuilder.Build()
	if err != nil {
		return nil, err
	}

	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}
	return &UnitController{client: c}, nil
}

// UpdateUnit implement the IPodControl.Interface.
func (p *UnitController) UpdateUnit(unit *unitv1alpha2.Unit) error {
	return p.client.Update(context.TODO(), unit)
}

// GetUnit implement the IPodControl.Interface.
func (p *UnitController) GetUnit(namespace, name string) (*unitv1alpha2.Unit, error) {
	unit := &unitv1alpha2.Unit{}
	err := p.client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, unit)
	return unit, err
}
