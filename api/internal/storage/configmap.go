package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/helm-version-manager/api/internal/model"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	configMapName     = "helm-version-manager-registry-mappings"
	configMapDataKey  = "mappings"
	defaultNamespace  = "default"
)

type RegistryStore struct {
	clientset *kubernetes.Clientset
	namespace string
	mu        sync.RWMutex
}

func NewRegistryStore() (*RegistryStore, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = defaultNamespace
	}

	return &RegistryStore{
		clientset: clientset,
		namespace: namespace,
	}, nil
}

func (s *RegistryStore) GetMapping(ctx context.Context, namespace, releaseName string) (*model.RegistryMapping, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mappings, err := s.loadMappings(ctx)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("%s/%s", namespace, releaseName)
	if mapping, ok := mappings[key]; ok {
		return &mapping, nil
	}

	return nil, nil
}

func (s *RegistryStore) SetMapping(ctx context.Context, mapping model.RegistryMapping) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mappings, err := s.loadMappings(ctx)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s/%s", mapping.Namespace, mapping.ReleaseName)
	mappings[key] = mapping

	return s.saveMappings(ctx, mappings)
}

func (s *RegistryStore) DeleteMapping(ctx context.Context, namespace, releaseName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mappings, err := s.loadMappings(ctx)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s/%s", namespace, releaseName)
	delete(mappings, key)

	return s.saveMappings(ctx, mappings)
}

func (s *RegistryStore) ListMappings(ctx context.Context) ([]model.RegistryMapping, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mappings, err := s.loadMappings(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.RegistryMapping, 0, len(mappings))
	for _, m := range mappings {
		result = append(result, m)
	}

	return result, nil
}

func (s *RegistryStore) loadMappings(ctx context.Context) (map[string]model.RegistryMapping, error) {
	cm, err := s.clientset.CoreV1().ConfigMaps(s.namespace).Get(ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return make(map[string]model.RegistryMapping), nil
		}
		return nil, fmt.Errorf("failed to get configmap: %w", err)
	}

	data, ok := cm.Data[configMapDataKey]
	if !ok || data == "" {
		return make(map[string]model.RegistryMapping), nil
	}

	var mappings map[string]model.RegistryMapping
	if err := json.Unmarshal([]byte(data), &mappings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mappings: %w", err)
	}

	return mappings, nil
}

func (s *RegistryStore) saveMappings(ctx context.Context, mappings map[string]model.RegistryMapping) error {
	data, err := json.Marshal(mappings)
	if err != nil {
		return fmt.Errorf("failed to marshal mappings: %w", err)
	}

	cm, err := s.clientset.CoreV1().ConfigMaps(s.namespace).Get(ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			cm = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      configMapName,
					Namespace: s.namespace,
				},
				Data: map[string]string{
					configMapDataKey: string(data),
				},
			}
			_, err = s.clientset.CoreV1().ConfigMaps(s.namespace).Create(ctx, cm, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create configmap: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to get configmap: %w", err)
	}

	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}
	cm.Data[configMapDataKey] = string(data)

	_, err = s.clientset.CoreV1().ConfigMaps(s.namespace).Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update configmap: %w", err)
	}

	return nil
}
