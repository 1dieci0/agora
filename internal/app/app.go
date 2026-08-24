package app

import (
	"database/sql"
	"net/http"

	"agora/internal/channels"
	"agora/internal/database"
	"agora/internal/messages"
	"agora/internal/realtime"
	"agora/internal/servers"
	"agora/internal/users"
)

type App struct {
	db       *sql.DB
	router   *http.ServeMux
	users    *users.Handler
	servers  *servers.Handler
	channels *channels.Handler
	messages *messages.Handler
	hub      *realtime.Hub
	realtime *realtime.Handler
}

func NewApp() (*App, error) {
	db, err := database.Open()
	if err != nil {
		return nil, err
	}

	if err := database.SetupDatabase(db); err != nil {
		db.Close()
		return nil, err
	}

	hub := realtime.NewHub()

	app := &App{
		db:       db,
		router:   http.NewServeMux(),
		users:    users.NewHandler(db),
		servers:  servers.NewHandler(db),
		channels: channels.NewHandler(db),
		messages: messages.NewHandler(db, hub),
		hub:      hub,
		realtime: realtime.NewHandler(db, hub),
	}

	app.RegisterRoutes()

	return app, nil
}

func (app *App) Close() error {
	return app.db.Close()
}

func (app *App) Router() http.Handler {
	return app.router
}

func (app *App) RegisterRoutes() {
	// Authentication
	app.router.HandleFunc(
		"/api/users",
		app.users.Register,
	)

	app.router.HandleFunc(
		"/api/login",
		app.users.Login,
	)

	app.router.HandleFunc(
		"/api/logout",
		app.users.Logout,
	)

	// Users
	app.router.HandleFunc(
		"/api/me",
		app.users.RequireAuth(app.users.GetMe),
	)

	app.router.HandleFunc(
		"GET /api/servers/{id}",
		app.users.RequireAuth(app.servers.Get),
	)

	app.router.HandleFunc(
		"GET /api/servers/{id}/members",
		app.users.RequireAuth(app.servers.GetMembers),
	)

	app.router.HandleFunc(
		"POST /api/servers",
		app.users.RequireAuth(app.servers.Create),
	)

	app.router.HandleFunc(
		"GET /api/servers",
		app.users.RequireAuth(app.servers.GetServers),
	)

	// app.router.HandleFunc(
	// 	"POST /api/servers/{id}/join",
	// 	app.users.RequireAuth(app.servers.Join),
	// )

	app.router.HandleFunc(
		"DELETE /api/servers/{id}/members/me",
		app.users.RequireAuth(app.servers.Leave),
	)

	app.router.HandleFunc(
		"POST /api/servers/{id}/invites",
		app.users.RequireAuth(app.servers.CreateInvite),
	)

	app.router.HandleFunc(
		"POST /api/invites/{code}/join",
		app.users.RequireAuth(app.servers.JoinInvite),
	)

	app.router.HandleFunc(
		"POST /api/servers/{serverID}/channels",
		app.users.RequireAuth(app.channels.Create),
	)

	app.router.HandleFunc(
		"GET /api/servers/{serverID}/channels",
		app.users.RequireAuth(app.channels.GetChannels),
	)

	app.router.HandleFunc(
		"POST /api/channels/{channelID}/messages",
		app.users.RequireAuth(app.messages.Create),
	)

	app.router.HandleFunc(
		"GET /api/channels/{channelID}/messages",
		app.users.RequireAuth(app.messages.GetMessages),
	)

	app.router.HandleFunc(
		"GET /ws/channels/{channelID}",
		app.users.RequireAuth(app.realtime.Connect),
	)
}
