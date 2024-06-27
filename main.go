package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

var savemessages bool

type formattedMessage struct {
	Msg template.HTML
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var clients []*websocket.Conn

func main() {
	savemessages = true
	db, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		log.Fatal(err)
	}
	statement, _ := db.Prepare("CREATE TABLE IF NOT EXISTS messages (id INTEGER PRIMARY KEY, message TEXT)")
	statement.Exec()
	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)
		clients = append(clients, conn)
		defer func() {
			if err := conn.Close(); err != nil {
				log.Fatal(err)
			}
			for i, c := range clients {
				if c == conn {
					clients = append(clients[:i], clients[i+1:]...)
					break
				}
			}
		}()
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				break
			}
			fmt.Printf("%s: %s\n", conn.RemoteAddr(), string(msg))
			for _, client := range clients {
				if err = client.WriteMessage(msgType, msg); err != nil {
					log.Printf("error: %v", err)
					break
				}
			}
			db, err := sql.Open("sqlite3", "./test.db")
			if err != nil {
				log.Fatal(err)
			}
			defer db.Close()
			if savemessages == true {
				insertVar, _ := db.Prepare("INSERT INTO messages (message) VALUES (?)")
				insertVar.Exec(msg)
			}
		}
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		db, err := sql.Open("sqlite3", "./test.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		rows, _ := db.Query("SELECT message FROM messages")
		var msg string
		var formattedTempMessage []string
		for rows.Next() {
			rows.Scan(&msg)
			htmlmsgform := msg + "<br>"
			formattedTempMessage = append(formattedTempMessage, htmlmsgform)
		}
		FinalSatetMessage := template.HTML(strings.Join(formattedTempMessage, ""))
		p := formattedMessage{Msg: FinalSatetMessage}
		t, _ := template.ParseFiles("index.html")
		t.Execute(w, p)
	})
	println("Server on: 8080...")
	http.ListenAndServe(":8080", nil)
}
