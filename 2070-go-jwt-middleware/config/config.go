package config

import (
	"time"
)

const (
	Issuer               = "jwt-middleware"
	Audience             = "jwt-middleware-audience"
	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 7 * 24 * time.Hour
	KeyTransitionPeriod  = 5 * time.Minute
	ServerPort           = 9404
	DBPath               = "./jwt.db"
)
