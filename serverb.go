package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"todo/models"
	"github.com/boltdb/bolt"
	"github.com/codegangsta/negroni"
	"github.com/eknkc/amber"
	"github.com/gorilla/mux"
	// "github.com/jehiah/go-strftime"
	"github.com/unrolled/render"
)

func checkError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func fatalIfError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

var (
	compiler *amber.Compiler
)

type Db struct {
	db *bolt.DB
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if compiler == nil {
		compiler = amber.New()
	}
}

func main() {
	// Create router
	r := mux.NewRouter()
	s := http.StripPrefix("/web/", http.FileServer(http.Dir("web")))
	r.PathPrefix("/web/").Handler(s)

	compiler = amber.New()
	// GET /todo, mendapatkan daftar todo
	r.HandleFunc("/", IndexHandler).Methods("GET")
	// POST /todo, tambah todo
	r.HandleFunc("/todo", AddTodoHandler).Methods("POST")
	// DELETE /todo/1, hapus todo
	r.HandleFunc("/todo/{id:[0-9]+}", DelTodoHandler).Methods("DELETE")
	// POST /todo/1, ubah status done
	r.HandleFunc("/todo/done", DoneTodoHandler).Methods("POST")
	// POST /todo/1, ubah todo
	r.HandleFunc("/todo/{id:[0-9]+}", EditTodoHandler).Methods("POST")

	n := negroni.Classic()
	n.UseHandler(r)

	// User env PORT if exists, otherwise use default from code
	// Linux: export PORT=3000
	// Windows: set PORT=3000
	port := "3000"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	host := "localhost"
	if os.Getenv("HOST") != "" {
		host = os.Getenv("HOST")
	}
	n.Run(host + ":" + port)
}

func (d *Db) Connect() error {
	// db, err := bolt.Open("my.db", 0777, &bolt.Options{Timeout: 30 * time.Second})
	db, err := bolt.Open("my.db", 0777, nil)
	if err != nil {
		return fmt.Errorf("Tidak dapat membuka database: %s.", err)
	}
	d.db = db
	return nil
}

func (d *Db) getNextId(bucketName string) int64 {
	var result int64 = 1
	d.db.Update(func(tx *bolt.Tx) error {
		bck := tx.Bucket([]byte(bucketName))
		if bck == nil {
			log.Fatalln("Bucket todo tidak ada.")
		}

		id, _ := bck.NextSequence()
		if id > 0 {
			result = int64(id)
		}
		return nil
	})

	return result
}

func intToByte(id int64) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, id)
	if err != nil {
		return nil, fmt.Errorf("intToByte: %s", err)
	}
	return buf.Bytes(), nil
}

func IntToByte32(id int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, id)
	if err != nil {
		return nil, fmt.Errorf("intToByte: %s", err)
	}
	return buf.Bytes(), nil
}

func renderToBytes(viewName string, data interface{}) []byte {
	var buf bytes.Buffer
	err := compiler.ParseFile(viewName)
	if err != nil {
		log.Fatalln(err)
	}

	tpl, err := compiler.Compile()
	if err != nil {
		log.Fatalln(err)
	}

	tpl.Execute(&buf, data)
	return buf.Bytes()
}

func IndexHandler(w http.ResponseWriter, req *http.Request) {
	type (
		ListTodo struct {
			Todos []models.Todo `json:"todos"`
		}

		Response struct {
			Error int    `json:"error"`
			Msg   string `json:"msg"`
		}
	)

	conn := new(Db)
	err := conn.Connect()
	fatalIfError(err)
	defer conn.db.Close()

	err = compiler.ParseFile("./views/index.amber")
	if err != nil {
		log.Fatalln(err)
	}

	tpl, err := compiler.Compile()
	if err != nil {
		log.Fatalln(err)
	}

	listTodo := new(ListTodo)
	todos := make([]models.Todo, 0)
	err = conn.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("todo"))
		b.ForEach(func(k, v []byte) error {
			todo := models.Todo{}
			_ = json.Unmarshal(v, &todo)
			todos = append(todos, todo)
			return nil
		})
		return nil
	})
	listTodo.Todos = todos
	tpl.Execute(w, listTodo)
}

func AddTodoHandler(w http.ResponseWriter, req *http.Request) {
	type Response struct {
		Error int           `json:"error"`
		Msg   string        `json:"msg"`
		Todos []models.Todo `json:"todos"`
	}

	resp := new(Response)
	resp.Error = 0
	resp.Msg = "Todo sukses ditambahkan."

	conn := new(Db)
	err := conn.Connect()
	fatalIfError(err)
	defer conn.db.Close()

	rnd := render.New()
	name := req.PostFormValue("name")

	if name == "" {
		resp.Error = 1
		resp.Msg = "Judul tidak boleh kosong."
	}

	if resp.Error == 0 {
		err = conn.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("todo"))

			id, _ := b.NextSequence()
			todo := &models.Todo{Title: name, CategoryId: 1, Done: 0}
			todo.ID = int64(id)
			todo.DateCreated = time.Now()
			todo.DateUpdated = todo.DateCreated

			buf, err := json.Marshal(todo)
			if err != nil {
				return fmt.Errorf("json marshal: %s", err)
			}

			bId := new(bytes.Buffer)
			binary.Write(bId, binary.LittleEndian, id)
			return b.Put(bId.Bytes(), buf)
		})
	}
	conn.db.Close()

	todo, err := GetLastRow()
	checkError(err)

	todos := make([]models.Todo, 0)
	todos = append(todos, todo)
	resp.Todos = todos

	resp.Msg = string(renderToBytes("./views/todo_add.amber", todo))
	rnd.JSON(w, http.StatusOK, resp)
}

