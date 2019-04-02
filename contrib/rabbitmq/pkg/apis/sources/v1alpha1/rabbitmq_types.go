/*
Copyright 2019 The Knative Authors

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

package v1alpha1

import (
	"github.com/knative/pkg/apis"
	"github.com/knative/pkg/apis/duck"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RabbitMQSource is the Schema for the RabbitMQsources API.
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:categories=all,knative,eventing,sources
type RabbitMQSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RabbitMQSourceSpec   `json:"spec,omitempty"`
	Status RabbitMQSourceStatus `json:"status,omitempty"`
}

// Check that RabbitMQSource can be validated and can be defaulted.
var _ runtime.Object = (*RabbitMQSource)(nil)

// Check that RabbitMQSource will be checked for immutable fields.
var _ apis.Immutable = (*RabbitMQSource)(nil)

// Check that RabbitMQSource implements the Conditions duck type.
var _ = duck.VerifyType(&RabbitMQSource{}, &duckv1alpha1.Conditions{})

type RabbitMQSourceSASLSpec struct {
	Enable   bool   `json:"enable,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

type RabbitMQSourceTLSSpec struct {
	Enable bool `json:"enable,omitempty"`
}

type RabbitMQSourceNetSpec struct {
	SASL RabbitMQSourceSASLSpec `json:"sasl,omitempty"`
	TLS  RabbitMQSourceTLSSpec  `json:"tls,omitempty"`
}

// RabbitMQSourceSpec defines the desired state of the RabbitMQSource.
type RabbitMQSourceSpec struct {
	// RabbitMQUser is the user name, shall be moved to Net.
	// TODO: shall be moved to Net and enable tsl
	// +required
	RabbitMQUser string `json:"rabbitmq_user"`

	// RabbitMQPassword is the password
	// TODO: shall be moved to Net and enable tsl
	// +required
	RabbitMQPassword string `json:"rabbitmq_password"`

	// ExchangeName.
	// +required
	ExchangeName string `json:"exchange_name"`

	// BootstrapServers
	// +required
	BootstrapServers string `json:"exchange_name"`

	// RabbitMQSourceNetSpec
	// +required
	Net RabbitMQSourceNetSpec `json:"net,omitempty"`

	// Sink is a reference to an object that will resolve to a domain name to use as the sink.
	// +optional
	Sink *corev1.ObjectReference `json:"sink,omitempty"`

	// ServiceAccoutName is the name of the ServiceAccount that will be used to run the Receive
	// Adapter Deployment.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

const (
	// RabbitMQConditionReady has status True when the RabbitMQSource is ready to send events.
	RabbitMQConditionReady = duckv1alpha1.ConditionReady

	// RabbitMQConditionSinkProvided has status True when the RabbitMQSource has been configured with a sink target.
	RabbitMQConditionSinkProvided duckv1alpha1.ConditionType = "SinkProvided"

	// RabbitMQConditionDeployed has status True when the RabbitMQSource has had it's receive adapter deployment created.
	RabbitMQConditionDeployed duckv1alpha1.ConditionType = "Deployed"
)

var RabbitMQSourceCondSet = duckv1alpha1.NewLivingConditionSet(
	RabbitMQConditionSinkProvided,
	RabbitMQConditionDeployed)

// RabbitMQSourceStatus defines the observed state of RabbitMQSource.
type RabbitMQSourceStatus struct {
	// inherits duck/v1alpha1 Status, which currently provides:
	// * ObservedGeneration - the 'Generation' of the Service that was last processed by the controller.
	// * Conditions - the latest available observations of a resource's current state.
	duckv1alpha1.Status `json:",inline"`

	// SinkURI is the current active sink URI that has been configured for the RabbitMQSource.
	// +optional
	SinkURI string `json:"sinkUri,omitempty"`
}

// GetCondition returns the condition currently associated with the given type, or nil.
func (s *RabbitMQSourceStatus) GetCondition(t duckv1alpha1.ConditionType) *duckv1alpha1.Condition {
	return RabbitMQSourceCondSet.Manage(s).GetCondition(t)
}

// IsReady returns true if the resource is ready overall.
func (s *RabbitMQSourceStatus) IsReady() bool {
	return RabbitMQSourceCondSet.Manage(s).IsHappy()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *RabbitMQSourceStatus) InitializeConditions() {
	RabbitMQSourceCondSet.Manage(s).InitializeConditions()
}

// MarkSink sets the condition that the source has a sink configured.
func (s *RabbitMQSourceStatus) MarkSink(uri string) {
	s.SinkURI = uri
	if len(uri) > 0 {
		RabbitMQSourceCondSet.Manage(s).MarkTrue(RabbitMQConditionSinkProvided)
	} else {
		RabbitMQSourceCondSet.Manage(s).MarkUnknown(RabbitMQConditionSinkProvided, "SinkEmpty", "Sink has resolved to empty.%s", "")
	}
}

// MarkNoSink sets the condition that the source does not have a sink configured.
func (s *RabbitMQSourceStatus) MarkNoSink(reason, messageFormat string, messageA ...interface{}) {
	RabbitMQSourceCondSet.Manage(s).MarkFalse(RabbitMQConditionSinkProvided, reason, messageFormat, messageA...)
}

// MarkDeployed sets the condition that the source has been deployed.
func (s *RabbitMQSourceStatus) MarkDeployed() {
	RabbitMQSourceCondSet.Manage(s).MarkTrue(RabbitMQConditionDeployed)
}

// MarkDeploying sets the condition that the source is deploying.
func (s *RabbitMQSourceStatus) MarkDeploying(reason, messageFormat string, messageA ...interface{}) {
	RabbitMQSourceCondSet.Manage(s).MarkUnknown(RabbitMQConditionDeployed, reason, messageFormat, messageA...)
}

// MarkNotDeployed sets the condition that the source has not been deployed.
func (s *RabbitMQSourceStatus) MarkNotDeployed(reason, messageFormat string, messageA ...interface{}) {
	RabbitMQSourceCondSet.Manage(s).MarkFalse(RabbitMQConditionDeployed, reason, messageFormat, messageA...)
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RabbitMQSourceList contains a list of RabbitMQSources.
type RabbitMQSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RabbitMQSource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RabbitMQSource{}, &RabbitMQSourceList{})
}