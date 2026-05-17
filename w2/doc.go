// Package w2 contains the core request and response types for serving w2ui
// widgets from Go.
//
// The package covers the JSON protocol used by w2grid, w2form, dropdown
// controls, and drag-and-drop grid reordering. Use the Parse functions to
// decode incoming w2ui payloads, build responses with the New constructors,
// and call Write to send JSON back through an http.ResponseWriter.
//
// The package is intentionally framework-neutral. Query-string parsers accept
// the raw "request" parameter value used by w2ui, body parsers accept an
// io.Reader, and responses write to the standard net/http response interface.
//
// The Field type is useful for editable grids and forms where the client may
// omit unchanged values. It keeps "field was not sent", "field was sent as
// null", and "field was sent with a value" separate so database helpers can
// update only the columns the user actually changed.
package w2
