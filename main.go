package main

import (
	"github.com/gin-gonic/gin"
	"github.com/bugsnag/bugsnag-go/gin"
	"github.com/bugsnag/bugsnag-go"
	"os"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"regexp"
	"time"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"fmt"
	"net/url"
	"encoding/json"
	"bytes"
	"strings"
)

var timezoneMap = map[string]string{
	"America/Chicago":     "Central",
	"America/Los_Angeles": "Western",
	"America/New_York":   "Eastern",
}

func main() {
	godotenv.Load()
	
	r := gin.Default()
	
	r.Use(bugsnaggin.AutoNotify(bugsnag.Configuration{
		// Your Bugsnag project API key
		APIKey: os.Getenv("BUGSNAG_API_KEY"),
		// The import paths for the Go packages containing your source files
		ProjectPackages: []string{"main"},
	}))
	
	r.POST("/events", handleEvent)
	
	r.Run()
}

func handleEvent(c *gin.Context) {
	var payload Payload
	
	rawPayload := c.PostForm("payload")
	rawPayload, _ = url.QueryUnescape(rawPayload)
	
	log.Println(rawPayload)
	
	err := json.Unmarshal([]byte(rawPayload), &payload)
	
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Invalid JSON",
			"reason": err.Error(),
		})
	}
	
	if payload.isAction() && payload.CallbackId == "convert_time" {
		response := createConvertTimezoneResponse(payload.Message.UserId, payload.Message.Text)
		c.JSON(http.StatusOK, response)
		actionResponse, _ := json.Marshal(response)
		http.Post(payload.ResponseUrl, "application/json", bytes.NewBuffer(actionResponse))
	}
}

func createConvertTimezoneResponse(UserId string, Message string) gin.H {
	timeRegex := regexp.MustCompile(`(?m)\b((1[0-2]|0?[1-9])(:([0-5][0-9]))*[[:space:]]*([AaPp][Mm]))`)
	matches := timeRegex.FindAllStringSubmatch(Message, len(Message))[0]
	
	hour := matches[2]
	minutes := matches[3]
	timeOfDay := matches[5]
	
	if minutes == "" {
		minutes = "00"
	}
	
	raidTimeString := fmt.Sprintf("%s:%s%s", hour, minutes, strings.ToUpper(timeOfDay))
	
	User := getUserFromDatabase(UserId)
	
	responseText := fmt.Sprintf("<@%s> said %s in the timezone *%s*, converted to other timezones, that would be:", User.Name, raidTimeString, timezoneMap[User.TimeZone])
	
	raidTimeCentral := timeIn(raidTimeString, User.TimeZone, "America/Chicago")
	raidTimeWestern := timeIn(raidTimeString, User.TimeZone, "America/Los_Angeles")
	raidTimeEaster := timeIn(raidTimeString, User.TimeZone, "America/New_York")
	
	return gin.H{
		"response_type": "in_channel",
		"text":          responseText,
		"attachments": []gin.H{
			{
				"title": "Central",
				"text":  raidTimeCentral.Format(time.Kitchen),
			},
			{
				"title": "Pacific",
				"text":  raidTimeWestern.Format(time.Kitchen),
			},
			{
				"title": "Eastern",
				"text":  raidTimeEaster.Format(time.Kitchen),
			},
		},
	}
}

func timeIn(Time string, TimeZone string, ToTimeZone string) time.Time {
	loc, err := time.LoadLocation(TimeZone)
	if err != nil {
		panic(err)
	}
	
	toLoc, err := time.LoadLocation(ToTimeZone)
	if err != nil {
		panic(err)
	}
	
	parsedTime, err := time.ParseInLocation(time.Kitchen, Time, loc)
	if err != nil {
		panic(err)
	}
	
	return parsedTime.In(toLoc)
}

func getUserFromDatabase(UserId string) User {
	user := env("DB_USERNAME", "root")
	password := env("DB_PASSWORD", "")
	dsn := fmt.Sprintf("%s:%s@/reinbot?charset=utf8&parseTime=True&loc=Local", user, password)
	db, err := gorm.Open("mysql", dsn)
	
	if err != nil {
		panic(err)
	}
	
	defer db.Close()
	
	var User User
	
	db.Find(&User, "slack_user_id = ?", UserId)
	
	return User
}

func env(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	
	return defaultValue
}
