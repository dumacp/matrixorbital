package gtt43a

import "errors"

var ErrorDevClosed = errors.New("dev is closed")
var ErrorDevNull = errors.New("dev is null")
var ErrorDevTimeout = errors.New("dev timeout")
var ErrorDevEmptyWrite = errors.New("write bytes in dev is empty")
var ErrorDevEmptyRead = errors.New("read bytes in dev is empty")
