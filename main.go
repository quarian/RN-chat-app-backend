package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
  "strconv"
	"io/ioutil"
	"time"
	"encoding/json"
  "database/sql"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/russross/blackfriday"
	_"github.com/lib/pq"
)

var (
	repeat int
	db     *sql.DB
)

type Joke struct {
	Type string `json:"type"`
	Value struct {
		Id int `json:"id"`
		Joke string `json:"joke"`
		Categories []string `json:"categories"`
	} `json:"value"`
}

type ChatMessage struct {
		Name1 string `json:"name1"`
		Name2 string `json:"name2"`
		Message string `json:"message"`
}

type ChatParticipants struct {
		Name1 string `json:"name1"`
		Name2 string `json:"name2"`
}


func repeatHandler(c *gin.Context) {
	var buffer bytes.Buffer
   	for i := 0; i < repeat; i++ {
        	buffer.WriteString("Hello from Go!\n")
    	}
    	c.String(http.StatusOK, buffer.String())
}

func quoteHandler(c *gin.Context) {
	quoteBody, err := getNorrisQuoteBody("http://api.icndb.com/jokes/random")
	if err != nil {
		c.String(http.StatusOK, err.Error())
	} else {
		joke, err := getJoke(quoteBody)
		if err != nil {
			c.String(http.StatusOK, err.Error())
		} else {
			body := c.Request.Body
			bodyContent, _ := ioutil.ReadAll(body)
			var message ChatMessage
			err := json.Unmarshal(bodyContent, &message)
			if err != nil {
				c.String(http.StatusOK, err.Error())
			} else {
				c.String(http.StatusOK, joke)
				addMessageToDB(message.Name1, message.Name2,
					strings.Replace(message.Message, "'", "''", -1))
				addMessageToDB(message.Name2, message.Name1,
					strings.Replace(joke, "'", "''", -1))
			}
			//c.String(http.StatusOK, joke)
		}
	}
}

func getJoke(quoteBody []byte) (string, error) {
	var joke Joke
	log.Println(quoteBody)
	err := json.Unmarshal(quoteBody, &joke)
	log.Println(joke)
	if err != nil {
		return "", err
	}
	return joke.Value.Joke, err

}

func getNorrisQuoteBody(url string) ([]byte, error) {
	quote, err := queryQuoteAPI(url)
	return quote, err
}

func queryQuoteAPI(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	client :=&http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	queryBody, err := readResponse(resp)
	return queryBody, err
}

func readResponse(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, err
}

func instantianteDB() {
	db.Exec("DROP TABLE IF EXISTS users")
	db.Exec("DROP TABLE IF EXISTS chats")
	db.Exec("CREATE TABLE IF NOT EXISTS users (name text not null)")
	db.Exec("CREATE TABLE IF NOT EXISTS chats (name1 text not null, name2 text not null, message text not null, time timestamp)")
	db.Exec("INSERT INTO users (name) VALUES ('Champ')")
	db.Exec("INSERT INTO users (name) VALUES ('Friend1')")
	db.Exec("INSERT INTO users (name) VALUES ('Friend2')")
	db.Exec("INSERT INTO users (name) VALUES ('Friend3')")
	db.Exec("INSERT INTO users (name) VALUES ('Friend4')")
	db.Exec("INSERT INTO users (name) VALUES ('Friend5')")
}

func addMessageToDB(name1, name2, message string) {
	var values string =
		"VALUES ('" + name1 + "', '" + name2 + "', '" + message + "', now())"
	var command string = "INSERT INTO chats (name1,name2,message,time) " + values
	log.Println(command)
	result, err := db.Exec(command)
	if err != nil {
		log.Println("Error putting into the db" + err.Error())
	} else {
		rowsAffected, _ := result.RowsAffected()
		log.Println("Result " + string(rowsAffected))
	}
}

func readUsersFromDB(c *gin.Context) {
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
	    c.String(http.StatusInternalServerError,
	        fmt.Sprintf("Error reading ticks: %q", err))
	    return
	}

	var users [6]string
	var i int = 0
	defer rows.Close()
	for rows.Next() {
	    var name string
	    if err := rows.Scan(&name); err != nil {
	      c.String(http.StatusInternalServerError,
	        fmt.Sprintf("Error scanning ticks: %q", err))
	        return
	    }
			users[i] = name
			i++
	}
	jsonResponse, _ := json.Marshal(users)
	c.String(http.StatusOK, string(jsonResponse))
}

func chatHandler(c *gin.Context) {
	body := c.Request.Body
	bodyContent, err := ioutil.ReadAll(body)
	if err == nil {
		c.String(http.StatusInternalServerError,
			        fmt.Sprintf("Error opening json body: %s", err))
		return
	}
	log.Println(bodyContent)
	var participants ChatParticipants
	err = json.Unmarshal(bodyContent, &participants)
	if err != nil {
		readChatFromDB(participants.Name1, participants.Name2, c)
	} else {
	c.String(http.StatusInternalServerError,
		        fmt.Sprintf("Error opening json: %q", err))
	}
}

func readChatFromDB(name1, name2 string, c *gin.Context) {
	var query string =
		"SELECT name1, message, time FROM chats WHERE (name1 = '" + name1 + "' AND name2 = '" + name2 + "') OR (name1 = '" + name2 + "' AND name2 = '" + name1 + "')"
	//"SELECT * FROM chats")
	log.Println(query)
	rows, err := db.Query(query)
	if err != nil {
	    c.String(http.StatusInternalServerError,
	        fmt.Sprintf("Error reading ticks: %q", err))
	    return
	}
	messages := make([][3]string, 0)
	defer rows.Close()
	for rows.Next() {
			var name, message string
			var time time.Time
			if err := rows.Scan(&name, &message, &time); err != nil {
				c.String(http.StatusInternalServerError,
					fmt.Sprintf("Error scanning ticks: %q", err))
					return
			}
			messages = append(messages, [3]string{name, message, time.String()})
	}
	jsonResponse, _ := json.Marshal(messages)
	c.String(http.StatusOK, string(jsonResponse))
	log.Println(messages)
}

func main() {
  var err error
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	tStr := os.Getenv("REPEAT")
	repeat, err = strconv.Atoi(tStr)
	if err != nil {
		log.Printf("Error converting $REPEAT to an int: %q - Using default\n", err)
		repeat = 5
	}

	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}

	instantianteDB()

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) { c.HTML(http.StatusOK, "index.tmpl.html", nil)
	})
	router.GET("/mark", func(c *gin.Context) {
  	c.String(http.StatusOK, string(blackfriday.MarkdownBasic([]byte("**hi!**"))))
	})

  router.GET("/repeat", repeatHandler)

	router.POST("/quote", quoteHandler)

  router.GET("/db", readUsersFromDB)

	router.POST("/chat", chatHandler)

	router.Run(":" + port)
}
