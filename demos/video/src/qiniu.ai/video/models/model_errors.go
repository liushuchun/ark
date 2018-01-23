package models

import (
	"errors"

	"gopkg.in/mgo.v2"
)

var (
	ErrInvalidId         = errors.New("Invalid BSON object id!")
	ErrNotFound          = mgo.ErrNotFound
	ErrNotPersisted      = errors.New("Record has not persisted!")
	ErrPartOrder         = errors.New("The order of part is invalid!")
	ErrPartData          = errors.New("The data of part is invailid!")
	ErrWaitPartTimeout   = errors.New("Waiting for previous part timeout!")
	ErrInvalidEtag       = errors.New("Invalid etag!")
	ErrEtagDuplicated    = errors.New("CANNOT update existed etag!")
	ErrBillingDuplicated = errors.New("CANNOT upate existed billing!")
	ErrInvalidDuration   = errors.New("The duration of start and end is too large.")
	ErrTokenExpired      = errors.New("Access token has expired.")
)
