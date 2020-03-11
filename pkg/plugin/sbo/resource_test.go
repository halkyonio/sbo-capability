package sbo

import (
	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"halkyon.io/api/v1beta1"
	"reflect"
	"testing"
)

func TestCreateServiceLocator(t *testing.T) {
	tests := []struct {
		name       string
		input      v1beta1.NameValuePair
		expected   *v1alpha1.BackingServiceSelector
		shouldFail bool
	}{
		{
			name:  "valid",
			input: v1beta1.NameValuePair{Name: sboTargetedParameterPrefix + "postgresql.baiju.dev:v1alpha1:Database", Value: "db-demo"},
			expected: &v1alpha1.BackingServiceSelector{
				Group:       "postgresql.baiju.dev",
				Version:     "v1alpha1",
				Kind:        "Database",
				ResourceRef: "db-demo",
			},
			shouldFail: false,
		},
		{
			name:       "invalid format",
			input:      v1beta1.NameValuePair{Name: sboTargetedParameterPrefix + "postgresql:Database", Value: "db-demo"},
			expected:   nil,
			shouldFail: true,
		},
		{
			name:       "non SBO parameter",
			input:      v1beta1.NameValuePair{Name: "postgresql.baiju.dev:v1alpha1:Database", Value: "db-demo"},
			expected:   nil,
			shouldFail: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			selector, err := createServiceSelector(test.input)
			if test.shouldFail && err == nil {
				t.Errorf("expected error, got none")
			}
			if !reflect.DeepEqual(test.expected, selector) {
				t.Errorf("expected %v, got %v", test.expected, selector)
			}
		})
	}
}
