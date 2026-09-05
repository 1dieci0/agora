package app

import (
	"database/sql"
	"net/http"

	"agora/internal/channels"
	"agora/internal/database"
	"agora/internal/dms"
	"agora/internal/messages"
	"agora/internal/notifications"
	"agora/internal/realtime"
	"agora/internal/servers"
	"agora/internal/unread"
	"agora/internal/users"
	"agora/internal/voice"
)

type App struct {
	db       *sql.DB
	router   *http.ServeMux
	users    *users.Handler
	servers  *servers.Handler
	channels *channels.Handler
	messages *messages.Handler

	textHub *realtime.Hub
	userHub *realtime.UserHub

	realtime     *realtime.Handler
	userRealtime *realtime.UserHandler

	voiceHub *voice.Hub
	voice    *voice.Handler

	unread        *unread.Handler
	notifications *notifications.Handler

	dms *dms.Handler
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

	//realtime
	hub := realtime.NewHub()
	userHub := realtime.NewUserHub()
	voiceHub := voice.NewHub()
	// Repositories
	serverRepo := servers.NewRepository(db)
	channelRepo := channels.NewRepository(db)
	messageRepo := messages.NewRepository(db)
	usersRepo := users.NewRepository(db)
	unreadRepo := unread.NewRepository(db)
	notificationsRepo := notifications.NewRepository(db)
	dmsRepo := dms.NewRepository(db)

	app := &App{
		db:     db,
		router: http.NewServeMux(),

		users: users.NewHandler(usersRepo, hub),

		servers: servers.NewHandler(serverRepo),

		channels: channels.NewHandler(
			channelRepo,
			serverRepo,
		),

		messages: messages.NewHandler(
			messageRepo,
			channelRepo,
			serverRepo,
			unreadRepo,
			hub,
			userHub,
			notificationsRepo,
		),

		textHub: hub,
		userHub: userHub,

		realtime:     realtime.NewHandler(channelRepo, serverRepo, hub),
		userRealtime: realtime.NewUserHandler(userHub),

		voiceHub: voiceHub,
		voice:    voice.NewHandler(voiceHub, channelRepo, serverRepo),

		unread: unread.NewHandler(
			unreadRepo,
			channelRepo,
			serverRepo,
		),

		notifications: notifications.NewHandler(
			notificationsRepo,
		),

		dms: dms.NewHandler(dmsRepo),
	}

	app.RegisterRoutes()

	return app, nil
}

func (app *App) Close() error {
	return app.db.Close()
}

func (app *App) Router() http.Handler {
	return cors(app.router)
}

func (app *App) RegisterRoutes() {
	// Authentication
	app.router.HandleFunc(
		"/api/register",
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
		"PATCH /api/servers/{id}/members/{userID}",
		app.users.RequireAuth(app.servers.UpdateMemberRole),
	)

	app.router.HandleFunc(
		"PUT /api/me/avatar",
		app.users.RequireAuth(app.users.UploadAvatar),
	)

	//servers

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
		"DELETE /api/servers/{id}",
		app.users.RequireAuth(app.servers.Delete),
	)

	//channel

	app.router.HandleFunc(
		"POST /api/servers/{serverID}/channels",
		app.users.RequireAuth(app.channels.Create),
	)

	app.router.HandleFunc(
		"GET /api/servers/{serverID}/channels",
		app.users.RequireAuth(app.channels.GetChannels),
	)

	app.router.HandleFunc(
		"PATCH /api/channels/{channelID}",
		app.users.RequireAuth(app.channels.Update),
	)

	app.router.HandleFunc(
		"DELETE /api/channels/{channelID}",
		app.users.RequireAuth(app.channels.Delete),
	)

	//messages

	app.router.HandleFunc(
		"POST /api/channels/{channelID}/messages",
		app.users.RequireAuth(app.messages.Create),
	)

	app.router.HandleFunc(
		"GET /api/channels/{channelID}/messages",
		app.users.RequireAuth(app.messages.GetMessages),
	)

	app.router.HandleFunc(
		"DELETE /api/messages/{messageID}",
		app.users.RequireAuth(app.messages.Delete),
	)

	app.router.HandleFunc(
		"PATCH /api/messages/{messageID}",
		app.users.RequireAuth(app.messages.Update),
	)

	//realtime
	app.router.HandleFunc(
		"GET /ws/servers/{serverID}",
		app.users.RequireAuth(app.realtime.Connect),
	)

	app.router.HandleFunc(
		"GET /ws/realtime",
		app.users.RequireAuth(app.userRealtime.Connect),
	)

	//uploads

	app.router.Handle(
		"/uploads/",
		http.StripPrefix(
			"/uploads/",
			http.FileServer(http.Dir("uploads")),
		),
	)

	//voice

	app.router.HandleFunc(
		"POST /api/voice/token",
		app.users.RequireAuth(app.voice.CreateToken),
	)

	//unread
	app.router.HandleFunc(
		"GET /api/unread",
		app.users.RequireAuth(app.unread.GetUnread),
	)

	app.router.HandleFunc(
		"PUT /api/channels/{channelID}/read",
		app.users.RequireAuth(app.unread.MarkChannelRead),
	)

	//notifs

	app.router.HandleFunc(
		"GET /api/notifications",
		app.users.RequireAuth(
			app.notifications.GetNotifications,
		),
	)

	app.router.HandleFunc(
		"POST /api/notifications/{id}/read",
		app.users.RequireAuth(
			app.notifications.MarkRead,
		),
	)

	app.router.HandleFunc(
		"POST /api/notifications/channel/{id}/read",
		app.users.RequireAuth(
			app.notifications.MarkChannelRead,
		),
	)

	// DMs

	app.router.HandleFunc(
		"POST /api/dms",
		app.users.RequireAuth(app.dms.CreateConversation),
	)

	app.router.HandleFunc(
		"GET /api/dms",
		app.users.RequireAuth(app.dms.GetConversations),
	)

	app.router.HandleFunc(
		"GET /api/dms/{id}/messages",
		app.users.RequireAuth(app.dms.GetMessages),
	)

	app.router.HandleFunc(
		"POST /api/dms/{id}/messages",
		app.users.RequireAuth(app.dms.CreateMessage),
	)
}
