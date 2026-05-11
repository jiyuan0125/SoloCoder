use clap::{Parser, Subcommand};
use purchase_return_core::{
    CreateExchangeRequest, CreateProductRequest, CreatePurchaseRequest, CreateReturnRequest,
    CreateSupplierRequest,
};
use reqwest::blocking::Client;
use serde::{Deserialize, Serialize};
use std::error::Error;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
pub struct Args {
    #[arg(long, env = "SERVER_URL", default_value = "http://127.0.0.1:8080")]
    pub server_url: String,

    #[command(subcommand)]
    pub command: Commands,
}

#[derive(Subcommand, Debug)]
pub enum Commands {
    ListProducts,
    ListSuppliers,
    ListBatches,
    ListReturns,
    ListExchanges,
    CreateProduct { name: String },
    CreateSupplier { name: String },
    Purchase {
        product_id: String,
        supplier_id: String,
        batch_number: String,
        price: String,
        quantity: u32,
    },
    Return {
        batch_id: String,
        quantity: u32,
    },
    Exchange {
        old_product_id: String,
        old_quantity: u32,
        new_product_id: String,
        new_quantity: u32,
    },
}

#[derive(Debug, Serialize, Deserialize)]
struct ApiError {
    error: String,
}

pub fn run_client(args: &Args) -> Result<(), Box<dyn Error>> {
    let client = Client::new();

    match &args.command {
        Commands::ListProducts => list_products(&client, &args.server_url),
        Commands::ListSuppliers => list_suppliers(&client, &args.server_url),
        Commands::ListBatches => list_batches(&client, &args.server_url),
        Commands::ListReturns => list_returns(&client, &args.server_url),
        Commands::ListExchanges => list_exchanges(&client, &args.server_url),
        Commands::CreateProduct { name } => create_product(&client, &args.server_url, name),
        Commands::CreateSupplier { name } => create_supplier(&client, &args.server_url, name),
        Commands::Purchase {
            product_id,
            supplier_id,
            batch_number,
            price,
            quantity,
        } => create_purchase(
            &client,
            &args.server_url,
            product_id,
            supplier_id,
            batch_number,
            price,
            *quantity,
        ),
        Commands::Return { batch_id, quantity } => {
            create_return(&client, &args.server_url, batch_id, *quantity)
        }
        Commands::Exchange {
            old_product_id,
            old_quantity,
            new_product_id,
            new_quantity,
        } => create_exchange(
            &client,
            &args.server_url,
            old_product_id,
            *old_quantity,
            new_product_id,
            *new_quantity,
        ),
    }
}

fn list_products(client: &Client, server_url: &str) -> Result<(), Box<dyn Error>> {
    let resp = client
        .get(&format!("{}/products", server_url))
        .send()?
        .error_for_status()?;
    let json: serde_json::Value = resp.json()?;
    println!("{}", serde_json::to_string_pretty(&json)?);
    Ok(())
}

fn list_suppliers(client: &Client, server_url: &str) -> Result<(), Box<dyn Error>> {
    let resp = client
        .get(&format!("{}/suppliers", server_url))
        .send()?
        .error_for_status()?;
    let json: serde_json::Value = resp.json()?;
    println!("{}", serde_json::to_string_pretty(&json)?);
    Ok(())
}

fn list_batches(client: &Client, server_url: &str) -> Result<(), Box<dyn Error>> {
    let resp = client
        .get(&format!("{}/batches", server_url))
        .send()?
        .error_for_status()?;
    let json: serde_json::Value = resp.json()?;
    println!("{}", serde_json::to_string_pretty(&json)?);
    Ok(())
}

fn list_returns(client: &Client, server_url: &str) -> Result<(), Box<dyn Error>> {
    let resp = client
        .get(&format!("{}/returns", server_url))
        .send()?
        .error_for_status()?;
    let json: serde_json::Value = resp.json()?;
    println!("{}", serde_json::to_string_pretty(&json)?);
    Ok(())
}

