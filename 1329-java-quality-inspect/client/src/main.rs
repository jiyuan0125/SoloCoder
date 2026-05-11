use clap::{Parser, Subcommand};
use dialoguer::{theme::ColorfulTheme, Confirm, Input, Select};
use reqwest::Client;
use std::time::Duration;
use uuid::Uuid;

use qa_core::{
    CompleteInspectionRequest, CreateOrderRequest, DefectRecord, InspectionOrder,
    InspectionResult, InspectItemRequest, ReinspectionRequest, UpdateDefectRequest,
};

#[derive(Parser, Debug)]
#[command(author, version, about = "QA Inspection Client", long_about = None)]
struct Cli {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://127.0.0.1:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Create {
        #[arg(short, long)]
        product: Option<String>,
        #[arg(short, long)]
        batch_size: Option<u32>,
        #[arg(short, long)]
        aql: Option<String>,
        #[arg(short, long)]
        inspector: Option<String>,
    },
    List,
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
    Inspect {
        #[arg(short, long)]
        id: Uuid,
    },
    Defects {
        #[arg(short, long)]
        id: Uuid,
    },
    UpdateDefect {
        #[arg(short, long)]
        defect_id: Uuid,
        #[arg(short, long)]
        description: Option<String>,
    },
    Complete {
        #[arg(short, long)]
        id: Uuid,
    },
    Reinspect {
        #[arg(short, long)]
        id: Uuid,
        #[arg(short, long)]
        inspector: Option<String>,
    },
}

struct ApiClient {
    client: Client,
    base_url: String,
}

impl ApiClient {
    fn new(base_url: &str) -> Self {
        ApiClient {
            client: Client::builder()
                .timeout(Duration::from_secs(30))
                .build()
                .unwrap(),
            base_url: base_url.to_string(),
        }
    }

    async fn create_order(&self, request: CreateOrderRequest) -> Result<InspectionOrder, Box<dyn std::error::Error>> {
        let url = format!("{}/api/orders", self.base_url);
        let response = self.client.post(&url).json(&request).send().await?;
        if !response.status().is_success() {
            let error_text = response.text().await?;
            return Err(format!("Server error: {}", error_text).into());
        }
        let order = response.json().await?;
        Ok(order)
    }

    async fn get_orders(&self) -> Result<Vec<InspectionOrder>, Box<dyn std::error::Error>> {
        let url = format!("{}/api/orders", self.base_url);
        let response = self.client.get(&url).send().await?;
        let orders = response.json().await?;
        Ok(orders)
    }

    async fn get_order(&self, id: Uuid) -> Result<InspectionOrder, Box<dyn std::error::Error>> {
        let url = format!("{}/api/orders/{}", self.base_url, id);
        let response = self.client.get(&url).send().await?;
        if !response.status().is_success() {
            let error_text = response.text().await?;
            return Err(format!("Server error: {}", error_text).into());
        }
        let order = response.json().await?;
        Ok(order)
    }

    async fn start_inspection(&self, id: Uuid) -> Result<InspectionOrder, Box<dyn std::error::Error>> {
        let url = format!("{}/api/orders/{}/start", self.base_url, id);
        let response = self.client.post(&url).send().await?;
        if !response.status().is_success() {
            let error_text = response.text().await?;
            return Err(format!("Server error: {}", error_text).into());
        }
        let order = response.json().await?;
        Ok(order)
    }

    async fn inspect_item(&self, request: InspectItemRequest) -> Result<Option<DefectRecord>, Box<dyn std::error::Error>> {
        let url = format!("{}/api/inspect", self.base_url);
        let response = self.client.post(&url).json(&request).send().await?;
        if !response.status().is_success() {
            let error_text = response.text().await?;
            return Err(format!("Server error: {}", error_text).into());
        }
        let defect = response.json().await?;
        Ok(defect)
    }

    async fn get_defects(&self, order_id: Uuid) -> Result<Vec<DefectRecord>, Box<dyn std::error::Error>> {
        let url = format!("{}/api/orders/{}/defects", self.base_url, order_id);
        let response = self.client.get(&url).send().await?;
        let defects = response.json().await?;
        Ok(defects)
    }

