use std::fs::{File, OpenOptions, create_dir_all, read_dir};
use std::io::{self, Seek, SeekFrom, Write, Read};
use std::path::{Path, PathBuf};

use parking_lot::RwLock;
use serde::{Deserialize, Serialize};

const MAX_SEGMENT_SIZE: u64 = 1024 * 1024 * 1024;
const INDEX_ENTRY_SIZE: usize = 16;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Message {
    pub offset: u64,
    pub payload: Vec<u8>,
    pub timestamp: u64,
}

#[derive(Debug, Clone)]
pub struct LogRecord {
    pub offset: u64,
    pub payload: Vec<u8>,
    pub timestamp: u64,
}

pub struct LogStore {
    base_dir: PathBuf,
    segments: RwLock<Vec<Segment>>,
    active_segment_idx: RwLock<usize>,
    next_offset: RwLock<u64>,
}

#[derive(Debug)]
struct Segment {
    start_offset: u64,
    log_file: File,
    index_file: File,
    log_path: PathBuf,
    index_path: PathBuf,
    current_offset: u64,
    size: u64,
}

impl LogStore {
    pub fn new(base_dir: PathBuf) -> io::Result<Self> {
        create_dir_all(&base_dir)?;
        
        let segments = Self::load_segments(&base_dir)?;
        let (segments, active_segment_idx, next_offset) = if segments.is_empty() {
            let first_segment = Segment::create(&base_dir, 0)?;
            (vec![first_segment], 0, 0)
        } else {
            let idx = segments.len() - 1;
            let next = segments[idx].current_offset;
            (segments, idx, next)
        };
        
        Ok(LogStore {
            base_dir,
            segments: RwLock::new(segments),
            active_segment_idx: RwLock::new(active_segment_idx),
            next_offset: RwLock::new(next_offset),
        })
    }
    
    fn load_segments(base_dir: &Path) -> io::Result<Vec<Segment>> {
        let mut segments = Vec::new();
        
        if let Ok(entries) = read_dir(base_dir) {
            let mut segment_files: Vec<(u64, PathBuf)> = Vec::new();
            
            for entry in entries {
                let entry = entry?;
                let path = entry.path();
                if path.extension().map_or(false, |e| e == "log") {
                    if let Some(name) = path.file_stem() {
                        if let Ok(offset) = name.to_string_lossy().parse::<u64>() {
                            segment_files.push((offset, path));
                        }
                    }
                }
            }
            
            segment_files.sort_by_key(|(o, _)| *o);
            
            for (start_offset, log_path) in segment_files {
                let index_path = log_path.with_extension("index");
                let segment = Segment::open(&log_path, &index_path, start_offset)?;
                segments.push(segment);
            }
        }
        
        Ok(segments)
    }
    
    pub fn append(&self, payload: &[u8]) -> io::Result<u64> {
        let offset = {
            let mut next_offset = self.next_offset.write();
            let offset = *next_offset;
            *next_offset += 1;
            offset
        };
        
        let timestamp = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_millis() as u64;
        
        let record = LogRecord {
            offset,
            payload: payload.to_vec(),
            timestamp,
        };
        
        let encoded = self.encode_record(&record);
        
        loop {
            let should_roll = {
                let segments = self.segments.read();
                let active_idx = *self.active_segment_idx.read();
                segments[active_idx].size + encoded.len() as u64 > MAX_SEGMENT_SIZE
            };
            
            if should_roll {
                self.roll_segment(offset)?;
            } else {
                break;
            }
        }
        
        {
            let mut segments = self.segments.write();
            let active_idx = *self.active_segment_idx.read();
            let active = &mut segments[active_idx];
            let pos = active.log_file.seek(SeekFrom::End(0))?;
            
            active.log_file.write_all(&encoded)?;
            active.log_file.flush()?;
            
            let index_entry = self.encode_index_entry(offset, pos);
            active.index_file.seek(SeekFrom::End(0))?;
            active.index_file.write_all(&index_entry)?;
            active.index_file.flush()?;
            
            active.size += encoded.len() as u64;
            active.current_offset = offset + 1;
        }
        
        Ok(offset)
    }
    
