package main

import (
	"bytes"
	"log"
	"net/http"
	"os"
  "strconv"
	"io/ioutil"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/russross/blackfriday"
)

var (
	repeat int
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

	router.Run(":" + port)
}