fn list_exchanges(client: &Client, server_url: &str) -> Result<(), Box<dyn Error>> {
    let resp = client
        .get(&format!("{}/exchanges", server_url))
        .send()?
        .error_for_status()?;
    let json: serde_json::Value = resp.json()?;
    println!("{}", serde_json::to_string_pretty(&json)?);
    Ok(())
}

fn create_product(client: &Client, server_url: &str, name: &str) -> Result<(), Box<dyn Error>> {
    let req = CreateProductRequest {
        name: name.to_string(),
    };
    let resp = client
        .post(&format!("{}/products", server_url))
        .json(&req)
        .send()?;

    if resp.status().is_success() {
        let json: serde_json::Value = resp.json()?;
        println!("Created product:");
        println!("{}", serde_json::to_string_pretty(&json)?);
    } else {
        let err: ApiError = resp.json()?;
        eprintln!("Error: {}", err.error);
        std::process::exit(1);
    }
    Ok(())
}

fn create_supplier(client: &Client, server_url: &str, name: &str) -> Result<(), Box<dyn Error>> {
    let req = CreateSupplierRequest {
        name: name.to_string(),
    };
    let resp = client
        .post(&format!("{}/suppliers", server_url))
        .json(&req)
        .send()?;

    if resp.status().is_success() {
        let json: serde_json::Value = resp.json()?;
        println!("Created supplier:");
        println!("{}", serde_json::to_string_pretty(&json)?);
    } else {
        let err: ApiError = resp.json()?;
        eprintln!("Error: {}", err.error);
        std::process::exit(1);
    }
    Ok(())
}

fn create_purchase(
    client: &Client,
    server_url: &str,
    product_id: &str,
    supplier_id: &str,
    batch_number: &str,
    price: &str,
    quantity: u32,
) -> Result<(), Box<dyn Error>> {
    let req = CreatePurchaseRequest {
        product_id: product_id.to_string(),
        supplier_id: supplier_id.to_string(),
        batch_number: batch_number.to_string(),
        price: price.to_string(),
        quantity,
    };
    let resp = client
        .post(&format!("{}/batches", server_url))
        .json(&req)
        .send()?;

    if resp.status().is_success() {
        let json: serde_json::Value = resp.json()?;
        println!("Created purchase batch:");
        println!("{}", serde_json::to_string_pretty(&json)?);
    } else {
        let err: ApiError = resp.json()?;
        eprintln!("Error: {}", err.error);
        std::process::exit(1);
    }
    Ok(())
}

fn create_return(
    client: &Client,
    server_url: &str,
    batch_id: &str,
    quantity: u32,
) -> Result<(), Box<dyn Error>> {
    let req = CreateReturnRequest {
        batch_id: batch_id.to_string(),
        quantity,
    };
    let resp = client
        .post(&format!("{}/returns", server_url))
        .json(&req)
        .send()?;

    if resp.status().is_success() {
        let json: serde_json::Value = resp.json()?;
        println!("Created return:");
        println!("{}", serde_json::to_string_pretty(&json)?);
    } else {
        let err: ApiError = resp.json()?;
        eprintln!("Error: {}", err.error);
        std::process::exit(1);
    }
    Ok(())
}

fn create_exchange(
    client: &Client,
    server_url: &str,
    old_product_id: &str,
    old_quantity: u32,
    new_product_id: &str,
    new_quantity: u32,
) -> Result<(), Box<dyn Error>> {
    let req = CreateExchangeRequest {
        old_product_id: old_product_id.to_string(),
        old_quantity,
        new_product_id: new_product_id.to_string(),
        new_quantity,
    };
    let resp = client
        .post(&format!("{}/exchanges", server_url))
        .json(&req)
        .send()?;

    if resp.status().is_success() {
        let json: serde_json::Value = resp.json()?;
        println!("Created exchange:");
        println!("{}", serde_json::to_string_pretty(&json)?);
    } else {
        let err: ApiError = resp.json()?;
        eprintln!("Error: {}", err.error);
        std::process::exit(1);
    }
    Ok(())
}
