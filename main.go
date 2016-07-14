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

	"github.com/gin-gonic/gin"
	"github.com/russross/blackfriday"
	"github.com/lib/pq"
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
			c.String(http.StatusOK, joke)
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

func dbFunc(c *gin.Context) {
    if _, err := db.Exec("CREATE TABLE IF NOT EXISTS ticks (tick timestamp)"); err != nil {
        c.String(http.StatusInternalServerError,
            fmt.Sprintf("Error creating database table: %q", err))
        return
    }

    if _, err := db.Exec("INSERT INTO ticks VALUES (now())"); err != nil {
        c.String(http.StatusInternalServerError,
            fmt.Sprintf("Error incrementing tick: %q", err))
        return
    }

    rows, err := db.Query("SELECT tick FROM ticks")
    if err != nil {
        c.String(http.StatusInternalServerError,
            fmt.Sprintf("Error reading ticks: %q", err))
        return
    }

    defer rows.Close()
    for rows.Next() {
        var tick time.Time
        if err := rows.Scan(&tick); err != nil {
          c.String(http.StatusInternalServerError,
            fmt.Sprintf("Error scanning ticks: %q", err))
            return
        }
        c.String(http.StatusOK, fmt.Sprintf("Read from DB: %s\n", tick.String()))
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

	router.GET("/quote", quoteHandler)

  router.GET("/db", dbFunc)

	router.Run(":" + port)
}
