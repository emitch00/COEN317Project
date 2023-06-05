package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	db1ConnectionString = "host=127.0.0.1 port=1433 user=postgres password=password dbname=db1 sslmode=disable"
	db2ConnectionString = "host=127.0.0.1 port=1433 user=postgres password=password dbname=db2 sslmode=disable"
)

func main() {

	db1, err := sqlx.Connect("postgres", db1ConnectionString)
	if err != nil {
		log.Fatalln(err)
	}
	defer db1.Close()

	db2, err := sqlx.Connect("postgres", db2ConnectionString)
	if err != nil {
		log.Fatalln(err)
	}
	defer db2.Close()

	coordinator := &Coordinator{DB: db1}
	participant1 := &Participant{DB: db1}
	participant2 := &Participant{DB: db2}
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	register(router, coordinator, []*Participant{participant1, participant2})

	router.Run(":8080")
}
