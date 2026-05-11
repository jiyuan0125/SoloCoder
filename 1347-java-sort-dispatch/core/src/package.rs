use crate::models::Package;

#[derive(Debug, Clone)]
pub struct PackageSplitter {
    max_weight_kg: f64,
    max_dimensions_sum_m: f64,
}

impl PackageSplitter {
    pub fn new(max_weight_kg: f64, max_dimensions_sum_m: f64) -> Self {
        Self {
            max_weight_kg,
            max_dimensions_sum_m,
        }
    }

    pub fn split_into_packages(
        &self,
        order_id: &str,
        total_quantity: u32,
        weight_per_item: f64,
        length_per_item: f64,
        width_per_item: f64,
        height_per_item: f64,
    ) -> Vec<Package> {
        if total_quantity == 0 {
            return Vec::new();
        }

        let mut packages = Vec::new();
        let mut remaining_quantity = total_quantity;

        while remaining_quantity > 0 {
            let max_items_by_weight = (self.max_weight_kg / weight_per_item) as u32;
            let dimension_sum_per_item = length_per_item + width_per_item + height_per_item;
            let max_items_by_dimension = (self.max_dimensions_sum_m / dimension_sum_per_item) as u32;
            
            let max_items_per_package = std::cmp::min(max_items_by_weight, max_items_by_dimension);
            let max_items_per_package = std::cmp::max(1, max_items_per_package);

            let package_quantity = std::cmp::min(remaining_quantity, max_items_per_package);
            
            let pkg = Package::new(
                order_id,
                weight_per_item * package_quantity as f64,
                length_per_item,
                width_per_item,
                height_per_item,
                package_quantity,
            );

            packages.push(pkg);
            remaining_quantity -= package_quantity;
        }

        packages
    }
}

impl Default for PackageSplitter {
    fn default() -> Self {
        Self::new(50.0, 1.5)
    }
}
