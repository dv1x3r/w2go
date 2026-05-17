// Package w2db provides database helpers for common w2ui endpoints.
//
// The helpers build SQL with go-sqlbuilder, execute it through database/sql,
// scan records into caller-defined Go structs, and return the response types
// from package w2. They are useful when a handler mostly needs to load, save,
// remove, or reorder records for w2grid, w2form, or dropdown controls.
//
// All query helpers accept QueryExecer, which is satisfied by both *sql.DB and
// *sql.Tx. That keeps handlers easy to write while still letting you wrap
// several operations in one transaction.
//
// Table names, column names, and SQL expressions in option structs are treated
// as trusted SQL fragments. Do not copy client-provided strings into those
// fields. For grid search and sort values, use WhereMapping and OrderByMapping
// to whitelist the client field names that may be translated into SQL.
package w2db
