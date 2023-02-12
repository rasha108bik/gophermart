package errors

import "errors"

var ErrBadOrderNum = errors.New("wrong order number")
var ErrAlreadyLoaded = errors.New("already loaded order number")
var ErrLoadedByOtherUser = errors.New("already loaded by another user")
var ErrEmptyQueue = errors.New("queue is empty")
var ErrNothingChanged = errors.New("nothing changes in order")
var ErrNotFound = errors.New("not found")
var ErrLoginExists = errors.New("login already exists")
var ErrNoMoney = errors.New("not enough money on balance")
