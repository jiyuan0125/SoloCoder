use kv_engine::StorageEngine;
use std::fs::remove_file;

const TEST_FILE: &str = "test_kv.dat";

fn cleanup() {
    let _ = remove_file(TEST_FILE);
}

fn main() {
    cleanup();
    
    println!("=== Test 1: Basic put/get ===");
    {
        let mut engine = StorageEngine::new(TEST_FILE).unwrap();
        
        engine.put("key1".to_string(), b"value1".to_vec()).unwrap();
        engine.put("key2".to_string(), b"value2".to_vec()).unwrap();
        engine.put("key3".to_string(), b"value3".to_vec()).unwrap();
        
        let v1 = engine.get("key1").unwrap();
        assert_eq!(v1, Some(b"value1".to_vec()), "key1 should be value1");
        println!("get key1: {:?}", String::from_utf8(v1.unwrap()));
        
        let v2 = engine.get("key2").unwrap();
        assert_eq!(v2, Some(b"value2".to_vec()), "key2 should be value2");
        println!("get key2: {:?}", String::from_utf8(v2.unwrap()));
        
        let count = engine.count();
        assert_eq!(count, 3, "count should be 3");
        println!("count: {}", count);
    }
    
    println!("\n=== Test 2: Restart and verify persistence ===");
    {
        let mut engine = StorageEngine::new(TEST_FILE).unwrap();
        
        let count = engine.count();
        assert_eq!(count, 3, "count should still be 3 after restart");
        println!("count after restart: {}", count);
        
        let v1 = engine.get("key1").unwrap();
        assert_eq!(v1, Some(b"value1".to_vec()), "key1 should persist");
        println!("get key1 after restart: {:?}", String::from_utf8(v1.unwrap()));
    }
    
    println!("\n=== Test 3: Update and delete ===");
    {
        let mut engine = StorageEngine::new(TEST_FILE).unwrap();
        
        engine.put("key1".to_string(), b"updated_value1".to_vec()).unwrap();
        engine.delete("key2".to_string()).unwrap();
        
        let v1 = engine.get("key1").unwrap();
        assert_eq!(v1, Some(b"updated_value1".to_vec()), "key1 should be updated");
        println!("get key1 after update: {:?}", String::from_utf8(v1.unwrap()));
        
        let v2 = engine.get("key2").unwrap();
        assert_eq!(v2, None, "key2 should be deleted");
        println!("get key2 after delete: {:?}", v2);
        
        let count = engine.count();
        assert_eq!(count, 2, "count should be 2");
        println!("count: {}", count);
    }
    
    println!("\n=== Test 4: Scan prefix ===");
    {
        let mut engine = StorageEngine::new(TEST_FILE).unwrap();
        
        engine.put("user:1".to_string(), b"alice".to_vec()).unwrap();
        engine.put("user:2".to_string(), b"bob".to_vec()).unwrap();
        engine.put("post:1".to_string(), b"hello".to_vec()).unwrap();
        
        let users = engine.scan("user:");
        println!("scan 'user:': {:?}", users);
        assert_eq!(users.len(), 2, "should find 2 users");
        assert!(users.contains(&"user:1".to_string()));
        assert!(users.contains(&"user:2".to_string()));
        
        let all = engine.scan("");
        println!("scan '': {:?}", all);
        assert_eq!(all.len(), 4, "should find 4 keys total");
    }
    
    println!("\n=== Test 5: Restart and verify all ===");
    {
        let mut engine = StorageEngine::new(TEST_FILE).unwrap();
        
        let count = engine.count();
        assert_eq!(count, 4, "count should be 4");
        println!("final count: {}", count);
        
        let all = engine.scan("");
        println!("final keys: {:?}", all);
    }
    
    cleanup();
    println!("\n✓ All tests passed!");
}
