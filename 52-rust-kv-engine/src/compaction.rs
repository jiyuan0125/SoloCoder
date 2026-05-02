use crate::index::Index;
use crate::storage::{append_record, create_temp_file, open_data_file, read_record, Record, RecordType, COMPACTION_THRESHOLD};
use std::collections::HashMap;
use std::fs::{self, File};
use std::io::{self, Seek, SeekFrom};
use std::path::Path;

pub fn should_compact(file: &mut File) -> io::Result<bool> {
    let size = file.seek(io::SeekFrom::End(0))?;
    Ok(size >= COMPACTION_THRESHOLD)
}

pub fn perform_compaction(
    data_path: &Path,
    index: &mut Index,
) -> io::Result<()> {
    let dir = data_path.parent().unwrap_or_else(|| Path::new("."));

    let mut old_file = open_data_file(data_path)?;
    let (mut temp_file, temp_path) = create_temp_file(dir)?;

    let mut latest_entries: HashMap<String, (Record, u64)> = HashMap::new();
    let mut offset = 0u64;

    loop {
        match read_record(&mut old_file, offset) {
            Ok(Some((record, record_size))) => {
                latest_entries.insert(record.key.clone(), (record, record_size));
                offset += record_size;
            }
            Ok(None) => break,
            Err(e) => {
                if e.kind() == io::ErrorKind::InvalidData {
                    break;
                }
                return Err(e);
            }
        }
    }

    let mut new_index = Index::new();
    let mut sorted_entries: Vec<_> = latest_entries.into_iter().collect();
    sorted_entries.sort_by(|a, b| a.0.cmp(&b.0));

    for (_key, (record, _old_size)) in sorted_entries {
        if record.op_type == RecordType::Put {
            let new_offset = append_record(&mut temp_file, &record)?;
            let record_size = record.serialized_size();
            
            new_index.update(&record, new_offset, record_size);
        }
    }

    drop(temp_file);
    drop(old_file);

    fs::remove_file(data_path)?;
    fs::rename(&temp_path, data_path)?;

    *index = new_index;

    Ok(())
}

pub fn compact_if_needed(
    data_path: &Path,
    file: &mut File,
    index: &mut Index,
) -> io::Result<bool> {
    if should_compact(file)? {
        perform_compaction(data_path, index)?;
        Ok(true)
    } else {
        Ok(false)
    }
}