    pub fn read(&self, start_offset: u64, limit: usize) -> io::Result<Vec<Message>> {
        let segments = self.segments.read();
        let mut messages = Vec::new();
        let mut current_offset = start_offset;
        
        let next_offset = *self.next_offset.read();
        if start_offset >= next_offset {
            return Ok(Vec::new());
        }
        
        let start_segment_idx = self.find_segment_idx(start_offset);
        if start_segment_idx >= segments.len() {
            return Ok(Vec::new());
        }
        
        for seg_idx in start_segment_idx..segments.len() {
            if messages.len() >= limit {
                break;
            }
            
            let segment = &segments[seg_idx];
            
            if seg_idx == start_segment_idx {
                let entries = self.read_from_segment(segment, current_offset, limit - messages.len())?;
                if let Some(last) = entries.last() {
                    current_offset = last.offset + 1;
                }
                messages.extend(entries);
            } else {
                let entries = self.read_from_segment(segment, segment.start_offset, limit - messages.len())?;
                if let Some(last) = entries.last() {
                    current_offset = last.offset + 1;
                }
                messages.extend(entries);
            }
        }
        
        Ok(messages)
    }
    
    fn find_segment_idx(&self, offset: u64) -> usize {
        let segments = self.segments.read();
        let mut idx = 0;
        for i in 0..segments.len() {
            if segments[i].start_offset <= offset {
                idx = i;
            } else {
                break;
            }
        }
        idx
    }
    
    fn read_from_segment(&self, segment: &Segment, start_offset: u64, limit: usize) -> io::Result<Vec<Message>> {
        let mut messages = Vec::new();
        let mut log_file = File::open(&segment.log_path)?;
        
        let index_count = segment.index_file.metadata()?.len() / INDEX_ENTRY_SIZE as u64;
        if index_count == 0 {
            return Ok(Vec::new());
        }
        
        let mut idx_file = File::open(&segment.index_path)?;
        let mut left = 0;
        let mut right = index_count as i64 - 1;
        let mut start_pos = None;
        
        while left <= right {
            let mid = (left + right) / 2;
            idx_file.seek(SeekFrom::Start(mid as u64 * INDEX_ENTRY_SIZE as u64))?;
            let mut buf = [0u8; INDEX_ENTRY_SIZE];
            idx_file.read_exact(&mut buf)?;
            
            let (entry_offset, _pos) = self.decode_index_entry(&buf);
            
            if entry_offset == start_offset {
                idx_file.seek(SeekFrom::Start(mid as u64 * INDEX_ENTRY_SIZE as u64))?;
                idx_file.read_exact(&mut buf)?;
                let (_, pos) = self.decode_index_entry(&buf);
                start_pos = Some(pos);
                break;
            } else if entry_offset < start_offset {
                left = mid + 1;
            } else {
                right = mid - 1;
            }
        }
        
        if start_pos.is_none() && right >= 0 {
            idx_file.seek(SeekFrom::Start(right as u64 * INDEX_ENTRY_SIZE as u64))?;
            let mut buf = [0u8; INDEX_ENTRY_SIZE];
            idx_file.read_exact(&mut buf)?;
            let (entry_offset, pos) = self.decode_index_entry(&buf);
            if entry_offset <= start_offset {
                start_pos = Some(pos);
            }
        }
        
        let start_pos = match start_pos {
            Some(pos) => pos,
            None => return Ok(Vec::new()),
        };
        
        log_file.seek(SeekFrom::Start(start_pos))?;
        
        let mut remaining = limit;
        while remaining > 0 {
            let mut size_buf = [0u8; 8];
            match log_file.read_exact(&mut size_buf) {
                Ok(_) => {}
                Err(e) if e.kind() == io::ErrorKind::UnexpectedEof => break,
                Err(e) => return Err(e),
            }
            
            let record_size = u64::from_le_bytes(size_buf) as usize;
            let mut record_buf = vec![0u8; record_size];
            log_file.read_exact(&mut record_buf)?;
            
            let record = self.decode_record(&record_buf)?;
            if record.offset >= start_offset {
                messages.push(Message {
                    offset: record.offset,
                    payload: record.payload,
                    timestamp: record.timestamp,
                });
                remaining -= 1;
            }
            
            if record.offset + 1 >= segment.current_offset {
                break;
            }
        }
        
        Ok(messages)
    }
    
    fn roll_segment(&self, start_offset: u64) -> io::Result<()> {
        let new_segment = Segment::create(&self.base_dir, start_offset)?;
        
        let mut segments = self.segments.write();
        let mut active_idx = self.active_segment_idx.write();
        
        segments.push(new_segment);
        *active_idx = segments.len() - 1;
        
        Ok(())
    }
    
