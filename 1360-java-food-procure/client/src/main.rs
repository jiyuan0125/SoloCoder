use clap::{Parser, Subcommand, ValueEnum};
use food_procure_core::models::*;
use reqwest::blocking::Client;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://127.0.0.1:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Health,
    
    CreateSupplier {
        #[arg(long)]
        name: String,
        #[arg(long)]
        contact: String,
        #[arg(long)]
        phone: String,
    },
    
    ListSuppliers,
    
    GetSupplier {
        #[arg(long)]
        id: Uuid,
    },
    
    UpdateSupplierScore {
        #[arg(long)]
        id: Uuid,
        #[arg(long)]
        month: String,
        #[arg(long)]
        score: f64,
    },
    
    CreateIngredient {
        #[arg(long)]
        name: String,
        #[arg(long)]
        unit: String,
        #[arg(long)]
        category: String,
        #[arg(long)]
        weight: f64,
    },
    
    ListIngredients,
    
    CreateStore {
        #[arg(long)]
        name: String,
        #[arg(long)]
        address: String,
        #[arg(long)]
        contact: String,
        #[arg(long)]
        phone: String,
    },
    
    ListStores,
    
    CreatePurchaseRequest {
        #[arg(long)]
        store_id: Uuid,
        #[arg(long)]
        type_: PurchaseTypeArg,
        #[arg(long, value_parser = parse_ingredient_quantity)]
        items: Vec<(Uuid, f64)>,
    },
    
    ListPurchaseRequests,
    
    GetPurchaseRequest {
        #[arg(long)]
        id: Uuid,
    },
    
    RequestQuotations {
        #[arg(long)]
        id: Uuid,
    },
    
    ListQuotations {
        #[arg(long)]
        request_id: Uuid,
    },
    
    PlaceOrder {
        #[arg(long)]
        quotation_id: Uuid,
        #[arg(long)]
        reason: Option<String>,
    },
    
    GetOrder {
        #[arg(long)]
        id: Uuid,
    },
    
    ConfirmOrder {
        #[arg(long)]
        id: Uuid,
    },
    
    PlaceEmergencyOrder {
        #[arg(long)]
        store_id: Uuid,
        #[arg(long)]
        supplier_id: Uuid,
        #[arg(long, value_parser = parse_ingredient_quantity)]
        items: Vec<(Uuid, f64)>,
    },
    
    GetEmergencyStats {
        #[arg(long)]
        store_id: Uuid,
    },
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, ValueEnum)]
enum PurchaseTypeArg {
    Normal,
    Emergency,
}

impl From<PurchaseTypeArg> for PurchaseType {
    fn from(arg: PurchaseTypeArg) -> Self {
        match arg {
            PurchaseTypeArg::Normal => PurchaseType::Normal,
            PurchaseTypeArg::Emergency => PurchaseType::Emergency,
        }
    }
}

fn parse_ingredient_quantity(s: &str) -> Result<(Uuid, f64), String> {
    let parts: Vec<&str> = s.split(':').collect();
    if parts.len() != 2 {
        return Err("Format must be ingredient_id:quantity".to_string());
    }
    let id = Uuid::parse_str(parts[0]).map_err(|e| e.to_string())?;
    let qty = parts[1].parse::<f64>().map_err(|e| e.to_string())?;
    Ok((id, qty))
}

#[derive(Serialize)]
struct CreateSupplierReq {
    name: String,
    contact: String,
    phone: String,
    ingredients: Vec<SupplierIngredient>,
}

#[derive(Serialize)]
struct CreateIngredientReq {
    name: String,
    unit: String,
    category: String,
    weight_per_unit: f64,
}

#[derive(Serialize)]
struct CreateStoreReq {
    name: String,
    address: String,
    contact: String,
    phone: String,
}

#[derive(Serialize)]
struct CreatePurchaseReq {
    store_id: Uuid,
    purchase_type: PurchaseType,
    items: Vec<PurchaseItem>,
}

#[derive(Serialize)]
struct PlaceOrderReq {
    quotation_id: Uuid,
    non_recommended_reason: Option<String>,
}

#[derive(Serialize)]
struct PlaceEmergencyOrderReq {
    store_id: Uuid,
    supplier_id: Uuid,
    items: Vec<PurchaseItem>,
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();
    let client = Client::new();

