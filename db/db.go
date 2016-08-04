package db

var (
	db     *sql.DB
)

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

func readChatFromDB(name1, name2 string, c *gin.Context) {
	var query string =
		"SELECT name1, message, time FROM chats WHERE (name1 = '" + name1 + "' AND name2 = '" + name2 + "') OR (name1 = '" + name2 + "' AND name2 = '" + name1 + "')"
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