    fn encode_record(&self, record: &LogRecord) -> Vec<u8> {
        let mut buf = Vec::new();
        
        buf.extend_from_slice(&record.offset.to_le_bytes());
        buf.extend_from_slice(&record.timestamp.to_le_bytes());
        buf.extend_from_slice(&(record.payload.len() as u32).to_le_bytes());
        buf.extend_from_slice(&record.payload);
        
        let record_size = buf.len() as u64;
        let mut result = Vec::new();
        result.extend_from_slice(&record_size.to_le_bytes());
        result.extend_from_slice(&buf);
        
        result
    }
    
    fn decode_record(&self, buf: &[u8]) -> io::Result<LogRecord> {
        if buf.len() < 20 {
            return Err(io::Error::new(io::ErrorKind::InvalidData, "Record too short"));
        }
        
        let mut offset_bytes = [0u8; 8];
        offset_bytes.copy_from_slice(&buf[0..8]);
        let offset = u64::from_le_bytes(offset_bytes);
        
        let mut timestamp_bytes = [0u8; 8];
        timestamp_bytes.copy_from_slice(&buf[8..16]);
        let timestamp = u64::from_le_bytes(timestamp_bytes);
        
        let mut len_bytes = [0u8; 4];
        len_bytes.copy_from_slice(&buf[16..20]);
        let payload_len = u32::from_le_bytes(len_bytes) as usize;
        
        if buf.len() < 20 + payload_len {
            return Err(io::Error::new(io::ErrorKind::InvalidData, "Payload too short"));
        }
        
        let payload = buf[20..20 + payload_len].to_vec();
        
        Ok(LogRecord {
            offset,
            payload,
            timestamp,
        })
    }
    
    fn encode_index_entry(&self, offset: u64, position: u64) -> [u8; INDEX_ENTRY_SIZE] {
        let mut buf = [0u8; INDEX_ENTRY_SIZE];
        buf[0..8].copy_from_slice(&offset.to_le_bytes());
        buf[8..16].copy_from_slice(&position.to_le_bytes());
        buf
    }
    
    fn decode_index_entry(&self, buf: &[u8]) -> (u64, u64) {
        let mut offset_bytes = [0u8; 8];
        offset_bytes.copy_from_slice(&buf[0..8]);
        let offset = u64::from_le_bytes(offset_bytes);
        
        let mut pos_bytes = [0u8; 8];
        pos_bytes.copy_from_slice(&buf[8..16]);
        let pos = u64::from_le_bytes(pos_bytes);
        
        (offset, pos)
    }
    
    pub fn get_last_offset(&self) -> u64 {
        *self.next_offset.read()
    }
}

impl Segment {
    fn create(base_dir: &Path, start_offset: u64) -> io::Result<Self> {
        let log_path = base_dir.join(format!("{:020}.log", start_offset));
        let index_path = base_dir.join(format!("{:020}.index", start_offset));
        
        let log_file = OpenOptions::new()
            .create(true)
            .read(true)
            .write(true)
            .open(&log_path)?;
        
        let index_file = OpenOptions::new()
            .create(true)
            .read(true)
            .write(true)
            .open(&index_path)?;
        
        Ok(Segment {
            start_offset,
            log_file,
            index_file,
            log_path,
            index_path,
            current_offset: start_offset,
            size: 0,
        })
    }
    
    fn open(log_path: &Path, index_path: &Path, start_offset: u64) -> io::Result<Self> {
        let log_file = OpenOptions::new()
            .create(true)
            .read(true)
            .write(true)
            .open(log_path)?;
        
        let index_file = OpenOptions::new()
            .create(true)
            .read(true)
            .write(true)
            .open(index_path)?;
        
        let log_size = log_file.metadata()?.len();
        
        let current_offset = if let Ok(mut idx_file) = File::open(index_path) {
            let index_len = idx_file.metadata()?.len();
            if index_len >= INDEX_ENTRY_SIZE as u64 {
                idx_file.seek(SeekFrom::Start(index_len - INDEX_ENTRY_SIZE as u64))?;
                let mut buf = [0u8; INDEX_ENTRY_SIZE];
                idx_file.read_exact(&mut buf)?;
                let offset = u64::from_le_bytes(buf[0..8].try_into().unwrap());
                offset + 1
            } else {
                start_offset
            }
        } else {
            start_offset
        };
        
        Ok(Segment {
            start_offset,
            log_file,
            index_file,
            log_path: log_path.to_path_buf(),
            index_path: index_path.to_path_buf(),
            current_offset,
            size: log_size,
        })
    }
}
