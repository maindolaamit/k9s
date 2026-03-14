// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ Accessor = (*ConfigMap)(nil)

// ConfigMap represents a configmap resource.
type ConfigMap struct {
	Resource
}

// ExtractConfigMap extracts configmap data and binaryData.
func ExtractConfigMap(o runtime.Object) (map[string]interface{}, error) {
	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("expecting *unstructured.Unstructured but got %T", o)
	}
	var cm v1.ConfigMap
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &cm)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})

	// Add regular data (string values)
	if len(cm.Data) > 0 {
		data := make(map[string]string, len(cm.Data))
		for k, v := range cm.Data {
			data[k] = v
		}
		result["data"] = data
	}

	// Add binary data (base64 will be decoded when displayed)
	if len(cm.BinaryData) > 0 {
		binaryData := make(map[string]string, len(cm.BinaryData))
		for k, v := range cm.BinaryData {
			binaryData[k] = string(v)
		}
		result["binaryData"] = binaryData
	}

	return result, nil
}
