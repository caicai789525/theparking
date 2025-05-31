package models

import "errors"

var (
	ErrParkingSpotNotFound   = errors.New("车位不存在")
	ErrParkingNotBoundToUser = errors.New("车位未绑定给指定用户")
)
