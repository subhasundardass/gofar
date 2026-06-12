package form

import (
	"encoding/json"
	"fmt"
	"sync"
)

// =============================================================================
// FormState — Datastar-mutable form snapshot
// =============================================================================
//
// Datastar stores its reactive state in a single flat JSON object called
// the "signals" store. When a user types in an input, Datastar patches
// that signals object and posts it back to the server. The server then
// merges the new state into whatever it had.
//
// FormState is the Go-side mirror of that signals object. It is:
//
//   1.  JSON-serialisable with stable field names (= the Form's Keys)
//   2.  Safe to mutate concurrently (sync.RWMutex)
//   3.  Round-trippable: Marshal / Unmarshal produce the same shape
//   4.  Fully decoupled from any *Instance — you can build a FormState
//       without ever constructing an Instance, which is what a Datastar
//       POST handler wants to do.
//
// Typical Datastar flow:
//
//	// Server → client  (initial render)
//	state := form.NewFormState("signup", inst)
//	json.NewEncoder(w).Encode(state)        // signals body
//
//	// Client → server  (user typed something)
//	patch, _ := datastar.ReadFormState(c, "signup")
//	state.Merge(patch)
//	inst := form.NewInstance(form).Apply(state)
//	inst.Validate()
//
// Naming convention: the signals object is namespaced under
// "<formName>.<fieldKey>" so multiple forms can coexist on the same page
// without their keys clashing.

const signalsKey = "_signals"

// FormState is the live, mutable form snapshot. It is safe for concurrent
// use; reads take an RLock, writes take the full Lock.
//
// IMPORTANT: always pass *FormState around, never FormState by value —
// the embedded mutex is a sync.RWMutex which must not be copied.
type FormState struct {
	mu sync.RWMutex

	// Name is the form this state belongs to. Two FormStates with the same
	// field keys but different Names are considered distinct and can
	// coexist in a single Datastar signals payload.
	Name string `json:"name"`

	// Values is the current value of every field the form declares.
	// Keys are the Form's field keys (PascalCase, snake, etc. — whatever
	// the schema uses). Values are kept as `any` to round-trip through
	// JSON without lossy conversions (numbers stay numbers, bools stay
	// bools, "" stays "" instead of becoming nil).
	Values map[string]any `json:"values"`

	// Touched is the set of fields the user has interacted with at least
	// once. Datastar sets a flag on `blur` so we know whether to surface
	// an error for that field yet (don't yell at a user for a field
	// they haven't visited).
	Touched map[string]bool `json:"touched"`

	// Errors is the most recent set of per-field validation messages.
	// Mirrors the shape of form.Errors so a renderer can read from
	// FormState and Instance interchangeably.
	Errors Errors `json:"errors"`

	// Meta is a free-form bag for renderer-specific state — pending
	// submission, last-submitted-at timestamps, focus, etc. — that
	// doesn't deserve a first-class field.
	Meta map[string]any `json:"meta,omitempty"`
}

// =============================================================================
// Constructors
// =============================================================================

// NewFormState builds a FormState from a *Form + *Instance pair. The
// instance may be nil; in that case the state starts empty.
//
// Defaults from the form's fields are NOT included automatically — pass
// them through inst first. This keeps NewFormState predictable.
func NewFormState(name string, inst *Instance) *FormState {
	fs := &FormState{
		Name:    name,
		Values:  map[string]any{},
		Touched: map[string]bool{},
		Errors:  Errors{},
		Meta:    map[string]any{},
	}
	if inst != nil {
		fs.MergeFromInstance(inst)
	}
	return fs
}

// NewFormStateFromValues builds a FormState from a raw values map. Useful
// when you don't have a full Instance (e.g. inside a Datastar POST handler
// where you only got a JSON body).
func NewFormStateFromValues(name string, values map[string]any) *FormState {
	return &FormState{
		Name:    name,
		Values:  cloneAnyMap(values),
		Touched: map[string]bool{},
		Errors:  Errors{},
		Meta:    map[string]any{},
	}
}

// =============================================================================
// Mutation
// =============================================================================

// Set stores a value in the Values bag, marking the field as Touched.
// Returns the receiver for chaining.
func (fs *FormState) Set(key string, value any) *FormState {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.Values[key] = value
	fs.Touched[key] = true
	return fs
}