func DelTodoHandler(w http.ResponseWriter, req *http.Request) {
	type Response struct {
		Error int           `json:"error"`
		Msg   string        `json:"msg"`
		Todos []models.Todo `json:"todos"`
	}
	resp := Response{}

	params := mux.Vars(req)
	conn := new(Db)
	err := conn.Connect()
	fatalIfError(err)
	defer conn.db.Close()

	err = conn.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("todo"))
		c := b.Cursor()

		id, _ := strconv.ParseInt(params["id"], 10, 64)
		bId, _ := intToByte(id)
		k, _ := c.Seek(bId)
		if k != nil {
			return b.Delete(k)
		} else {
			return fmt.Errorf("Data tidak ditemukan, hapus dibatalkan.")
		}
		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	conn.db.Close()

	todos, err := GetAllTodo()
	if err != nil {
		resp.Error = 1
		resp.Msg = "Error: " + err.Error()
	}
	resp.Todos = todos

	rnd := render.New()
	rnd.JSON(w, http.StatusOK, resp)
}

func DoneTodoHandler(w http.ResponseWriter, req *http.Request) {
	type Response struct {
		Error int    `json:"error"`
		Msg   string `json:"msg"`
	}

	conn := Db{}
	err := conn.Connect()
	fatalIfError(err)
	defer conn.db.Close()

	resp := new(Response)
	resp.Error = 0
	resp.Msg = "Todo sukses dirubah ke done."

	rnd := render.New()
	id := req.PostFormValue("id")
	status := req.PostFormValue("status")

	if id == "" || id == "0" {
		resp.Error = 1
		resp.Msg = "Id tidak boleh kosong."
	}

	if status == "" {
		status = "1"
	}

	_status, _ := strconv.Atoi(status)
	_id, _ := strconv.ParseInt(id, 10, 64)
	err = conn.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("todo"))
		c := b.Cursor()

		bId, err := intToByte(_id)
		if err != nil {
			return fmt.Errorf("Error saat mengubah int ke byte.")
		}

		k, v := c.Seek(bId)
		if k == nil {
			return fmt.Errorf("Data tidak ditemukan.")
		}
		todo := models.Todo{}
		err = json.Unmarshal(v, &todo)
		if err != nil {
			return fmt.Errorf("Error saat unmarshal data.")
		}

		todo.Done = _status
		buf, err := json.Marshal(todo)
		if err != nil {
			return fmt.Errorf("Error json marshal: %s", err)
		}
		return b.Put(k, buf)
	})

	if err != nil {
		resp.Error = 1
		resp.Msg = err.Error()
	}

	rnd.JSON(w, http.StatusOK, resp)
}

func EditTodoHandler(w http.ResponseWriter, req *http.Request) {
	type Response struct {
		Error int    `json:"error"`
		Msg   string `json:"msg"`
	}

	conn := Db{}
	err := conn.Connect()
	fatalIfError(err)
	defer conn.db.Close()

	params := mux.Vars(req)
	name := req.PostFormValue("title")

	_id, _ := strconv.ParseInt(params["id"], 10, 64)
	id, _ := intToByte(_id)

	err = conn.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("todo"))
		c := b.Cursor()

		k, v := c.Seek(id)
		todo := models.Todo{}
		err := json.Unmarshal(v, &todo)
		if err != nil {
			return fmt.Errorf("json unmarshal error: %s\n", err.Error())
		}

		todo.Title = name
		todo.DateUpdated = todo.DateCreated

		buf, err := json.Marshal(todo)
		if err != nil {
			return fmt.Errorf("json marshal: %s", err)
		}

		return b.Put(k, buf)
	})

	resp := Response{Error: 0}
	resp.Msg = "Todo sukses diperbarui."
	if err != nil {
		resp.Error = 1
		resp.Msg = "Data dapat mengupdate todo."
	}

	rnd := render.New()
	rnd.JSON(w, http.StatusOK, resp)
}

func IsTodoDone(id int) (bool, error) {
	conn := new(Db)
	err := conn.Connect()
	fatalIfError(err)
	defer conn.db.Close()

	result := false
	err = conn.db.View(func(tx *bolt.Tx) error {
		bId, _ := IntToByte32(id)
		b := tx.Bucket([]byte("todo"))
		v := b.Get(bId)

		if v == nil {
			log.Println("IsTodoDone: Data tidak ditemukan")
			return fmt.Errorf("Data tidak ditemukan")
		}

		todo := models.Todo{}
		json.Unmarshal(v, &todo)
		if todo.Done == 1 {
			result = true
		}
		return nil
	})

	if err != nil {
		return false, err
	}

	return result, nil
}

func GetLastRow() (models.Todo, error) {
	conn := new(Db)
	err := conn.Connect()
	fatalIfError(err)
	defer conn.db.Close()

	todo := models.Todo{}
	err = conn.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("todo"))
		c := b.Cursor()

		_, v := c.Last()
		err := json.Unmarshal(v, &todo)
		if err != nil {
			return fmt.Errorf("json unmarshal: %s", err)
		}
		return nil
	})

	if err != nil {
		return todo, err
	}
	return todo, nil
}

func GetAllTodo() ([]models.Todo, error) {
	conn := new(Db)
	err := conn.Connect()
	fatalIfError(err)
	defer conn.db.Close()

	todos := make([]models.Todo, 0)
	err = conn.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("todo"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			todo := models.Todo{}
			err := json.Unmarshal(v, &todo)
			if err != nil {
				log.Fatalln(err)
			}
			todos = append(todos, todo)
		}
		return nil
	})

	if err != nil {
		return todos, err
	}
	return todos, nil
}
