/*
Copyright 2018 The Kubernetes Authors.

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

package shareoptions

import (
	"fmt"

	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	volumeutil "k8s.io/kubernetes/pkg/volume/util"
)

// ShareOptions contains options for provisioning and attaching a share
type ShareOptions struct {
	ShareName string

	CommonOptions    // Required common options
	ProtocolOptions  // Protocol specific options
	BackendOptions   // Backend specific options
	OpenStackOptions // OpenStack credentials
}

// Sets a default value in params in case the field `fieldName` is absent.
func setDefaultValue(fieldName string, params map[string]string, defaultValue string) {
	if _, ok := params[fieldName]; !ok {
		params[fieldName] = defaultValue
	}
}

// NewShareOptions creates new share options
func NewShareOptions(volOptions *controller.VolumeOptions, c clientset.Interface) (*ShareOptions, error) {
	params := volOptions.Parameters
	opts := &ShareOptions{}
	nParams := len(params)

	opts.ShareName = "pvc-" + string(volOptions.PVC.GetUID())

	// Set default values

	setDefaultValue("type", params, "default")
	setDefaultValue("zones", params, "nova")

	var (
		n   int
		err error
	)

	// Required common options
	n, err = extractParams(&optionConstraints{}, params, &opts.CommonOptions)
	if err != nil {
		return nil, err
	}
	nParams -= n

	constraints := optionConstraints{protocol: opts.Protocol, backend: opts.Backend}

	// Protocol specific options
	n, err = extractParams(&constraints, params, &opts.ProtocolOptions)
	if err != nil {
		return nil, err
	}
	nParams -= n

	// Backend specific options
	n, err = extractParams(&constraints, params, &opts.BackendOptions)
	if err != nil {
		return nil, err
	}
	nParams -= n

	if nParams != 0 {
		return nil, fmt.Errorf("parameters contain invalid field(s)")
	}

	setOfZones, err := volumeutil.ZonesToSet(opts.Zones)
	if err != nil {
		return nil, err
	}

	opts.Zones = volumeutil.ChooseZoneForVolume(setOfZones, volOptions.PVC.GetName())

	// Retrieve and parse secrets

	err = buildOpenStackOptionsTo(c, &opts.OpenStackOptions, &v1.SecretReference{
		Name:      opts.OSSecretName,
		Namespace: opts.OSSecretNamespace,
	})
	if err != nil {
		return nil, err
	}

	return opts, nil
}
