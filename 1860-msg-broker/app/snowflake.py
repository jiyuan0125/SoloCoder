import time
import threading

class Snowflake:
    def __init__(self, worker_id: int = 1):
        self.worker_id = worker_id & 0x1F
        self.sequence = 0
        self.last_timestamp = -1
        self.lock = threading.Lock()
        
        self.epoch = 1609459200000
        
        self.timestamp_bits = 41
        self.worker_id_bits = 5
        self.sequence_bits = 12
        
        self.max_sequence = (1 << self.sequence_bits) - 1
        self.worker_id_shift = self.sequence_bits
        self.timestamp_shift = self.worker_id_shift + self.worker_id_bits

    def _current_timestamp(self) -> int:
        return int(time.time() * 1000)

    def _wait_next_millisecond(self, last_timestamp: int) -> int:
        timestamp = self._current_timestamp()
        while timestamp <= last_timestamp:
            timestamp = self._current_timestamp()
        return timestamp

    def generate(self) -> int:
        with self.lock:
            timestamp = self._current_timestamp()
            
            if timestamp < self.last_timestamp:
                raise Exception("Clock moved backwards. Refusing to generate id")
            
            if timestamp == self.last_timestamp:
                self.sequence = (self.sequence + 1) & self.max_sequence
                if self.sequence == 0:
                    timestamp = self._wait_next_millisecond(self.last_timestamp)
            else:
                self.sequence = 0
            
            self.last_timestamp = timestamp
            
            id_value = ((timestamp - self.epoch) << self.timestamp_shift) | \
                       (self.worker_id << self.worker_id_shift) | \
                       self.sequence
            
            return id_value


id_generator = Snowflake(worker_id=1)


def generate_id() -> int:
    return id_generator.generate()
