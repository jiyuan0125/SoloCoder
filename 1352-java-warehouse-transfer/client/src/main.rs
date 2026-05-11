use clap::{Parser, Subcommand};
use warehouse_core::{
    AddInventoryRequest, CreateTransferRequest, DiscrepancyReason, ReceiveRequest,
};
use reqwest::Client;
use serde::de::DeserializeOwned;

#[derive(Parser, Debug)]
#[command(version, about, long_about = None)]
struct Args {
    #[arg(long, env = "SERVER_URL", default_value = "http://127.0.0.1:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Warehouse {
        #[command(subcommand)]
        action: WarehouseCommands,
    },
    Inventory {
        #[command(subcommand)]
        action: InventoryCommands,
    },
    Transfer {
        #[command(subcommand)]
        action: TransferCommands,
    },
    Logs {
        #[command(subcommand)]
        action: LogsCommands,
    },
}

#[derive(Subcommand, Debug)]
enum WarehouseCommands {
    Create {
        #[arg(long)]
        name: String,
        #[arg(long)]
        city: String,
    },
    List,
    Get {
        id: String,
    },
}

#[derive(Subcommand, Debug)]
enum InventoryCommands {
    Add {
        #[arg(long)]
        warehouse_id: String,
        #[arg(long)]
        product_id: String,
        #[arg(long)]
        product_name: String,
        #[arg(long)]
        quantity: u32,
    },
    List {
        #[arg(long)]
        warehouse_id: Option<String>,
    },
    Get {
        warehouse_id: String,
        product_id: String,
    },
}

#[derive(Subcommand, Debug)]
enum TransferCommands {
    Create {
        #[arg(long)]
        source: String,
        #[arg(long)]
        target: String,
        #[arg(long)]
        product: String,
        #[arg(long)]
        quantity: u32,
    },
    List,
    Get {
        id: String,
    },
    Ship {
        id: String,
    },
    Receive {
        #[arg(long)]
        transfer_id: String,
        #[arg(long)]
        quantity: u32,
        #[arg(long)]
        reason: Option<DiscrepancyReasonArg>,
    },
    Return {
        id: String,
    },
    Cancel {
        id: String,
    },
}

#[derive(Subcommand, Debug)]
enum LogsCommands {
    List,
    ByTransfer {
        id: String,
    },
}

#[derive(clap::ValueEnum, Debug, Clone, Copy)]
enum DiscrepancyReasonArg {
    Lost,
    Damaged,
    Rejected,
}

impl From<DiscrepancyReasonArg> for DiscrepancyReason {
    fn from(arg: DiscrepancyReasonArg) -> Self {
        match arg {
            DiscrepancyReasonArg::Lost => DiscrepancyReason::Lost,
            DiscrepancyReasonArg::Damaged => DiscrepancyReason::Damaged,
            DiscrepancyReasonArg::Rejected => DiscrepancyReason::Rejected,
        }
    }
}

struct ApiClient {
    base_url: String,
    client: Client,
}

impl ApiClient {
    fn new(base_url: String) -> Self {
        Self {
            base_url,
            client: Client::new(),
        }
    }

    async fn get<T: DeserializeOwned>(&self, path: &str) -> Result<T, reqwest::Error> {
        let url = format!("{}{}", self.base_url, path);
        self.client.get(&url).send().await?.json().await
    }

    async fn post<T: DeserializeOwned, B: serde::Serialize>(
        &self,
        path: &str,
        body: Option<&B>,
    ) -> Result<T, reqwest::Error> {
        let url = format!("{}{}", self.base_url, path);
        let mut req = self.client.post(&url);
        if let Some(b) = body {
            req = req.json(b);
        }
        req.send().await?.json().await
    }

    async fn post_no_body<T: DeserializeOwned>(&self, path: &str) -> Result<T, reqwest::Error> {
        let url = format!("{}{}", self.base_url, path);
        self.client.post(&url).send().await?.json().await
    }
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    let api = ApiClient::new(args.server_url);

