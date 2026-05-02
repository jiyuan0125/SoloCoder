use mpmc_queue::{channel, Queue};
use std::collections::HashSet;
use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::Arc;
use std::thread;
use std::time::Duration;

fn test_queue_basic() {
    println!("=== Test 1: Basic push/pop ===");
    
    let q = Queue::new(4);
    
    assert_eq!(q.len(), 0);
    assert!(q.is_empty());
    assert_eq!(q.capacity(), 4);
    
    q.push(1).unwrap();
    q.push(2).unwrap();
    assert_eq!(q.len(), 2);
    
    assert_eq!(q.pop().unwrap(), 1);
    assert_eq!(q.pop().unwrap(), 2);
    assert_eq!(q.len(), 0);
    
    println!("PASSED");
}

fn test_try_push_try_pop() {
    println!("\n=== Test 2: try_push/try_pop ===");
    
    let q = Queue::new(2);
    
    assert!(q.try_push(1).is_ok());
    assert!(q.try_push(2).is_ok());
    assert!(q.try_push(3).is_err());
    
    assert_eq!(q.try_pop().unwrap(), 1);
    assert_eq!(q.try_pop().unwrap(), 2);
    assert!(q.try_pop().is_err());
    
    println!("PASSED");
}

fn test_multithreaded() {
    println!("\n=== Test 3: Multi-threaded producer/consumer ===");
    
    let q = Queue::new(10);
    
    let producer_q = q.clone();
    let producer = thread::spawn(move || {
        for i in 0..5 {
            producer_q.push(i).unwrap();
            println!("Producer pushed: {}", i);
            thread::sleep(Duration::from_millis(10));
        }
        producer_q.close();
        println!("Producer closed queue");
    });
    
    let consumer_q = q.clone();
    let consumer = thread::spawn(move || {
        loop {
            match consumer_q.pop() {
                Ok(v) => println!("Consumer popped: {}", v),
                Err(_) => {
                    println!("Consumer got closed");
                    break;
                }
            }
        }
    });
    
    producer.join().unwrap();
    consumer.join().unwrap();
    
    println!("PASSED");
}

fn test_channel() {
    println!("\n=== Test 4: Channel API ===");
    
    let (tx, rx) = channel::<i32>(10);
    
    let tx2 = tx.clone();
    let producer1 = thread::spawn(move || {
        for i in 0..3 {
            tx.push(i * 10).unwrap();
            thread::sleep(Duration::from_millis(5));
        }
    });
    
    let producer2 = thread::spawn(move || {
        for i in 0..3 {
            tx2.push(i * 10 + 1).unwrap();
            thread::sleep(Duration::from_millis(5));
        }
    });
    
    producer1.join().unwrap();
    producer2.join().unwrap();
    
    assert_eq!(rx.len(), 6);
    rx.close();
    
    let mut received = Vec::new();
    while let Ok(v) = rx.try_pop() {
        received.push(v);
    }
    
    println!("Received: {:?}", received);
    assert_eq!(received.len(), 6);
    
    println!("PASSED");
}

fn test_into_iter() {
    println!("\n=== Test 5: IntoIterator ===");
    
    let q = Queue::new(4);
    q.push(1).unwrap();
    q.push(2).unwrap();
    q.push(3).unwrap();
    
    let items: Vec<i32> = q.into_iter().collect();
    assert_eq!(items, vec![1, 2, 3]);
    
    println!("PASSED");
}

fn test_stress_4p4c() {
    println!("\n=== Test 6: STRESS - 4 producers, 4 consumers, 20000 messages ===");
    
    const NUM_PRODUCERS: usize = 4;
    const NUM_CONSUMERS: usize = 4;
    const MESSAGES_PER_PRODUCER: usize = 5000;
    const TOTAL_MESSAGES: usize = NUM_PRODUCERS * MESSAGES_PER_PRODUCER;
    
    let q = Queue::new(100);
    
    let counter = Arc::new(AtomicUsize::new(0));
    let received_set = Arc::new(std::sync::Mutex::new(HashSet::new()));
    
    let mut producers = Vec::with_capacity(NUM_PRODUCERS);
    for p in 0..NUM_PRODUCERS {
        let q = q.clone();
        producers.push(thread::spawn(move || {
            for m in 0..MESSAGES_PER_PRODUCER {
                let value = p * MESSAGES_PER_PRODUCER + m;
                q.push(value).unwrap();
            }
            println!("Producer {} finished", p);
        }));
    }
    
    let mut consumers = Vec::with_capacity(NUM_CONSUMERS);
    for c_id in 0..NUM_CONSUMERS {
        let q = q.clone();
        let counter = counter.clone();
        let received_set = received_set.clone();
        consumers.push(thread::spawn(move || {
            loop {
                match q.pop() {
                    Ok(v) => {
                        counter.fetch_add(1, Ordering::SeqCst);
                        let mut set = received_set.lock().unwrap();
                        if !set.insert(v) {
                            panic!("Consumer {}: Duplicate value received: {}", c_id, v);
                        }
                    }
                    Err(_) => {
                        println!("Consumer {} exiting", c_id);
                        break;
                    }
                }
            }
        }));
    }
    
    for p in producers {
        p.join().unwrap();
    }
    
    println!("All producers finished, closing queue...");
    q.close();
    
    for c in consumers {
        c.join().unwrap();
    }
    
    let received_count = counter.load(Ordering::SeqCst);
    let set_size = received_set.lock().unwrap().len();
    
    println!("TOTAL_MESSAGES:  {}", TOTAL_MESSAGES);
    println!("Received count:  {}", received_count);
    println!("Unique in set:   {}", set_size);
    
    assert_eq!(received_count, TOTAL_MESSAGES, "Lost {} messages!", TOTAL_MESSAGES - received_count);
    assert_eq!(set_size, TOTAL_MESSAGES, "Set size mismatch!");
    
    println!("PASSED - No data loss!");
}

fn main() {
    println!("Running MPMC Queue tests...\n");
    
    test_queue_basic();
    test_try_push_try_pop();
    test_multithreaded();
    test_channel();
    test_into_iter();
    
    test_stress_4p4c();
    
    println!("\n=== All tests passed! ===");
}
