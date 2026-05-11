use clap::{Parser, Subcommand};
use rental_core::{
    CheckoutRequest, CheckoutRecord, Contract, CreateContractRequest, CreatePropertyRequest,
    CreateTenantRequest, Property, RenewContractRequest, Tenant,
};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(long, default_value = "http://localhost:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Property {
        #[command(subcommand)]
        subcommand: PropertyCommands,
    },
    Tenant {
        #[command(subcommand)]
        subcommand: TenantCommands,
    },
    Contract {
        #[command(subcommand)]
        subcommand: ContractCommands,
    },
}

#[derive(Subcommand, Debug)]
enum PropertyCommands {
    List,
    Get {
        #[arg(long)]
        id: Uuid,
    },
    Create {
        #[arg(long)]
        property_no: String,
        #[arg(long)]
        area: f64,
        #[arg(long)]
        monthly_rent: f64,
    },
}

#[derive(Subcommand, Debug)]
enum TenantCommands {
    List,
    Get {
        #[arg(long)]
        id: Uuid,
    },
    Create {
        #[arg(long)]
        name: String,
        #[arg(long)]
        phone: String,
    },
}

#[derive(Subcommand, Debug)]
enum ContractCommands {
    List,
    Get {
        #[arg(long)]
        id: Uuid,
    },
    Create {
        #[arg(long)]
        tenant_id: Uuid,
        #[arg(long)]
        property_id: Uuid,
        #[arg(long)]
        start_date: String,
        #[arg(long)]
        end_date: String,
        #[arg(long)]
        monthly_rent: Option<f64>,
        #[arg(long)]
        deposit_amount: f64,
    },
    Expiring {
        #[arg(long, default_value_t = 30)]
        days: i64,
    },
    Renew {
        #[arg(long)]
        contract_id: Uuid,
        #[arg(long)]
        new_monthly_rent: Option<f64>,
        #[arg(long)]
        new_deposit_amount: Option<f64>,
        #[arg(long)]
        extend_years: Option<i32>,
    },
    Checkout {
        #[arg(long)]
        contract_id: Uuid,
        #[arg(long)]
        actual_end_date: Option<String>,
        #[arg(long)]
        damage_cost: f64,
    },
    CheckoutRecord {
        #[arg(long)]
        contract_id: Uuid,
    },
}

#[derive(Debug, Serialize, Deserialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    error: Option<String>,
}

struct RentalClient {
    base_url: String,
    client: reqwest::Client,
}

impl RentalClient {
    fn new(base_url: String) -> Self {
        Self {
            base_url,
            client: reqwest::Client::new(),
        }
    }

    async fn get<T: for<'de> Deserialize<'de>>(&self, path: &str) -> Result<T, String> {
        let url = format!("{}{}", self.base_url, path);
        let response = self
            .client
            .get(&url)
            .send()
            .await
            .map_err(|e| format!("Request failed: {}", e))?;
        
        let api_response: ApiResponse<T> = response
            .json()
            .await
            .map_err(|e| format!("Failed to parse response: {}", e))?;
        
        if api_response.success {
            api_response.data.ok_or_else(|| "No data returned".to_string())
        } else {
            Err(api_response.error.unwrap_or_else(|| "Unknown error".to_string()))
        }
    }

    async fn post<T: Serialize, R: for<'de> Deserialize<'de>>(
        &self,
        path: &str,
        body: &T,
    ) -> Result<R, String> {
        let url = format!("{}{}", self.base_url, path);
        let response = self
            .client
            .post(&url)
            .json(body)
            .send()
            .await
            .map_err(|e| format!("Request failed: {}", e))?;
        
        let api_response: ApiResponse<R> = response
            .json()
            .await
            .map_err(|e| format!("Failed to parse response: {}", e))?;
        
        if api_response.success {
            api_response.data.ok_or_else(|| "No data returned".to_string())
        } else {
            Err(api_response.error.unwrap_or_else(|| "Unknown error".to_string()))
        }
    }
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    let client = RentalClient::new(cli.server);

    let result = match cli.command {
        Commands::Property { subcommand } => handle_property(&client, subcommand).await,
        Commands::Tenant { subcommand } => handle_tenant(&client, subcommand).await,
        Commands::Contract { subcommand } => handle_contract(&client, subcommand).await,
    };

    if let Err(e) = result {
        eprintln!("Error: {}", e);
        std::process::exit(1);
    }
}

async fn handle_property(client: &RentalClient, subcommand: PropertyCommands) -> Result<(), String> {
    match subcommand {
        PropertyCommands::List => {
            let properties: Vec<Property> = client.get("/properties").await?;
            if properties.is_empty() {
                println!("No properties found");
            } else {
                    for p in properties {
                    print_property(&p);
                }
            }
        }
        PropertyCommands::Get { id } => {
            let property: Property = client.get(&format!("/properties/{}", id)).await?;
            print_property(&property);
        }
        PropertyCommands::Create {
            property_no,
            area,
            monthly_rent,
        } => {
            let req = CreatePropertyRequest {
                property_no,
                area,
                monthly_rent,
            };
            let property: Property = client.post("/properties", &req).await?;
            println!("Created property:");
            print_property(&property);
        }
    }
    Ok(())
}

