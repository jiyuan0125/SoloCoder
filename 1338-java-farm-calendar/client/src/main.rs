use anyhow::{anyhow, Context, Result};
use clap::{Parser, Subcommand};
use colored::*;
use agri_core as core;
use core::{
    BatchCreateSowingPlanRequest, BatchCreateSowingPlanResult, CreateCropRequest,
    CreateCropStageRequest, CreatePlotRequest, CreateRegionRequest, CreateSowingPlanRequest,
};
use reqwest::Client;
use serde::{de::DeserializeOwned, Deserialize, Serialize};
use std::fmt;
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://127.0.0.1:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Crop {
        #[command(subcommand)]
        sub: CropCommands,
    },
    Region {
        #[command(subcommand)]
        sub: RegionCommands,
    },
    Plot {
        #[command(subcommand)]
        sub: PlotCommands,
    },
    Plan {
        #[command(subcommand)]
        sub: PlanCommands,
    },
    Health,
}

#[derive(Subcommand, Debug)]
enum CropCommands {
    Create {
        #[arg(long)]
        name: String,
        #[arg(long)]
        total_days: u32,
        #[arg(long, value_delimiter = ',')]
        stage_names: Vec<String>,
        #[arg(long, value_delimiter = ',')]
        stage_days: Vec<u32>,
        #[arg(long, default_value = "")]
        description: String,
    },
    List,
    Get {
        #[arg(long)]
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum RegionCommands {
    Create {
        #[arg(long)]
        name: String,
        #[arg(long)]
        avg_temp: f64,
        #[arg(long, value_delimiter = ',')]
        suitable_months: Vec<u32>,
        #[arg(long, default_value = "")]
        description: String,
    },
    List,
    Get {
        #[arg(long)]
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum PlotCommands {
    Create {
        #[arg(long)]
        name: String,
        #[arg(long)]
        region_id: Uuid,
        #[arg(long)]
        area: f64,
        #[arg(long, default_value = "亩")]
        unit: String,
        #[arg(long, default_value = "")]
        description: String,
    },
    List,
    Get {
        #[arg(long)]
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum PlanCommands {
    Create {
        #[arg(long)]
        crop_id: Uuid,
        #[arg(long)]
        plot_id: Uuid,
        #[arg(long)]
        sowing_date: String,
    },
    BatchCreate {
        #[arg(long, value_delimiter = ';')]
        crops: Vec<Uuid>,
        #[arg(long, value_delimiter = ';')]
        plots: Vec<Uuid>,
        #[arg(long, value_delimiter = ';')]
        dates: Vec<String>,
    },
    List,
    Get {
        #[arg(long)]
        id: Uuid,
    },
    ListByPlot {
        #[arg(long)]
        plot_id: Uuid,
    },
}

#[derive(Debug, Deserialize)]
struct ApiError {
    error: String,
    message: String,
}

impl fmt::Display for ApiError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}: {}", self.error.red(), self.message)
    }
}

struct ApiClient {
    client: Client,
    base_url: String,
}

impl ApiClient {
    fn new(base_url: String) -> Self {
        Self {
            client: Client::new(),
            base_url,
        }
    }

    async fn post<T: Serialize, R: DeserializeOwned>(&self, path: &str, body: &T) -> Result<R> {
        let url = format!("{}{}", self.base_url, path);
        let response = self
            .client
            .post(&url)
            .json(body)
            .send()
            .await
            .with_context(|| format!("Failed to send POST request to {}", url))?;

        self.handle_response(response).await
    }

    async fn get<R: DeserializeOwned>(&self, path: &str) -> Result<R> {
        let url = format!("{}{}", self.base_url, path);
        let response = self
            .client
            .get(&url)
            .send()
            .await
            .with_context(|| format!("Failed to send GET request to {}", url))?;

        self.handle_response(response).await
    }

    async fn handle_response<R: DeserializeOwned>(&self, response: reqwest::Response) -> Result<R> {
        let status = response.status();
        if status.is_success() || status == reqwest::StatusCode::MULTI_STATUS {
            response
                .json::<R>()
                .await
                .with_context(|| "Failed to parse response body")
        } else {
            let error: ApiError = response
                .json()
                .await
                .with_context(|| "Failed to parse error response")?;
            Err(anyhow!(error.to_string()))
        }
    }
}

fn print_separator(title: &str) {
    println!("\n{}", "=".repeat(60).bright_blue());
    println!("  {}", title.bright_white().bold());
    println!("{}", "=".repeat(60).bright_blue());
}

fn print_list_header(columns: &[&str]) {
    let line: Vec<String> = columns.iter().map(|c| format!("{:<20}", c)).collect();
    println!("{}", line.join("").cyan());
    println!("{}", "-".repeat(columns.len() * 20).cyan());
}

#[tokio::main]
async fn main() -> Result<()> {
    let args = Args::parse();
    let client = ApiClient::new(args.server_url);

    match args.command {
        Commands::Crop { sub } => handle_crop_commands(&client, sub).await,
        Commands::Region { sub } => handle_region_commands(&client, sub).await,
        Commands::Plot { sub } => handle_plot_commands(&client, sub).await,
        Commands::Plan { sub } => handle_plan_commands(&client, sub).await,
        Commands::Health => handle_health(&client).await,
    }
}

async fn handle_health(client: &ApiClient) -> Result<()> {
    let result: serde_json::Value = client.get("/health").await?;
    println!("{} {}", "Health status:".bold(), result["status"].to_string().green());
    Ok(())
}

async fn handle_crop_commands(client: &ApiClient, cmd: CropCommands) -> Result<()> {
    match cmd {
        CropCommands::Create {
            name,
            total_days,
            stage_names,
            stage_days,
            description,
        } => {
            if stage_names.len() != stage_days.len() {
                return Err(anyhow!("Stage names and days count must match"));
            }

            let stages: Vec<CreateCropStageRequest> = stage_names
                .into_iter()
                .zip(stage_days.into_iter())
                .map(|(n, d)| CreateCropStageRequest {
                    name: n,
                    days: d,
                    description: String::new(),
                })
                .collect();

            let request = CreateCropRequest {
                name,
                total_growth_days: total_days,
                stages,
                description,
            };

            let crop: serde_json::Value = client.post("/crops", &request).await?;
            print_separator("Crop Created");
            println!("{}", serde_json::to_string_pretty(&crop)?);
        }
        CropCommands::List => {
            let crops: Vec<serde_json::Value> = client.get("/crops").await?;
            print_separator("Crops List");
            print_list_header(&["ID", "Name", "Total Days", "Stages"]);
            for crop in &crops {
                let id = crop["id"].as_str().unwrap_or("").to_string();
                let name = crop["name"].as_str().unwrap_or("").to_string();
                let days = crop["total_growth_days"].as_u64().unwrap_or(0);
                let stage_count = crop["stages"].as_array().map(|a| a.len()).unwrap_or(0);
                println!(
                    "{:<20} {:<20} {:<20} {:<20}",
                    id.split('-').next().unwrap_or(""),
                    name,
                    days,
                    stage_count
                );
            }
            if crops.is_empty() {
                println!("{}", "No crops found".yellow());
            }
        }
        CropCommands::Get { id } => {
            let crop: serde_json::Value = client.get(&format!("/crops/{}", id)).await?;
            print_separator("Crop Details");
            println!("{}", serde_json::to_string_pretty(&crop)?);
        }
    }
    Ok(())
}

async fn handle_region_commands(client: &ApiClient, cmd: RegionCommands) -> Result<()> {
    match cmd {
        RegionCommands::Create {
            name,
            avg_temp,
            suitable_months,
            description,
        } => {
            let request = CreateRegionRequest {
                name,
                average_effective_temperature: avg_temp,
                suitable_sowing_months: suitable_months,
                description,
            };

            let region: serde_json::Value = client.post("/regions", &request).await?;
            print_separator("Region Created");
            println!("{}", serde_json::to_string_pretty(&region)?);
        }
        RegionCommands::List => {
            let regions: Vec<serde_json::Value> = client.get("/regions").await?;
            print_separator("Regions List");
            print_list_header(&["ID", "Name", "Avg Temp", "Suitable Months"]);
            for region in &regions {
                let id = region["id"].as_str().unwrap_or("").to_string();
                let name = region["name"].as_str().unwrap_or("").to_string();
                let temp = region["average_effective_temperature"].as_f64().unwrap_or(0.0);
                let months = region["suitable_sowing_months"]
                    .as_array()
                    .map(|a| {
                        a.iter()
                            .map(|m| m.as_u64().unwrap_or(0).to_string())
                            .collect::<Vec<_>>()
                            .join(",")
                    })
                    .unwrap_or_default();
                println!(
                    "{:<20} {:<20} {:<20} {:<20}",
                    id.split('-').next().unwrap_or(""),
                    name,
                    format!("{:.1}", temp),
                    months
                );
            }
            if regions.is_empty() {
                println!("{}", "No regions found".yellow());
            }
        }
        RegionCommands::Get { id } => {
            let region: serde_json::Value = client.get(&format!("/regions/{}", id)).await?;
            print_separator("Region Details");
            println!("{}", serde_json::to_string_pretty(&region)?);
        }
    }
    Ok(())
}

async fn handle_plot_commands(client: &ApiClient, cmd: PlotCommands) -> Result<()> {
    match cmd {
        PlotCommands::Create {
            name,
            region_id,
            area,
            unit,
            description,
        } => {
            let request = CreatePlotRequest {
                name,
                region_id,
                area,
                unit,
                description,
            };

            let plot: serde_json::Value = client.post("/plots", &request).await?;
            print_separator("Plot Created");
            println!("{}", serde_json::to_string_pretty(&plot)?);
        }
        PlotCommands::List => {
            let plots: Vec<serde_json::Value> = client.get("/plots").await?;
            print_separator("Plots List");
            print_list_header(&["ID", "Name", "Area", "Unit"]);
            for plot in &plots {
                let id = plot["id"].as_str().unwrap_or("").to_string();
                let name = plot["name"].as_str().unwrap_or("").to_string();
                let area = plot["area"].as_f64().unwrap_or(0.0);
                let unit = plot["unit"].as_str().unwrap_or("").to_string();
                println!(
                    "{:<20} {:<20} {:<20} {:<20}",
                    id.split('-').next().unwrap_or(""),
                    name,
                    format!("{:.1}", area),
                    unit
                );
            }
            if plots.is_empty() {
                println!("{}", "No plots found".yellow());
            }
        }
        PlotCommands::Get { id } => {
            let plot: serde_json::Value = client.get(&format!("/plots/{}", id)).await?;
            print_separator("Plot Details");
            println!("{}", serde_json::to_string_pretty(&plot)?);
        }
    }
    Ok(())
}

async fn handle_plan_commands(client: &ApiClient, cmd: PlanCommands) -> Result<()> {
    match cmd {
        PlanCommands::Create {
            crop_id,
            plot_id,
            sowing_date,
        } => {
            let request = CreateSowingPlanRequest {
                crop_id,
                plot_id,
                sowing_date,
            };

            let plan: serde_json::Value = client.post("/plans", &request).await?;
            print_separator("Sowing Plan Created");
            print_plan_details(&plan);
        }
        PlanCommands::BatchCreate { crops, plots, dates } => {
            if crops.len() != plots.len() || crops.len() != dates.len() {
                return Err(anyhow!(
                    "Number of crops, plots, and dates must match ({} crops, {} plots, {} dates)",
                    crops.len(),
                    plots.len(),
                    dates.len()
                ));
            }

            let plans: Vec<CreateSowingPlanRequest> = crops
                .into_iter()
                .zip(plots.into_iter())
                .zip(dates.into_iter())
                .map(|((c, p), d)| CreateSowingPlanRequest {
                    crop_id: c,
                    plot_id: p,
                    sowing_date: d,
                })
                .collect();

            let request = BatchCreateSowingPlanRequest { plans };
            let result: BatchCreateSowingPlanResult = client.post("/plans/batch", &request).await?;

            print_separator("Batch Create Results");
            if result.has_conflicts {
                println!("{}", "Warning: Some plans have conflicts!".yellow().bold());
            }

            for (idx, (_key, plan_result)) in result.results.iter().enumerate() {
                println!("\n{}", format!("--- Plan {} ---", idx + 1).bold());
                if let Some(plan) = &plan_result.plan {
                    println!("{}", "✓ Successfully created".green());
                    let plan_json = serde_json::to_value(plan)?;
                    print_plan_details(&plan_json);
                }
                if !plan_result.errors.is_empty() {
                    println!("{}", "✗ Errors:".red());
                    for error in &plan_result.errors {
                        println!("  - {}", error.red());
                    }
                }
            }
        }
        PlanCommands::List => {
            let plans: Vec<serde_json::Value> = client.get("/plans").await?;
            print_separator("Sowing Plans List");
            print_list_header(&["ID", "Sowing Date", "Start", "End", "Warnings"]);
            for plan in &plans {
                let id = plan["id"].as_str().unwrap_or("").to_string();
                let sowing = plan["sowing_date"].as_str().unwrap_or("").to_string();
                let start = plan["start_date"].as_str().unwrap_or("").to_string();
                let end = plan["end_date"].as_str().unwrap_or("").to_string();
                let warnings = plan["warnings"]
                    .as_array()
                    .map(|a| a.len())
                    .unwrap_or(0);
                println!(
                    "{:<20} {:<20} {:<20} {:<20} {:<20}",
                    id.split('-').next().unwrap_or(""),
                    sowing,
                    start,
                    end,
                    if warnings > 0 {
                        format!("{} warnings", warnings).yellow().to_string()
                    } else {
                        "None".to_string()
                    }
                );
            }
            if plans.is_empty() {
                println!("{}", "No plans found".yellow());
            }
        }
        PlanCommands::Get { id } => {
            let plan: serde_json::Value = client.get(&format!("/plans/{}", id)).await?;
            print_separator("Sowing Plan Details");
            print_plan_details(&plan);
        }
        PlanCommands::ListByPlot { plot_id } => {
            let plans: Vec<serde_json::Value> =
                client.get(&format!("/plans/plot/{}", plot_id)).await?;
            print_separator(&format!("Plans for Plot {}", plot_id));
            for plan in &plans {
                print_plan_details(plan);
                println!();
            }
            if plans.is_empty() {
                println!("{}", "No plans found for this plot".yellow());
            }
        }
    }
    Ok(())
}

fn print_plan_details(plan: &serde_json::Value) {
    println!("{} {}", "Plan ID:".bold(), plan["id"]);
    println!("{} {}", "Crop ID:".bold(), plan["crop_id"]);
    println!("{} {}", "Plot ID:".bold(), plan["plot_id"]);
    println!("{} {}", "Sowing Date:".bold(), plan["sowing_date"]);
    println!("{} {} - {}", "Growth Period:".bold(), plan["start_date"], plan["end_date"]);

    let warnings = plan["warnings"].as_array();
    if let Some(warnings) = warnings {
        if !warnings.is_empty() {
            println!("\n{}", "Warnings:".yellow().bold());
            for w in warnings {
                println!("  ! {}", w.as_str().unwrap_or("").yellow());
            }
        }
    }

    let stages = plan["stage_schedules"].as_array();
    if let Some(stages) = stages {
        println!("\n{}", "Stage Schedules:".cyan().bold());
        for stage in stages {
            println!(
                "  {}: {} - {} ({} days)",
                stage["stage_name"].as_str().unwrap_or("").bold(),
                stage["start_date"],
                stage["end_date"],
                stage["days"]
            );
            println!("    Advice: {}", stage["advice"].as_str().unwrap_or(""));
        }
    }
}
