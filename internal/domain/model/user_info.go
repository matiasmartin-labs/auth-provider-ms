package model

import "github.com/matiasmartin-labs/auth-provider-ms/pkg"

type UserInfo struct {
	Email     string
	FirstName string
	LastName  string
	Picture   string
}

func (u *UserInfo) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

func (u *UserInfo) IsEmailAllowed() bool {
	securityCfg := pkg.App.Config.GetSecurityConfig().GetLoginConfig()
	allowedEmails := securityCfg.GetAllowedEmails()
	for _, allowedEmail := range allowedEmails {
		if allowedEmail == u.Email {
			return true
		}
	}
	return false
}
