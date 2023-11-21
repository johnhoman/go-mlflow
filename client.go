package mlflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/pkg/errors"
)

type DeleteOptions struct {
	// IgnoreMissing will ignore the error if the experiment is not found
	// and will not return an error. This is useful for idempotent deletes
	// where you do not want to return an error if the experiment is already
	// deleted.
	IgnoreMissing bool

	// Namespace is for deleting the experiment in a specific namespace. If
	// the namespace is not set, then the default namespace is used.
	Namespace string
}

// CreateOptions are options that can be passed to the CreateExperiment
type CreateOptions struct {
	// Namespace is for creating the experiment in a specific namespace.
	// A namespace is a Kubernetes construct, but is useful for managing
	// experiments in a multi-tenant environment, running on Kubernetes.
	Namespace string
}

// ListOptions are options that can be passed to the ListExperiments.
type ListOptions struct {
	// Namespace is for listing experiments in a specific namespace.
	// A namespace is a Kubernetes construct, but is useful for managing
	// experiments in a multi-tenant environment, running on Kubernetes.
	Namespace string
}

type GetOptions struct {
	// Namespace is only required for getting an experiment by name. If
	// ID is set, then the namespace is ignored. If the namespace is not
	// set, then the default namespace is used.
	Namespace string
}

type CreateOption interface {
	ApplyToCreate(*CreateOptions)
}

type DeleteOption interface {
	ApplyToDelete(*DeleteOptions)
}

type GetOption interface {
	ApplyToGet(*GetOptions)
}

// InNamespace is a CreateOption that sets the target namespace of the
// resource being created.
type InNamespace string

func (i InNamespace) ApplyToCreate(o *CreateOptions) {
	o.Namespace = string(i)
}

func (i InNamespace) ApplyToGet(o *GetOptions) {
	o.Namespace = string(i)
}

func (i InNamespace) ApplyToDelete(o *DeleteOptions) {
}

// IgnoreMissing is a DeleteOption that will force the client to
// not return an error if the experiment is not found.
type IgnoreMissing bool

func (i IgnoreMissing) ApplyToDelete(o *DeleteOptions) {
	o.IgnoreMissing = bool(i)
}

// Client is the interface for interacting with the MLFlow API
type Client interface {
	// CreateExperiment creates a new experiment. If the experiment
	// name already exists, then an error is returned. The experiment
	// will set all computed fields with the response
	CreateExperiment(experiment *Experiment, opts ...CreateOption) error
	// DeleteExperiment deletes the provided experiment with the given ID. Experiments
	// can be deleted by name, but this requires an additional lookup step to find the
	// experiment ID. If the experiment is not found, then an error is returned.
	DeleteExperiment(experiment *Experiment, opts ...DeleteOption) error
	// GetExperiment gets the experiment with the given ID or name. If the experiment
	// is not found, then an error is returned. At least one of the ID or Name must
	// be set on the experiment. The remaining fields will be set from the response.
	GetExperiment(experiment *Experiment, opts ...GetOption) error
	// UpdateExperiment updates the experiment with the given ID or name. If the
	// ID and Name are both set, then the ID will be used, and if the name doesn't
	// match the stored Name, it will be updated. Name, Tags, and LifecycleStage
	// can all be updated. All other fields are ignored.
	UpdateExperiment(experiment *Experiment) error
}

type client struct {
	address    *url.URL
	httpClient *http.Client

	authenticator func(*http.Request)
}

