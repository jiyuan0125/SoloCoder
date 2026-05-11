package api

type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Mercator struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Tile struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
	Z int   `json:"z"`
}

type TileBounds struct {
	Tile    Tile    `json:"tile"`
	North   float64 `json:"north"`
	South   float64 `json:"south"`
	East    float64 `json:"east"`
	West    float64 `json:"west"`
}

type MapView struct {
	Center LatLng `json:"center"`
	Zoom   int    `json:"zoom"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Pixel struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type ViewBounds struct {
	NW LatLng `json:"nw"`
	NE LatLng `json:"ne"`
	SW LatLng `json:"sw"`
	SE LatLng `json:"se"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Warning string `json:"warning,omitempty"`
}

type LatLngToMercatorRequest struct {
	Coordinates []LatLng `json:"coordinates"`
}

type LatLngToMercatorResponse struct {
	Results []MercatorResult `json:"results"`
}

type MercatorResult struct {
	Mercator Mercator `json:"mercator"`
	Warning  string   `json:"warning,omitempty"`
}

type MercatorToLatLngRequest struct {
	Coordinates []Mercator `json:"coordinates"`
}

type MercatorToLatLngResponse struct {
	Results []LatLngResult `json:"results"`
}

type LatLngResult struct {
	LatLng  LatLng `json:"latlng"`
	Warning string `json:"warning,omitempty"`
}

type LatLngToTileRequest struct {
	Coordinates []LatLng `json:"coordinates"`
	Zoom        int      `json:"zoom"`
}

type LatLngToTileResponse struct {
	Results []TileResult `json:"results"`
}

type TileResult struct {
	Tile    Tile   `json:"tile"`
	Warning string `json:"warning,omitempty"`
}

type TileToBoundsRequest struct {
	Tiles []Tile `json:"tiles"`
}

type TileToBoundsResponse struct {
	Results []TileBounds `json:"results"`
}

type LatLngToPixelRequest struct {
	View        MapView   `json:"view"`
	Coordinates []LatLng  `json:"coordinates"`
}

type LatLngToPixelResponse struct {
	Results []PixelResult `json:"results"`
}

type PixelResult struct {
	Pixel   Pixel  `json:"pixel"`
	Warning string `json:"warning,omitempty"`
}

type PixelToLatLngRequest struct {
	View   MapView `json:"view"`
	Pixels []Pixel `json:"pixels"`
}

type PixelToLatLngResponse struct {
	Results []LatLngResult `json:"results"`
}

type ViewBoundsRequest struct {
	View MapView `json:"view"`
}

type ViewBoundsResponse struct {
	Bounds  ViewBounds `json:"bounds"`
	Warning string     `json:"warning,omitempty"`
}

type DistanceRequest struct {
	From LatLng `json:"from"`
	To   LatLng `json:"to"`
}

type DistanceResponse struct {
	DistanceMeters       float64 `json:"distance_meters"`
	DistanceKilometers   float64 `json:"distance_kilometers"`
	Warning              string  `json:"warning,omitempty"`
}

type AreaRequest struct {
	Polygon []LatLng `json:"polygon"`
}

type AreaResponse struct {
	AreaSquareMeters   float64 `json:"area_square_meters"`
	AreaSquareKilometers float64 `json:"area_square_kilometers"`
	Warning            string  `json:"warning,omitempty"`
}
