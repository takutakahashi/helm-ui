package helm

import (
	"os"
	"testing"

	"helm.sh/helm/v3/pkg/cli"
)

func TestBuildConfigFlagsUsesExplicitNamespace(t *testing.T) {
	// Set HELM_NAMESPACE to ensure our explicit namespace takes precedence
	originalHelmNamespace := os.Getenv("HELM_NAMESPACE")
	os.Setenv("HELM_NAMESPACE", "helm-ui-namespace")
	defer func() {
		if originalHelmNamespace != "" {
			os.Setenv("HELM_NAMESPACE", originalHelmNamespace)
		} else {
			os.Unsetenv("HELM_NAMESPACE")
		}
	}()

	c := &Client{settings: cli.New()}

	// Test with explicit namespace that differs from HELM_NAMESPACE
	targetNamespace := "my-release-namespace"
	configFlags := c.buildConfigFlags(targetNamespace)

	if configFlags.Namespace == nil {
		t.Fatal("expected Namespace to be set, got nil")
	}
	if *configFlags.Namespace != targetNamespace {
		t.Errorf("expected namespace %q, got %q", targetNamespace, *configFlags.Namespace)
	}
}

func TestBuildConfigFlagsWithDifferentNamespaces(t *testing.T) {
	c := &Client{settings: cli.New()}

	testCases := []string{
		"default",
		"kube-system",
		"my-app-namespace",
		"production",
	}

	for _, ns := range testCases {
		t.Run(ns, func(t *testing.T) {
			configFlags := c.buildConfigFlags(ns)

			if configFlags.Namespace == nil {
				t.Fatal("expected Namespace to be set, got nil")
			}
			if *configFlags.Namespace != ns {
				t.Errorf("expected namespace %q, got %q", ns, *configFlags.Namespace)
			}
		})
	}
}

func TestBuildConfigFlagsRespectsKubeconfigEnv(t *testing.T) {
	c := &Client{settings: cli.New()}

	// Test without KUBECONFIG set
	originalKubeconfig := os.Getenv("KUBECONFIG")
	os.Unsetenv("KUBECONFIG")

	configFlags := c.buildConfigFlags("test-ns")
	if configFlags.KubeConfig != nil && *configFlags.KubeConfig != "" {
		t.Errorf("expected KubeConfig to be nil or empty when KUBECONFIG not set, got %q", *configFlags.KubeConfig)
	}

	// Test with KUBECONFIG set
	testKubeconfig := "/tmp/test-kubeconfig"
	os.Setenv("KUBECONFIG", testKubeconfig)
	defer func() {
		if originalKubeconfig != "" {
			os.Setenv("KUBECONFIG", originalKubeconfig)
		} else {
			os.Unsetenv("KUBECONFIG")
		}
	}()

	configFlags = c.buildConfigFlags("test-ns")
	if configFlags.KubeConfig == nil || *configFlags.KubeConfig != testKubeconfig {
		t.Errorf("expected KubeConfig to be %q, got %v", testKubeconfig, configFlags.KubeConfig)
	}
}

func TestSearchChartVersions(t *testing.T) {
	c := &Client{
		settings: cli.New(),
	}

	versions, err := c.searchChartVersions("oci://ghcr.io/takutakahashi/charts", "agentapi-ui")
	if err != nil {
		t.Fatalf("searchChartVersions failed: %v", err)
	}

	if len(versions) == 0 {
		t.Fatal("expected at least one version, got none")
	}

	t.Logf("Found %d versions", len(versions))
	for i, v := range versions {
		if i < 5 {
			t.Logf("  Version: %s", v.Version)
		}
	}
}

func TestLocateChart(t *testing.T) {
	settings := cli.New()

	c := &Client{
		settings: settings,
	}

	chartPath, err := c.locateChart(nil, "oci://ghcr.io/takutakahashi/charts", "agentapi-ui", "0.1.0")
	if err != nil {
		t.Fatalf("locateChart failed: %v", err)
	}

	if chartPath == "" {
		t.Fatal("expected chart path, got empty string")
	}

	t.Logf("Chart downloaded to: %s", chartPath)

	// Verify the file exists
	if _, err := os.Stat(chartPath); os.IsNotExist(err) {
		t.Fatalf("chart file does not exist: %s", chartPath)
	}
}

