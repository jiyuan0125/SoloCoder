use std::fs::{File, OpenOptions, remove_file, rename};
use std::io::{Read, Write, Seek, SeekFrom};
use std::path::Path;
use crc::{Crc, CRC_32_ISCSI};

use crate::index::IndexManager;

pub const MAX_FILE_SIZE: u64 = 16 * 1024 * 1024;
pub const MAX_VALUE_SIZE: usize = 4 * 1024 * 1024;

const CRC32: Crc<u32> = Crc::<u32>::new(&CRC_32_ISCSI);

pub const OP_PUT: u8 = 0;
pub const OP_DELETE: u8 = 1;

const HEADER_SIZE: usize = 4 + 4 + 1 + 2;

#[derive(Debug, Clone)]
pub struct Record {
    pub op_type: u8,
    pub key: String,
    pub value: Option<Vec<u8>>,
}

impl Record {
    pub fn put(key: String, value: Vec<u8>) -> Self {
        Record {
            op_type: OP_PUT,
            key,
            value: Some(value),
        }
    }

    pub fn delete(key: String) -> Self {
        Record {
            op_type: OP_DELETE,
            key,
            value: None,
        }
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let key_bytes = self.key.as_bytes();
        let key_len = key_bytes.len() as u16;
        let value_len = self.value.as_ref().map(|v| v.len()).unwrap_or(0);
        
        let total_len = (1 + 2 + key_bytes.len() + value_len) as u32;
        
        let mut buffer = Vec::with_capacity((HEADER_SIZE + value_len) as usize);
        buffer.extend_from_slice(&[0u8; 4]);
        buffer.extend_from_slice(&total_len.to_le_bytes());
        buffer.push(self.op_type);
        buffer.extend_from_slice(&key_len.to_le_bytes());
        buffer.extend_from_slice(key_bytes);
        if let Some(ref v) = self.value {
            buffer.extend_from_slice(v);
        }
        
        let crc = CRC32.checksum(&buffer[4..]);
        buffer[0..4].copy_from_slice(&crc.to_le_bytes());
        
        buffer
    }

    pub fn from_bytes(data: &[u8]) -> Option<Self> {
        if data.len() < 7 {
            return None;
        }
        
        let stored_crc = u32::from_le_bytes([data[0], data[1], data[2], data[3]]);
        let computed_crc = CRC32.checksum(&data[4..]);
        
        if stored_crc != computed_crc {
            return None;
        }
        
        let total_len = u32::from_le_bytes([data[4], data[5], data[6], data[7]]) as usize;
        if data.len() != HEADER_SIZE + total_len {
            return None;
        }
        
        let op_type = data[8];
        let key_len = u16::from_le_bytes([data[9], data[10]]) as usize;
        
        if HEADER_SIZE + key_len > data.len() {
            return None;
        }
        
        let key_bytes = &data[HEADER_SIZE..HEADER_SIZE + key_len];
        let key = String::from_utf8_lossy(key_bytes).to_string();
        
        let value = if op_type == OP_PUT {
            let value_start = HEADER_SIZE + key_len;
            Some(data[value_start..].to_vec())
        } else {
            None
        };
        
        Some(Record {
            op_type,
            key,
            value,
        })
    }
}

pub struct StorageEngine {
    file: File,
    file_path: String,
    index: IndexManager,
}

impl StorageEngine {
    pub fn new(file_path: &str) -> std::io::Result<Self> {
        let path = Path::new(file_path);
        let file = OpenOptions::new()
            .read(true)
            .write(true)
            .create(true)
            .open(path)?;
        
        let mut engine = StorageEngine {
            file,
            file_path: file_path.to_string(),
            index: IndexManager::new(),
        };
        
        engine.build_index()?;
        
        Ok(engine)
    }

    fn build_index(&mut self) -> std::io::Result<()> {
        let mut offset: u64 = 0;
        let mut temp_records: Vec<(String, u64, usize, u8)> = Vec::new();
        
        loop {
            self.file.seek(SeekFrom::Start(offset))?;
            
            let mut header = [0u8; HEADER_SIZE];
            match self.file.read_exact(&mut header) {
                Ok(_) => {}
                Err(_) => break,
            }
            
            let total_len = u32::from_le_bytes([header[4], header[5], header[6], header[7]]) as usize;
            let record_size = HEADER_SIZE + total_len;
            
            let mut record_buf = vec![0u8; record_size];
            record_buf[0..HEADER_SIZE].copy_from_slice(&header);
            
            match self.file.read_exact(&mut record_buf[HEADER_SIZE..]) {
                Ok(_) => {}
                Err(_) => break,
            }
            
            let stored_crc = u32::from_le_bytes([record_buf[0], record_buf[1], record_buf[2], record_buf[3]]);
            let computed_crc = CRC32.checksum(&record_buf[4..]);
            
            if stored_crc != computed_crc {
                break;
            }
            
            let op_type = record_buf[8];
            let key_len = u16::from_le_bytes([record_buf[9], record_buf[10]]) as usize;
            
            if HEADER_SIZE + key_len > record_buf.len() {
                break;
            }
            
            let key_bytes = &record_buf[HEADER_SIZE..HEADER_SIZE + key_len];
            let key = String::from_utf8_lossy(key_bytes).to_string();
            
            temp_records.push((key, offset, record_size, op_type));
            
            offset += record_size as u64;
        }
        
        for (key, offset, length, op_type) in temp_records {
            if op_type == OP_DELETE {
                self.index.remove(&key);
            } else {
                self.index.put(key, offset, length);
            }
        }
        
        Ok(())
    }

