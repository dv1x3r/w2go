// Package w2file contains helpers for multipart uploads sent by w2ui widgets.
//
// The helpers parse the "files[]" multipart field name used by the JavaScript
// upload helper and enforce a simple per-file size limit before the caller opens
// and stores each file.
package w2file
