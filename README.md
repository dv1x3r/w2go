# Go bindings for w2ui

This package offers Go bindings for the [w2ui JavaScript UI Library](https://github.com/vitmalina/w2ui).

- [Features](#features)
- [Install](#install)
- [Usage](#usage)
  - [w2grid](#w2grid)
  - [w2form](#w2form)
  - [Dropdown](#dropdown)
  - [SQL Builder](#sql-builder)
  - [Utils](#utils)
- [Example](#example)
- [Changelog](#changelog)
- [License](#license)

## Features

- `w2grid` support in `JSON` mode:
  - **Pagination**, **sorting**, and **search**;
  - **Inline editing** with typed updates;
  - **Batch delete**;
  - **Drag and Drop row reordering**;

- `w2form` support in `JSON` mode:
  - **Record retrieval**;
  - **Form submission (create/update)**;

- **Dropdowns**:
  - Reusable across `w2grid` and `w2form` components;
  - **Searchable** and **dynamic**;

- **SQL Builder integration**:
  - Translate `w2grid` data request into SQL with `go-sqlbuilder`;

## Install

```shell
go get github.com/dv1x3r/w2go
```

## Usage

`w2go` types are JSON-serializable and work seamlessly with frameworks like `Echo` or `Fiber`, or standard `net/http`.
The following examples use Go's standard `net/http` library: `(w http.ResponseWriter, r *http.Request)`.

### w2grid

**Get Records**

```go
req, err := w2.ParseGridDataRequest(r.URL.Query().Get("request"))
if err != nil {
	res := w2.NewErrorResponse(err.Error())
	res.Write(w, http.StatusBadRequest)
	return
}

// Process the request...

records := []Todo{{ID: 1, Name: "Buy groceries"}}
res := w2.NewGridDataResponse(records, len(records))
res.Write(w)
```

**Save Changes**

```go
req, err := w2.ParseGridSaveRequest[Todo](r.Body)
if err != nil {
	res := w2.NewErrorResponse(err.Error())
	res.Write(w, http.StatusBadRequest)
	return
}

// Apply updates for req.Changes slice...

res := w2.NewSuccessResponse()
res.Write(w, http.StatusOK)
```

**Remove Records**

```go
req, err := w2.ParseGridRemoveRequest(r.Body)
if err != nil {
	res := w2.NewErrorResponse(err.Error())
	res.Write(w, http.StatusBadRequest)
	return
}

// Apply updates for req.ID slice...

res := w2.NewSuccessResponse()
res.Write(w, http.StatusOK)
```

**Reorder Rows**

```go
req, err := w2.ParseGridReorderRequest(r.Body)
if err != nil {
	res := w2.NewErrorResponse(err.Error())
	res.Write(w, http.StatusBadRequest)
	return
}

// Apply updates based on req.RecID, req.MoveBefore and req.Last...

res := w2.NewSuccessResponse()
res.Write(w, http.StatusOK)
```

### w2form

**Get Record**

```go
req, err := w2.ParseFormGetRequest(r.URL.Query().Get("request"))
if err != nil {
	res := w2.NewErrorResponse(err.Error())
	res.Write(w, http.StatusBadRequest)
	return
}

// Process the request based on req.RecID...

record := Todo{ID: 1, Name: "Example"}
res := w2.NewFormGetResponse(record)
res.Write(w)
```

**Save Record**

```go
req, err := w2.ParseFormSaveRequest[Todo](r.Body)
if err != nil {
	res := w2.NewErrorResponse(err.Error())
	res.Write(w, http.StatusBadRequest)
	return
}

// Insert or update based on req.RecId...

res := w2.NewFormSaveResponse(req.RecID)
res.Write(w)
```

### Dropdown

**Get Records**

```go
req, err := w2.ParseDropdownRequest(r.URL.Query().Get("request"))
if err != nil {
	res := w2.NewErrorResponse(err.Error())
	res.Write(w, http.StatusBadRequest)
	return
}

// Process the request based on req.Max and req.Search...

records := []w2.DropdownValue{{ID: 1, Text: "Open"}, {ID: 2, Text: "Closed"}}
res := w2.NewDropdownResponse(records)
res.Write(w)
```

### SQL Builder

Use `w2sqlbuilder` with `github.com/huandu/go-sqlbuilder` to apply filters, sorters, limits, and updates.

**Apply Filters**

```go
req, _ := w2.ParseGridDataRequest(r.URL.Query().Get("request"))

// Create a SelectBuilder
sb := sqlbuilder.NewSelectBuilder()
sb.Select("t.id", "t.name")
sb.From("todo as t")

// Define the w2ui field to database field name mapping.
mapping := map[string]string{
	"id":   "t.id",
	"name": "t.name",
}

// Apply query filters based on the request and mapping.
w2sqlbuilder.Where(sb, req, mapping)
```

**Apply Sorters**

```go
req, _ := w2.ParseGridDataRequest(r.URL.Query().Get("request"))

// Create a SelectBuilder
sb := sqlbuilder.NewSelectBuilder()
sb.Select("t.id", "t.name")
sb.From("todo as t")

// Define the w2ui field to database field name mapping.
mapping := map[string]string{
	"id":   "t.id",
	"name": "t.name",
}

// Apply query sorters based on the request and mapping.
w2sqlbuilder.OrderBy(sb, req, mapping)
```

**Apply Limits**

```go
req, _ := w2.ParseGridDataRequest(r.URL.Query().Get("request"))

// Create a SelectBuilder
sb := sqlbuilder.NewSelectBuilder()
sb.Select("t.id", "t.name")
sb.From("todo as t")

// Apply limit and offset based on the request.
w2sqlbuilder.Limit(sb, req)
w2sqlbuilder.Offset(sb, req)
```

**Apply Updates**

Use this approach to apply **inline changes** to editable `w2grid` fields.

To track whether a field was included in the request, wrap it with the `w2.Editable` type:

```go
type Todo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Description w2.Editable[string] `json:"description"`
	Quantity    w2.Editable[int]    `json:"quantity"`
}
```

Update functions:

- `w2sqlbuilder.SetEditable`: updates the field only if a value **is provided**. Uses **null** if the value **is empty**.
- `w2sqlbuilder.SetEditableWithDefault`: updates the field only if a value **is provided**. Uses the **zero value** if value **is empty**.

```go
req, _ := w2.ParseGridSaveRequest[Todo](r.Body)

for _, change := range req.Changes {
	ub := sqlbuilder.Update("todo")
	ub.Where(ub.EQ("id", change.ID))

	w2sqlbuilder.SetEditable(ub, change.Description, "description")
	w2sqlbuilder.SetEditable(ub, change.Quantity, "quantity")

	// ...
}
```

### Utils

- `w2sort.ReorderArray`: reorders a slice of integers based on `w2grid` drag and drop.

## Example

Run a **full** CRUD demo using in-memory SQLite:

```shell
go run example/main.go
```

Starts the server on `http://localhost:3000`

## Changelog

- v0.2.8 2025-12-06 - Update go-sqlbuilder to v1.38
- v0.2.7 2025-10-19 - EditableDropdown support for clean w2form records
- v0.2.6 2025-05-24 - Write methods now return errors
- v0.2.5 2025-05-16 - First stable release

## License

Licensed under the MIT license.