    pub fn put(&mut self, key: String, value: Vec<u8>) -> std::io::Result<()> {
        if value.len() > MAX_VALUE_SIZE {
            return Err(std::io::Error::new(
                std::io::ErrorKind::InvalidInput,
                format!("Value size exceeds maximum limit of {} bytes", MAX_VALUE_SIZE),
            ));
        }
        
        let record = Record::put(key, value);
        self.append_record(record)
    }

    pub fn delete(&mut self, key: String) -> std::io::Result<()> {
        let record = Record::delete(key);
        self.append_record(record)
    }

    fn append_record(&mut self, record: Record) -> std::io::Result<()> {
        let key = record.key.clone();
        let op_type = record.op_type;
        let bytes = record.to_bytes();
        
        let offset = self.file.seek(SeekFrom::End(0))?;
        self.file.write_all(&bytes)?;
        self.file.flush()?;
        
        if op_type == OP_DELETE {
            self.index.remove(&key);
        } else {
            self.index.put(key, offset, bytes.len());
        }
        
        let file_size = self.file.metadata()?.len();
        if file_size > MAX_FILE_SIZE {
            self.compact()?;
        }
        
        Ok(())
    }

    pub fn get(&mut self, key: &str) -> std::io::Result<Option<Vec<u8>>> {
        let entry = self.index.get(key);
        
        if entry.is_none() {
            return Ok(None);
        }
        
        let (offset, length) = entry.unwrap();
        
        self.file.seek(SeekFrom::Start(offset))?;
        let mut buf = vec![0u8; length];
        self.file.read_exact(&mut buf)?;
        
        let record = Record::from_bytes(&buf)
            .ok_or_else(|| std::io::Error::new(std::io::ErrorKind::InvalidData, "Failed to parse record"))?;
        
        if record.op_type == OP_DELETE {
            Ok(None)
        } else {
            Ok(record.value)
        }
    }

    pub fn scan(&self, prefix: &str) -> Vec<String> {
        self.index.scan_prefix(prefix)
    }

    pub fn count(&self) -> usize {
        self.index.count()
    }

    pub fn compact(&mut self) -> std::io::Result<()> {
        let temp_path = format!("{}.tmp", self.file_path);
        
        let all_records = self.read_all_records_backward()?;
        let mut active_keys = std::collections::HashMap::new();
        
        for (record, offset, length) in all_records {
            if active_keys.contains_key(&record.key) {
                continue;
            }
            
            if record.op_type == OP_DELETE {
                active_keys.insert(record.key.clone(), None);
            } else {
                active_keys.insert(record.key.clone(), Some((record, offset, length)));
            }
        }
        
        let mut active_records: Vec<Record> = Vec::new();
        for (_, entry) in active_keys {
            if let Some((record, _, _)) = entry {
                active_records.push(record);
            }
        }
        
        active_records.sort_by(|a, b| a.key.cmp(&b.key));
        
        let mut temp_file = OpenOptions::new()
            .write(true)
            .create(true)
            .truncate(true)
            .open(&temp_path)?;
        
        let mut new_index = IndexManager::new();
        let mut current_offset: u64 = 0;
        
        for record in active_records {
            let key = record.key.clone();
            let bytes = record.to_bytes();
            
            temp_file.write_all(&bytes)?;
            new_index.put(key, current_offset, bytes.len());
            current_offset += bytes.len() as u64;
        }
        
        temp_file.flush()?;
        drop(temp_file);
        
        remove_file(&self.file_path)?;
        rename(&temp_path, &self.file_path)?;
        
        self.file = OpenOptions::new()
            .read(true)
            .write(true)
            .open(&self.file_path)?;
        self.index = new_index;
        
        Ok(())
    }

    fn read_all_records_backward(&mut self) -> std::io::Result<Vec<(Record, u64, usize)>> {
        let file_size = self.file.metadata()?.len();
        let mut records = Vec::new();
        let mut position = file_size;
        
        while position > 0 {
            let mut header = [0u8; HEADER_SIZE];
            if position < HEADER_SIZE as u64 {
                break;
            }
            
            position -= HEADER_SIZE as u64;
            self.file.seek(SeekFrom::Start(position))?;
            
            match self.file.read_exact(&mut header) {
                Ok(_) => {}
                Err(_) => break,
            }
            
            let total_len = u32::from_le_bytes([header[4], header[5], header[6], header[7]]) as usize;
            let record_size = HEADER_SIZE + total_len;
            
            if position < record_size as u64 {
                break;
            }
            
            let record_offset = position - total_len as u64;
            self.file.seek(SeekFrom::Start(record_offset))?;
            
            let mut record_buf = vec![0u8; record_size];
            match self.file.read_exact(&mut record_buf) {
                Ok(_) => {}
                Err(_) => {
                    position = record_offset;
                    continue;
                }
            }
            
            let stored_crc = u32::from_le_bytes([record_buf[0], record_buf[1], record_buf[2], record_buf[3]]);
            let computed_crc = CRC32.checksum(&record_buf[4..]);
            
            if stored_crc != computed_crc {
                position = record_offset;
                continue;
            }
            
            if let Some(record) = Record::from_bytes(&record_buf) {
                records.push((record, record_offset, record_size));
            }
            
            position = record_offset;
        }
        
        Ok(records)
    }
}
