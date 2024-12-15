package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

var oauthCofig *oauth2.Config

func main() {

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	initOAuthConfig()

	e.GET("/", home)
	e.GET("/auth", startAuth)
	e.GET("/callback", handleCallback)

	e.Logger.Fatal(e.Start(":8080"))
}

func initOAuthConfig() {
	credentials, err := os.ReadFile(".credentials/calendar_credentials.json")
	if err != nil {
		log.Fatalf("Failed to read credentials fle : %v", err)
	}

	config, err := google.ConfigFromJSON(credentials, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Failed to parse client secrets : %v", err)
	}

	oauthCofig = config
}

func home(c echo.Context) error {
	return c.HTML(http.StatusOK, `<a href="/auth">Login with Google Calendar</a>`)
}

func startAuth(c echo.Context) error {
	url := oauthCofig.AuthCodeURL("start-token", oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func handleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		return c.String(http.StatusBadRequest, "Code not found")
	}

	// get token
	token, err := oauthCofig.Exchange(context.Background(), code)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Token exchange failed: %v", err))
	}
	// create client
	client := oauthCofig.Client(context.Background(), token)

	// create calendar service
	srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to create calendar client: %v", err))
	}

	// Get events from calendar
	events, err := srv.Events.List("primary").
		TimeMin(time.Now().Format(time.RFC3339)).
		TimeMax(time.Now().AddDate(0, 1, 0).Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Do()
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve events: %v", err))
	}

	// No events found
	if len(events.Items) == 0 {
		return c.String(http.StatusOK, "No upcoming events found")
	}

	// Format and display events
	var output strings.Builder
	for _, item := range events.Items {
		date := item.Start.DateTime
		if date == "" {
			date = item.Start.Date
		}
		fmt.Fprintf(&output, "%v: %v\n", date, item.Summary)
	}

	return c.String(http.StatusOK, output.String())

}
