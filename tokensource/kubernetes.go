package tokensource

import (
	"context"
	"fmt"
	v1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type KubernetesTokenSource struct {
	namespace      string
	serviceAccount string
	client         corev1.CoreV1Interface
}

func NewKubernetesTokenSource(namespace, serviceAccount string) (*KubernetesTokenSource, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes client config: %w", err)
	}

	// Create a clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes clientset: %w", err)
	}

	return &KubernetesTokenSource{
		namespace:      namespace,
		serviceAccount: serviceAccount,
		client:         clientset.CoreV1(),
	}, nil
}

func (t *KubernetesTokenSource) Create(ctx context.Context) (*Token, error) {
	req := &v1.TokenRequest{
		Spec: v1.TokenRequestSpec{
			ExpirationSeconds: ptr(int64(3600)), // Token valid for 1 hour
		},
	}

	s := t.client.ServiceAccounts(t.namespace)
	resp, err := s.CreateToken(ctx, t.serviceAccount, req, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating service account token: %w", err)
	}

	return &Token{
		Value:     resp.Status.Token,
		ExpiresAt: resp.Status.ExpirationTimestamp.Time,
	}, nil
}

func ptr[T any](v T) *T {
	return &v
}
