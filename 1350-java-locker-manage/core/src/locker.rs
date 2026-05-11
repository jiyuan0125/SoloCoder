use crate::fee::*;
use crate::models::*;
use crate::PickupCodeGenerator;
use chrono::Utc;
use std::collections::HashMap;
use std::sync::{Arc, Mutex, RwLock};

pub type LockResult<T> = std::result::Result<T, crate::LockerError>;

pub struct LockerSystem {
    lockers: RwLock<HashMap<String, Locker>>,
    pickup_code_generator: Mutex<PickupCodeGenerator>,
    code_to_package: RwLock<HashMap<String, (String, String)>>,
}

impl LockerSystem {
    pub fn new() -> Self {
        Self {
            lockers: RwLock::new(HashMap::new()),
            pickup_code_generator: Mutex::new(PickupCodeGenerator::new()),
            code_to_package: RwLock::new(HashMap::new()),
        }
    }

    pub fn add_locker(&self, locker: Locker) {
        let mut lockers = self.lockers.write().unwrap();
        lockers.insert(locker.id.clone(), locker);
    }

    pub fn get_locker_info(&self, locker_id: &str) -> LockResult<LockerInfo> {
        let lockers = self.lockers.read().unwrap();
        let locker = lockers
            .get(locker_id)
            .ok_or_else(|| crate::LockerError::LockerNotFound(locker_id.to_string()))?;

        let total = locker.compartments.len();
        let mut available = 0;
        let mut available_small = 0;
        let mut available_medium = 0;
        let mut available_large = 0;

        for compartment in &locker.compartments {
            if compartment.current_package.is_none() {
                available += 1;
                match compartment.size {
                    CompartmentSize::Small => available_small += 1,
                    CompartmentSize::Medium => available_medium += 1,
                    CompartmentSize::Large => available_large += 1,
                }
            }
        }

        Ok(LockerInfo {
            id: locker.id.clone(),
            name: locker.name.clone(),
            total_compartments: total,
            available_compartments: available,
            available_small,
            available_medium,
            available_large,
        })
    }

    pub fn list_lockers(&self) -> Vec<LockerInfo> {
        let lockers = self.lockers.read().unwrap();
        lockers
            .keys()
            .filter_map(|id| self.get_locker_info(id).ok())
            .collect()
    }

    pub fn get_locker_details(&self, locker_id: &str) -> LockResult<LockerDetails> {
        let lockers = self.lockers.read().unwrap();
        let locker = lockers
            .get(locker_id)
            .ok_or_else(|| crate::LockerError::LockerNotFound(locker_id.to_string()))?;

        let compartments = locker
            .compartments
            .iter()
            .map(|c| CompartmentDetails {
                id: c.id.clone(),
                size: c.size,
                occupied: c.current_package.is_some(),
                package: c.current_package.as_ref().map(|p| PackageInfo {
                    id: p.id.clone(),
                    phone: p.phone.clone(),
                    courier_name: p.courier_name.clone(),
                    pickup_code: p.pickup_code.clone(),
                    deposited_at: p.deposited_at,
                }),
            })
            .collect();

        Ok(LockerDetails {
            id: locker.id.clone(),
            name: locker.name.clone(),
            compartments,
        })
    }

    pub fn deposit_package(
        &self,
        request: &DepositRequest,
    ) -> LockResult<DepositResponse> {
        let mut lockers = self.lockers.write().unwrap();
        let locker = lockers
            .get_mut(&request.locker_id)
            .ok_or_else(|| crate::LockerError::LockerNotFound(request.locker_id.clone()))?;

        let compartment_id = Self::find_best_compartment(locker, &request.package_size)?;

        let pickup_code = {
            let mut generator = self.pickup_code_generator.lock().unwrap();
            generator.generate()?
        };

        let compartment = locker
            .compartments
            .iter_mut()
            .find(|c| c.id == compartment_id)
            .unwrap();

        let package_id = uuid::Uuid::new_v4().to_string();
        let deposited_at = Utc::now();

        let package = Package {
            id: package_id.clone(),
            size: request.package_size,
            phone: request.phone.clone(),
            courier_name: request.courier_name.clone(),
            compartment_id: compartment_id.clone(),
            pickup_code: pickup_code.clone(),
            deposited_at,
            pickup_attempts: 0,
            locked_until: None,
        };

        compartment.current_package = Some(package);

        {
            let mut code_map = self.code_to_package.write().unwrap();
            code_map.insert(
                pickup_code.clone(),
                (request.locker_id.clone(), compartment_id.clone()),
            );
        }

        Ok(DepositResponse {
            locker_id: request.locker_id.clone(),
            compartment_id,
            pickup_code,
            deposited_at,
        })
    }

    fn find_best_compartment(
        locker: &Locker,
        package_size: &PackageSize,
    ) -> LockResult<String> {
        let mut best: Option<&Compartment> = None;

        for compartment in &locker.compartments {
            if compartment.current_package.is_some() {
                continue;
            }

            if !compartment.size.can_fit(package_size) {
                continue;
            }

            match best {
                None => {
                    best = Some(compartment);
                }
                Some(current_best) => {
                    if compartment.size.is_better_than(&current_best.size, package_size) {
                        best = Some(compartment);
                    }
                }
            }
        }

        match best {
            Some(compartment) => Ok(compartment.id.clone()),
            None => Err(crate::LockerError::NoAvailableCompartment(*package_size)),
        }
    }

