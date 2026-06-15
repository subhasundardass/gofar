package form

// =============================================================================
// Reactive Form Runtime
// =============================================================================
//
// The runtime evaluates the Rule Engine against a live *FormState. All
// functions accept *FormState (never FormState by value) to avoid copying
// the embedded sync.RWMutex.
//
// Call order in a typical Datastar POST handler:
//
//  1. ApplyComputedFields(form, state)   — run ComputeWith functions
//  2. IsVisible / IsEnabled / IsRequired — used during rendering
//  3. ValidateWithRules(inst, state)     — skip invisible/disabled fields,
//                                          inject dynamic RequiredValidators

// ApplyComputedFields runs every field's ComputeWith function against state
// and writes the result back. Call this BEFORE any rule evaluation so that
// rules depending on computed values see the correct data.
func ApplyComputedFields(f *Form, state *FormState) {
	if f == nil || state == nil {
		return
	}
	for _, field := range f.GetFields() {
		fn := field.GetComputeWith()
		if fn == nil {
			continue
		}
		// Snapshot gives the compute fn a stable read view while we
		// write the result back to the live state.
		computed := fn(state.Snapshot())
		state.SetSilent(field.GetKey(), computed)
	}
}

// IsVisible reports whether field should be rendered given the current state.
// Returns true if no VisibleIf rule is set.
func IsVisible(f Field, state *FormState) bool {
	r := f.GetVisibleIf()
	if r == nil {
		return true
	}
	return r.Eval(state)
}

// IsEnabled reports whether field should be interactive given the current state.
// EnabledIf overrides the static Disabled flag when set.
func IsEnabled(f Field, state *FormState) bool {
	r := f.GetEnabledIf()
	if r == nil {
		return !f.GetDisabled()
	}
	return r.Eval(state)
}

// IsRequired reports whether field is required given the current state.
// RequiredIf overrides the static Required flag when set.
func IsRequired(f Field, state *FormState) bool {
	r := f.GetRequiredIf()
	if r == nil {
		return f.GetRequired()
	}
	return r.Eval(state)
}

// IsComputed reports whether field has a compute function attached.
func IsComputed(f Field) bool {
	return f.GetComputeWith() != nil
}

// FieldEval bundles all rule results for a single field so renderers only
// pay the evaluation cost once per field per render pass.
type FieldEval struct {
	Visible  bool
	Enabled  bool
	Required bool
	Computed bool
}

// EvalField evaluates all rules for a field in one shot.
func EvalField(f Field, state *FormState) FieldEval {
	return FieldEval{
		Visible:  IsVisible(f, state),
		Enabled:  IsEnabled(f, state),
		Required: IsRequired(f, state),
		Computed: IsComputed(f),
	}
}

// ValidateWithRules is a rule-aware replacement for Instance.Validate().
// It:
//   - Skips fields that are invisible (VisibleIf → false)
//   - Skips fields that are disabled (EnabledIf → false)
//   - Injects a dynamic RequiredValidator when RequiredIf → true
//
// The instance's Errors map is replaced on every call.
func ValidateWithRules(inst *Instance, state *FormState) *Instance {
	inst.Errors = Errors{}
	if inst.Form == nil {
		return inst
	}

	for _, field := range inst.Form.GetFields() {
		key := field.GetKey()

		if !IsVisible(field, state) {
			continue
		}
		if !IsEnabled(field, state) {
			continue
		}

		value := inst.Values[key]

		if IsRequired(field, state) {
			rv := RequiredValidator{Field: key}
			if err := rv.Validate(value); err != nil {
				inst.collectError(key, rv.Name(), err)
			}
		}

		for _, v := range field.GetValidators() {
			// Skip static RequiredValidator when RequiredIf already ran it.
			if _, ok := v.(RequiredValidator); ok && field.GetRequiredIf() != nil {
				continue
			}
			if err := v.Validate(value); err != nil {
				inst.collectError(key, v.Name(), err)
			}
		}
	}
	return inst
}
