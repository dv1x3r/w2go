package main

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/dv1x3r/w2go/w2"
	"github.com/dv1x3r/w2go/w2sort"
	"github.com/dv1x3r/w2go/w2sql"
	"github.com/dv1x3r/w2go/w2ui"

	"github.com/huandu/go-sqlbuilder"
	_ "modernc.org/sqlite"
)

const address = "localhost:3000"

//go:embed index.html
var htmlFS embed.FS

var db *sql.DB

type Todo struct {
	ID          int              `json:"id"`
	Name        string           `json:"name"`
	Description w2.Field[string] `json:"description"`
	Quantity    w2.Field[int]    `json:"quantity"`
	Status      w2.Dropdown      `json:"status"`
}

type Status struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	var err error
	if db, err = sql.Open("sqlite", ":memory:"); err != nil {
		log.Fatalln(err)
	}

	// :memory: database is bound to the connection
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		create table status (
			[id] integer primary key,
			[name] text not null,
			[position] integer not null
		) strict;

		insert into status ([name], [position]) values
			('pending', 0),
			('in progress', 1),
			('completed', 2);

		create table todo (
			[id] integer primary key,
			[name] text not null,
			[description] text not null,
			[quantity] integer not null,
			[status_id] integer not null references status(id) on delete restrict
		) strict;

		insert into todo ([name], [description], [quantity], [status_id]) values
			('buy groceries', 'go to the store and buy some food and drinks', 4, 3),
			('throw out the trash', 'ew, it stinks', 1, 1),
			('build a house', 'build a solid one for your family', 1, 2),
			('plant a tree', 'it is not that hard', 2, 2),
			('raise a son', 'so you can enjoy your food and drinks together', 1, 1);
	`)

	if err != nil {
		log.Fatalln(err)
	}

	router := http.NewServeMux()

	router.Handle("GET /{$}", http.FileServerFS(htmlFS))
	router.Handle("GET /lib/", http.StripPrefix("/lib/", w2ui.FileServerFS()))

	v1 := http.NewServeMux()

	v1.HandleFunc("GET /todo/grid/records", getTodoGridRecords)
	v1.HandleFunc("POST /todo/grid/save", postTodoGridSave)
	v1.HandleFunc("POST /todo/grid/remove", postTodoGridRemove)

	v1.HandleFunc("GET /todo/form", getTodoForm)
	v1.HandleFunc("POST /todo/form", postTodoForm)

	v1.HandleFunc("GET /status/dropdown", getStatusDropdown)
	v1.HandleFunc("GET /status/grid/records", getStatusGridRecords)
	v1.HandleFunc("POST /status/grid/reorder", postStatusGridReorder)

	router.Handle("/api/v1/", http.StripPrefix("/api/v1", v1))

	log.Println("listening on: " + address)
	if err := http.ListenAndServe(address, router); err != nil {
		log.Fatalln(err)
	}
}

func getTodoGridRecords(w http.ResponseWriter, r *http.Request) {
	req, err := w2.ParseGetGridRequest(r.URL.Query().Get("request"))
	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusBadRequest)
		return
	}

	var total int
	var records []Todo

	sb := sqlbuilder.Select("count(*)").From("todo as t")
	w2sql.Where(sb, req, map[string]string{
		"id":          "t.id",
		"name":        "t.name",
		"description": "t.description",
		"quantity":    "t.quantity",
		"status":      "t.status_id",
	})

	// query total number of rows with w2grid filters applied
	query, args := sb.BuildWithFlavor(sqlbuilder.SQLite)
	row := db.QueryRow(query, args...)
	if err := row.Scan(&total); err != nil && !errors.Is(err, sql.ErrNoRows) {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}

	// reuse the same query builder for data records
	sb.Select(
		"t.id",
		"t.name",
		"t.description",
		"t.quantity",
		"t.status_id",
		"s.name as status_name",
	)
	sb.JoinWithOption(sqlbuilder.LeftJoin, "status as s", "s.id = t.status_id")
	w2sql.OrderBy(sb, req, map[string]string{
		"id":          "t.id",
		"name":        "t.name",
		"description": "t.description",
		"quantity":    "t.quantity",
		"status":      "s.name",
	})

	w2sql.Limit(sb, req)
	w2sql.Offset(sb, req)

	query, args = sb.BuildWithFlavor(sqlbuilder.SQLite)
	rows, err := db.Query(query, args...)
	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var record Todo
		err = rows.Scan(
			&record.ID,
			&record.Name,
			&record.Description,
			&record.Quantity,
			&record.Status.ID,
			&record.Status.Text,
		)
		if err != nil {
			res := w2.NewErrorResponse(err.Error())
			res.Write(w, http.StatusInternalServerError)
			return
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}

	res := w2.NewGetGridResponse(records, total)
	res.Write(w)
}

func postTodoGridSave(w http.ResponseWriter, r *http.Request) {
	req, err := w2.ParseSaveGridRequest[Todo](r.Body)
	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	for _, change := range req.Changes {
		ub := sqlbuilder.Update("todo")
		ub.Where(ub.EQ("id", change.ID))

		w2sql.SetNoNull(ub, change.Description, "description")
		w2sql.Set(ub, change.Quantity, "quantity")
		w2sql.Set(ub, change.Status.ID, "status_id")

		query, args := ub.BuildWithFlavor(sqlbuilder.SQLite)
		if _, err := tx.Exec(query, args...); err != nil {
			res := w2.NewErrorResponse(err.Error())
			res.Write(w, http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}

	res := w2.NewSuccessResponse()
	res.Write(w, http.StatusOK)
}

func postTodoGridRemove(w http.ResponseWriter, r *http.Request) {
	req, err := w2.ParseRemoveGridRequest(r.Body)
	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusBadRequest)
		return
	}

	dlb := sqlbuilder.DeleteFrom("todo")
	dlb.Where(dlb.In("id", sqlbuilder.List(req.ID)))

	query, args := dlb.BuildWithFlavor(sqlbuilder.SQLite)
	if _, err := db.Exec(query, args...); err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}

	res := w2.NewSuccessResponse()
	res.Write(w, http.StatusOK)
}

func getTodoForm(w http.ResponseWriter, r *http.Request) {
	req, err := w2.ParseGetFormRequest(r.URL.Query().Get("request"))
	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusBadRequest)
		return
	}

	var record Todo

	sb := sqlbuilder.Select(
		"t.id",
		"t.name",
		"t.description",
		"t.quantity",
		"t.status_id",
		"s.name as status_name",
	).From("todo as t")
	sb.JoinWithOption(sqlbuilder.LeftJoin, "status as s", "s.id = t.status_id")
	sb.Where(sb.EQ("t.id", req.RecID))

	query, args := sb.BuildWithFlavor(sqlbuilder.SQLite)
	row := db.QueryRow(query, args...)
	err = row.Scan(
		&record.ID,
		&record.Name,
		&record.Description,
		&record.Quantity,
		&record.Status.ID,
		&record.Status.Text,
	)
	if errors.Is(err, sql.ErrNoRows) {
		res := w2.NewErrorResponse(http.StatusText(http.StatusNotFound))
		res.Write(w, http.StatusNotFound)
		return
	} else if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}

	res := w2.NewGetFormResponse(record)
	res.Write(w)
}

func postTodoForm(w http.ResponseWriter, r *http.Request) {
	req, err := w2.ParseSaveFormRequest[Todo](r.Body)
	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusBadRequest)
		return
	}

	if req.RecID == 0 {
		const query = "insert into todo (name, description, quantity, status_id) values (?, ?, ?, ?);"
		res, err := db.Exec(query, req.Record.Name, req.Record.Description, req.Record.Quantity, req.Record.Status.ID)
		if err != nil {
			res := w2.NewErrorResponse(err.Error())
			res.Write(w, http.StatusInternalServerError)
			return
		}
		lastInsertId, err := res.LastInsertId()
		if err != nil {
			res := w2.NewErrorResponse(err.Error())
			res.Write(w, http.StatusInternalServerError)
			return
		}
		req.RecID = int(lastInsertId)
	} else {
		const query = "update todo set name = ?, description = ?, quantity = ?, status_id = ? where id = ?;"
		_, err := db.Exec(query, req.Record.Name, req.Record.Description, req.Record.Quantity, req.Record.Status.ID, req.RecID)
		if err != nil {
			res := w2.NewErrorResponse(err.Error())
			res.Write(w, http.StatusInternalServerError)
			return
		}
	}

	res := w2.NewSaveFormResponse(req.RecID)
	res.Write(w)
}

func getStatusDropdown(w http.ResponseWriter, r *http.Request) {
	req, err := w2.ParseGetDropdownRequest(r.URL.Query().Get("request"))
	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusBadRequest)
		return
	}

	var records []w2.Dropdown

	const query = "select id, name from status where name like ? order by position limit ?;"
	rows, err := db.Query(query, fmt.Sprintf("%%%s%%", req.Search), req.Max)
	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var record w2.Dropdown
		if err := rows.Scan(&record.ID, &record.Text); err != nil {
			res := w2.NewErrorResponse(err.Error())
			res.Write(w, http.StatusInternalServerError)
			return
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}

	res := w2.NewGetDropdownResponse(records)
	res.Write(w)
}

func getStatusGridRecords(w http.ResponseWriter, r *http.Request) {
	req, err := w2.ParseGetGridRequest(r.URL.Query().Get("request"))
	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusBadRequest)
		return
	}

	var total int
	var records []Status

	sb := sqlbuilder.Select("count(*)").From("status")
	query, args := sb.BuildWithFlavor(sqlbuilder.SQLite)
	row := db.QueryRow(query, args...)
	if err := row.Scan(&total); err != nil && !errors.Is(err, sql.ErrNoRows) {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}

	sb.Select("id", "name")
	sb.OrderByAsc("position")

	w2sql.Limit(sb, req)
	w2sql.Offset(sb, req)

	query, args = sb.BuildWithFlavor(sqlbuilder.SQLite)
	rows, err := db.Query(query, args...)
	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var record Status
		if err := rows.Scan(&record.ID, &record.Name); err != nil {
			res := w2.NewErrorResponse(err.Error())
			res.Write(w, http.StatusInternalServerError)
			return
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}

	res := w2.NewGetGridResponse(records, total)
	res.Write(w)
}

func postStatusGridReorder(w http.ResponseWriter, r *http.Request) {
	req, err := w2.ParseReorderGridRequest(r.Body)
	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var ids []int

	rows, err := tx.Query("select id from status order by position;")
	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			res := w2.NewErrorResponse(err.Error())
			res.Write(w, http.StatusInternalServerError)
			return
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}

	if err := w2sort.ReorderArray(ids, req); err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}

	for i, id := range ids {
		if _, err := tx.Exec("update status set position = ? where id = ?;", i, id); err != nil {
			res := w2.NewErrorResponse(err.Error())
			res.Write(w, http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}

	res := w2.NewSuccessResponse()
	res.Write(w, http.StatusOK)
}