    pub fn pickup_package(&self, request: &PickupRequest) -> LockResult<PickupResponse> {
        let location = {
            let code_map = self.code_to_package.read().unwrap();
            code_map
                .get(&request.pickup_code)
                .cloned()
                .ok_or(crate::LockerError::InvalidPickupCode)?
        };

        let (locker_id, compartment_id) = location;

        let mut lockers = self.lockers.write().unwrap();
        let locker = lockers
            .get_mut(&locker_id)
            .ok_or(crate::LockerError::PackageNotFound)?;

        let compartment = locker
            .compartments
            .iter_mut()
            .find(|c| c.id == compartment_id)
            .ok_or(crate::LockerError::PackageNotFound)?;

        let package = compartment
            .current_package
            .as_mut()
            .ok_or(crate::LockerError::PackageNotFound)?;

        if is_locked(&package.locked_until) {
            let remaining = package
                .locked_until
                .unwrap()
                .signed_duration_since(Utc::now());
            let remaining_minutes = (remaining.num_seconds() + 59) / 60;
            return Err(crate::LockerError::PickupLocked(remaining_minutes as u64));
        }

        if package.pickup_code != request.pickup_code {
            package.pickup_attempts += 1;
            if package.pickup_attempts >= max_attempts() {
                package.locked_until = Some(create_lock_time());
                package.pickup_attempts = 0;
            }
            return Err(crate::LockerError::InvalidPickupCode);
        }

        let picked_at = Utc::now();
        let (hours, fee) = FeeCalculator::calculate(&package.deposited_at, &picked_at);

        let response = PickupResponse {
            locker_id: locker_id.clone(),
            compartment_id: compartment_id.clone(),
            package_id: package.id.clone(),
            phone: package.phone.clone(),
            deposited_at: package.deposited_at,
            picked_at,
            storage_hours: hours,
            fee,
        };

        let pickup_code = package.pickup_code.clone();
        compartment.current_package = None;

        {
            let mut code_map = self.code_to_package.write().unwrap();
            code_map.remove(&pickup_code);
        }

        {
            let mut generator = self.pickup_code_generator.lock().unwrap();
            generator.release(&pickup_code);
        }

        Ok(response)
    }

    pub fn list_packages_by_phone(&self, phone: &str) -> Vec<PackageInfo> {
        let lockers = self.lockers.read().unwrap();
        let mut packages = Vec::new();

        for locker in lockers.values() {
            for compartment in &locker.compartments {
                if let Some(package) = &compartment.current_package {
                    if package.phone == phone {
                        packages.push(PackageInfo {
                            id: package.id.clone(),
                            phone: package.phone.clone(),
                            courier_name: package.courier_name.clone(),
                            pickup_code: package.pickup_code.clone(),
                            deposited_at: package.deposited_at,
                        });
                    }
                }
            }
        }

        packages
    }
}

impl Default for LockerSystem {
    fn default() -> Self {
        Self::new()
    }
}

pub fn create_sample_locker_system() -> Arc<LockerSystem> {
    let system = Arc::new(LockerSystem::new());

    let locker1 = Locker {
        id: "locker-001".to_string(),
        name: "A栋快递柜".to_string(),
        compartments: vec![
            Compartment {
                id: "comp-001-01".to_string(),
                size: CompartmentSize::Small,
                locker_id: "locker-001".to_string(),
                current_package: None,
            },
            Compartment {
                id: "comp-001-02".to_string(),
                size: CompartmentSize::Small,
                locker_id: "locker-001".to_string(),
                current_package: None,
            },
            Compartment {
                id: "comp-001-03".to_string(),
                size: CompartmentSize::Medium,
                locker_id: "locker-001".to_string(),
                current_package: None,
            },
            Compartment {
                id: "comp-001-04".to_string(),
                size: CompartmentSize::Medium,
                locker_id: "locker-001".to_string(),
                current_package: None,
            },
            Compartment {
                id: "comp-001-05".to_string(),
                size: CompartmentSize::Large,
                locker_id: "locker-001".to_string(),
                current_package: None,
            },
        ],
    };

    let locker2 = Locker {
        id: "locker-002".to_string(),
        name: "B栋快递柜".to_string(),
        compartments: vec![
            Compartment {
                id: "comp-002-01".to_string(),
                size: CompartmentSize::Small,
                locker_id: "locker-002".to_string(),
                current_package: None,
            },
            Compartment {
                id: "comp-002-02".to_string(),
                size: CompartmentSize::Medium,
                locker_id: "locker-002".to_string(),
                current_package: None,
            },
            Compartment {
                id: "comp-002-03".to_string(),
                size: CompartmentSize::Large,
                locker_id: "locker-002".to_string(),
                current_package: None,
            },
        ],
    };

    system.add_locker(locker1);
    system.add_locker(locker2);

    system
}
