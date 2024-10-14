// Package mod_management provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen/v2 version v2.1.0 DO NOT EDIT.
package mod_management

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/oapi-codegen/runtime"
)

const (
	Access_tokenScopes = "access_token.Scopes"
)

// Defines values for ModuleStatus.
const (
	Running ModuleStatus = "running"
	Stopped ModuleStatus = "stopped"
)

// Defines values for ModuleUiFormalityType.
const (
	ModuleFormalityModal  ModuleUiFormalityType = "modal"
	ModuleFormalityNewTab ModuleUiFormalityType = "newtab"
)

// Defines values for ModuleStartStopJSONBodyAction.
const (
	Start ModuleStartStopJSONBodyAction = "start"
	Stop  ModuleStartStopJSONBodyAction = "stop"
)

// BaseResponse defines model for base_response.
type BaseResponse struct {
	// Message message returned by server side if there is any
	Message *string `json:"message,omitempty"`
}

// Module defines model for module.
type Module struct {
	Name *string `json:"name,omitempty"`

	// Services a module can have one or more backend services
	Services *[]ModuleService `json:"services,omitempty"`
	UI       *ModuleUI        `json:"ui,omitempty"`
}

// ModuleId defines model for module_id.
type ModuleId struct {
	Name string `json:"name"`
}

// ModuleService defines model for module_service.
type ModuleService struct {
	Name *string `json:"name,omitempty"`
}

// ModuleStatus defines model for module_status.
type ModuleStatus string

// ModuleUI defines model for module_ui.
type ModuleUI struct {
	Description *string            `json:"description,omitempty"`
	Entry       *string            `json:"entry,omitempty"`
	Formality   *ModuleUIFormality `json:"formality,omitempty"`
	Icon        *string            `json:"icon,omitempty"`
	Prefetch    *bool              `json:"prefetch,omitempty"`
	Show        *bool              `json:"show,omitempty"`
	Title       *map[string]string `json:"title,omitempty"`
}

// ModuleUIFormality defines model for module_ui_formality.
type ModuleUIFormality struct {
	Props *ModuleUIFormalityProperties `json:"props,omitempty"`
	Type  *ModuleUiFormalityType       `json:"type,omitempty"`
}

// ModuleUiFormalityType defines model for ModuleUiFormality.Type.
type ModuleUiFormalityType string

// ModuleUIFormalityProperties defines model for module_ui_formality_properties.
type ModuleUIFormalityProperties struct {
	Animation    *string `json:"animation,omitempty"`
	HasModalCard *bool   `json:"hasModalCard,omitempty" yaml:"hasModalCard,omitempty"`
	Height       *string `json:"height,omitempty"`
	Width        *string `json:"width,omitempty"`
}

// RemoteModule defines model for remote_module.
type RemoteModule struct {
	Name  string `json:"name"`
	Repo  string `json:"repo"`
	Title string `json:"title"`
}

// Name defines model for name.
type Name = string

// InstallableModuleListOK defines model for installable_module_list_ok.
type InstallableModuleListOK struct {
	Data *[]RemoteModule `json:"data,omitempty"`

	// Message message returned by server side if there is any
	Message *string `json:"message,omitempty"`
}

// ModuleInstallOk defines model for module_install_ok.
type ModuleInstallOk = BaseResponse

// ModuleListOK defines model for module_list_ok.
type ModuleListOK struct {
	Data *[]Module `json:"data,omitempty"`

	// Message message returned by server side if there is any
	Message *string `json:"message,omitempty"`
}

// ModuleStartStopOk defines model for module_start_stop_ok.
type ModuleStartStopOk = BaseResponse

// ModuleStatusOk defines model for module_status_ok.
type ModuleStatusOk struct {
	Data *struct {
		Status *ModuleStatus `json:"status,omitempty"`
	} `json:"data,omitempty"`

	// Message message returned by server side if there is any
	Message *string `json:"message,omitempty"`
}

// ModuleUninstallOk defines model for module_uninstall_ok.
type ModuleUninstallOk = BaseResponse

// ResponseBadRequest defines model for response_bad_request.
type ResponseBadRequest = BaseResponse

