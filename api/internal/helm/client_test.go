package helm

import (
	"os"
	"testing"

	"helm.sh/helm/v3/pkg/cli"
)

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

