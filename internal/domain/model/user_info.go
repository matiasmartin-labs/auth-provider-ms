package model

type UserInfo struct {
	Email         string
	FirstName     string
	LastName      string
	Picture       string
	AllowedEmails []string
}

func (u *UserInfo) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

func (u *UserInfo) IsEmailAllowed() bool {
	for _, allowed := range u.AllowedEmails {
		if allowed == u.Email {
			return true
		}
	}
	return false
}
