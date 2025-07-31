package ui

import (
	"strings"

	"k8sgo/internal/types"
)

func (a *App) filterResources(resources []types.Resource) []types.Resource {
	if a.state.SearchQuery == "" && a.state.FilterCriteria.Status == "" {
		return resources
	}

	var filtered []types.Resource

	for _, resource := range resources {
		if a.matchesSearchCriteria(resource) {
			filtered = append(filtered, resource)
		}
	}

	return filtered
}

func (a *App) matchesSearchCriteria(resource types.Resource) bool {
	if !a.matchesSearchQuery(resource) {
		return false
	}

	if !a.matchesStatusFilter(resource) {
		return false
	}

	if !a.matchesNamespaceFilter(resource) {
		return false
	}

	return true
}

func (a *App) matchesSearchQuery(resource types.Resource) bool {
	if a.state.SearchQuery == "" {
		return true
	}

	query := strings.ToLower(a.state.SearchQuery)

	fields := []string{
		strings.ToLower(resource.Name),
		strings.ToLower(resource.Namespace),
		strings.ToLower(resource.Type),
		strings.ToLower(resource.Status),
	}

	for _, field := range fields {
		if strings.Contains(field, query) {
			return true
		}
	}

	return false
}

func (a *App) matchesStatusFilter(resource types.Resource) bool {
	if a.state.FilterCriteria.Status == "" || a.state.FilterCriteria.Status == "All" {
		return true
	}

	return strings.EqualFold(resource.Status, a.state.FilterCriteria.Status)
}

func (a *App) matchesNamespaceFilter(resource types.Resource) bool {
	if a.state.FilterCriteria.Namespace == "" || a.state.FilterCriteria.Namespace == "All" {
		return true
	}

	return strings.EqualFold(resource.Namespace, a.state.FilterCriteria.Namespace)
}