// ResponseInternalServerError defines model for response_internal_server_error.
type ResponseInternalServerError = BaseResponse

// ResponseOK defines model for response_ok.
type ResponseOK = BaseResponse

// RequestModuleInstall defines model for request_module_install.
type RequestModuleInstall = ModuleId

// RequestModuleUninstall defines model for request_module_uninstall.
type RequestModuleUninstall = ModuleId

// ModuleStartStopJSONBody defines parameters for ModuleStartStop.
type ModuleStartStopJSONBody struct {
	Action ModuleStartStopJSONBodyAction `json:"action"`
}

// ModuleStartStopJSONBodyAction defines parameters for ModuleStartStop.
type ModuleStartStopJSONBodyAction string

// ModuleUninstallJSONRequestBody defines body for ModuleUninstall for application/json ContentType.
type ModuleUninstallJSONRequestBody = ModuleId

// ModuleInstallJSONRequestBody defines body for ModuleInstall for application/json ContentType.
type ModuleInstallJSONRequestBody = ModuleId

// ModuleStartStopJSONRequestBody defines body for ModuleStartStop for application/json ContentType.
type ModuleStartStopJSONRequestBody ModuleStartStopJSONBody

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// ModuleUninstallWithBody request with any body
	ModuleUninstallWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	ModuleUninstall(ctx context.Context, body ModuleUninstallJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// InstallableModuleList request
	InstallableModuleList(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ModuleInstallWithBody request with any body
	ModuleInstallWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	ModuleInstall(ctx context.Context, body ModuleInstallJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ModuleList request
	ModuleList(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// RefreshModule request
	RefreshModule(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ModuleStatus request
	ModuleStatus(ctx context.Context, name Name, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ModuleStartStopWithBody request with any body
	ModuleStartStopWithBody(ctx context.Context, name Name, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	ModuleStartStop(ctx context.Context, name Name, body ModuleStartStopJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) ModuleUninstallWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewModuleUninstallRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ModuleUninstall(ctx context.Context, body ModuleUninstallJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewModuleUninstallRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) InstallableModuleList(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewInstallableModuleListRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ModuleInstallWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewModuleInstallRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ModuleInstall(ctx context.Context, body ModuleInstallJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewModuleInstallRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ModuleList(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewModuleListRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) RefreshModule(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewRefreshModuleRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ModuleStatus(ctx context.Context, name Name, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewModuleStatusRequest(c.Server, name)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ModuleStartStopWithBody(ctx context.Context, name Name, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewModuleStartStopRequestWithBody(c.Server, name, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ModuleStartStop(ctx context.Context, name Name, body ModuleStartStopJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewModuleStartStopRequest(c.Server, name, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewModuleUninstallRequest calls the generic ModuleUninstall builder with application/json body
func NewModuleUninstallRequest(server string, body ModuleUninstallJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewModuleUninstallRequestWithBody(server, "application/json", bodyReader)
}

// NewModuleUninstallRequestWithBody generates requests for ModuleUninstall with any type of body
func NewModuleUninstallRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/management/modules")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewInstallableModuleListRequest generates requests for InstallableModuleList
func NewInstallableModuleListRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/management/modules")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewModuleInstallRequest calls the generic ModuleInstall builder with application/json body
func NewModuleInstallRequest(server string, body ModuleInstallJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewModuleInstallRequestWithBody(server, "application/json", bodyReader)
}

// NewModuleInstallRequestWithBody generates requests for ModuleInstall with any type of body
func NewModuleInstallRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/management/modules")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewModuleListRequest generates requests for ModuleList
func NewModuleListRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/modules")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewRefreshModuleRequest generates requests for RefreshModule
func NewRefreshModuleRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/modules")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewModuleStatusRequest generates requests for ModuleStatus
func NewModuleStatusRequest(server string, name Name) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/modules/%s/status", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewModuleStartStopRequest calls the generic ModuleStartStop builder with application/json body
func NewModuleStartStopRequest(server string, name Name, body ModuleStartStopJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewModuleStartStopRequestWithBody(server, name, "application/json", bodyReader)
}

// NewModuleStartStopRequestWithBody generates requests for ModuleStartStop with any type of body
func NewModuleStartStopRequestWithBody(server string, name Name, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/modules/%s/status", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// ModuleUninstallWithBodyWithResponse request with any body
	ModuleUninstallWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ModuleUninstallResponse, error)

	ModuleUninstallWithResponse(ctx context.Context, body ModuleUninstallJSONRequestBody, reqEditors ...RequestEditorFn) (*ModuleUninstallResponse, error)

	// InstallableModuleListWithResponse request
	InstallableModuleListWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*InstallableModuleListResponse, error)

	// ModuleInstallWithBodyWithResponse request with any body
	ModuleInstallWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ModuleInstallResponse, error)

	ModuleInstallWithResponse(ctx context.Context, body ModuleInstallJSONRequestBody, reqEditors ...RequestEditorFn) (*ModuleInstallResponse, error)

	// ModuleListWithResponse request
	ModuleListWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ModuleListResponse, error)

	// RefreshModuleWithResponse request
	RefreshModuleWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*RefreshModuleResponse, error)

	// ModuleStatusWithResponse request
	ModuleStatusWithResponse(ctx context.Context, name Name, reqEditors ...RequestEditorFn) (*ModuleStatusResponse, error)

	// ModuleStartStopWithBodyWithResponse request with any body
	ModuleStartStopWithBodyWithResponse(ctx context.Context, name Name, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ModuleStartStopResponse, error)

	ModuleStartStopWithResponse(ctx context.Context, name Name, body ModuleStartStopJSONRequestBody, reqEditors ...RequestEditorFn) (*ModuleStartStopResponse, error)
}

type ModuleUninstallResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *ModuleUninstallOk
	JSON500      *ResponseInternalServerError
}

// Status returns HTTPResponse.Status
func (r ModuleUninstallResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ModuleUninstallResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type InstallableModuleListResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *InstallableModuleListOK
	JSON500      *ResponseInternalServerError
}

// Status returns HTTPResponse.Status
func (r InstallableModuleListResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r InstallableModuleListResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ModuleInstallResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *ModuleInstallOk
	JSON500      *ResponseInternalServerError
}

// Status returns HTTPResponse.Status
func (r ModuleInstallResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ModuleInstallResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ModuleListResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *ModuleListOK
	JSON500      *ResponseInternalServerError
}

// Status returns HTTPResponse.Status
func (r ModuleListResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ModuleListResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type RefreshModuleResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *ResponseOK
	JSON500      *ResponseInternalServerError
}

// Status returns HTTPResponse.Status
func (r RefreshModuleResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r RefreshModuleResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ModuleStatusResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *ModuleStatusOk
}

// Status returns HTTPResponse.Status
func (r ModuleStatusResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ModuleStatusResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ModuleStartStopResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *ModuleStartStopOk
	JSON400      *ResponseBadRequest
	JSON500      *ResponseInternalServerError
}

// Status returns HTTPResponse.Status
func (r ModuleStartStopResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ModuleStartStopResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// ModuleUninstallWithBodyWithResponse request with arbitrary body returning *ModuleUninstallResponse
func (c *ClientWithResponses) ModuleUninstallWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ModuleUninstallResponse, error) {
	rsp, err := c.ModuleUninstallWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseModuleUninstallResponse(rsp)
}

func (c *ClientWithResponses) ModuleUninstallWithResponse(ctx context.Context, body ModuleUninstallJSONRequestBody, reqEditors ...RequestEditorFn) (*ModuleUninstallResponse, error) {
	rsp, err := c.ModuleUninstall(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseModuleUninstallResponse(rsp)
}

// InstallableModuleListWithResponse request returning *InstallableModuleListResponse
func (c *ClientWithResponses) InstallableModuleListWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*InstallableModuleListResponse, error) {
	rsp, err := c.InstallableModuleList(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseInstallableModuleListResponse(rsp)
}

// ModuleInstallWithBodyWithResponse request with arbitrary body returning *ModuleInstallResponse
func (c *ClientWithResponses) ModuleInstallWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ModuleInstallResponse, error) {
	rsp, err := c.ModuleInstallWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseModuleInstallResponse(rsp)
}

func (c *ClientWithResponses) ModuleInstallWithResponse(ctx context.Context, body ModuleInstallJSONRequestBody, reqEditors ...RequestEditorFn) (*ModuleInstallResponse, error) {
	rsp, err := c.ModuleInstall(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseModuleInstallResponse(rsp)
}

// ModuleListWithResponse request returning *ModuleListResponse
func (c *ClientWithResponses) ModuleListWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ModuleListResponse, error) {
	rsp, err := c.ModuleList(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseModuleListResponse(rsp)
}

// RefreshModuleWithResponse request returning *RefreshModuleResponse
func (c *ClientWithResponses) RefreshModuleWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*RefreshModuleResponse, error) {
	rsp, err := c.RefreshModule(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseRefreshModuleResponse(rsp)
}

// ModuleStatusWithResponse request returning *ModuleStatusResponse
func (c *ClientWithResponses) ModuleStatusWithResponse(ctx context.Context, name Name, reqEditors ...RequestEditorFn) (*ModuleStatusResponse, error) {
	rsp, err := c.ModuleStatus(ctx, name, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseModuleStatusResponse(rsp)
}

// ModuleStartStopWithBodyWithResponse request with arbitrary body returning *ModuleStartStopResponse
func (c *ClientWithResponses) ModuleStartStopWithBodyWithResponse(ctx context.Context, name Name, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ModuleStartStopResponse, error) {
	rsp, err := c.ModuleStartStopWithBody(ctx, name, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseModuleStartStopResponse(rsp)
}

func (c *ClientWithResponses) ModuleStartStopWithResponse(ctx context.Context, name Name, body ModuleStartStopJSONRequestBody, reqEditors ...RequestEditorFn) (*ModuleStartStopResponse, error) {
	rsp, err := c.ModuleStartStop(ctx, name, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseModuleStartStopResponse(rsp)
}

// ParseModuleUninstallResponse parses an HTTP response from a ModuleUninstallWithResponse call
func ParseModuleUninstallResponse(rsp *http.Response) (*ModuleUninstallResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ModuleUninstallResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest ModuleUninstallOk
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest ResponseInternalServerError
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest

	}

	return response, nil
}

// ParseInstallableModuleListResponse parses an HTTP response from a InstallableModuleListWithResponse call
func ParseInstallableModuleListResponse(rsp *http.Response) (*InstallableModuleListResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &InstallableModuleListResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest InstallableModuleListOK
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest ResponseInternalServerError
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest

	}

	return response, nil
}

// ParseModuleInstallResponse parses an HTTP response from a ModuleInstallWithResponse call
func ParseModuleInstallResponse(rsp *http.Response) (*ModuleInstallResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ModuleInstallResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest ModuleInstallOk
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest ResponseInternalServerError
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest

	}

	return response, nil
}

// ParseModuleListResponse parses an HTTP response from a ModuleListWithResponse call
func ParseModuleListResponse(rsp *http.Response) (*ModuleListResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ModuleListResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest ModuleListOK
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest ResponseInternalServerError
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest

	}

	return response, nil
}

// ParseRefreshModuleResponse parses an HTTP response from a RefreshModuleWithResponse call
func ParseRefreshModuleResponse(rsp *http.Response) (*RefreshModuleResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &RefreshModuleResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest ResponseOK
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest ResponseInternalServerError
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest

	}

	return response, nil
}

// ParseModuleStatusResponse parses an HTTP response from a ModuleStatusWithResponse call
func ParseModuleStatusResponse(rsp *http.Response) (*ModuleStatusResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ModuleStatusResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest ModuleStatusOk
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseModuleStartStopResponse parses an HTTP response from a ModuleStartStopWithResponse call
func ParseModuleStartStopResponse(rsp *http.Response) (*ModuleStartStopResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ModuleStartStopResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest ModuleStartStopOk
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest ResponseBadRequest
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest ResponseInternalServerError
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest

	}

	return response, nil
}
