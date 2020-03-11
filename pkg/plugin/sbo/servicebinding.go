package sbo

import (
	"fmt"
	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest"
	v1beta12 "halkyon.io/api/capability/v1beta1"
	"halkyon.io/api/v1beta1"
	framework "halkyon.io/operator-framework"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ framework.DependentResource = &servicebinding{}
var gvk = v1alpha1.SchemeGroupVersion.WithKind(servicebindingrequest.ServiceBindingRequestKind)

func NewServiceBinding(owner framework.SerializableResource) *servicebinding {
	config := framework.NewConfig(gvk)
	config.CheckedForReadiness = true
	return &servicebinding{BaseDependentResource: framework.NewConfiguredBaseDependentResource(owner, config)}
}

func ownerAsCapability(res framework.DependentResource) *v1beta12.Capability {
	return res.Owner().(*v1beta12.Capability)
}

type servicebinding struct {
	*framework.BaseDependentResource
}

func (s servicebinding) Name() string {
	return framework.DefaultSecretNameFrom(s.Owner())
}

func (s servicebinding) Fetch() (runtime.Object, error) {
	return framework.DefaultFetcher(s)
}

func (s servicebinding) Build(empty bool) (runtime.Object, error) {
	sb := &v1alpha1.ServiceBindingRequest{}
	if !empty {
		c := ownerAsCapability(s)
		sb.ObjectMeta = v1.ObjectMeta{
			Name:      s.Name(),
			Namespace: c.Namespace,
		}
		paramNb := len(c.Spec.Parameters)
		serviceSelectors := make([]v1alpha1.BackingServiceSelector, 0, paramNb)
		var targetDeploymentName string
		targetDeploymentFound := false
		for _, parameter := range c.Spec.Parameters {
			// error is ignored because it should have been raised during validity check
			selector, _ := createServiceSelector(parameter)
			if selector == nil {
				if !targetDeploymentFound && targetDeploymentParamName == parameter.Name {
					targetDeploymentName = parameter.Value
					targetDeploymentFound = true
				}
				continue
			}
			serviceSelectors = append(serviceSelectors, *selector)
		}
		gvr := v12.SchemeGroupVersion.WithResource("deployments")
		sb.Spec = v1alpha1.ServiceBindingRequestSpec{
			BackingServiceSelectors: &serviceSelectors,
			ApplicationSelector: v1alpha1.ApplicationSelector{
				GroupVersionResource: v1.GroupVersionResource{
					Group:    gvr.Group,
					Version:  gvr.Version,
					Resource: gvr.Resource,
				},
				ResourceRef: targetDeploymentName,
			},
		}
	}
	return sb, nil
}

func (s servicebinding) Update(toUpdate runtime.Object) (bool, runtime.Object, error) {
	panic("implement me")
}

func (s servicebinding) GetCondition(underlying runtime.Object, err error) *v1beta1.DependentCondition {
	return framework.DefaultCustomizedGetConditionFor(s, err, underlying, func(underlying runtime.Object, cond *v1beta1.DependentCondition) {
		sb := underlying.(*v1alpha1.ServiceBindingRequest)
		ready := sb.Status.BindingStatus == servicebindingrequest.BindingSuccess
		if !ready {
			cond.Type = v1beta1.DependentFailed
			cond.Message = fmt.Sprintf("%s binding has failed", s.Name())
		} else {
			cond.Type = v1beta1.DependentReady
			cond.Message = fmt.Sprintf("%s is ready", s.Name())
		}
	})
}
