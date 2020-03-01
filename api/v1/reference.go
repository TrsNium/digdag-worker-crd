package v1

import (
	"context"
	"errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretKeyRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
}

func (r *SecretKeyRef) getSecretValue(clt client.Client) (string, error) {
	ctx := context.Background()
	secretKey := client.ObjectKey{Namespace: r.Namespace, Name: r.Name}
	secret := &corev1.Secret{}
	if err := clt.Get(ctx, secretKey, secret); err != nil {
		if !apierrors.IsNotFound(err) {
			return "", errors.New("secret is not found")
		}
	}

	bytes := secret.Data[r.Key]
	return string(bytes[:]), nil
}

type Ref struct {
	Value           *string       `json:"value,omitempty"`
	ValueFromSecret *SecretKeyRef `json:"valueFromSecretKeyRef,omitempty"`
}

func (r *Ref) GetValue(client client.Client) (string, error) {
	if r.Value != nil {
		return *r.Value, nil
	}
	if r.ValueFromSecret != nil {
		return r.ValueFromSecret.getSecretValue(client)
	}
	return "", errors.New("Should Include value or valueFromSecretKeyRef")
}

type Postgresql struct {
	Host     Ref `json:"host"`
	Port     Ref `json:"port"`
	Database Ref `json:"database"`
	User     Ref `json:"user"`
	Password Ref `json:"password"`
}

type Deployment struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	// +kubebuilder:validation:Minimum=1
	MaxTaskThreads int32 `json:"maxTaskThreads"`
	// +kubebuilder:validation:Minimum=0
	MinReplicas *int32 `json:"minReplicas,omitempty"`
}
