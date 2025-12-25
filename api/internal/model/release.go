package model

import "time"

type Release struct {
	Name         string    `json:"name"`
	Namespace    string    `json:"namespace"`
	Chart        string    `json:"chart"`
	ChartVersion string    `json:"chartVersion"`
	AppVersion   string    `json:"appVersion"`
	Status       string    `json:"status"`
	Updated      time.Time `json:"updated"`
	Revision     int       `json:"revision"`
}

type VersionUpgradeRequest struct {
	ChartVersion string         `json:"chartVersion"`
	Values       map[string]any `json:"values,omitempty"`
}

type ChartVersion struct {
	Version     string `json:"version"`
	AppVersion  string `json:"appVersion"`
	Description string `json:"description"`
}

type ReleaseHistory struct {
	Revision    int       `json:"revision"`
	Updated     time.Time `json:"updated"`
	Status      string    `json:"status"`
	Chart       string    `json:"chart"`
	AppVersion  string    `json:"appVersion"`
	Description string    `json:"description"`
}
