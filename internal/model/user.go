package model

import "time"

type User struct {
	ID          string     `db:"id" json:"id"`
	Issuer      string     `db:"issuer" json:"-"`
	Subject     string     `db:"subject" json:"-"`
	Email       *string    `db:"email" json:"email"`
	DisplayName string     `db:"display_name" json:"displayName"`
	Status      string     `db:"status" json:"status"`
	LastLoginAt *time.Time `db:"last_login_at" json:"lastLoginAt"`
}

type UserIdentity struct {
	Issuer      string
	Subject     string
	Email       string
	DisplayName string
}

type Organization struct {
	ID     string `db:"id" json:"id"`
	Name   string `db:"name" json:"name"`
	Slug   string `db:"slug" json:"slug"`
	Status string `db:"status" json:"status"`
}
