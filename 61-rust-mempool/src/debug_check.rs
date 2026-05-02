const CANARY_VALUE: u32 = 0xDEADBEEF;
const CANARY_COUNT: usize = 2;
const SINGLE_CANARY_SIZE: usize = std::mem::size_of::<u32>();
pub const CANARY_SIZE: usize = CANARY_COUNT * SINGLE_CANARY_SIZE;

#[derive(Debug, Clone)]
pub struct DebugAllocator {
    #[cfg_attr(not(feature = "debug"), allow(dead_code))]
    canary_bytes: [u8; CANARY_SIZE],
}

impl Default for DebugAllocator {
    fn default() -> Self {
        Self::new()
    }
}

impl DebugAllocator {
    pub fn new() -> Self {
        let mut canary_bytes = [0u8; CANARY_SIZE];
        let canary_u32 = CANARY_VALUE.to_ne_bytes();

        for i in 0..CANARY_COUNT {
            let start = i * SINGLE_CANARY_SIZE;
            let end = start + SINGLE_CANARY_SIZE;
            canary_bytes[start..end].copy_from_slice(&canary_u32);
        }

        DebugAllocator { canary_bytes }
    }

    pub fn actual_size(&self, requested_size: usize) -> usize {
        #[cfg(feature = "debug")]
        {
            requested_size + 2 * CANARY_SIZE
        }
        #[cfg(not(feature = "debug"))]
        {
            requested_size
        }
    }

    pub fn user_offset(&self) -> usize {
        #[cfg(feature = "debug")]
        {
            CANARY_SIZE
        }
        #[cfg(not(feature = "debug"))]
        {
            0
        }
    }

    pub fn write_canaries(&self, buffer: &mut [u8], user_size: usize) {
        #[cfg(feature = "debug")]
        {
            let total_needed = self.actual_size(user_size);
            if buffer.len() < total_needed {
                return;
            }

            buffer[..CANARY_SIZE].copy_from_slice(&self.canary_bytes);

            let suffix_offset = CANARY_SIZE + user_size;
            buffer[suffix_offset..suffix_offset + CANARY_SIZE].copy_from_slice(&self.canary_bytes);
        }
        #[cfg(not(feature = "debug"))]
        {
            let _ = buffer;
            let _ = user_size;
        }
    }

    pub fn verify_canaries(
        &self,
        buffer: &[u8],
        user_size: usize,
        base_offset: usize,
    ) -> Result<(), CanaryError> {
        #[cfg(feature = "debug")]
        {
            let total_needed = self.actual_size(user_size);
            if buffer.len() < total_needed {
                return Err(CanaryError::BufferTooSmall);
            }

            for i in 0..CANARY_SIZE {
                if buffer[i] != self.canary_bytes[i] {
                    return Err(CanaryError::PrefixCorrupted {
                        offset_from_start: base_offset + i,
                        byte_offset: i,
                    });
                }
            }

            let suffix_offset = CANARY_SIZE + user_size;
            for i in 0..CANARY_SIZE {
                let idx = suffix_offset + i;
                if buffer[idx] != self.canary_bytes[i] {
                    return Err(CanaryError::SuffixCorrupted {
                        offset_from_start: base_offset + idx,
                        byte_offset: i,
                        user_size,
                    });
                }
            }

            Ok(())
        }
        #[cfg(not(feature = "debug"))]
        {
            let _ = buffer;
            let _ = user_size;
            let _ = base_offset;
            Ok(())
        }
    }
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum CanaryError {
    BufferTooSmall,
    PrefixCorrupted {
        offset_from_start: usize,
        byte_offset: usize,
    },
    SuffixCorrupted {
        offset_from_start: usize,
        byte_offset: usize,
        user_size: usize,
    },
}

impl std::fmt::Display for CanaryError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            CanaryError::BufferTooSmall => write!(f, "Buffer too small for canary check"),
            CanaryError::PrefixCorrupted {
                offset_from_start,
                byte_offset,
            } => write!(
                f,
                "Prefix canary corrupted at byte offset {} (absolute address offset {})",
                byte_offset, offset_from_start
            ),
            CanaryError::SuffixCorrupted {
                offset_from_start,
                byte_offset,
                user_size,
            } => write!(
                f,
                "Suffix canary corrupted at byte offset {} (absolute address offset {}), user data size was {} bytes - overflow detected at +{} from user data end",
                byte_offset, offset_from_start, user_size, byte_offset
            ),
        }
    }
}

impl std::error::Error for CanaryError {}
