package sbo

import (
	"fmt"
	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"halkyon.io/api/capability/v1beta1"
	v1beta12 "halkyon.io/api/v1beta1"
	"halkyon.io/operator-framework"
	"halkyon.io/operator-framework/plugins/capability"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

var _ capability.PluginResource = &Resource{}

const (
	separator                  = ":"
	sboTargetedParameterPrefix = "sbo" + separator
	prefixLength               = len(sboTargetedParameterPrefix)
	parameterFormat            = sboTargetedParameterPrefix + "group" + separator + "version" + separator + "kind"
	targetDeploymentParamName  = "halkyon.target.deployment"
)

type Resource struct {
	capability.SimplePluginResourceStem
}

func (m Resource) CheckValidity(owner framework.SerializableResource) []string {
	c := owner.(*v1beta1.Capability)

	errors := make([]string, 0, len(c.Spec.Parameters))
	hasSBOParams := false
	targetDeploymentFound := false
	for _, parameter := range c.Spec.Parameters {
		selector, err := createServiceSelector(parameter)
		if err != nil {
			errors = append(errors, m.GetPrefixedValidationMessage(err.Error()))
			continue
		}
		if selector != nil {
			hasSBOParams = true
		} else {
			// check if the parameter is the target deployment
			if !targetDeploymentFound && targetDeploymentParamName == parameter.Name {
				targetDeploymentFound = true
			}
		}
	}

	if !hasSBOParams {
		errors = append(errors, m.GetPrefixedValidationMessage("no ServiceBinding parameters were found, define them using "+parameterFormat))
	}

	if !targetDeploymentFound {
		errors = append(errors, m.GetPrefixedValidationMessage("no '"+targetDeploymentParamName+"' parameter was found, needed to identify target deployment in which to inject secret"))
	}

	return errors
}

func createServiceSelector(parameter v1beta12.NameValuePair) (*v1alpha1.BackingServiceSelector, error) {
	prefix := strings.Index(parameter.Name, sboTargetedParameterPrefix)
	if prefix > -1 {
		combinedGVK := parameter.Name[prefix+prefixLength:]
		split := strings.Split(combinedGVK, separator)
		if len(split) != 3 {
			return nil, fmt.Errorf("valid parameter format is %s, got: %s", parameterFormat, combinedGVK)
		}
		return &v1alpha1.BackingServiceSelector{
			GroupVersionKind: v1.GroupVersionKind{
				Group:   split[0],
				Version: split[1],
				Kind:    split[3],
			},
			ResourceRef: parameter.Value,
		}, nil
	}
	return nil, nil
}

func (m Resource) GetDependentResourcesWith(owner framework.SerializableResource) []framework.DependentResource {
	return []framework.DependentResource{NewServiceBinding(owner)}
}

func NewPluginResource() capability.PluginResource {
	return &Resource{capability.NewSimplePluginResourceStem("service binding", capability.TypeInfo{Type: "operator", Versions: []string{"pre-0.24"}})}
}
