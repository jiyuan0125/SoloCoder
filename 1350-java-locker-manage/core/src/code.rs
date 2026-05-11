use rand::rngs::StdRng;
use rand::{Rng, SeedableRng};
use std::collections::HashSet;

const CODE_LENGTH: usize = 6;
const MAX_ATTEMPTS: usize = 100;

pub struct PickupCodeGenerator {
    used_codes: HashSet<String>,
    rng: StdRng,
}

impl PickupCodeGenerator {
    pub fn new() -> Self {
        Self {
            used_codes: HashSet::new(),
            rng: StdRng::from_entropy(),
        }
    }

    pub fn generate(&mut self) -> Result<String, crate::LockerError> {
        for _ in 0..MAX_ATTEMPTS {
            let code = self.generate_single();
            if !self.used_codes.contains(&code) {
                self.used_codes.insert(code.clone());
                return Ok(code);
            }
        }
        Err(crate::LockerError::PickupCodeGenerationFailed)
    }

    fn generate_single(&mut self) -> String {
        let code: String = (0..CODE_LENGTH)
            .map(|_| (self.rng.gen_range(0..10) as u8 + b'0') as char)
            .collect();
        code
    }

    pub fn release(&mut self, code: &str) {
        self.used_codes.remove(code);
    }

    pub fn is_used(&self, code: &str) -> bool {
        self.used_codes.contains(code)
    }
}

impl Default for PickupCodeGenerator {
    fn default() -> Self {
        Self::new()
    }
}
