package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"io/ioutil"
	"encoding/json"
	"strings"

	"github.com/gorilla/websocket"

	"github.com/gin-gonic/gin"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

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
				addMessageToDB(message.Name2, message.Name1,
					strings.Replace(joke, "'", "''", -1))
			}
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

func chatHandler(c *gin.Context) {
	body := c.Request.Body
	bodyContent, err := ioutil.ReadAll(body)
	if err != nil {
		c.String(http.StatusInternalServerError,
			        fmt.Sprintf("Error opening json body: %q", err))
		return
	}
	log.Println(bodyContent)
	var participants ChatParticipants
	err = json.Unmarshal(bodyContent, &participants)
	if err == nil {
		readChatFromDB(participants.Name1, participants.Name2, c)
	} else {
	c.String(http.StatusInternalServerError,
		        fmt.Sprintf("Error opening json: %q", err))
	}
}

func webSocketHandler(c *gin.Context) {
	connection, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("CONNECTION OPEN")
	var first bool = true
	var name1, name2 string
	for {
    t, msg, err := connection.ReadMessage()
    if err != nil {
				log.Println("ERROR IN WEBSOCKET LOOP")
        break
    }
		if first {
			splits := strings.Split(string(msg), " ")
			name1 = splits[0]
			name2 = splits[1]
			first = false
		} else {
			if (strings.Contains(string(msg), "__IMAGE__")) {
				response := "What a fine picture!"
				url := string(msg)
				connection.WriteMessage(t, []byte(response))
				addMessageToDB(name1, name2, strings.Replace(url, "'", "''", -1))
				addMessageToDB(name2, name1, response)
			} else {
				echo := append([]byte("I see, you wrote: "), msg...)
				jokePreface := []byte("Remninds me of a joke - ")
	      connection.WriteMessage(t, echo)
				connection.WriteMessage(t, jokePreface)
				addMessageToDB(name1, name2, strings.Replace(string(msg), "'", "''", -1))
				addMessageToDB(name2, name1, strings.Replace(string(echo), "'", "''", -1))
				addMessageToDB(name2, name1, string(jokePreface))
			}
		}
  }
}
