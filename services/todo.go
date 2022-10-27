package services

import (
	"net/http"
	"strconv"
	"time"

	"todo/models"
	"todo/repo"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
)

func TodoService(app *gin.Engine, db *bolt.DB) {
	todoRepo := repo.NewTodoRepo(db)
	api := app.Group("/todo")
	// Index
	api.GET("", func(c *gin.Context) {
		resp := NewResponse(0, "success", nil)
		rows, err := todoRepo.FindAll()
		resp.Items = nil
		if err != nil {
			resp.Error = 1
			resp.SetErrMsg(err.Error())
			c.JSON(http.StatusOK, resp)
			return
		}

		resp.SetData(rows)
		c.JSON(http.StatusOK, resp)
	})

	// Hapus group
	api.DELETE("/:id", func(c *gin.Context) {
		resp := NewResponse(0, "Todo sukses dihapus", nil)
		ids := c.Param("id")
		id, err := strconv.Atoi(ids)
		if err != nil {
			resp.SetErrMsg(err.Error())
			c.JSON(http.StatusBadRequest, resp)
			return
		}

		err = todoRepo.Delete(id)
		if err != nil {
			resp.SetErrMsg(err.Error())
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	// Create
	api.POST("", func(c *gin.Context) {
		var todo models.Todo
		resp := NewResponse(0, "Todo sukses disimpan", nil)
		if err := c.BindJSON(&todo); err != nil {
			resp.SetErrMsg(err.Error())
			c.JSON(http.StatusOK, resp)
			return
		}

		todo.DateCreated = time.Now()
		todo.DateUpdated = time.Now()
		res, err := todoRepo.Create(&todo)
		if err != nil {
			resp.SetErrMsg(err.Error())
			c.JSON(http.StatusOK, resp)
			return
		}

		resp.SetData(res)
		c.JSON(http.StatusOK, resp)
	})

	// Update
	api.PATCH("/:id", func(c *gin.Context) {
		resp := NewResponse(0, "Todo sukses diupdate", nil)
		ids := c.Param("id")
		id, err := strconv.Atoi(ids)
		if err != nil {
			resp.SetErrMsg(err.Error())
			c.JSON(http.StatusOK, resp)
			return
		}

		todo := new(models.Todo)
		if err := c.BindJSON(todo); err != nil {
			resp.SetErrMsg(err.Error())
			c.JSON(http.StatusOK, resp)
			return
		}

		_, err = todoRepo.FindOne(id)
		if err != nil {
			resp.SetErrMsg(err.Error())
			c.JSON(http.StatusOK, resp)
			return
		}

		res, err := todoRepo.Update(id, todo)
		if err != nil {
			resp.SetErrMsg(err.Error())
			c.JSON(http.StatusOK, resp)
			return
		}

		resp.SetData(res)
		c.JSON(http.StatusOK, resp)
	})

	// View
	api.GET("/:id", func(c *gin.Context) {
		resp := NewResponse(0, "success", nil)
		ids := c.Param("id")
		id, err := strconv.Atoi(ids)
		if err != nil {
			resp.SetErrMsg(err.Error())
			c.JSON(http.StatusOK, resp)
			return
		}
		todo, err := todoRepo.FindOne(id)
		if err != nil {
			resp.SetErrMsg(err.Error())
			c.JSON(http.StatusNotFound, resp)
			return
		}

		resp.SetData(todo)
		c.JSON(http.StatusOK, resp)
	})
}