async fn handle_tenant(client: &RentalClient, subcommand: TenantCommands) -> Result<(), String> {
    match subcommand {
        TenantCommands::List => {
            let tenants: Vec<Tenant> = client.get("/tenants").await?;
            if tenants.is_empty() {
                println!("No tenants found");
            } else {
                for t in tenants {
                    print_tenant(&t);
                }
            }
        }
        TenantCommands::Get { id } => {
            let tenant: Tenant = client.get(&format!("/tenants/{}", id)).await?;
            print_tenant(&tenant);
        }
        TenantCommands::Create { name, phone } => {
            let req = CreateTenantRequest { name, phone };
            let tenant: Tenant = client.post("/tenants", &req).await?;
            println!("Created tenant:");
            print_tenant(&tenant);
        }
    }
    Ok(())
}

async fn handle_contract(client: &RentalClient, subcommand: ContractCommands) -> Result<(), String> {
    match subcommand {
        ContractCommands::List => {
            let contracts: Vec<Contract> = client.get("/contracts").await?;
            if contracts.is_empty() {
                println!("No contracts found");
            } else {
                for c in contracts {
                    print_contract(&c);
                }
            }
        }
        ContractCommands::Get { id } => {
            let contract: Contract = client.get(&format!("/contracts/{}", id)).await?;
            print_contract(&contract);
        }
        ContractCommands::Create {
            tenant_id,
            property_id,
            start_date,
            end_date,
            monthly_rent,
            deposit_amount,
        } => {
            let start_date = chrono::NaiveDate::parse_from_str(&start_date, "%Y-%m-%d")
                .map_err(|e| format!("Invalid start date: {}", e))?;
            let end_date = chrono::NaiveDate::parse_from_str(&end_date, "%Y-%m-%d")
                .map_err(|e| format!("Invalid end date: {}", e))?;

            let req = CreateContractRequest {
                tenant_id,
                property_id,
                start_date,
                end_date,
                monthly_rent,
                deposit_amount,
            };
            let contract: Contract = client.post("/contracts", &req).await?;
            println!("Created contract:");
            print_contract(&contract);
        }
        ContractCommands::Expiring { days } => {
            let contracts: Vec<Contract> = client.get(&format!("/contracts/expiring?days={}", days)).await?;
            if contracts.is_empty() {
                println!("No expiring contracts in next {} days", days);
            } else {
                println!("Contracts expiring in next {} days:", days);
                for c in contracts {
                    print_contract(&c);
                }
            }
        }
        ContractCommands::Renew {
            contract_id,
            new_monthly_rent,
            new_deposit_amount,
            extend_years,
        } => {
            let req = RenewContractRequest {
                contract_id,
                new_monthly_rent,
                new_deposit_amount,
                extend_years,
            };
            let contract: Contract = client.post("/contracts/renew", &req).await?;
            println!("Renewed contract:");
            print_contract(&contract);
        }
        ContractCommands::Checkout {
            contract_id,
            actual_end_date,
            damage_cost,
        } => {
            let actual_end_date = match actual_end_date {
                Some(s) => Some(
                    chrono::NaiveDate::parse_from_str(&s, "%Y-%m-%d")
                        .map_err(|e| format!("Invalid actual end date: {}", e))?,
                ),
                None => None,
            };

            let req = CheckoutRequest {
                contract_id,
                actual_end_date,
                damage_cost,
            };
            let record: CheckoutRecord = client.post("/contracts/checkout", &req).await?;
            println!("Checkout completed:");
            print_checkout_record(&record);
        }
        ContractCommands::CheckoutRecord { contract_id } => {
            let record: CheckoutRecord = client
                .get(&format!("/checkout/{}", contract_id))
                .await?;
            print_checkout_record(&record);
        }
    }
    Ok(())
}

fn print_property(p: &Property) {
    println!("Property:");
    println!("  ID: {}", p.id);
    println!("  Number: {}", p.property_no);
    println!("  Area: {:.2} sqm", p.area);
    println!("  Monthly Rent: ${:.2}", p.monthly_rent);
    println!();
}

fn print_tenant(t: &Tenant) {
    println!("Tenant:");
    println!("  ID: {}", t.id);
    println!("  Name: {}", t.name);
    println!("  Phone: {}", t.phone);
    println!();
}

fn print_contract(c: &Contract) {
    println!("Contract:");
    println!("  ID: {}", c.id);
    println!("  Tenant ID: {}", c.tenant_id);
    println!("  Property ID: {}", c.property_id);
    println!("  Start Date: {}", c.start_date);
    println!("  End Date: {}", c.end_date);
    println!("  Monthly Rent: ${:.2}", c.monthly_rent);
    println!("  Deposit: ${:.2}", c.deposit_amount);
    println!("  Status: {:?}", c.status);
    if let Some(original) = c.original_contract_id {
        println!("  Original Contract: {}", original);
    }
    println!();
}

fn print_checkout_record(r: &CheckoutRecord) {
    println!("Checkout Record:");
    println!("  ID: {}", r.id);
    println!("  Contract ID: {}", r.contract_id);
    println!("  Actual End Date: {}", r.actual_end_date);
    println!("  Damage Cost: ${:.2}", r.damage_cost);
    println!("  Refund Amount: ${:.2}", r.refund_amount);
    println!("  Outstanding Debt: ${:.2}", r.outstanding_debt);
    println!();
}
