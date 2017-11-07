/*
Copyright 2014 The Kubernetes Authors.

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

package admission

import "time"

// chainAdmissionHandler is an instance of admission.Interface that performs admission control using
// a chain of admission handlers
type chainAdmissionHandler []NamedHandler

// NewChainHandler creates a new chain handler from an array of handlers. Used for testing.
func NewChainHandler(handlers ...NamedHandler) chainAdmissionHandler {
	return chainAdmissionHandler(handlers)
}

const (
	stepValidating = "validating"
	stepMutating   = "mutating"
)

// Admit performs an admission control check using a chain of handlers, and returns immediately on first error
func (admissionHandler chainAdmissionHandler) Admit(a Attributes) error {
	var err error
	start := time.Now()
	defer func() {
		Metrics.ObserveAdmissionStep(time.Since(start), err != nil, a, stepMutating)
	}()

	for _, handler := range admissionHandler {
		if !handler.Handles(a.GetOperation()) {
			continue
		}
		if mutator, ok := handler.(MutationInterface); ok {
			t := time.Now()
			err = mutator.Admit(a)
			Metrics.ObserveAdmissionController(time.Since(t), err != nil, handler, a)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Validate performs an admission control check using a chain of handlers, and returns immediately on first error
func (admissionHandler chainAdmissionHandler) Validate(a Attributes) error {
	var err error
	start := time.Now()
	defer func() {
		Metrics.ObserveAdmissionStep(time.Since(start), err != nil, a, stepValidating)
	}()

	for _, handler := range admissionHandler {
		if !handler.Handles(a.GetOperation()) {
			continue
		}
		if validator, ok := handler.(ValidationInterface); ok {
			t := time.Now()
			err = validator.Validate(a)
			Metrics.ObserveAdmissionController(time.Since(t), err != nil, handler, a)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Handles will return true if any of the handlers handles the given operation
func (admissionHandler chainAdmissionHandler) Handles(operation Operation) bool {
	for _, handler := range admissionHandler {
		if handler.Handles(operation) {
			return true
		}
	}
	return false
}
