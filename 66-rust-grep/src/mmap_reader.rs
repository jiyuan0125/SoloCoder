use memmap2::Mmap;
use std::fs::File;
use std::io::{self, Read};
use std::path::Path;

pub const LARGE_FILE_THRESHOLD: u64 = 1024 * 1024;
pub const CHUNK_SIZE: usize = 4 * 1024 * 1024;

pub enum FileReader {
    Mmap(MmapReader),
    InMemory(InMemoryReader),
}

pub struct MmapReader {
    mmap: Mmap,
    file_size: usize,
}

pub struct InMemoryReader {
    data: Vec<u8>,
}

pub struct ChunkInfo {
    pub start: usize,
    pub data: Vec<u8>,
    pub line_offset: usize,
}

impl FileReader {
    pub fn new(path: &Path) -> io::Result<Self> {
        let file = File::open(path)?;
        let metadata = file.metadata()?;
        let file_size = metadata.len();
        
        if file_size > LARGE_FILE_THRESHOLD {
            let mmap = unsafe { Mmap::map(&file)? };
            Ok(FileReader::Mmap(MmapReader {
                mmap,
                file_size: file_size as usize,
            }))
        } else {
            let mut data = Vec::with_capacity(file_size as usize);
            let mut file = File::open(path)?;
            file.read_to_end(&mut data)?;
            Ok(FileReader::InMemory(InMemoryReader { data }))
        }
    }

    pub fn get_chunks(&self) -> Vec<ChunkInfo> {
        match self {
            FileReader::Mmap(reader) => reader.get_chunks(),
            FileReader::InMemory(reader) => reader.get_chunks(),
        }
    }
}

impl MmapReader {
    pub fn get_chunks(&self) -> Vec<ChunkInfo> {
        let mut chunks = Vec::new();
        let mut current_start = 0;
        let mut line_offset = 0;
        
        while current_start < self.file_size {
            let mut end = std::cmp::min(current_start + CHUNK_SIZE, self.file_size);
            
            if end < self.file_size {
                while end < self.file_size && self.mmap[end] != b'\n' {
                    end += 1;
                }
                if end < self.file_size {
                    end += 1;
                }
            }
            
            let chunk_data = self.mmap[current_start..end].to_vec();
            
            chunks.push(ChunkInfo {
                start: current_start,
                data: chunk_data,
                line_offset,
            });
            
            line_offset += count_lines(&self.mmap[current_start..end]);
            current_start = end;
        }
        
        chunks
    }
}

impl InMemoryReader {
    pub fn get_chunks(&self) -> Vec<ChunkInfo> {
        vec![ChunkInfo {
            start: 0,
            data: self.data.clone(),
            line_offset: 0,
        }]
    }
}

fn count_lines(data: &[u8]) -> usize {
    data.iter().filter(|&&b| b == b'\n').count()
}

pub fn is_binary_file(path: &Path) -> io::Result<bool> {
    const CHECK_SIZE: usize = 8192;
    
    let mut file = File::open(path)?;
    let mut buffer = [0u8; CHECK_SIZE];
    let bytes_read = file.read(&mut buffer)?;
    
    Ok(buffer[..bytes_read].contains(&b'\0'))
}