// SetSilent stores a value WITHOUT marking the field as touched. Use this
// for server-driven updates (e.g. pre-filling a form from a database
// record) so the user isn't shown errors for fields they haven't visited.
func (fs *FormState) SetSilent(key string, value any) *FormState {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.Values[key] = value
	return fs
}

// SetName overwrites the FormState's Name. The framework/datastar package
// uses this to force the server-known form name onto an incoming patch
// so a malicious client cannot route data into the wrong form.
func (fs *FormState) SetName(name string) *FormState {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.Name = name
	return fs
}

// MarkTouched flips the touched flag for a field. Call this from a Datastar
// `@blur` handler: state.MarkTouched('email').
func (fs *FormState) MarkTouched(key string) *FormState {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.Touched[key] = true
	return fs
}

// SetError sets a single error message for a field, overwriting any prior
// messages. Pass an empty string to clear.
func (fs *FormState) SetError(key, message string) *FormState {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if message == "" {
		delete(fs.Errors, key)
		return fs
	}
	fs.Errors[key] = []string{message}
	return fs
}

// SetMeta stores a free-form value in the Meta bag. Use for renderer-only
// state (last-submitted timestamp, focus, pending spinner flag, etc.).
func (fs *FormState) SetMeta(key string, value any) *FormState {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if fs.Meta == nil {
		fs.Meta = map[string]any{}
	}
	fs.Meta[key] = value
	return fs
}

// AddError appends a message to a field's existing error list.
func (fs *FormState) AddError(key, message string) *FormState {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.Errors[key] = append(fs.Errors[key], message)
	return fs
}

// ClearErrors drops every error.
func (fs *FormState) ClearErrors() *FormState {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.Errors = Errors{}
	return fs
}

// ResetTouched drops every touched flag. The form re-enters its
// "pristine" state.
func (fs *FormState) ResetTouched() *FormState {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.Touched = map[string]bool{}
	return fs
}

// Merge applies a patch produced by Datastar (or by another FormState)
// into this one. The patch is taken by pointer so the caller can build it
// via a literal and the embedded mutex is never copied. Nil-valued
// entries in the patch REMOVE the key — this matches Datastar's
// "null = delete" convention so resetting a field on the client actually
// clears it server-side.
func (fs *FormState) Merge(patch *FormState) *FormState {
	if patch == nil {
		return fs
	}
	fs.mu.Lock()
	defer fs.mu.Unlock()

	patch.mu.RLock()
	patchValues := patch.Values
	patchTouched := patch.Touched
	patchErrors := patch.Errors
	patchMeta := patch.Meta
	patch.mu.RUnlock()

	for k, v := range patchValues {
		if v == nil {
			delete(fs.Values, k)
		} else {
			fs.Values[k] = v
		}
	}
	for k := range patchTouched {
		fs.Touched[k] = true
	}
	for k, msgs := range patchErrors {
		if len(msgs) == 0 {
			delete(fs.Errors, k)
			continue
		}
		fs.Errors[k] = append(fs.Errors[k], msgs...)
	}
	for k, v := range patchMeta {
		fs.Meta[k] = v
	}
	return fs
}

// MergeFromInstance copies values and errors from an *Instance into the
// FormState. This is what you call right before pushing state back to the
// client so the next render reflects what the server just validated.
func (fs *FormState) MergeFromInstance(inst *Instance) *FormState {
	if inst == nil {
		return fs
	}
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for k, v := range inst.Values {
		fs.Values[k] = v
	}
	for k, msgs := range inst.Errors {
		fs.Errors[k] = append(fs.Errors[k], msgs...)
	}
	return fs
}

// =============================================================================
// Reads
// =============================================================================

// Get returns the current value of key, or nil if absent.
func (fs *FormState) Get(key string) any {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.Values[key]
}

// IsTouched reports whether the field has been touched at least once.
func (fs *FormState) IsTouched(key string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.Touched[key]
}

// HasError reports whether the field has any current error.
func (fs *FormState) HasError(key string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return len(fs.Errors[key]) > 0
}

// FirstError returns the first error message for a field, or "" if none.
func (fs *FormState) FirstError(key string) string {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	if msgs := fs.Errors[key]; len(msgs) > 0 {
		return msgs[0]
	}
	return ""
}

