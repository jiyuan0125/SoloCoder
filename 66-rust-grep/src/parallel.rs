use std::path::PathBuf;
use std::sync::{mpsc, Arc, Mutex};
use std::thread;

use crate::cli::Config;
use crate::mmap_reader::{FileReader, is_binary_file};
use crate::regex_matcher::Matcher;

#[derive(Debug, Clone)]
pub struct FileMatch {
    pub path: PathBuf,
    pub matches: Vec<LineMatch>,
    pub total_matches: usize,
}

#[derive(Debug, Clone)]
pub struct LineMatch {
    pub line_number: usize,
    pub content: String,
}

pub struct SearchResult {
    pub path: PathBuf,
    pub chunk_results: Vec<ChunkResult>,
    pub has_matches: bool,
}

pub struct ChunkResult {
    pub matches: Vec<(usize, String)>,
}

pub struct ThreadPool {
    workers: Vec<Worker>,
    sender: Option<mpsc::Sender<Job>>,
}

type Job = Box<dyn FnOnce() + Send + 'static>;

struct Worker {
    id: usize,
    thread: Option<thread::JoinHandle<()>>,
}

impl Worker {
    fn new(id: usize, receiver: Arc<Mutex<mpsc::Receiver<Job>>>) -> Worker {
        let thread = thread::spawn(move || loop {
            let job = receiver.lock().unwrap().recv();
            
            match job {
                Ok(job) => {
                    job();
                }
                Err(_) => break,
            }
        });
        
        Worker {
            id,
            thread: Some(thread),
        }
    }
}

impl ThreadPool {
    pub fn new(size: usize) -> ThreadPool {
        assert!(size > 0);
        
        let (sender, receiver) = mpsc::channel();
        let receiver = Arc::new(Mutex::new(receiver));
        
        let mut workers = Vec::with_capacity(size);
        
        for id in 0..size {
            workers.push(Worker::new(id, Arc::clone(&receiver)));
        }
        
        ThreadPool {
            workers,
            sender: Some(sender),
        }
    }

    pub fn execute<F>(&self, f: F)
    where
        F: FnOnce() + Send + 'static,
    {
        let job = Box::new(f);
        self.sender.as_ref().unwrap().send(job).unwrap();
    }
}

impl Drop for ThreadPool {
    fn drop(&mut self) {
        drop(self.sender.take());
        
        for worker in &mut self.workers {
            if let Some(thread) = worker.thread.take() {
                thread.join().unwrap();
            }
        }
    }
}

pub fn search_file_parallel(
    path: &PathBuf,
    matcher: Arc<Matcher>,
    config: Arc<Config>,
) -> Option<FileMatch> {
    if is_binary_file(path).unwrap_or(false) {
        return None;
    }
    
    let reader = match FileReader::new(path) {
        Ok(r) => r,
        Err(_) => return None,
    };
    
    let chunks = reader.get_chunks();
    let num_workers = num_cpus::get();
    let pool = ThreadPool::new(num_workers);
    
    let (result_sender, result_receiver) = mpsc::channel();
    
    for (chunk_idx, chunk) in chunks.into_iter().enumerate() {
        let matcher = Arc::clone(&matcher);
        let result_sender = result_sender.clone();
        let _path = path.clone();
        
        pool.execute(move || {
            let matches = matcher.find_matches_in_buffer(&chunk.data, chunk.start);
            let line_matches: Vec<(usize, String)> = matches
                .into_iter()
                .map(|m| (chunk.line_offset + m.line_number, m.content))
                .collect();
            
            result_sender.send((chunk_idx, line_matches)).unwrap();
        });
    }
    
    drop(result_sender);
    
    let mut all_matches: Vec<(usize, Vec<(usize, String)>)> = Vec::new();
    for (chunk_idx, matches) in result_receiver {
        all_matches.push((chunk_idx, matches));
    }
    
    all_matches.sort_by_key(|(idx, _)| *idx);
    
    let mut line_matches: Vec<LineMatch> = Vec::new();
    for (_, matches) in all_matches {
        for (line_num, content) in matches {
            line_matches.push(LineMatch {
                line_number: line_num,
                content,
            });
        }
    }
    
    let total_matches = line_matches.len();
    
    if total_matches > 0 || config.count {
        Some(FileMatch {
            path: path.clone(),
            matches: line_matches,
            total_matches,
        })
    } else {
        None
    }
}

pub fn collect_files(
    paths: &[PathBuf],
    exclude_patterns: &[String],
    gitignore_matchers: &[crate::gitignore::GitignoreMatcher],
) -> Vec<PathBuf> {
    let mut files = Vec::new();
    
    for path in paths {
        if path.is_file() {
            if should_include_file(path, exclude_patterns, gitignore_matchers) {
                files.push(path.clone());
            }
        } else if path.is_dir() {
            let walker = walkdir::WalkDir::new(path);
            for entry in walker {
                if let Ok(entry) = entry {
                    let entry_path = entry.path();
                    if entry_path.is_file() {
                        if should_include_file(entry_path, exclude_patterns, gitignore_matchers) {
                            files.push(entry_path.to_path_buf());
                        }
                    }
                }
            }
        }
    }
    
    files
}

fn should_include_file(
    path: &std::path::Path,
    exclude_patterns: &[String],
    gitignore_matchers: &[crate::gitignore::GitignoreMatcher],
) -> bool {
    if crate::gitignore::should_exclude_by_exclude_pattern(path, exclude_patterns) {
        return false;
    }
    
    for matcher in gitignore_matchers {
        if matcher.matches(path) {
            return false;
        }
    }
    
    true
}
