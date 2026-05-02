mod packet;
mod topic;
mod session;
mod broker;

use std::net::TcpListener;
use std::thread;
use std::sync::{Arc, Mutex};
use broker::Broker;

fn main() -> std::io::Result<()> {
    let listener = TcpListener::bind("0.0.0.0:1883")?;
    println!("MQTT Broker listening on port 1883");

    let broker = Arc::new(Mutex::new(Broker::new()));

    for stream in listener.incoming() {
        match stream {
            Ok(stream) => {
                let broker_clone = Arc::clone(&broker);
                thread::spawn(move || {
                    broker::handle_client(stream, broker_clone);
                });
            }
            Err(e) => eprintln!("Connection failed: {}", e),
        }
    }

    Ok(())
}
