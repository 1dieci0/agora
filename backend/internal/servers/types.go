package servers

type Server struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	OwnerID int    `json:"owner_id"`
}

type CreateServerRequest struct {
	Name string `json:"name"`
}

type ServerResponse struct {
	Message string `json:"message"`
	Server  Server `json:"server"`
}

type ServersResponse struct {
	Servers []Server `json:"servers"`
}

type Member struct {
	ID        int     `json:"id"`
	Username  string  `json:"username"`
	AvatarURL *string `json:"avatar_url"`
	Role      Role    `json:"role"`
}

type MembersResponse struct {
	Members []Member `json:"members"`
}

type CreateInviteResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}
type Channel struct {
	ID        int    `json:"id"`
	ServerID  int    `json:"server_id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type ChannelsResponse struct {
	Channels []Channel `json:"channels"`
}

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

type UpdateMemberRoleRequest struct {
	Role Role `json:"role"`
}
