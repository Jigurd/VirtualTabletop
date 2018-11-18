package tabletop

/*
User contains various information about the user
*/
type User struct {
	Username    string      `json:"username"`
	Password    string      `json:"password"`
	Email       string      `json:"email"`
	Description string      `json:"description"`
	Options     UserOptions `json:"options"`
}

/*
UserOptions contains options for the user
*/
type UserOptions struct {
	VisibleInDirectory bool `json:"visibleindirectory"`
}