// Snapshot returns a deep copy of the FormState's mutable maps. Useful
// when you want to hand data off to a goroutine without holding the lock.
//
// Always returns a *FormState so callers can use it identically to the
// original (no value-copy of the mutex).
func (fs *FormState) Snapshot() *FormState {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return &FormState{
		Name:    fs.Name,
		Values:  cloneAnyMap(fs.Values),
		Touched: cloneBoolMap(fs.Touched),
		Errors:  cloneErrors(fs.Errors),
		Meta:    cloneAnyMap(fs.Meta),
	}
}

// ToInstance builds an *Instance backed by the given *Form, populating it
// with this state's values. The instance's errors are NOT overwritten —
// call inst.Validate() afterwards to re-derive them.
func (fs *FormState) ToInstance(f *Form) *Instance {
	inst := f.NewInstance()
	fs.mu.RLock()
	for k, v := range fs.Values {
		inst.Values[k] = v
	}
	fs.mu.RUnlock()
	return inst
}

// =============================================================================
// Serialisation
// =============================================================================

// MarshalJSON serialises the FormState. We keep the JSON shape stable so
// the same payload can be (a) sent to Datastar as a signals merge and
// (b) round-tripped through UnmarshalJSON.
func (fs *FormState) MarshalJSON() ([]byte, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// Always emit non-nil empty maps so the client gets a predictable
	// shape and never has to nil-check before iterating.
	values := fs.Values
	if values == nil {
		values = map[string]any{}
	}
	touched := fs.Touched
	if touched == nil {
		touched = map[string]bool{}
	}
	errors := fs.Errors
	if errors == nil {
		errors = Errors{}
	}

	type wire struct {
		Name    string          `json:"name"`
		Values  map[string]any  `json:"values"`
		Touched map[string]bool `json:"touched"`
		Errors  Errors          `json:"errors"`
		Meta    map[string]any  `json:"meta,omitempty"`
	}
	return json.Marshal(wire{
		Name:    fs.Name,
		Values:  values,
		Touched: touched,
		Errors:  errors,
		Meta:    fs.Meta,
	})
}

// UnmarshalJSON parses a Datastar-style payload into the FormState.
// Unknown fields are tolerated so a payload carrying extra metadata
// does not break the parse.
func (fs *FormState) UnmarshalJSON(data []byte) error {
	type wire struct {
		Name    string          `json:"name"`
		Values  map[string]any  `json:"values"`
		Touched map[string]bool `json:"touched"`
		Errors  Errors          `json:"errors"`
		Meta    map[string]any  `json:"meta"`
	}
	var w wire
	if err := json.Unmarshal(data, &w); err != nil {
		return fmt.Errorf("form.FormState: %w", err)
	}
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.Name = w.Name
	fs.Values = cloneAnyMap(w.Values)
	fs.Touched = cloneBoolMap(w.Touched)
	fs.Errors = cloneErrors(w.Errors)
	fs.Meta = cloneAnyMap(w.Meta)
	return nil
}

// =============================================================================
// Datastar integration helpers
// =============================================================================
//
// These are the bridge between FormState and your existing framework/datastar
// package. They live here (not there) to avoid a circular import: datastar
// is at the framework's top level and shouldn't depend on the form package.

// SignalKey returns the Datastar signals path for a given field. Use it
// when generating `data-bind` or `data-signals-*` attributes in your
// templ renderer, e.g.:
//
//	data-bind={"_signals." + form.SignalKey(state, "email")}
//
// The leading "_signals." prefix is what Datastar expects.
func SignalKey(fs *FormState, fieldKey string) string {
	if fs == nil || fs.Name == "" {
		return signalsKey + "." + fieldKey
	}
	return fmt.Sprintf("%s.%s.%s", signalsKey, fs.Name, fieldKey)
}

// SignalValue returns the JSON-encoded literal that should appear in a
// Datastar `data-signals-foo='{...}'` attribute to initialise a field.
// It is a thin wrapper around json.Marshal so it never produces invalid JS.
func SignalValue(value any) string {
	b, err := json.Marshal(value)
	if err != nil {
		return "null"
	}
	return string(b)
}

// =============================================================================
// Internal helpers
// =============================================================================

func cloneAnyMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func cloneBoolMap(m map[string]bool) map[string]bool {
	out := make(map[string]bool, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func cloneErrors(e Errors) Errors {
	out := make(Errors, len(e))
	for k, v := range e {
		copied := make([]string, len(v))
		copy(copied, v)
		out[k] = copied
	}
	return out
}
