package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
  "strconv"
	"io/ioutil"
	//"time"
	"encoding/json"
  "database/sql"

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
				addMessageToDB(message.Name1, message.Name2, message.Message)
			}
			//c.String(http.StatusOK, joke)
		}
	}
}

func getJoke(quoteBody []byte) (string, error) {
	var joke Joke
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
	db.Exec("CREATE TABLE IF NOT EXISTS chats (name1 text not null, name2 text not null, message text not null)")
	db.Exec("INSERT INTO users (name) VALUES ('Champ')")
	db.Exec("INSERT INTO users (name) VALUES ('Friend1')")
	db.Exec("INSERT INTO users (name) VALUES ('Friend2')")
	db.Exec("INSERT INTO users (name) VALUES ('Friend3')")
	db.Exec("INSERT INTO users (name) VALUES ('Friend4')")
	db.Exec("INSERT INTO users (name) VALUES ('Friend5')")
}

func addMessageToDB(name1, name2, message string) {
	var values string =
		"VALUES (''" + name1 + "'', ''" + name2 + "'', ''" + message + "')"
	result, err := db.Exec("INSERT INTO chats (name1,name2,message) " + values)
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

func readChat(c *gin.Context) {
	readChatFromDB("Champ", "Friend3", c)
}

func readChatFromDB(name1, name2 string, c *gin.Context) {
	rows, err := db.Query(
		"SELECT * FROM chats")
		//"SELECT * FROM chats WHERE name1 = " + name1 + " AND name2 = " + name2)
	if err != nil {
	    c.String(http.StatusInternalServerError,
	        fmt.Sprintf("Error reading ticks: %q", err))
	    return
	}
	defer rows.Close()
	for rows.Next() {
			var message string
			if err := rows.Scan(&message); err != nil {
				c.String(http.StatusInternalServerError,
					fmt.Sprintf("Error scanning ticks: %q", err))
					return
			}
			c.String(http.StatusOK, fmt.Sprintf("Read from DB: %s\n", message))
	}

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

	router.GET("/chat", readChat)

	router.Run(":" + port)
}
