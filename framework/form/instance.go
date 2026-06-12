package form

// Instance is a per-request snapshot of a Form. It holds three things:
//
//   - Form:     pointer back to the static schema (so a renderer can ask
//     "what type is this field?" without re-deriving it).
//   - Values:   the current value of every field. The form is "complete"
//     when every key the schema declares has a value here.
//   - Errors:   per-field validation errors collected by the most recent
//     call to Validate. Empty after a successful validation.
//
// Typical flow:
//
//	form := form.New("signup").Fields(...)
//	inst := form.NewInstance()              // defaults populated
//	inst.Set("email", "alice@example.com")  // user input
//	inst.Validate()                          // runs every validator
//	render(inst)                             // templ/Datastar sees Values + Errors
type Instance struct {
	Form   *Form
	Values map[string]any
	Errors Errors
}

// NewInstance builds a fresh instance, pre-filling every value with the
// field's DefaultValue. It is the same as form.NewInstance(); kept as a
// package-level helper for callers who only have a *Form in hand.
func NewInstance(form *Form) *Instance {
	if form == nil {
		return &Instance{
			Values: map[string]any{},
			Errors: Errors{},
		}
	}
	return form.NewInstance()
}

// Set stores value under key, overwriting any previous value. The key is
// NOT validated against the schema — passing a key the form does not
// declare is allowed and the value will simply be ignored by Validate.
func (i *Instance) Set(key string, value any) *Instance {
	if i.Values == nil {
		i.Values = map[string]any{}
	}
	i.Values[key] = value
	return i
}

// Get returns the current value for key (or nil if absent).
func (i *Instance) Get(key string) any {
	if i.Values == nil {
		return nil
	}
	return i.Values[key]
}

// HasError reports whether the given field has any validation errors.
func (i *Instance) HasError(key string) bool {
	return len(i.Errors[key]) > 0
}

// FirstError returns the first error message for a field, or "" if none.
func (i *Instance) FirstError(key string) string {
	if msgs := i.Errors[key]; len(msgs) > 0 {
		return msgs[0]
	}
	return ""
}

// ClearErrors drops all errors. Useful when re-displaying a form after
// a successful submission.
func (i *Instance) ClearErrors() *Instance {
	i.Errors = Errors{}
	return i
}

// Apply copies values, touched flags, and errors from a FormState into
// the Instance. It is the inverse of FormState.ToInstance and is what a
// Datastar POST handler calls after unmarshalling the incoming patch:
//
//	form  := lookupForm()       // static, registered at boot
//	state := form.NewFormState(form.Name(), nil)
//	_ = json.Unmarshal(c.Body(), state)
//	inst  := form.NewInstance(form).Apply(state)
//	inst.Validate()
//
// Apply does NOT trigger validation — call inst.Validate() afterwards.
func (i *Instance) Apply(state *FormState) *Instance {
	if i == nil || state == nil {
		return i
	}
	state.mu.RLock()
	defer state.mu.RUnlock()
	for k, v := range state.Values {
		if v == nil {
			delete(i.Values, k)
			continue
		}
		i.Values[k] = v
	}
	return i
}

// Validate runs every validator on every field of the underlying form and
// collects errors into i.Errors. It returns the receiver so it can be
// chained: inst.Set(...).Validate().
//
// The receiver's Errors map is *replaced* on each call, not appended to,
// so a second Validate() will not double-count errors from the first.
func (i *Instance) Validate() *Instance {
	i.Errors = Errors{}
	if i.Form == nil {
		return i
	}
	for _, field := range i.Form.GetFields() {
		key := field.GetKey()
		value := i.Values[key]
		for _, v := range field.GetValidators() {
			if err := v.Validate(value); err != nil {
				i.collectError(key, v.Name(), err)
			}
		}
	}
	return i
}

// Valid reports whether the most recent Validate() pass produced no errors.
func (i *Instance) Valid() bool {
	return !i.Errors.HasAny()
}

// collectError extracts a field/code/message tuple from any error and
// appends it to the Errors map. Plain `error` values fall back to a
// generic "invalid" code so custom validators that just return
// fmt.Errorf(...) still get surfaced.
func (i *Instance) collectError(key, validatorName string, err error) {
	if err == nil {
		return
	}
	if ve, ok := err.(*ValidationError); ok {
		i.Errors[key] = append(i.Errors[key], ve.Message)
		return
	}
	_ = validatorName // reserved for future per-validator metadata
	i.Errors[key] = append(i.Errors[key], err.Error())
}
