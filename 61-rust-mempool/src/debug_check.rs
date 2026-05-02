const CANARY_VALUE: u32 = 0xDEADBEEF;
const CANARY_SIZE: usize = std::mem::size_of::<u32>();
const PADDING_SIZE: usize = 8;

#[derive(Debug, Clone)]
pub struct DebugAllocator {
    canary_prefix: [u8; CANARY_SIZE],
    canary_suffix: [u8; CANARY_SIZE],
}

impl Default for DebugAllocator {
    fn default() -> Self {
        Self::new()
    }
}

impl DebugAllocator {
    pub fn new() -> Self {
        let canary = CANARY_VALUE.to_ne_bytes();
        DebugAllocator {
            canary_prefix: canary,
            canary_suffix: canary,
        }
    }

    pub fn actual_size(&self, requested_size: usize) -> usize {
        #[cfg(feature = "debug")]
        {
            requested_size + PADDING_SIZE
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
            if buffer.len() < self.actual_size(user_size) {
                return;
            }

            for i in 0..CANARY_SIZE {
                if i < buffer.len() {
                    buffer[i] = self.canary_prefix[i];
                }
            }

            let suffix_offset = CANARY_SIZE + user_size;
            for i in 0..CANARY_SIZE {
                let idx = suffix_offset + i;
                if idx < buffer.len() {
                    buffer[idx] = self.canary_suffix[i];
                }
            }
        }
    }

    pub fn verify_canaries(
        &self,
        buffer: &[u8],
        user_size: usize,
        offset: usize,
    ) -> Result<(), CanaryError> {
        #[cfg(feature = "debug")]
        {
            if buffer.len() < self.actual_size(user_size) {
                return Err(CanaryError::BufferTooSmall);
            }

            for i in 0..CANARY_SIZE {
                if i < buffer.len() && buffer[i] != self.canary_prefix[i] {
                    return Err(CanaryError::PrefixCorrupted {
                        offset_from_start: offset + i,
                        byte_offset: i,
                    });
                }
            }

            let suffix_offset = CANARY_SIZE + user_size;
            for i in 0..CANARY_SIZE {
                let idx = suffix_offset + i;
                if idx < buffer.len() && buffer[idx] != self.canary_suffix[i] {
                    return Err(CanaryError::SuffixCorrupted {
                        offset_from_start: offset + idx,
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
            let _ = offset;
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