    match cli.command {
        Commands::Health => {
            let resp = client.get(format!("{}/health", cli.server_url)).send()?;
            println!("{}", resp.text()?);
        }
        
        Commands::CreateSupplier { name, contact, phone } => {
            let req = CreateSupplierReq {
                name,
                contact,
                phone,
                ingredients: vec![],
            };
            let resp = client.post(format!("{}/suppliers", cli.server_url))
                .json(&req)
                .send()?;
            println!("{}", serde_json::to_string_pretty(&resp.json::<serde_json::Value>()?)?);
        }
        
        Commands::ListSuppliers => {
            let resp = client.get(format!("{}/suppliers", cli.server_url)).send()?;
            println!("{}", serde_json::to_string_pretty(&resp.json::<serde_json::Value>()?)?);
        }
        
        Commands::GetSupplier { id } => {
            let resp = client.get(format!("{}/suppliers/{}", cli.server_url, id)).send()?;
            println!("{}", serde_json::to_string_pretty(&resp.json::<serde_json::Value>()?)?);
        }
        
        Commands::UpdateSupplierScore { id, month, score } => {
            let body = serde_json::json!({ "month": month, "score": score });
            let resp = client.post(format!("{}/suppliers/{}/score", cli.server_url, id))
                .json(&body)
                .send()?;
            println!("{}", serde_json::to_string_pretty(&resp.json::<serde_json::Value>()?)?);
        }
        
        Commands::CreateIngredient { name, unit, category, weight } => {
            let req = CreateIngredientReq {
                name,
                unit,
                category,
                weight_per_unit: weight,
            };
            let resp = client.post(format!("{}/ingredients", cli.server_url))
                .json(&req)
                .send()?;
            println!("{}", serde_json::to_string_pretty(&resp.json::<serde_json::Value>()?)?);
        }
        
        Commands::ListIngredients => {
            let resp = client.get(format!("{}/ingredients", cli.server_url)).send()?;
            println!("{}", serde_json::to_string_pretty(&resp.json::<serde_json::Value>()?)?);
        }
        
        Commands::CreateStore { name, address, contact, phone } => {
            let req = CreateStoreReq {
                name,
                address,
                contact,
                phone,
            };
            let resp = client.post(format!("{}/stores", cli.server_url))
                .json(&req)
                .send()?;
            println!("{}", serde_json::to_string_pretty(&resp.json::<serde_json::Value>()?)?);
        }
        
        Commands::ListStores => {
            let resp = client.get(format!("{}/stores", cli.server_url)).send()?;
            println!("{}", serde_json::to_string_pretty(&resp.json::<serde_json::Value>()?)?);
        }
        
        Commands::CreatePurchaseRequest { store_id, type_, items } => {
            let items: Vec<PurchaseItem> = items.into_iter()
                .map(|(id, qty)| PurchaseItem {
                    id: Uuid::new_v4(),
                    ingredient_id: id,
                    requested_quantity: qty,
                })
                .collect();
            let req = CreatePurchaseReq {
                store_id,
                purchase_type: type_.into(),
                items,
            };
            let resp = client.post(format!("{}/purchase-requests", cli.server_url))
                .json(&req)
                .send()?;
            println!("{}", serde_json::to_string_pretty(&resp.json::<serde_json::Value>()?)?);
        }
        
        Commands::ListPurchaseRequests => {
            let resp = client.get(format!("{}/purchase-requests", cli.server_url)).send()?;
            println!("{}", serde_json::to_string_pretty(&resp.json::<serde_json::Value>()?)?);
        }
        
        Commands::GetPurchaseRequest { id } => {
            let resp = client.get(format!("{}/purchase-requests/{}", cli.server_url, id)).send()?;
            println!("{}", serde_json::to_string_pretty(&resp.json::<serde_json::Value>()?)?);
        }
        
        Commands::RequestQuotations { id } => {
            let resp = client.post(format!("{}/purchase-requests/{}/quotations", cli.server_url, id))
                .send()?;
            println!("{}", serde_json::to_string_pretty(&resp.json::<serde_json::Value>()?)?);
        }
        
        Commands::ListQuotations { request_id } => {
            let resp = client.get(format!("{}/purchase-requests/{}/quotations", cli.server_url, request_id)).send()?;
            println!("{}", serde_json::to_string_pretty(&resp.json::<serde_json::Value>()?)?);
        }
        
        Commands::PlaceOrder { quotation_id, reason } => {
            let req = PlaceOrderReq {
                quotation_id,
                non_recommended_reason: reason,
            };
            let resp = client.post(format!("{}/orders", cli.server_url))
                .json(&req)
                .send()?;
            println!("{}", serde_json::to_string_pretty(&resp.json::<serde_json::Value>()?)?);
        }
        
        Commands::GetOrder { id } => {
            let resp = client.get(format!("{}/orders/{}", cli.server_url, id)).send()?;
            println!("{}", serde_json::to_string_pretty(&resp.json::<serde_json::Value>()?)?);
        }
        
        Commands::ConfirmOrder { id } => {
            let resp = client.post(format!("{}/orders/{}", cli.server_url, id)).send()?;
            println!("{}", serde_json::to_string_pretty(&resp.json::<serde_json::Value>()?)?);
        }
        
        Commands::PlaceEmergencyOrder { store_id, supplier_id, items } => {
            let items: Vec<PurchaseItem> = items.into_iter()
                .map(|(id, qty)| PurchaseItem {
                    id: Uuid::new_v4(),
                    ingredient_id: id,
                    requested_quantity: qty,
                })
                .collect();
            let req = PlaceEmergencyOrderReq {
                store_id,
                supplier_id,
                items,
            };
            let resp = client.post(format!("{}/emergency-orders", cli.server_url))
                .json(&req)
                .send()?;
            println!("{}", serde_json::to_string_pretty(&resp.json::<serde_json::Value>()?)?);
        }
        
        Commands::GetEmergencyStats { store_id } => {
            let resp = client.get(format!("{}/stores/{}/emergency-stats", cli.server_url, store_id)).send()?;
            println!("{}", serde_json::to_string_pretty(&resp.json::<serde_json::Value>()?)?);
        }
    }

    Ok(())
}
