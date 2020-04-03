package service

import (
	"encoding/json"
	"errors"
	"github.com/databrickslabs/databricks-terraform/client/model"
	"net/http"
)

// SecretsAPI exposes the Secrets API
type SecretScopesAPI struct {
	Client DBApiClient
}

func (a SecretScopesAPI) init(client DBApiClient) SecretScopesAPI {
	a.Client = client
	return a
}

// CreateSecretScope creates a new secret scope
func (a SecretScopesAPI) Create(scope string, initialManagePrincipal string) error {
	data := struct {
		Scope                  string `json:"scope,omitempty"`
		InitialManagePrincipal string `json:"initial_manage_principal,omitempty"`
	}{
		scope,
		initialManagePrincipal,
	}
	_, err := a.Client.performQuery(http.MethodPost, "/secrets/scopes/create", "2.0", nil, data)
	return err
}

// DeleteSecretScope deletes a secret scope
func (a SecretScopesAPI) Delete(scope string) error {
	data := struct {
		Scope string `json:"scope,omitempty" `
	}{
		scope,
	}
	_, err := a.Client.performQuery(http.MethodPost, "/secrets/scopes/delete", "2.0", nil, data)
	return err
}

// List lists all secret scopes available in the workspace
func (a SecretScopesAPI) List() ([]model.SecretScope, error) {
	var listSecretScopesResponse struct {
		Scopes []model.SecretScope `json:"scopes,omitempty"`
	}

	resp, err := a.Client.performQuery(http.MethodGet, "/secrets/scopes/list", "2.0", nil, nil)
	if err != nil {
		return listSecretScopesResponse.Scopes, err
	}

	err = json.Unmarshal(resp, &listSecretScopesResponse)
	return listSecretScopesResponse.Scopes, err
}

func (a SecretScopesAPI) Read(scopeName string) (model.SecretScope, error) {
	var secretScope model.SecretScope
	scopes, err := a.List()
	if err != nil {
		return secretScope, err
	}
	for _, scope := range scopes {
		if scope.Name == scopeName {
			return scope, nil
		}
	}
	return secretScope, errors.New("No Secret Scope found with scope name " + scopeName)
}
