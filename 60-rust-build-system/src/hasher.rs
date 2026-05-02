use sha2::{Sha256, Digest};
use std::collections::HashMap;
use std::path::Path;
use std::fs::File;
use std::io::Read;

pub struct FileHasher;

impl FileHasher {
    pub fn compute_file_hash<P: AsRef<Path>>(path: P) -> Result<String, String> {
        let path = path.as_ref();
        let mut file = File::open(path)
            .map_err(|e| format!("Failed to open file {}: {}", path.display(), e))?;
        
        let mut hasher = Sha256::new();
        let mut buffer = [0u8; 8192];
        
        loop {
            let bytes_read = file.read(&mut buffer)
                .map_err(|e| format!("Failed to read file {}: {}", path.display(), e))?;
            if bytes_read == 0 {
                break;
            }
            hasher.update(&buffer[..bytes_read]);
        }
        
        let result = hasher.finalize();
        Ok(format!("{:x}", result))
    }

    pub fn compute_files_hash<P: AsRef<Path>>(files: &[P]) -> Result<HashMap<String, String>, String> {
        let mut hashes = HashMap::new();
        for file in files {
            let path_str = file.as_ref().to_string_lossy().to_string();
            let hash = Self::compute_file_hash(file)?;
            hashes.insert(path_str, hash);
        }
        Ok(hashes)
    }

    pub fn compute_target_inputs_hash(inputs: &[String]) -> Result<HashMap<String, String>, String> {
        Self::compute_files_hash(inputs)
    }
}
