package api

import "github.com/zorder-index/pkg/zorder"

type Encode2DRequest struct {
	X uint32 `json:"x"`
	Y uint32 `json:"y"`
}

type Encode2DResponse struct {
	Code       zorder.Code `json:"code"`
	Method     string      `json:"method"`
	Coordinate zorder.Point2D `json:"coordinate"`
}

type Decode2DRequest struct {
	Code zorder.Code `json:"code"`
}

type Decode2DResponse struct {
	Coordinate zorder.Point2D `json:"coordinate"`
	Code       zorder.Code `json:"code"`
}

type Encode3DRequest struct {
	X uint32 `json:"x"`
	Y uint32 `json:"y"`
	Z uint32 `json:"z"`
}

type Encode3DResponse struct {
	Code       zorder.Code `json:"code"`
	Coordinate zorder.Point3D `json:"coordinate"`
}

type Decode3DRequest struct {
	Code zorder.Code `json:"code"`
}

type Decode3DResponse struct {
	Coordinate zorder.Point3D `json:"coordinate"`
	Code       zorder.Code `json:"code"`
}

type QueryRangesRequest struct {
	MinX uint32 `json:"min_x"`
	MinY uint32 `json:"min_y"`
	MaxX uint32 `json:"max_x"`
	MaxY uint32 `json:"max_y"`
}

type QueryRangesResponse struct {
	Rect   zorder.Rect2D `json:"rect"`
	Ranges []zorder.Range `json:"ranges"`
	Count  int          `json:"count"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
