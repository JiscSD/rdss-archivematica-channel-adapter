package message

type Coverage struct {
	GeospatialCoverage    []GeospatialCoverage `json:"geospatialCoverage"`
	TemporalCoverageStart Timestamp            `json:"temporalCoverageStart"`
	TemporalCoverageEnd   Timestamp            `json:"temporalCoverageEnd"`
}

type GeospatialCoverage struct {
	GeolocationPoint          *GeolocationPoint  `json:"geolocationPoint,omitempty"`
	GeolocationPolygon        []GeolocationPoint `json:"geolocationPolygon,omitempty"`
	GeolocationPlace          string             `json:"geolocationPlace,omitempty"`
	CoordinateReferenceSystem string             `json:"coordinateReferenceSystem,omitempty"`
}

type GeolocationPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
