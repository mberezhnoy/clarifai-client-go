package clarifai

import "net/http"

type Input struct {
	Data      *Image         `json:"data,omitempty"`
	ID        string         `json:"id,omitempty"`
	CreatedAt string         `json:"created_at,omitempty"`
	Status    *ServiceStatus `json:"status,omitempty"`
}

type Inputs struct {
	Inputs  []*Input `json:"inputs"`
	modelID string   `json:"-"`
}

// InitInputs returns a default inputs object.
func InitInputs() *Inputs {
	return &Inputs{
		modelID: PublicModelGeneral,
	}
}

// AddInput adds an image input to a request.
func (i *Inputs) AddInput(im *Image, id string) error {
	if len(i.Inputs) >= InputLimit {
		return ErrInputLimitReached
	}

	in := &Input{
		Data: im,
	}

	// Add custom ID if provided.
	if id != "" {
		in.ID = id
	}

	i.Inputs = append(i.Inputs, in)
	return nil
}

// SetModel is an optional model setter for predict calls.
func (i *Inputs) SetModel(m string) {
	i.modelID = m
}

// AddConcept adds concepts to input.
func (i *Input) AddConcept(id string, value interface{}) {

	if i.Data == nil {
		i.Data = &Image{}
	}

	i.Data.Concepts = append(i.Data.Concepts, map[string]interface{}{
		"name":  id,
		"value": value,
	})
}

// SetMetadata adds metadata to a query input item ("input" -> "data" -> "metadata").
func (q *Input) SetMetadata(i interface{}) {
	if q.Data == nil {
		q.Data = &Image{}
	}
	q.Data.Metadata = i
}

// AddInputs builds a request to add inputs to the API.
func (s *Session) AddInputs(p *Inputs) *Request {

	r := NewRequest(s, http.MethodPost, "inputs")
	r.SetPayload(p)

	return r
}

// GetAllInputs fetches a list of all inputs.
func (s *Session) GetAllInputs() *Request {

	return NewRequest(s, http.MethodGet, "inputs")
}

// GetInput fetches one input.
func (s *Session) GetInput(id string) *Request {

	return NewRequest(s, http.MethodGet, "inputs/"+id)
}

// GetInputStatuses fetches statuses of all inputs.
func (s *Session) GetInputStatuses() *Request {

	return NewRequest(s, http.MethodGet, "inputs/status")
}

// Payload for update/delete concepts of input
type patchInputsPayload struct {
	Action string        `json:"action"`
	Inputs []*patchInput `json:"inputs"`
}

func newPatchInputsPayload(action string) *patchInputsPayload {
	return &patchInputsPayload{
		Action: action,
		Inputs: make([]*patchInput, 0),
	}
}

type patchInput struct {
	Id   string `json:"id"`
	Data struct {
		Concepts []interface{} `json:"concepts"`
	} `json:"data"`
}

func newPatchInput(id string) *patchInput {
	p := &patchInput{Id: id}
	p.Data.Concepts = make([]interface{}, 0)
	return p
}

func (p *patchInput) addConcept(id string, val, ignoreVal bool) {
	c := map[string]interface{}{
		"id": id,
	}
	if !ignoreVal {
		if val {
			c["value"] = 1
		} else {
			c["value"] = 0
		}
	}
	p.Data.Concepts = append(p.Data.Concepts, c)
}

// DeleteInputConcepts remove concepts that were already added to an input.
func (s *Session) DeleteInputConcepts(id string, concepts []string) *Request {

	// 1. Build a request.
	r := NewRequest(s, http.MethodPatch, "inputs")

	// 2. Add payload.
	p := newPatchInputsPayload("remove")
	i := newPatchInput(id)

	for _, v := range concepts {
		i.addConcept(v, false, true)
	}
	p.Inputs = append(p.Inputs, i)

	r.SetPayload(p)

	return r
}

// UpdateInputConcepts updates existing and/or adds new concepts to an input by its ID.
func (s *Session) UpdateInputConcepts(id string, userConcepts map[string]bool) *Request {

	// 1. Build a request.
	r := NewRequest(s, http.MethodPatch, "inputs")

	// 2. Add payload.
	// Convert an input map into a map of concepts.
	p := newPatchInputsPayload("merge")
	i := newPatchInput(id)

	for id, value := range userConcepts {
		i.addConcept(id, value, false)
	}
	p.Inputs = append(p.Inputs, i)

	r.SetPayload(p)

	return r
}

// DeleteInput deletes a single input by its ID.
func (s *Session) DeleteInput(id string) *Request {

	return NewRequest(s, http.MethodDelete, "inputs/"+id)
}

// DeleteInputs deletes multiple inputs by their IDs.
func (s *Session) DeleteInputs(ids []string) *Request {

	// 1. Build a request.
	r := NewRequest(s, http.MethodDelete, "inputs")

	// 2. Add a payload.
	r.SetPayload(struct {
		Inputs []string `json:"ids"`
	}{
		Inputs: ids,
	})

	return r
}

// DeleteAllInputs deletes all inputs.
func (s *Session) DeleteAllInputs() *Request {

	return NewRequest(s, http.MethodDelete, "inputs")
}
