package datastar

import ds "github.com/starfederation/datastar-go/datastar"

type PatchElementOption = ds.PatchElementOption

// WithSelector targets an element using a CSS selector.
//
// Use this when the target element does not have a unique ID or when
// you want to target elements using class names, attributes, or other
// CSS selector expressions.
//
// Example:
//
//	return datastar.MergeFragmentTempl(
//		c,
//		views.CustomerRow(customer),
//		datastar.WithSelector("#customer-table-body"),
//		datastar.WithModeAppend(),
//	)
//
// Result:
// The rendered fragment will be appended inside the element matching
// the selector "#customer-table-body".
func WithSelector(selector string) PatchElementOption {
	return ds.WithSelector(selector)
}

// WithSelectorID targets an element by its HTML id attribute.
//
// This is the preferred option when the target element has a unique ID,
// as it is simpler and slightly more efficient than a CSS selector.
//
// Example:
//
//	return datastar.MergeFragmentTempl(
//		c,
//		views.CustomerRow(customer),
//		datastar.WithSelectorID("customer-table-body"),
//		datastar.WithModeAppend(),
//	)
//
// HTML:
//
//	<tbody id="customer-table-body"></tbody>
func WithSelectorID(id string) PatchElementOption {
	return ds.WithSelectorID(id)
}

// WithModeAppend appends the fragment as the last child of the target element.
//
// Common use cases:
//   - Add a row to a table
//   - Add an item to a list
//   - Load more records
//
// Example:
//
//	return datastar.MergeFragmentTempl(
//		c,
//		views.JournalEntryRow(index, ledgers),
//		datastar.WithSelectorID("entry-body"),
//		datastar.WithModeAppend(),
//	)
func WithModeAppend() PatchElementOption {
	return ds.WithModeAppend()
}

// WithModePrepend inserts the fragment as the first child of the target element.
//
// Common use cases:
//   - Show newest notifications first
//   - Add chat messages at the top
//
// Example:
//
//	return datastar.MergeFragmentTempl(
//		c,
//		views.Notification(notification),
//		datastar.WithSelectorID("notifications"),
//		datastar.WithModePrepend(),
//	)
func WithModePrepend() PatchElementOption {
	return ds.WithModePrepend()
}

// WithModeAfter inserts the fragment immediately after the target element.
//
// Common use cases:
//   - Insert a new table row after the current row
//   - Add dynamic content after a section
//
// Example:
//
//	return datastar.MergeFragmentTempl(
//		c,
//		views.JournalEntryRow(index, ledgers),
//		datastar.WithSelectorID("row-1"),
//		datastar.WithModeAfter(),
//	)
func WithModeAfter() PatchElementOption {
	return ds.WithModeAfter()
}

// WithModeBefore inserts the fragment immediately before the target element.
//
// Common use cases:
//   - Insert a row before another row
//   - Insert validation messages before a form section
//
// Example:
//
//	return datastar.MergeFragmentTempl(
//		c,
//		views.WarningMessage(),
//		datastar.WithSelectorID("form-section"),
//		datastar.WithModeBefore(),
//	)
func WithModeBefore() PatchElementOption {
	return ds.WithModeBefore()
}

// WithModeReplace replaces the target element with the new fragment.
//
// Common use cases:
//   - Refresh a form
//   - Update a table row
//   - Replace a card or widget
//
// Example:
//
//	return datastar.MergeFragmentTempl(
//		c,
//		views.CustomerForm(customer),
//		datastar.WithSelectorID("customer-form"),
//		datastar.WithModeReplace(),
//	)
//
// HTML:
//
//	<div id="customer-form"></div>
func WithModeReplace() PatchElementOption {
	return ds.WithModeReplace()
}
