use crate::models::GpsCoordinate;

pub fn haversine_distance(a: &GpsCoordinate, b: &GpsCoordinate) -> f64 {
    let earth_radius_meters = 6371000.0;

    let lat1 = a.latitude.to_radians();
    let lat2 = b.latitude.to_radians();
    let delta_lat = (b.latitude - a.latitude).to_radians();
    let delta_lon = (b.longitude - a.longitude).to_radians();

    let a = (delta_lat / 2.0).sin().powi(2)
        + lat1.cos() * lat2.cos() * (delta_lon / 2.0).sin().powi(2);
    let c = 2.0 * a.sqrt().atan2((1.0 - a).sqrt());

    earth_radius_meters * c
}

pub fn is_within_range(
    target: &GpsCoordinate,
    current: &GpsCoordinate,
    allowed_meters: f64,
) -> bool {
    let distance = haversine_distance(target, current);
    distance <= allowed_meters
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_haversine_distance_same_point() {
        let coord = GpsCoordinate {
            latitude: 39.9042,
            longitude: 116.4074,
        };
        let distance = haversine_distance(&coord, &coord);
        assert!(distance < 1.0);
    }

    #[test]
    fn test_haversine_distance_known() {
        let beijing = GpsCoordinate {
            latitude: 39.9042,
            longitude: 116.4074,
        };
        let shanghai = GpsCoordinate {
            latitude: 31.2304,
            longitude: 121.4737,
        };
        let distance = haversine_distance(&beijing, &shanghai);
        assert!(distance > 1000000.0 && distance < 1100000.0);
    }
}