    async fn update_defect(&self, request: UpdateDefectRequest) -> Result<DefectRecord, Box<dyn std::error::Error>> {
        let url = format!("{}/api/defects/update", self.base_url);
        let response = self.client.post(&url).json(&request).send().await?;
        if !response.status().is_success() {
            let error_text = response.text().await?;
            return Err(format!("Server error: {}", error_text).into());
        }
        let defect = response.json().await?;
        Ok(defect)
    }

    async fn complete_inspection(&self, request: CompleteInspectionRequest) -> Result<InspectionResult, Box<dyn std::error::Error>> {
        let url = format!("{}/api/complete", self.base_url);
        let response = self.client.post(&url).json(&request).send().await?;
        if !response.status().is_success() {
            let error_text = response.text().await?;
            return Err(format!("Server error: {}", error_text).into());
        }
        let result = response.json().await?;
        Ok(result)
    }

    async fn request_reinspection(&self, request: ReinspectionRequest) -> Result<InspectionOrder, Box<dyn std::error::Error>> {
        let url = format!("{}/api/reinspect", self.base_url);
        let response = self.client.post(&url).json(&request).send().await?;
        if !response.status().is_success() {
            let error_text = response.text().await?;
            return Err(format!("Server error: {}", error_text).into());
        }
        let order = response.json().await?;
        Ok(order)
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();
    let api_client = ApiClient::new(&cli.server);

    match cli.command {
        Commands::Create { product, batch_size, aql, inspector } => {
            handle_create(&api_client, product, batch_size, aql, inspector).await?;
        }
        Commands::List => {
            handle_list(&api_client).await?;
        }
        Commands::Get { id } => {
            handle_get(&api_client, id).await?;
        }
        Commands::Inspect { id } => {
            handle_inspect(&api_client, id).await?;
        }
        Commands::Defects { id } => {
            handle_defects(&api_client, id).await?;
        }
        Commands::UpdateDefect { defect_id, description } => {
            handle_update_defect(&api_client, defect_id, description).await?;
        }
        Commands::Complete { id } => {
            handle_complete(&api_client, id).await?;
        }
        Commands::Reinspect { id, inspector } => {
            handle_reinspect(&api_client, id, inspector).await?;
        }
    }

    Ok(())
}

async fn handle_create(
    api: &ApiClient,
    product: Option<String>,
    batch_size: Option<u32>,
    aql: Option<String>,
    inspector: Option<String>,
) -> Result<(), Box<dyn std::error::Error>> {
    let product_name = product.unwrap_or_else(|| {
        Input::with_theme(&ColorfulTheme::default())
            .with_prompt("Product name")
            .interact_text()
            .unwrap()
    });

    let batch_size = batch_size.unwrap_or_else(|| {
        Input::with_theme(&ColorfulTheme::default())
            .with_prompt("Batch size")
            .interact_text()
            .unwrap()
    });

    let aql_levels = &["0.65", "1.0", "1.5", "2.5", "4.0", "6.5"];
    let aql_level = aql.unwrap_or_else(|| {
        let selection = Select::with_theme(&ColorfulTheme::default())
            .with_prompt("Select AQL level")
            .items(aql_levels)
            .default(0)
            .interact()
            .unwrap();
        aql_levels[selection].to_string()
    });

    let inspector_name = inspector.unwrap_or_else(|| {
        Input::with_theme(&ColorfulTheme::default())
            .with_prompt("Inspector name")
            .interact_text()
            .unwrap()
    });

    let order = api
        .create_order(CreateOrderRequest {
            product_name,
            batch_size,
            aql_level: aql_level,
            inspector: inspector_name,
        })
        .await?;

    println!("\n✓ Order created successfully!");
    print_order(&order);
    Ok(())
}

async fn handle_list(api: &ApiClient) -> Result<(), Box<dyn std::error::Error>> {
    let orders = api.get_orders().await?;

    if orders.is_empty() {
        println!("No orders found.");
    } else {
        println!("\n┌─────────────────────────────────────────────────────────────────────────────────┐");
        println!("│                              Inspection Orders                                  │");
        println!("├─────────────────────────────────────────────────────────────────────────────────┤");
        for order in &orders {
            println!(
                "│ ID: {} │",
                order.id
            );
            println!(
                "│ Product: {:<20} Status: {:<20}           │",
                order.product_name,
                order.status.as_str()
            );
            println!(
                "│ Batch: {:<10} Sample: {:<10} AQL: {:<10} Ac: {:<10}     │",
                order.batch_size,
                order.sample_size,
                order.aql_level.as_str(),
                order.ac
            );
            println!(
                "│ Inspected: {}/{} Defects: {} Inspector: {:<20}     │",
                order.inspection_count,
                order.sample_size,
                order.defect_count,
                order.inspector
            );
            if order.is_reinspection {
                println!("│ ⚠ Reinspection from: {}                              │", order.parent_order_id.unwrap());
            }
            println!("├─────────────────────────────────────────────────────────────────────────────────┤");
        }
        println!("│ Total orders: {:<67}│", orders.len());
        println!("└─────────────────────────────────────────────────────────────────────────────────┘");
    }
    Ok(())
}

async fn handle_get(api: &ApiClient, id: Uuid) -> Result<(), Box<dyn std::error::Error>> {
    let order = api.get_order(id).await?;
    print_order(&order);
    Ok(())
}

fn print_order(order: &InspectionOrder) {
    println!("\n┌─────────────────────────────────────────────────────┐");
    println!("│                 Inspection Order                    │");
    println!("├─────────────────────────────────────────────────────┤");
    println!("│  ID: {:<44}│", order.id);
    println!("│  Product: {:<40}│", order.product_name);
    println!("│  Inspector: {:<38}│", order.inspector);
    println!("│  Status: {:<41}│", order.status.as_str());
    println!("├─────────────────────────────────────────────────────┤");
    println!("│  Batch Size: {:<15} Sample Size: {:<14}│", order.batch_size, order.sample_size);
    println!("│  AQL Level: {:<16} Ac (Accept): {:<12}│", order.aql_level.as_str(), order.ac);
    println!("│  Inspected: {}/{}                         Defects: {}         │", order.inspection_count, order.sample_size, order.defect_count);
    println!("├─────────────────────────────────────────────────────┤");
    println!("│  Created: {:<41}│", order.created_at.format("%Y-%m-%d %H:%M:%S"));
    println!("│  Updated: {:<41}│", order.updated_at.format("%Y-%m-%d %H:%M:%S"));
    if order.is_reinspection {
        println!("│  ⚠ Reinspection                                    │");
        println!("│  Parent: {:<39}│", order.parent_order_id.unwrap());
    }
    println!("└─────────────────────────────────────────────────────┘");
}

async fn handle_inspect(api: &ApiClient, id: Uuid) -> Result<(), Box<dyn std::error::Error>> {
    let mut order = api.get_order(id).await?;

    if order.status.as_str() == "pending" {
        order = api.start_inspection(id).await?;
        println!("✓ Inspection started!");
    }

    println!("\nStarting inspection for: {}", order.product_name);
    println!("Sample size: {}, Ac: {}", order.sample_size, order.ac);
    println!(
        "Progress: {}/{}\n",
        order.inspection_count, order.sample_size
    );

    let start_item = order.inspection_count + 1;

    for item_number in start_item..=order.sample_size {
        println!("────────────────────────────────────────");
        println!("Inspecting item {}/{}", item_number, order.sample_size);

        let is_defective = Confirm::with_theme(&ColorfulTheme::default())
            .with_prompt("Is this item defective?")
            .default(false)
            .interact()
            .unwrap();

        let (defect_type, defect_description) = if is_defective {
            let defect_type: String = Input::with_theme(&ColorfulTheme::default())
                .with_prompt("Defect type (e.g., scratch, crack, missing part)")
                .interact_text()
                .unwrap();

            let description: String = Input::with_theme(&ColorfulTheme::default())
                .with_prompt("Defect description")
                .interact_text()
                .unwrap();

            (Some(defect_type), Some(description))
        } else {
            (None, None)
        };

        let request = InspectItemRequest {
            order_id: id,
            item_number,
            is_defective,
            defect_type,
            defect_description,
        };

        match api.inspect_item(request).await {
            Ok(defect) => {
                if let Some(defect) = defect {
                    println!("✓ Defect recorded: {}", defect.defect_type);
                } else {
                    println!("✓ Item passed");
                }
            }
            Err(e) => {
                eprintln!("✗ Error: {}", e);
                return Ok(());
            }
        }
    }

    println!("\n────────────────────────────────────────");
    println!("All items inspected! Completing inspection...");

    let result = api
        .complete_inspection(CompleteInspectionRequest { order_id: id })
        .await?;

    println!("\n================ Inspection Result ================");
    println!("Defects found: {}/{}", result.defect_count, result.sample_size);
    println!("Acceptance criteria (Ac): {}", result.ac);
    println!("================================================");

    if result.accepted {
        println!("\n✓✓✓ BATCH ACCEPTED ✓✓✓");
    } else {
        println!("\n✗✗✗ BATCH REJECTED ✗✗✗");
        if !result.is_final {
            println!("\nHint: You can request reinspection with:");
            println!("  qa-client reinspect -i {}", id);
        }
    }

    if result.is_final {
        println!("\n⚠ This is the final result (reinspection)");
    }

    Ok(())
}

async fn handle_defects(api: &ApiClient, id: Uuid) -> Result<(), Box<dyn std::error::Error>> {
    let defects = api.get_defects(id).await?;

    if defects.is_empty() {
        println!("No defects found for this order.");
    } else {
        println!("\n┌─────────────────────────────────────────────────────────────────────────────────┐");
        println!("│                                Defect Records                                   │");
        println!("├─────────────────────────────────────────────────────────────────────────────────┤");
        for defect in &defects {
            println!("│ Defect ID: {}                           │", defect.id);
            println!("│ Sample Item: {:<69}│", defect.sample_item_number);
            println!("│ Type: {:<72}│", defect.defect_type);
            println!("│ Description: {:<67}│", defect.description);
            println!("│ Recorded: {:<69}│", defect.recorded_at.format("%Y-%m-%d %H:%M:%S"));
            if defect.updated_at != defect.recorded_at {
                println!("│ Updated: {:<70}│", defect.updated_at.format("%Y-%m-%d %H:%M:%S"));
            }
            println!("├─────────────────────────────────────────────────────────────────────────────────┤");
        }
        println!("│ Total defects: {:<66}│", defects.len());
        println!("└─────────────────────────────────────────────────────────────────────────────────┘");
    }
    Ok(())
}

async fn handle_update_defect(
    api: &ApiClient,
    defect_id: Uuid,
    description: Option<String>,
) -> Result<(), Box<dyn std::error::Error>> {
    let new_description = description.unwrap_or_else(|| {
        Input::with_theme(&ColorfulTheme::default())
            .with_prompt("New defect description")
            .interact_text()
            .unwrap()
    });

    let defect = api
        .update_defect(UpdateDefectRequest {
            defect_id,
            new_description,
        })
        .await?;

    println!("\n✓ Defect updated successfully!");
    println!("  Defect ID: {}", defect.id);
    println!("  New Description: {}", defect.description);
    Ok(())
}

async fn handle_complete(api: &ApiClient, id: Uuid) -> Result<(), Box<dyn std::error::Error>> {
    let result = api
        .complete_inspection(CompleteInspectionRequest { order_id: id })
        .await?;

    println!("\n================ Inspection Result ================");
    println!("Defects found: {}/{}", result.defect_count, result.sample_size);
    println!("Acceptance criteria (Ac): {}", result.ac);
    println!("================================================");

    if result.accepted {
        println!("\n✓✓✓ BATCH ACCEPTED ✓✓✓");
    } else {
        println!("\n✗✗✗ BATCH REJECTED ✗✗✗");
    }

    if result.is_final {
        println!("\n⚠ This is the final result (reinspection)");
    }
    Ok(())
}

async fn handle_reinspect(
    api: &ApiClient,
    id: Uuid,
    inspector: Option<String>,
) -> Result<(), Box<dyn std::error::Error>> {
    let inspector_name = inspector.unwrap_or_else(|| {
        Input::with_theme(&ColorfulTheme::default())
            .with_prompt("Inspector name")
            .interact_text()
            .unwrap()
    });

    let order = api
        .request_reinspection(ReinspectionRequest {
            order_id: id,
            inspector: inspector_name,
        })
        .await?;

    println!("\n✓ Reinspection requested!");
    println!("\nNew reinspection order:");
    print_order(&order);
    println!("\n⚠ Note: Reinspection sample size is doubled, but Ac remains the same.");
    println!("⚠ Reinspection result is final and cannot be further appealed.");
    Ok(())
}
