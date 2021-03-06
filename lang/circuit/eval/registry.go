//
// Copyright © 2018 Aljabr, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package eval

import (
	"log"
	"reflect"
	"sync"

	"github.com/kocircuit/kocircuit/lang/go/gate"
)

type Registry struct {
	sync.Mutex `ko:"name=mutex"`
	GateMacro  `ko:"name=gateMacro"`
	Faculty    Faculty `ko:"name=faculty"`
}

type GateMacro interface {
	GateMacro(gate.Gate) Macro
}

func NewRegistry(gateMacro GateMacro) *Registry {
	return &Registry{
		GateMacro: gateMacro,
		Faculty:   Faculty{},
	}
}

// RegisterGate registers a gate with the name and
// package of the given stub.
func (r *Registry) RegisterGate(stub interface{}) {
	gate, err := gate.BindGate(reflect.TypeOf(stub))
	if err != nil {
		panic(err)
	}
	r.RegisterPkgMacro(gate.GoPkgPath(), gate.GoName(), r.GateMacro.GateMacro(gate))
}

// RegisterNamedGate registers a gate with the given name in
// the package of the given stub.
func (r *Registry) RegisterNamedGate(name string, stub interface{}) {
	gate, err := gate.BindGate(reflect.TypeOf(stub))
	if err != nil {
		panic(err)
	}
	r.RegisterPkgMacro(gate.GoPkgPath(), name, r.GateMacro.GateMacro(gate))
}

// RegisterGateAt registers a gate with given name in the given package.
func (r *Registry) RegisterGateAt(pkg, name string, stub interface{}) {
	gate, err := gate.BindGate(reflect.TypeOf(stub))
	if err != nil {
		panic(err)
	}
	r.RegisterPkgMacro(pkg, name, r.GateMacro.GateMacro(gate))
}

func (r *Registry) RegisterMacro(name string, macro Macro) {
	r.RegisterPkgMacro("", name, macro)
}

func (r *Registry) RegisterPkgMacro(pkg, name string, macro Macro) {
	ideal := Ideal{Pkg: pkg, Name: name}
	r.Lock()
	defer r.Unlock()
	if _, ok := r.Faculty[ideal]; ok {
		log.Fatalf("re-registering eval macro %v", ideal)
	}
	r.Faculty[ideal] = macro
}

func (r *Registry) Snapshot() Faculty {
	r.Lock()
	defer r.Unlock()
	g := Faculty{}
	for k, v := range r.Faculty {
		g[k] = v
	}
	return g
}