    match args.command {
        Commands::Warehouse { action } => match action {
            WarehouseCommands::Create { name, city } => {
                let res: serde_json::Value = api
                    .post_no_body(&format!("/api/warehouses?name={}&city={}", name, city))
                    .await
                    .unwrap();
                println!("{}", serde_json::to_string_pretty(&res).unwrap());
            }
            WarehouseCommands::List => {
                let res: serde_json::Value = api.get("/api/warehouses").await.unwrap();
                println!("{}", serde_json::to_string_pretty(&res).unwrap());
            }
            WarehouseCommands::Get { id } => {
                let res: serde_json::Value = api.get(&format!("/api/warehouses/{}", id)).await.unwrap();
                println!("{}", serde_json::to_string_pretty(&res).unwrap());
            }
        },
        Commands::Inventory { action } => match action {
            InventoryCommands::Add {
                warehouse_id,
                product_id,
                product_name,
                quantity,
            } => {
                let req = AddInventoryRequest {
                    warehouse_id,
                    product_id,
                    product_name,
                    quantity,
                };
                let res: serde_json::Value = api.post("/api/inventory", Some(&req)).await.unwrap();
                println!("{}", serde_json::to_string_pretty(&res).unwrap());
            }
            InventoryCommands::List { warehouse_id } => {
                let path = match warehouse_id {
                    Some(id) => format!("/api/inventory?warehouse_id={}", id),
                    None => "/api/inventory".to_string(),
                };
                let res: serde_json::Value = api.get(&path).await.unwrap();
                println!("{}", serde_json::to_string_pretty(&res).unwrap());
            }
            InventoryCommands::Get {
                warehouse_id,
                product_id,
            } => {
                let res: serde_json::Value = api
                    .get(&format!("/api/inventory/{}/{}", warehouse_id, product_id))
                    .await
                    .unwrap();
                println!("{}", serde_json::to_string_pretty(&res).unwrap());
            }
        },
        Commands::Transfer { action } => match action {
            TransferCommands::Create {
                source,
                target,
                product,
                quantity,
            } => {
                let req = CreateTransferRequest {
                    source_warehouse_id: source,
                    target_warehouse_id: target,
                    product_id: product,
                    quantity,
                };
                let res: serde_json::Value = api.post("/api/transfers", Some(&req)).await.unwrap();
                println!("{}", serde_json::to_string_pretty(&res).unwrap());
            }
            TransferCommands::List => {
                let res: serde_json::Value = api.get("/api/transfers").await.unwrap();
                println!("{}", serde_json::to_string_pretty(&res).unwrap());
            }
            TransferCommands::Get { id } => {
                let res: serde_json::Value = api.get(&format!("/api/transfers/{}", id)).await.unwrap();
                println!("{}", serde_json::to_string_pretty(&res).unwrap());
            }
            TransferCommands::Ship { id } => {
                let res: serde_json::Value = api
                    .post_no_body(&format!("/api/transfers/{}/ship", id))
                    .await
                    .unwrap();
                println!("{}", serde_json::to_string_pretty(&res).unwrap());
            }
            TransferCommands::Receive {
                transfer_id,
                quantity,
                reason,
            } => {
                let req = ReceiveRequest {
                    transfer_id,
                    received_quantity: quantity,
                    discrepancy_reason: reason.map(|r| r.into()),
                };
                let res: serde_json::Value = api.post("/api/transfers/receive", Some(&req)).await.unwrap();
                println!("{}", serde_json::to_string_pretty(&res).unwrap());
            }
            TransferCommands::Return { id } => {
                let res: serde_json::Value = api
                    .post_no_body(&format!("/api/transfers/{}/return", id))
                    .await
                    .unwrap();
                println!("{}", serde_json::to_string_pretty(&res).unwrap());
            }
            TransferCommands::Cancel { id } => {
                let res: serde_json::Value = api
                    .post_no_body(&format!("/api/transfers/{}/cancel", id))
                    .await
                    .unwrap();
                println!("{}", serde_json::to_string_pretty(&res).unwrap());
            }
        },
        Commands::Logs { action } => match action {
            LogsCommands::List => {
                let res: serde_json::Value = api.get("/api/logs").await.unwrap();
                println!("{}", serde_json::to_string_pretty(&res).unwrap());
            }
            LogsCommands::ByTransfer { id } => {
                let res: serde_json::Value = api
                    .get(&format!("/api/logs/transfer/{}", id))
                    .await
                    .unwrap();
                println!("{}", serde_json::to_string_pretty(&res).unwrap());
            }
        },
    }
}
