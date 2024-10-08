/*
Copyright 2024 The Kubernetes Authors.

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

package compat

import (
	"context"
	"slices"

	"k8s.io/apimachinery/pkg/util/sets"
	"oras.land/oras-go/v2/registry"

	"sigs.k8s.io/node-feature-discovery/pkg/apis/nfd/nodefeaturerule"
	"sigs.k8s.io/node-feature-discovery/source"

	// register sources
	_ "sigs.k8s.io/node-feature-discovery/source/cpu"
	_ "sigs.k8s.io/node-feature-discovery/source/kernel"
	_ "sigs.k8s.io/node-feature-discovery/source/memory"
	_ "sigs.k8s.io/node-feature-discovery/source/network"
	_ "sigs.k8s.io/node-feature-discovery/source/pci"
	_ "sigs.k8s.io/node-feature-discovery/source/storage"
	_ "sigs.k8s.io/node-feature-discovery/source/system"
	_ "sigs.k8s.io/node-feature-discovery/source/usb"
)

type ValidationResult struct {
	RuleName string
	RuleTags []string
	IsValid  bool
}

func ValidateNode(ctx context.Context, ref *registry.Reference, tags sets.Set[string]) ([]ValidationResult, error) {

	spec, err := FetchSpec(ctx, ref)
	if err != nil {
		return nil, err
	}

	sources := source.GetAllFeatureSources()
	for _, s := range sources {
		if err := s.Discover(); err != nil {
			return nil, err
		}
	}

	features := source.GetAllFeatures()

	results := []ValidationResult{}
	for _, c := range spec.Compatibilties {
		if tags.Len() > 0 && len(c.Tags) > 0 {
			exist := slices.ContainsFunc(c.Tags, func(v string) bool {
				return tags.Has(v)
			})
			if !exist {
				continue
			}
		}

		for _, r := range c.Rules {
			out, err := nodefeaturerule.Execute(&r, features)
			if err != nil {
				return nil, err
			}
			results = append(results, ValidationResult{r.Name, c.Tags, out.Valid})
		}
	}

	return results, nil
}
