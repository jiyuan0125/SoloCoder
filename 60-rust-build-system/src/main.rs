mod config;
mod dag;
mod cache;
mod hasher;
mod executor;

use config::BuildConfig;
use dag::BuildDAG;
use cache::{BuildCache, TargetCacheEntry};
use hasher::FileHasher;
use executor::{ParallelExecutor, TargetResult, TargetStatus};

use std::collections::HashMap;
use std::env;
use std::process;
use std::time::Duration;

fn print_results(results: &[TargetResult], total_duration: Duration) {
    println!("{:-<30}{:-<15}{:-<10}", "", "", "");
    println!("{:<30} | {:<12} | {:<8}", "name", "status", "duration_ms");
    println!("{:-<30}{:-<15}{:-<10}", "", "", "");
    
    for result in results {
        println!(
            "{:<30} | {:<12} | {:<8}",
            result.name,
            result.status,
            result.duration_ms
        );
    }
    
    println!("{:-<30}{:-<15}{:-<10}", "", "", "");
    println!("Total time: {} ms", total_duration.as_millis());
}

fn determine_needs_rebuild(
    dag: &BuildDAG,
    cache: &BuildCache,
) -> Result<HashMap<String, bool>, String> {
    let mut needs_rebuild = HashMap::new();
    let mut input_hashes_map = HashMap::new();

    for node in dag.nodes.values() {
        let input_hashes = FileHasher::compute_target_inputs_hash(&node.config.inputs)?;
        input_hashes_map.insert(node.name.clone(), input_hashes);
    }

    for node_name in dag.topological_order().iter().rev() {
        let node = dag.nodes.get(node_name).unwrap();
        let input_hashes = input_hashes_map.get(node_name).unwrap();
        
        let is_cached = cache.is_target_cached(node_name, input_hashes);
        
        let dep_changed = node.dependencies.iter().any(|dep| {
            needs_rebuild.get(dep).copied().unwrap_or(false)
        });

        let rebuild = !is_cached || dep_changed;
        needs_rebuild.insert(node_name.clone(), rebuild);
    }

    Ok(needs_rebuild)
}

fn update_cache(
    cache: &mut BuildCache,
    dag: &BuildDAG,
    results: &[TargetResult],
) -> Result<(), String> {
    for result in results {
        if result.status == TargetStatus::Built || result.status == TargetStatus::Cached {
            let node = dag.nodes.get(&result.name).unwrap();
            
            let input_hashes = FileHasher::compute_target_inputs_hash(&node.config.inputs)?;
            let output_hashes = FileHasher::compute_files_hash(&node.config.outputs).unwrap_or_default();
            
            let entry = TargetCacheEntry {
                input_hashes,
                output_hashes,
            };
            
            cache.set_target_entry(result.name.clone(), entry);
        }
    }
    
    Ok(())
}

fn run_build(config_path: &str, parallelism: usize) -> Result<i32, String> {
    let config = BuildConfig::from_file(config_path)?;
    config.validate()?;

    let dag = match BuildDAG::from_config(&config) {
        Ok(dag) => dag,
        Err(e) => {
            if e.contains("Cycle") {
                return Err(e);
            }
            return Err(e);
        }
    };

    let mut cache = BuildCache::load();
    
    let needs_rebuild = determine_needs_rebuild(&dag, &cache)?;

    let executor = ParallelExecutor::new(parallelism);

    let should_rebuild = move |name: &str| -> bool {
        needs_rebuild.get(name).copied().unwrap_or(true)
    };

    let (results, total_duration) = executor.execute(&dag, should_rebuild);

    update_cache(&mut cache, &dag, &results)?;
    cache.save()?;

    print_results(&results, total_duration);

    let has_failure = results.iter().any(|r| r.status == TargetStatus::Failed);
    Ok(if has_failure { 1 } else { 0 })
}

fn main() {
    let args: Vec<String> = env::args().collect();
    
    let mut config_path = "build.json".to_string();
    let mut parallelism: usize = 4;
    
    let mut i = 1;
    while i < args.len() {
        match args[i].as_str() {
            "-j" | "--jobs" => {
                if i + 1 < args.len() {
                    parallelism = args[i + 1].parse().unwrap_or(4);
                    i += 2;
                    continue;
                }
            }
            "-f" | "--file" => {
                if i + 1 < args.len() {
                    config_path = args[i + 1].clone();
                    i += 2;
                    continue;
                }
            }
            _ => {
                config_path = args[i].clone();
            }
        }
        i += 1;
    }

    match run_build(&config_path, parallelism) {
        Ok(exit_code) => {
            process::exit(exit_code);
        }
        Err(e) => {
            eprintln!("Error: {}", e);
            process::exit(1);
        }
    }
}
