package services

import (
	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
)

type Service struct {
	DB *bolt.DB
}

func NewService(db *bolt.DB) *Service {
	return &Service{db}
}

func (self *Service) Init(app *gin.Engine) {
	TodoService(app, self.DB)
}
