use mpmc_queue::{MpmcQueue, QueueError};
use std::sync::Arc;
use std::thread;
use std::time::Duration;

fn main() {
    println!("=== Testing MPMC Queue ===\n");

    test_basic_push_pop();
    test_try_push_try_pop();
    test_blocking_push();
    test_blocking_pop();
    test_close();
    test_len();
    test_iterator();
    test_multiple_producers_consumers();

    println!("\n=== All tests passed! ===");
}

fn test_basic_push_pop() {
    println!("Test 1: Basic push/pop");
    let queue = MpmcQueue::new(3);
    
    assert!(queue.is_empty());
    assert_eq!(queue.len(), 0);
    
    queue.push(1).unwrap();
    queue.push(2).unwrap();
    queue.push(3).unwrap();
    
    assert!(queue.is_full());
    assert_eq!(queue.len(), 3);
    
    assert_eq!(queue.pop().unwrap(), 1);
    assert_eq!(queue.pop().unwrap(), 2);
    assert_eq!(queue.pop().unwrap(), 3);
    
    assert!(queue.is_empty());
    println!("  PASSED");
}

fn test_try_push_try_pop() {
    println!("Test 2: try_push/try_pop");
    let queue = MpmcQueue::new(2);
    
    assert_eq!(queue.try_push(1), Ok(()));
    assert_eq!(queue.try_push(2), Ok(()));
    assert_eq!(queue.try_push(3), Err(QueueError::Full));
    
    assert_eq!(queue.try_pop(), Ok(1));
    assert_eq!(queue.try_pop(), Ok(2));
    assert_eq!(queue.try_pop(), Err(QueueError::Empty));
    
    println!("  PASSED");
}

fn test_blocking_push() {
    println!("Test 3: Blocking push (wait when full)");
    let queue = Arc::new(MpmcQueue::new(2));
    let queue_clone = Arc::clone(&queue);
    
    queue.push(1).unwrap();
    queue.push(2).unwrap();
    
    let handle = thread::spawn(move || {
        thread::sleep(Duration::from_millis(100));
        let val = queue_clone.pop().unwrap();
        assert_eq!(val, 1);
        val
    });
    
    queue.push(3).unwrap();
    
    let result = handle.join().unwrap();
    assert_eq!(result, 1);
    println!("  PASSED");
}

fn test_blocking_pop() {
    println!("Test 4: Blocking pop (wait when empty)");
    let queue = Arc::new(MpmcQueue::new(2));
    let queue_clone = Arc::clone(&queue);
    
    let handle = thread::spawn(move || {
        let val = queue_clone.pop().unwrap();
        val
    });
    
    thread::sleep(Duration::from_millis(100));
    queue.push(42).unwrap();
    
    let result = handle.join().unwrap();
    assert_eq!(result, 42);
    println!("  PASSED");
}

fn test_close() {
    println!("Test 5: Close functionality");
    let queue = Arc::new(MpmcQueue::new(2));
    let queue_clone = Arc::clone(&queue);
    
    queue.push(1).unwrap();
    queue.push(2).unwrap();
    
    let pop_handle = thread::spawn(move || {
        let mut results = Vec::new();
        results.push(queue_clone.pop());
        results.push(queue_clone.pop());
        results.push(queue_clone.pop());
        results
    });
    
    let queue_clone2 = Arc::clone(&queue);
    let push_handle = thread::spawn(move || {
        thread::sleep(Duration::from_millis(50));
        queue_clone2.close();
        queue_clone2.try_push(3)
    });
    
    let pop_results = pop_handle.join().unwrap();
    let push_result = push_handle.join().unwrap();
    
    assert!(pop_results[0].is_ok() || pop_results[0] == Err(QueueError::Closed));
    assert!(pop_results[1].is_ok() || pop_results[1] == Err(QueueError::Closed));
    assert_eq!(pop_results[2], Err(QueueError::Closed));
    assert_eq!(push_result, Err(QueueError::Closed));
    
    let queue2 = MpmcQueue::new(2);
    queue2.close();
    assert_eq!(queue2.push(1), Err(QueueError::Closed));
    assert_eq!(queue2.try_push(1), Err(QueueError::Closed));
    
    println!("  PASSED");
}

fn test_len() {
    println!("Test 6: len() method");
    let queue = MpmcQueue::new(5);
    
    assert_eq!(queue.len(), 0);
    
    queue.push(1).unwrap();
    assert_eq!(queue.len(), 1);
    
    queue.push(2).unwrap();
    queue.push(3).unwrap();
    assert_eq!(queue.len(), 3);
    
    queue.pop().unwrap();
    assert_eq!(queue.len(), 2);
    
    queue.pop().unwrap();
    queue.pop().unwrap();
    assert_eq!(queue.len(), 0);
    
    println!("  PASSED");
}

fn test_iterator() {
    println!("Test 7: Iterator (into_iter)");
    let queue = MpmcQueue::new(5);
    
    queue.push(10).unwrap();
    queue.push(20).unwrap();
    queue.push(30).unwrap();
    
    let collected: Vec<i32> = queue.into_iter().collect();
    assert_eq!(collected, vec![10, 20, 30]);
    
    let queue2 = MpmcQueue::new(3);
    queue2.push("a").unwrap();
    queue2.push("b").unwrap();
    
    let mut iter = queue2.into_iter();
    assert_eq!(iter.next(), Some("a"));
    assert_eq!(iter.next(), Some("b"));
    assert_eq!(iter.next(), None);
    
    println!("  PASSED");
}

fn test_multiple_producers_consumers() {
    println!("Test 8: Multiple producers and consumers");
    let queue = Arc::new(MpmcQueue::new(100));
    let mut handles = Vec::new();
    
    let num_producers: usize = 4;
    let num_consumers: usize = 4;
    let items_per_producer: usize = 100;
    
    for i in 0..num_producers {
        let queue_clone = Arc::clone(&queue);
        handles.push(thread::spawn(move || {
            for j in 0..items_per_producer {
                queue_clone.push((i, j)).unwrap();
            }
        }));
    }
    
    let received: Arc<std::sync::Mutex<Vec<(usize, usize)>>> = Arc::new(std::sync::Mutex::new(Vec::new()));
    for _ in 0..num_consumers {
        let queue_clone = Arc::clone(&queue);
        let received_clone = Arc::clone(&received);
        handles.push(thread::spawn(move || {
            loop {
                match queue_clone.pop() {
                    Ok(item) => {
                        received_clone.lock().unwrap().push(item);
                    }
                    Err(_) => break,
                }
            }
        }));
    }
    
    for handle in handles.drain(..num_producers) {
        handle.join().unwrap();
    }
    
    queue.close();
    
    for handle in handles {
        handle.join().unwrap();
    }
    
    let received = received.lock().unwrap();
    assert_eq!(received.len(), num_producers * items_per_producer);
    
    for i in 0..num_producers {
        let mut producer_items: Vec<usize> = received
            .iter()
            .filter(|(prod, _)| *prod == i)
            .map(|(_, val)| *val)
            .collect();
        producer_items.sort();
        
        let expected: Vec<usize> = (0..items_per_producer).collect();
        assert_eq!(producer_items, expected);
    }
    
    println!("  PASSED");
}
