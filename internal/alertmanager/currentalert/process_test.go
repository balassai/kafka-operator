// Copyright © 2020 Banzai Cloud
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

package currentalert

import (
	"context"
	"testing"

	"github.com/banzaicloud/kafka-operator/api/v1beta1"
	"github.com/banzaicloud/kafka-operator/pkg/util"
	"github.com/prometheus/common/model"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_addPvc(t *testing.T) {
	testClient := fake.NewFakeClientWithScheme(scheme.Scheme)

	//Setup kafkacluster and pvc
	//TODO separate to different method
	storageResourceQuantity, err := resource.ParseQuantity("10Gi")
	if err != nil {
		t.Error(err)
	}

	err = testClient.Create(context.Background(), &v1beta1.KafkaCluster{
		ObjectMeta: v1.ObjectMeta{
			Name: "kafka",
		},
		Spec: v1beta1.KafkaClusterSpec{
			Brokers: []v1beta1.Broker{
				{
					Id:                0,
					BrokerConfigGroup: "default",
				},
				{
					Id:                1,
					BrokerConfigGroup: "default",
				},
				{
					Id:                2,
					BrokerConfigGroup: "default",
				},
			},
			BrokerConfigGroups: map[string]v1beta1.BrokerConfig{
				"default": {
					StorageConfigs: []v1beta1.StorageConfig{
						{
							MountPath: "/kafka-logs",
							PVCSpec: &corev1.PersistentVolumeClaimSpec{
								AccessModes: []corev1.PersistentVolumeAccessMode{
									"ReadWriteOnce",
								},
								Resources: corev1.ResourceRequirements{
									Requests: map[corev1.ResourceName]resource.Quantity{
										"storage": storageResourceQuantity,
									},
								},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
	}

	err = testClient.Create(context.Background(), &corev1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
			Name:      "testPvc",
			Namespace: "kafka",
			Labels: map[string]string{
				"app":      "kafka",
				"brokerId": "0",
				"kafka_cr": "kafka",
			},
			Annotations: map[string]string{
				"mountPath": "/kafka-logs",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: util.StringPointer("gp2"),
		},
	})
	if err != nil {
		t.Error(err)
	}

	testAlerts := []struct {
		name    string
		alert   model.Alert
		wantErr bool
	}{
		{
			name: "addPvc alert success",
			alert: model.Alert{
				Labels: model.LabelSet{
					"kafka_cr":              "kafka",
					"namespace":             "kafka",
					"persistentvolumeclaim": "testPvc",
				},
				Annotations: model.LabelSet{
					"command":   "addPvc",
					"mountPath": "/kafka-logs",
					"diskSize":  "2G",
				},
			},
			wantErr: false,
		},
		//{
		//	name: "addPvc failed because persistentvolumeclaim label is missing",
		//	alert: model.Alert{
		//		Labels: model.LabelSet{
		//			"kafka_cr":  "kafka",
		//			"namespace": "kafka",
		//		},
		//		Annotations: model.LabelSet{
		//			"command":   "addPvc",
		//			"mountPath": "/kafka-logs",
		//			"diskSize":  "2G",
		//		},
		//	},
		//	wantErr: true,
		//},
	}

	for _, tt := range testAlerts {
		//TODO check stuff got into kafkacluster
		t.Run(tt.name, func(t *testing.T) {
			err := addPvc(tt.alert.Labels, tt.alert.Annotations, testClient)
			if err != nil {
				t.Errorf("process.addPvc() error = %v, wantErr = %v", err, tt.wantErr)
			}

			var kafkaCluster v1beta1.KafkaCluster
			err = testClient.Get(
				context.Background(),
				types.NamespacedName{Namespace: kafkaCluster.Namespace, Name: kafkaCluster.Name},
				&kafkaCluster)
			if err != nil {
				t.Errorf("kafka cr was not found, error = %v", err)
			}

			//TODO remove 0 magic number into variable, make more detailed check for storage etc.
			if kafkaCluster.Spec.Brokers[0].BrokerConfig == nil {
				t.Errorf("BrokerConfig of updated broker should not be nil, error = %v", err)
			}
		})
	}

}

func TestNewBrokerConnection(t *testing.T) {

	//err := client.Create(context.Background(), &corev1.Namespace{
	//	ObjectMeta: v1.ObjectMeta{
	//		Name: "almafa",
	//	},
	//})
	//if err != nil {
	//	t.Error(err)
	//}
	//
	//var namespaces corev1.NamespaceList
	//err = client.List(context.Background(), &namespaces)
	//if err != nil {
	//	t.Error(err)
	//}

}
