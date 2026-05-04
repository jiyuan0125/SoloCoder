package server

import "errors"

var (
	ErrorPropertyNotFound  = errors.New("房源不存在")
	ErrorPropertySold      = errors.New("已成交房源不能修改")
	ErrorDuplicateProperty = errors.New("同一房东不能发布相同小区和户型的房源")
	ErrorAlreadyFavorited  = errors.New("该房源已收藏")
	ErrorFavoriteNotFound  = errors.New("收藏不存在")
	ErrorInvalidInput      = errors.New("输入参数无效")
)
