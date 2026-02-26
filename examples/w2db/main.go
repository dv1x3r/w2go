package main

import (
	"database/sql"
	"embed"
	"errors"
	"log"
	"net/http"

	"github.com/dv1x3r/w2go/w2"
	"github.com/dv1x3r/w2go/w2db"
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

	res, err := w2db.GetGrid(db, req, w2db.GetGridOptions[Todo]{
		From: "todo as t",
		Select: []string{
			"t.id",
			"t.name",
			"t.description",
			"t.quantity",
			"t.status_id",
			"s.name as status_name",
		},
		WhereMapping: map[string]string{
			"id":          "t.id",
			"name":        "t.name",
			"description": "t.description",
			"quantity":    "t.quantity",
			"status":      "t.status_id",
		},
		OrderByMapping: map[string]string{
			"id":          "t.id",
			"name":        "t.name",
			"description": "t.description",
			"quantity":    "t.quantity",
			"status":      "s.name",
		},
		Flavor: sqlbuilder.SQLite,
		BuildBase: func(sb *sqlbuilder.SelectBuilder) {
			sb.JoinWithOption(sqlbuilder.LeftJoin, "status as s", "s.id = t.status_id")
		},
		Scan: func(rows *sql.Rows) (Todo, error) {
			var record Todo
			return record, rows.Scan(
				&record.ID,
				&record.Name,
				&record.Description,
				&record.Quantity,
				&record.Status.ID,
				&record.Status.Text,
			)
		},
	})

	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}

	res.Write(w)
}

func postTodoGridSave(w http.ResponseWriter, r *http.Request) {
	req, err := w2.ParseSaveGridRequest[Todo](r.Body)
	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusBadRequest)
		return
	}

	err = w2db.WithinTransaction(db, func(tx *sql.Tx) error {
		_, err := w2db.SaveGrid(tx, req, w2db.SaveGridOptions[Todo]{
			Flavor: sqlbuilder.SQLite,
			BuildUpdate: func(change Todo) *sqlbuilder.UpdateBuilder {
				ub := sqlbuilder.Update("todo")
				ub.Where(ub.EQ("id", change.ID))
				w2sql.SetNoNull(ub, change.Description, "description")
				w2sql.Set(ub, change.Quantity, "quantity")
				w2sql.Set(ub, change.Status.ID, "status_id")
				return ub
			},
		})
		return err
	})

	if err != nil {
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

	_, err = w2db.RemoveGrid(db, req, w2db.RemoveGridOptions{
		From:    "todo",
		IDField: "id",
		Flavor:  sqlbuilder.SQLite,
	})

	if err != nil {
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

	res, err := w2db.GetForm(db, req, w2db.GetFormOptions[Todo]{
		From:    "todo as t",
		IDField: "t.id",
		Select: []string{
			"t.id",
			"t.name",
			"t.description",
			"t.quantity",
			"t.status_id",
			"s.name as status_name",
		},
		Flavor: sqlbuilder.SQLite,
		BuildSelect: func(sb *sqlbuilder.SelectBuilder) {
			sb.JoinWithOption(sqlbuilder.LeftJoin, "status as s", "s.id = t.status_id")
		},
		Scan: func(row *sql.Row) (Todo, error) {
			var record Todo
			return record, row.Scan(
				&record.ID,
				&record.Name,
				&record.Description,
				&record.Quantity,
				&record.Status.ID,
				&record.Status.Text,
			)
		},
	})

	if errors.Is(err, sql.ErrNoRows) {
		res := w2.NewErrorResponse(http.StatusText(http.StatusNotFound))
		res.Write(w, http.StatusNotFound)
		return
	} else if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}

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
		recID, err := w2db.InsertForm(db, req, w2db.InsertFormOptions{
			Into:   "todo",
			Cols:   []string{"name", "description", "quantity", "status_id"},
			Values: []any{req.Record.Name, req.Record.Description, req.Record.Quantity, req.Record.Status.ID},
			Flavor: sqlbuilder.SQLite,
		})
		if err != nil {
			res := w2.NewErrorResponse(err.Error())
			res.Write(w, http.StatusInternalServerError)
			return
		}
		req.RecID = recID
	} else {
		_, err := w2db.UpdateForm(db, req, w2db.UpdateFormOptions{
			Update:  "todo",
			IDField: "id",
			Cols:    []string{"name", "description", "quantity", "status_id"},
			Values:  []any{req.Record.Name, req.Record.Description, req.Record.Quantity, req.Record.Status.ID},
			Flavor:  sqlbuilder.SQLite,
		})
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

	res, err := w2db.GetDropdown(db, req, w2db.GetDropdownOptions{
		From:         "status",
		IDField:      "id",
		TextField:    "name",
		OrderByField: "position",
		Flavor:       sqlbuilder.SQLite,
	})

	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}

	res.Write(w)
}

func getStatusGridRecords(w http.ResponseWriter, r *http.Request) {
	req, err := w2.ParseGetGridRequest(r.URL.Query().Get("request"))
	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusBadRequest)
		return
	}

	res, err := w2db.GetGrid(db, req, w2db.GetGridOptions[Status]{
		From:   "status",
		Select: []string{"id", "name"},
		Flavor: sqlbuilder.SQLite,
		BuildSelect: func(sb *sqlbuilder.SelectBuilder) {
			sb.OrderByAsc("position")
		},
		Scan: func(rows *sql.Rows) (Status, error) {
			var record Status
			return record, rows.Scan(&record.ID, &record.Name)
		},
	})

	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}

	res.Write(w)
}

func postStatusGridReorder(w http.ResponseWriter, r *http.Request) {
	req, err := w2.ParseReorderGridRequest(r.Body)
	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusBadRequest)
		return
	}

	err = w2db.WithinTransaction(db, func(tx *sql.Tx) error {
		_, err := w2db.ReorderGrid(tx, req, w2db.ReorderGridOptions{
			Update:   "status",
			IDField:  "id",
			SetField: "position",
			Flavor:   sqlbuilder.SQLite,
		})
		return err
	})

	if err != nil {
		res := w2.NewErrorResponse(err.Error())
		res.Write(w, http.StatusInternalServerError)
		return
	}

	res := w2.NewSuccessResponse()
	res.Write(w, http.StatusOK)
}
