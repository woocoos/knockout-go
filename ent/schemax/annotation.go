package schemax

import (
	"entgo.io/ent/schema"
)

var (
	AnnotationName = "Knockout"
)

// Annotation is a schema annotation for Knockout projects.
type Annotation struct {
	// Resources is the list of resources that this annotation is applied to.
	// each resource name use field name
	Resources []string `json:"Resources,omitempty"`
	// TenantField is the name of the tenant field.if you want to use a name except tenant_id
	TenantField string `json:"TenantField,omitempty"`
	// ExcludeNodeQuery indicator whether to exclude node query. It is for Schema
	ExcludeNodeQuery bool `json:"ExcludeNodeQuery,omitempty"`
}

func (Annotation) Name() string {
	return AnnotationName
}

func (a Annotation) Merge(other schema.Annotation) schema.Annotation {
	var ant Annotation
	switch other := other.(type) {
	case Annotation:
		ant = other
	case *Annotation:
		if other != nil {
			ant = *other
		}
	default:
		return a
	}
	if ant.TenantField != "" {
		a.TenantField = ant.TenantField
	}
	if len(ant.Resources) != 0 {
		a.Resources = ant.Resources
	}
	if ant.ExcludeNodeQuery {
		a.ExcludeNodeQuery = true
	}
	return a
}

// Resources returns a new annotation with the given resources.
func Resources(fields []string) Annotation {
	return Annotation{
		Resources: fields,
	}
}

// TenantField returns a new annotation with the given tenant field.
func TenantField(field string) Annotation {
	return Annotation{
		TenantField: field,
	}
}

// ExcludeNodeQuery returns a new annotation with the given exclude node query.
func ExcludeNodeQuery() Annotation {
	return Annotation{
		ExcludeNodeQuery: true,
	}
}
