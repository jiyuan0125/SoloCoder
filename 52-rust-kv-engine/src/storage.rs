use crc32fast::Hasher;
use std::fs::{File, OpenOptions};
use std::io::{self, Read, Seek, SeekFrom, Write};
use std::path::Path;

pub const MAX_VALUE_SIZE: u64 = 4 * 1024 * 1024;
pub const COMPACTION_THRESHOLD: u64 = 16 * 1024 * 1024;

pub const OP_PUT: u8 = 0;
pub const OP_DELETE: u8 = 1;

pub const HEADER_SIZE: u64 = 4 + 4 + 1 + 2;

#[derive(Debug, Clone, PartialEq)]
pub enum RecordType {
    Put,
    Delete,
}

#[derive(Debug, Clone)]
pub struct Record {
    pub op_type: RecordType,
    pub key: String,
    pub value: Option<Vec<u8>>,
}

impl Record {
    pub fn new_put(key: String, value: Vec<u8>) -> Self {
        Record {
            op_type: RecordType::Put,
            key,
            value: Some(value),
        }
    }

    pub fn new_delete(key: String) -> Self {
        Record {
            op_type: RecordType::Delete,
            key,
            value: None,
        }
    }

    pub fn serialized_size(&self) -> u64 {
        let key_len = self.key.len() as u64;
        let value_len = match &self.value {
            Some(v) => v.len() as u64,
            None => 0,
        };
        HEADER_SIZE + key_len + value_len
    }

    pub fn serialize(&self) -> io::Result<Vec<u8>> {
        let key_bytes = self.key.as_bytes();
        let value_bytes = self.value.as_deref().unwrap_or(&[]);

        let key_len = key_bytes.len() as u16;
        let total_len = (HEADER_SIZE - 8 + key_len as u64 + value_bytes.len() as u64) as u32;

        let op_code = match self.op_type {
            RecordType::Put => OP_PUT,
            RecordType::Delete => OP_DELETE,
        };

        let mut buffer = Vec::with_capacity((HEADER_SIZE + key_len as u64 + value_bytes.len() as u64) as usize);

        buffer.extend_from_slice(&[0u8; 4]);
        buffer.extend_from_slice(&total_len.to_le_bytes());
        buffer.push(op_code);
        buffer.extend_from_slice(&key_len.to_le_bytes());
        buffer.extend_from_slice(key_bytes);
        buffer.extend_from_slice(value_bytes);

        let crc = calculate_crc(&buffer[4..]);
        buffer[0..4].copy_from_slice(&crc.to_le_bytes());

        Ok(buffer)
    }
}

pub fn calculate_crc(data: &[u8]) -> u32 {
    let mut hasher = Hasher::new();
    hasher.update(data);
    hasher.finalize()
}

pub fn read_record(file: &mut File, offset: u64) -> io::Result<Option<(Record, u64)>> {
    file.seek(SeekFrom::Start(offset))?;

    let mut header = [0u8; 11];
    match file.read_exact(&mut header) {
        Ok(_) => {}
        Err(e) if e.kind() == io::ErrorKind::UnexpectedEof => return Ok(None),
        Err(e) => return Err(e),
    }

    let crc_stored = u32::from_le_bytes(header[0..4].try_into().unwrap());
    let total_len = u32::from_le_bytes(header[4..8].try_into().unwrap());
    let op_code = header[8];
    let key_len = u16::from_le_bytes(header[9..11].try_into().unwrap());

    let value_len = total_len as u64 - (1 + 2 + key_len as u64);

    let mut key_and_value = vec![0u8; (key_len as u64 + value_len) as usize];
    match file.read_exact(&mut key_and_value) {
        Ok(_) => {}
        Err(e) if e.kind() == io::ErrorKind::UnexpectedEof => return Ok(None),
        Err(e) => return Err(e),
    }

    let mut crc_data = Vec::with_capacity((4 + 1 + 2 + key_len as u64 + value_len) as usize);
    crc_data.extend_from_slice(&header[4..11]);
    crc_data.extend_from_slice(&key_and_value);

    let crc_calculated = calculate_crc(&crc_data);
    if crc_calculated != crc_stored {
        return Err(io::Error::new(
            io::ErrorKind::InvalidData,
            format!("CRC mismatch at offset {}", offset),
        ));
    }

    let key = String::from_utf8(key_and_value[..key_len as usize].to_vec())
        .map_err(|e| io::Error::new(io::ErrorKind::InvalidData, format!("Invalid UTF-8 in key: {}", e)))?;

    let value = if value_len > 0 {
        Some(key_and_value[key_len as usize..].to_vec())
    } else {
        None
    };

    let op_type = match op_code {
        OP_PUT => RecordType::Put,
        OP_DELETE => RecordType::Delete,
        _ => return Err(io::Error::new(io::ErrorKind::InvalidData, format!("Unknown op code: {}", op_code))),
    };

    let record = Record {
        op_type,
        key,
        value,
    };

    let record_size = 4 + 4 + 1 + 2 + key_len as u64 + value_len;

    Ok(Some((record, record_size)))
}

pub fn append_record(file: &mut File, record: &Record) -> io::Result<u64> {
    let offset = file.seek(SeekFrom::End(0))?;
    let data = record.serialize()?;
    file.write_all(&data)?;
    file.flush()?;
    Ok(offset)
}

pub fn file_size(file: &mut File) -> io::Result<u64> {
    file.seek(SeekFrom::End(0))
}

pub fn open_data_file(path: &Path) -> io::Result<File> {
    OpenOptions::new()
        .read(true)
        .write(true)
        .create(true)
        .open(path)
}

pub fn create_temp_file(dir: &Path) -> io::Result<(File, std::path::PathBuf)> {
    let temp_path = dir.join("data.tmp");
    let file = OpenOptions::new()
        .read(true)
        .write(true)
        .create(true)
        .truncate(true)
        .open(&temp_path)?;
    Ok((file, temp_path))
}
