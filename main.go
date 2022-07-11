package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type Checkins struct {
	id          int
	eventId     int
	beganAt     time.Time
	completedAt time.Time
	meritUserId string
}

type Events struct {
	id                      int
	gratedMeritTemplateId   string
	name                    string
	qualifyingMeritTemplate string
}

type Merit []struct {
	id         int
	templateId *int
	userId     *int
}

func addCheckin(context *gin.Context) {
	eventId := context.Param("eventId")
	userId := context.Param("userId")

	// Call Merit API to get merits for template id, user id
	requestUrl := "http://localhost:3000/templates/" + eventId + "/merits?userId=" + userId
	response, err := http.Get(requestUrl)

	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	var merits Merit
	err = json.Unmarshal(body, &merits)
	if err != nil {
		log.Fatal(err)
	}

	// If user has the merit for this event, add a new checkin
	// ASSUMPTION: Only one merit returned for each event id
	if len(merits) > 0 {

		// Connect to sqlite db
		var DB *sql.DB

		var sqliteDbName = filepath.FromSlash("C:/sqlite_dbs/beiq.db")

		db, err := sql.Open("sqlite3", sqliteDbName)
		if err != nil {
			log.Fatal(err)
		}

		DB = db

		transaction, err := DB.Begin()
		if err != nil {
			log.Fatal(err)
		}

		sqlStmt := `INSERT INTO checkins (id, event_id, began_at, completed_at, merit_user_id) 
			VALUES ((select max(id) from checkins) + 1, ` + eventId + `, date('now'), NULL, ` + userId + `)`

		sqlPrepped, err := transaction.Prepare(sqlStmt)

		if err != nil {
			log.Fatal(err)
		}

		_, err = sqlPrepped.Exec()

		if err != nil {
			log.Fatal(err)
			context.JSON(http.StatusBadRequest, gin.H{"error": "err"})
		} else {
			context.JSON(http.StatusCreated, gin.H{"message": "Success"})
		}

		transaction.Commit()
	}
}

func main() {
	// default port: 8080 unless a
	// PORT environment variable is defined
	router := gin.Default()

	router.POST("events/:eventId/user/:userId", addCheckin)
	router.Run()
}
