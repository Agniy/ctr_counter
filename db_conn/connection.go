/*
   Пакет для создания и управления подключением к базе данных
*/
package db_conn

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

const (
	BASE_TYPE = "postgres"
	BASE_USER = "agniy"
	BASE_PASS = "12345678"
	BASE_NAME = "fandeco"
	BASE_HOST = "localhost"
	BASE_PORT = "5432"
)

var (
	Conn *connect = nil
)

//метод вызывается при импорте пакета
func init() {
	//создаем наш объект логирования
	newConnection()
}

func newConnection() {
	if Conn == nil {
		Conn = new(connect)
		Conn.initConnection()
	}
}

type connect struct {
	Db *sql.DB
}

//метод инициализации соединения
func (conn *connect) initConnection() {
	connect_string := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable", BASE_USER, BASE_PASS, BASE_NAME, BASE_HOST, BASE_PORT)
	db, err := sql.Open(BASE_TYPE, connect_string)
	if err != nil {
		fmt.Println("Conn error", err)
		log.Fatal(err)
	} else {
		db.SetMaxIdleConns(30)
		conn.Db = db
	}
}

//метод закрытия соединения
func (conn *connect) Close() {
	conn.Db.Close()
}
