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

package resources

import (
	"fmt"

	"github.com/knative/eventing-sources/contrib/rabbitmq/pkg/apis/sources/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReceiveAdapterArgs is the arguments
type ReceiveAdapterArgs struct {
	Image   string
	Source  *v1alpha1.RabbitMQSource
	Labels  map[string]string
	SinkURI string
}

// MakeReceiveAdapter will create a Deployment with image rabbitmq-receive-adapter
func MakeReceiveAdapter(args *ReceiveAdapterArgs) *v1.Deployment {
	replicas := int32(1)
	return &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    args.Source.Namespace,
			GenerateName: fmt.Sprintf("%s-", args.Source.Name),
			Labels:       args.Labels,
		},
		Spec: v1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: args.Labels,
			},
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "true",
					},
					Labels: args.Labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: args.Source.Spec.ServiceAccountName,
					Containers: []corev1.Container{
						{
							Name:            "rabbitmq-receive-adapter",
							Image:           args.Image,
							ImagePullPolicy: "Always",
							Env: []corev1.EnvVar{
								{
									Name:  "RABBITMQ_USER",
									Value: args.Source.Spec.RabbitMQUser,
								},
								{
									Name:  "RABBITMQ_PASSWORD",
									Value: args.Source.Spec.RabbitMQPassword,
								},
								{
									Name:  "RABBITMQ_BOOTSTRAP_SERVERS",
									Value: args.Source.Spec.BootstrapServers,
								},
								{
									Name:  "RABBITMQ_PORT",
									Value: args.Source.Spec.ConsumerGroup,
								},
								{
									Name:  "EXCHANGE_NAME",
									Value: args.Source.Spec.ExchangeName,
								},
								{
									Name:  "SINK_URI",
									Value: args.SinkURI,
								}
							},
						},
					},
				},
			},
		},
	}
}
