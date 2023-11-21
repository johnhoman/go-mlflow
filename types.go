package mlflow

import (
	"time"
)

type LifecycleStage string

const (
	LifecycleStageActive  LifecycleStage = "active"
	LifecycleStageDeleted LifecycleStage = "deleted"
)

// Experiment is a collection of runs used to track and/or
// document an experiment over time.
type Experiment struct {
	// ID is the unique ID for the experiment. The ID field
	// is computed by the server and should not be set when
	// created.
	ExperimentID string `json:"experiment_id,omitempty"`
	// Name is a human, readable identifier for an experiment. The name
	// field must be set when creating a new experiment.
	Name string `json:"name"`
	// ArtifactLocation is the location where artifacts for this
	// experiment are stored. If unset, then the default artifact
	// location is used.
	ArtifactLocation string `json:"artifact_location,omitempty"`
	// CreationTime is the unix timestamp (in milliseconds) of when
	// the experiment was created. This is computed by the server
	// and cannot be set when creating the experiment.
	CreationTime int64 `json:"creation_time,omitempty"`
	// LastUpdatedTime is the unix timestamp (in milliseconds) of when
	// the experiment was last updated. LastUpdatedTime is computed by
	// the server and cannot be set manually.
	LastUpdatedTime int64 `json:"last_updated_time,omitempty"`
	// LifecycleStage is the current stage of the experiment. Possible
	// values are "active" and "deleted". Deleted experiments are not
	// returned by the server by default in list operations. You can
	// restore a deleted with the Restore API endpoint.
	LifecycleStage LifecycleStage `json:"lifecycle_stage,omitempty"`
	// Tags is a list of key-value pairs that are associated with the
	// experiment. You can set and delete tags for an experiment.
	Tags Tags `json:"tags,omitempty"`
}

// DeepCopy returns a deep copy of the Experiment
func (e *Experiment) DeepCopy() *Experiment {
	out := &Experiment{}
	e.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies the attributes of the Experiment into the
// provided Experiment
func (e *Experiment) DeepCopyInto(out *Experiment) {
	*out = *e
	out.Tags = make(Tags, len(e.Tags))
	for index := range e.Tags {
		out.Tags[index] = Tag{
			Key:   e.Tags[index].Key,
			Value: e.Tags[index].Value,
		}
	}
}

// GetExperimentID returns the experiment ID
func (e *Experiment) GetExperimentID() string {
	return e.ExperimentID
}

// SetExperimentID sets the experiment ID attribute on the Experiment
func (e *Experiment) SetExperimentID(id string) {
	e.ExperimentID = id
}

// GetName returns the experiment name
func (e *Experiment) GetName() string {
	return e.Name
}

// SetName sets the experiment name attribute on the Experiment
func (e *Experiment) SetName(name string) {
	e.Name = name
}

// GetNamespace returns the experiment namespace tag
func (e *Experiment) GetNamespace() string {
	return e.Tags.Get("metadata.namespace")
}

// SetNamespace sets the experiment namespace tag
func (e *Experiment) SetNamespace(namespace string) {
	e.Tags.Set("metadata.namespace", namespace)
}

// GetArtifactLocation returns the experiment artifact location
func (e *Experiment) GetArtifactLocation() string {
	return e.ArtifactLocation
}

// SetArtifactLocation sets the experiment artifact location attribute
// on the Experiment
func (e *Experiment) SetArtifactLocation(location string) {
	e.ArtifactLocation = location
}

// GetCreationTimestamp returns the experiment creation timestamp
func (e *Experiment) GetCreationTimestamp() time.Time {
	return time.Unix(e.CreationTime, 0)
}

// SetCreationTimestamp sets the experiment creation timestamp
// attribute on the Experiment
func (e *Experiment) SetCreationTimestamp(t time.Time) {
	e.CreationTime = t.Unix()
}

// GetLastUpdatedTimestamp returns the experiment last updated timestamp
func (e *Experiment) GetLastUpdatedTimestamp() time.Time {
	return time.Unix(e.LastUpdatedTime, 0)
}

// SetLastUpdatedTimestamp sets the experiment last updated timestamp
// attribute on the Experiment
func (e *Experiment) SetLastUpdatedTimestamp(t time.Time) {
	e.LastUpdatedTime = t.Unix()
}

// GetLifecycleStage returns the experiment lifecycle stage
func (e *Experiment) GetLifecycleStage() LifecycleStage {
	return e.LifecycleStage
}

// SetLifecycleStage sets the experiment lifecycle stage attribute
// on the Experiment
func (e *Experiment) SetLifecycleStage(stage LifecycleStage) {
	e.LifecycleStage = stage
}

// GetTags returns the experiment tags
func (e *Experiment) GetTags() *Tags {
	return &e.Tags
}

// SetTags sets the experiment tags attribute on the Experiment
func (e *Experiment) SetTags(tags *Tags) {
	e.Tags = *tags
}

// Tag is a key-value pair associated with an entity,
// such as an experiment. Tags can be used for server-side
// filtering in list experiments queries.
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Tags is a list of Tag objects.
type Tags []Tag

// Contains returns true if the Tags list contains a Tag with the
// provided key. Otherwise, it returns false.
func (t *Tags) Contains(key string) bool {
	for _, tag := range *t {
		if tag.Key == key {
			return true
		}
	}
	return false
}

// Get returns the value of the Tag with the provided key. If the
// Tag is not found, then an empty string is returned.
func (t *Tags) Get(key string) string {
	for _, tag := range *t {
		if tag.Key == key {
			return tag.Value
		}
	}
	return ""
}

func (t *Tags) Set(key, value string) {
	for i, tag := range *t {
		if tag.Key == key {
			(*t)[i].Value = value
			return
		}
	}
	*t = append(*t, Tag{Key: key, Value: value})
}

// Len along with Less and Swap, implements the Sort interface
func (t *Tags) Len() int {
	return len(*t)
}

// Less along with Len and Swap, implements the Sort interface
func (t *Tags) Less(i, j int) bool {
	return (*t)[i].Key < (*t)[j].Key
}

// Swap along with Len and Less, implements the Sort interface
func (t *Tags) Swap(i, j int) {
	(*t)[i], (*t)[j] = (*t)[j], (*t)[i]
}
