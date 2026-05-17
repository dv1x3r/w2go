// Package w2sql translates w2ui grid requests into go-sqlbuilder clauses.
//
// Use it when you want to build SQL yourself but still apply w2grid paging,
// sorting, searching, and inline-edit updates consistently. Search and sort
// helpers require a field mapping so client-side field names are explicitly
// whitelisted before they become SQL identifiers.
package w2sql