// CreateExperiment creates a new experiment
func (c *client) CreateExperiment(experiment *Experiment, opts ...CreateOption) error {
	if experiment.Name == "" {
		return errors.Errorf("missing required attribute %q on experiment", "Name")
	}

	o := &CreateOptions{}
	for _, f := range opts {
		f.ApplyToCreate(o)
	}

	if o.Namespace == "" {
		o.Namespace = "default"
	}

	u := mustCopyURL(c.address)
	u.Path = path.Join(u.Path, "/api/2.0/mlflow/experiments/create")

	var in struct {
		Name             string `json:"name"`
		ArtifactLocation string `json:"artifact_location"`
		Tags             Tags   `json:"tags"`
	}
	in.Name = fmt.Sprintf("%s/%s", o.Namespace, experiment.Name)
	in.ArtifactLocation = experiment.ArtifactLocation
	in.Tags = experiment.Tags
	in.Tags.Set("metadata.namespace", o.Namespace)

	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(in)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), buf)
	if err != nil {
		return err
	}

	if c.authenticator != nil {
		c.authenticator(req)
	}

	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	body := res.Body
	defer body.Close()

	if res.StatusCode != http.StatusOK {
		var data []byte
		data, err = io.ReadAll(body)
		if err != nil {
			return err
		}
		return errors.Errorf("unexpected status code %d: %s", res.StatusCode, string(data))
	}

	var out struct {
		ExperimentID string `json:"experiment_id"`
	}

	err = json.NewDecoder(body).Decode(&out)
	if err != nil {
		return err
	}

	experiment.ExperimentID = out.ExperimentID
	return c.GetExperiment(experiment, InNamespace(o.Namespace))
}

func (c *client) GetExperiment(experiment *Experiment, opts ...GetOption) error {
	if experiment.ExperimentID == "" {
		return errors.Errorf("ExperimentID must be set")
	}

	o := &GetOptions{}
	for _, f := range opts {
		f.ApplyToGet(o)
	}

	if o.Namespace == "" {
		o.Namespace = "default"
	}

	u := mustCopyURL(c.address)
	u.Path = path.Join(u.Path, "/api/2.0/mlflow/experiments/get")

	q := u.Query()
	q.Set("experiment_id", experiment.ExperimentID)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	if c.authenticator != nil {
		c.authenticator(req)
	}

	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	body := res.Body
	defer body.Close()

	if res.StatusCode != http.StatusOK {
		var data []byte
		data, err = io.ReadAll(body)
		if err != nil {
			return err
		}
		return errors.Errorf("unexpected status code %d: %s", res.StatusCode, string(data))
	}

	var out struct {
		Experiment `json:"experiment"`
	}

	err = json.NewDecoder(body).Decode(&out)
	if err != nil {
		return err
	}

	out.DeepCopyInto(experiment)
	namespace := experiment.Tags.Get("metadata.namespace")
	if namespace != "" {
		prefix := fmt.Sprintf("%s/", namespace)
		experiment.Name = strings.TrimPrefix(experiment.Name, prefix)
	}
	return nil
}

func (c *client) DeleteExperiment(experiment *Experiment, opts ...DeleteOption) error {
	if experiment.ExperimentID == "" {
		return errors.Errorf("ExperimentID must be set")
	}

	o := &DeleteOptions{}
	for _, f := range opts {
		f.ApplyToDelete(o)
	}

	if o.Namespace == "" {
		o.Namespace = "default"
	}

	u := mustCopyURL(c.address)
	u.Path = path.Join(u.Path, "/api/2.0/mlflow/experiments/delete")

	var body struct {
		ExperimentID string `json:"experiment_id"`
	}
	body.ExperimentID = experiment.ExperimentID

	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), buf)
	if err != nil {
		return err
	}

	if c.authenticator != nil {
		c.authenticator(req)
	}

	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusOK {
		empty := &Experiment{}
		empty.DeepCopyInto(experiment)
		return nil
	}

	if res.StatusCode == http.StatusNotFound && o.IgnoreMissing {
		empty := &Experiment{}
		empty.DeepCopyInto(experiment)
		return nil
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return errors.Errorf("unexpected status code %d: %s", res.StatusCode, string(data))
}

func mustCopyURL(in *url.URL) *url.URL {
	out, err := url.Parse(in.String())
	if err != nil {
		panic(err)
	}
	return out
}

func mustParseURL(in string) *url.URL {
	out, err := url.Parse(in)
	if err != nil {
		panic(err)
	}
	return out
}
