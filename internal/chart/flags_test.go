package chart

import (
	"context"
	"reflect"
	"testing"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	testRegistrySecret = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Data: map[string][]byte{
			"username":        []byte("test-username"),
			"password":        []byte("test-password"),
			"registryAddress": []byte("test-registryAddress"),
			"serverAddress":   []byte("test-serverAddress"),
		},
	}
)

func TestBuildFlags(t *testing.T) {
	type args struct {
		ctx        context.Context
		client     client.Client
		serverless *v1alpha1.Serverless
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "build from resource",
			args: args{
				serverless: &v1alpha1.Serverless{
					Spec: v1alpha1.ServerlessSpec{
						DockerRegistry: &v1alpha1.DockerRegistry{
							EnableInternal: pointer.Bool(true),
						},
					},
				},
			},
			want: map[string]interface{}{
				"dockerRegistry": func() map[string]interface{} {
					return map[string]interface{}{
						"enableInternal":  true,
						"registryAddress": "k3d-kyma-registry:5000",
						"serverAddress":   "k3d-kyma-registry:5000",
					}
				}(),
			},
		},
		{
			name: "with secretName",
			args: args{
				ctx:    context.Background(),
				client: fake.NewFakeClient(&testRegistrySecret),
				serverless: &v1alpha1.Serverless{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: testRegistrySecret.Namespace,
					},
					Spec: v1alpha1.ServerlessSpec{
						DockerRegistry: &v1alpha1.DockerRegistry{
							EnableInternal: pointer.Bool(true),
							SecretName:     pointer.String(testRegistrySecret.Name),
						},
					},
				},
			},
			want: map[string]interface{}{
				"dockerRegistry": func() map[string]interface{} {
					return map[string]interface{}{
						"enableInternal":  true,
						"username":        "test-username",
						"password":        "test-password",
						"registryAddress": "test-registryAddress",
						"serverAddress":   "test-serverAddress",
					}
				}(),
			},
		},
		{
			name: "secret not found",
			args: args{
				ctx:    context.Background(),
				client: fake.NewFakeClient(),
				serverless: &v1alpha1.Serverless{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: testRegistrySecret.Namespace,
					},
					Spec: v1alpha1.ServerlessSpec{
						DockerRegistry: &v1alpha1.DockerRegistry{
							EnableInternal: pointer.Bool(true),
							SecretName:     pointer.String(testRegistrySecret.Name),
						},
					},
				},
			},
			wantErr: true,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildFlags(tt.args.ctx, tt.args.client, tt.args.serverless)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}