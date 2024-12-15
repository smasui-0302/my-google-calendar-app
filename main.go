package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
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

type Event struct {
	Date    string
	Summary string
}

func main() {

	e := echo.New()

	e.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseGlob("templates/*html")),
	}

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	initOAuthConfig()

	e.GET("/", home)
	e.GET("/auth", startAuth)
	e.GET("/callback", handleCallback)
	e.GET("/events", events)

	e.Logger.Fatal(e.Start(":8080"))
}

func initOAuthConfig() {
	// initOAuthConfig initializes OAuth2 configuration for Google Calendar API.
	//
	// Operation:
	// 1. Reads credentials from .credentials/calendar_credentials.json
	// 2. Generates OAuth2 configuration from credentials
	// 3. Stores configuration in global variable oauthConfig
	//
	// Parameters: none
	// Returns: none

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
	// home handles the root endpoint ("/") request.
	//
	// Operation:
	// 1. Displays a simple HTML page with Google Calendar login link
	//
	// Parameters:
	// - c echo.Context: The Echo context containing HTTP request/response data
	//
	// Returns:
	// - error: Returns nil if template renders successfully
	return c.Render(http.StatusOK, "home.html", nil)
}

func startAuth(c echo.Context) error {
	// startAuth handles the OAuth2 authentication initiation.
	//
	// Operation:
	// 1. Generates OAuth2 authorization URL with offline access
	// 2. Redirects user to Google's consent page
	//
	// Parameters:
	// - c echo.Context: The Echo context containing HTTP request/response data
	//
	// Returns:
	// - error: Returns nil on success, error on failure

	url := oauthCofig.AuthCodeURL("start-token", oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func handleCallback(c echo.Context) error {
	// handleCallback processes the OAuth2 callback after user authorization.
	//
	// Operation:
	// 1. Extracts authorization code from callback URL
	// 2. Exchanges code for OAuth2 token
	// 3. Saves token to cookie
	// 4. Redirects to events page
	//
	// Parameters:
	// - c echo.Context: The Echo context for handling HTTP request/response
	//
	// Returns:
	// - error: Returns nil on successful redirect
	//         Returns error with status code on failures:
	//         - StatusOK with redirect if no code
	//         - StatusInternalServerError if token exchange fails
	//         - StatusTemporaryRedirect to /events on success

	// extract OAuth2 code
	code := c.QueryParam("code")
	if code == "" {
		return c.Redirect(http.StatusOK, "/")
	}

	// get token
	token, err := oauthCofig.Exchange(context.Background(), code)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Token exchange failed: %v", err))
	}

	// Save token to cookie
	tokenCookie := &http.Cookie{
		Name:     "calendar_token",
		Value:    token.AccessToken,
		Expires:  token.Expiry,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(tokenCookie)

	return c.Redirect(http.StatusTemporaryRedirect, "/events")
}

func events(c echo.Context) error {
	// events handles the calendar events display endpoint ("/events").
	//
	// Operation:
	// 1. Gets OAuth2 token from cookie
	// 2. Creates Google Calendar API client and service
	// 3. Retrieves upcoming events for next month
	// 4. Formats and displays events in HTML template
	//
	// Parameters:
	// - c echo.Context: The Echo context containing HTTP request/response data
	//
	// Returns:
	// - error: Returns nil on successful render
	//         Returns redirect on failures:
	//         - StatusTemporaryRedirect to / if no token or API errors
	//         Returns render with empty events if no events found

	// Get token from cookie
	cookie, err := c.Cookie("calendar_token")
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	token := &oauth2.Token{
		AccessToken: cookie.Value,
		Expiry:      cookie.Expires,
	}
	// create client
	client := oauthCofig.Client(context.Background(), token)

	// create calendar service
	srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	// Get events from calendar
	events, err := srv.Events.List("primary").
		TimeMin(time.Now().Format(time.RFC3339)).
		TimeMax(time.Now().AddDate(0, 1, 0).Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Do()
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	// No events found
	if len(events.Items) == 0 {
		return c.Render(http.StatusOK, "events.html", map[string]interface{}{
			"Events": nil,
		})
	}

	// Format events data
	var eventsList []Event
	for _, item := range events.Items {
		date := item.Start.DateTime
		if date == "" {
			date = item.Start.Date
		}

		var formattedDate string
		if strings.Contains(date, "T") { // DateTime format
			parsedDate, _ := time.Parse(time.RFC3339, date)
			formattedDate = parsedDate.Format("2006/01/02 15:01")
		} else { //Date format
			parsedDate, _ := time.Parse("2006-01-02", date)
			formattedDate = parsedDate.Format("2006/01/02")
		}

		eventsList = append(eventsList, Event{
			Date:    formattedDate,
			Summary: item.Summary,
		})
	}

	return c.Render(http.StatusOK, "events.html", map[string]interface{}{
		"Events": eventsList,
	})
}
