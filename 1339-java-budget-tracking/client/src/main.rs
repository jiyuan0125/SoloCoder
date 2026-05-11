use clap::{Parser, Subcommand};
use reqwest::Client;
use uuid::Uuid;

use budget_tracking_core::{
    AlertConfig, BudgetCategory, BudgetCategoryTree, CreateBudgetCategoryRequest,
    CreateExpenseRequest, ExecutionSummary, Expense, UpdateBudgetCategoryRequest,
};

#[derive(Parser, Debug)]
#[command(version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "BUDGET_SERVER_URL", default_value = "http://localhost:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Category {
        #[command(subcommand)]
        action: CategoryCommands,
    },
    Expense {
        #[command(subcommand)]
        action: ExpenseCommands,
    },
    Alert {
        #[command(subcommand)]
        action: AlertCommands,
    },
    Execution {
        #[command(subcommand)]
        action: ExecutionCommands,
    },
}

#[derive(Subcommand, Debug)]
enum CategoryCommands {
    Create {
        #[arg(short, long)]
        name: String,
        #[arg(short, long)]
        parent: Option<Uuid>,
        #[arg(short, long)]
        budget: Option<f64>,
    },
    List,
    Tree,
    Update {
        #[arg(short, long)]
        id: Uuid,
        #[arg(short, long)]
        name: Option<String>,
        #[arg(short, long)]
        budget: Option<f64>,
    },
    Delete {
        #[arg(short, long)]
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum ExpenseCommands {
    Create {
        #[arg(short, long)]
        category: Uuid,
        #[arg(short, long)]
        amount: f64,
        #[arg(short, long)]
        date: Option<String>,
        #[arg(short, long)]
        description: String,
    },
    List,
}

#[derive(Subcommand, Debug)]
enum AlertCommands {
    Get,
    Set {
        #[arg(short, long)]
        warning: f64,
        #[arg(short, long)]
        critical: f64,
    },
}

#[derive(Subcommand, Debug)]
enum ExecutionCommands {
    Summary {
        #[arg(short, long)]
        category: Option<Uuid>,
        #[arg(long)]
        start: Option<String>,
        #[arg(long)]
        end: Option<String>,
    },
}

struct ApiClient {
    base_url: String,
    client: Client,
}

impl ApiClient {
    fn new(base_url: String) -> Self {
        ApiClient {
            base_url,
            client: Client::new(),
        }
    }

    async fn create_category(
        &self,
        req: CreateBudgetCategoryRequest,
    ) -> Result<BudgetCategory, Box<dyn std::error::Error>> {
        let url = format!("{}/api/categories", self.base_url);
        let res = self.client.post(&url).json(&req).send().await?;
        let status = res.status();
        if !status.is_success() {
            let body: serde_json::Value = res.json().await?;
            return Err(format!("API Error ({}): {}", status, body).into());
        }
        Ok(res.json().await?)
    }

    async fn list_categories(&self) -> Result<Vec<BudgetCategory>, Box<dyn std::error::Error>> {
        let url = format!("{}/api/categories", self.base_url);
        let res = self.client.get(&url).send().await?;
        let status = res.status();
        if !status.is_success() {
            let body: serde_json::Value = res.json().await?;
            return Err(format!("API Error ({}): {}", status, body).into());
        }
        Ok(res.json().await?)
    }

    async fn get_category_tree(
        &self,
    ) -> Result<Vec<BudgetCategoryTree>, Box<dyn std::error::Error>> {
        let url = format!("{}/api/categories/tree", self.base_url);
        let res = self.client.get(&url).send().await?;
        let status = res.status();
        if !status.is_success() {
            let body: serde_json::Value = res.json().await?;
            return Err(format!("API Error ({}): {}", status, body).into());
        }
        Ok(res.json().await?)
    }

    async fn update_category(
        &self,
        id: Uuid,
        req: UpdateBudgetCategoryRequest,
    ) -> Result<BudgetCategory, Box<dyn std::error::Error>> {
        let url = format!("{}/api/categories/{}", self.base_url, id);
        let res = self.client.put(&url).json(&req).send().await?;
        let status = res.status();
        if !status.is_success() {
            let body: serde_json::Value = res.json().await?;
            return Err(format!("API Error ({}): {}", status, body).into());
        }
        Ok(res.json().await?)
    }

    async fn delete_category(&self, id: Uuid) -> Result<(), Box<dyn std::error::Error>> {
        let url = format!("{}/api/categories/{}", self.base_url, id);
        let res = self.client.delete(&url).send().await?;
        let status = res.status();
        if !status.is_success() {
            let body: serde_json::Value = res.json().await?;
            return Err(format!("API Error ({}): {}", status, body).into());
        }
        Ok(())
    }

    async fn create_expense(
        &self,
        req: CreateExpenseRequest,
    ) -> Result<Expense, Box<dyn std::error::Error>> {
        let url = format!("{}/api/expenses", self.base_url);
        let res = self.client.post(&url).json(&req).send().await?;
        let status = res.status();
        if !status.is_success() {
            let body: serde_json::Value = res.json().await?;
            return Err(format!("API Error ({}): {}", status, body).into());
        }
        Ok(res.json().await?)
    }

    async fn list_expenses(&self) -> Result<Vec<Expense>, Box<dyn std::error::Error>> {
        let url = format!("{}/api/expenses", self.base_url);
        let res = self.client.get(&url).send().await?;
        let status = res.status();
        if !status.is_success() {
            let body: serde_json::Value = res.json().await?;
            return Err(format!("API Error ({}): {}", status, body).into());
        }
        Ok(res.json().await?)
    }

    async fn get_alert_config(&self) -> Result<AlertConfig, Box<dyn std::error::Error>> {
        let url = format!("{}/api/alert-config", self.base_url);
        let res = self.client.get(&url).send().await?;
        let status = res.status();
        if !status.is_success() {
            let body: serde_json::Value = res.json().await?;
            return Err(format!("API Error ({}): {}", status, body).into());
        }
        Ok(res.json().await?)
    }

    async fn set_alert_config(
        &self,
        config: AlertConfig,
    ) -> Result<(), Box<dyn std::error::Error>> {
        let url = format!("{}/api/alert-config", self.base_url);
        let res = self.client.put(&url).json(&config).send().await?;
        let status = res.status();
        if !status.is_success() {
            let body: serde_json::Value = res.json().await?;
            return Err(format!("API Error ({}): {}", status, body).into());
        }
        Ok(())
    }

    async fn get_execution_summary(
        &self,
        category_id: Option<Uuid>,
        start_date: Option<String>,
        end_date: Option<String>,
    ) -> Result<Vec<ExecutionSummary>, Box<dyn std::error::Error>> {
        let mut url = format!("{}/api/execution", self.base_url);
        let mut first = true;

        if let Some(id) = category_id {
            url.push_str(&format!("?category_id={}", id));
            first = false;
        }
        if let Some(start) = start_date {
            let prefix = if first { "?" } else { "&" };
            url.push_str(&format!("{}start_date={}", prefix, start));
            first = false;
        }
        if let Some(end) = end_date {
            let prefix = if first { "?" } else { "&" };
            url.push_str(&format!("{}end_date={}", prefix, end));
        }

        let res = self.client.get(&url).send().await?;
        let status = res.status();
        if !status.is_success() {
            let body: serde_json::Value = res.json().await?;
            return Err(format!("API Error ({}): {}", status, body).into());
        }
        Ok(res.json().await?)
    }
}

fn print_category_tree(nodes: &[BudgetCategoryTree], indent: usize) {
    for node in nodes {
        let prefix = "  ".repeat(indent);
        let budget_info = if node.children.is_empty() {
            format!(" (预算: {:.2})", node.computed_budget)
        } else {
            format!(" (汇总预算: {:.2})", node.computed_budget)
        };
        println!("{}{}{}", prefix, node.name, budget_info);
        print_category_tree(&node.children, indent + 1);
    }
}

fn alert_level_to_string(level: budget_tracking_core::AlertLevel) -> &'static str {
    use budget_tracking_core::AlertLevel::*;
    match level {
        None => "正常",
        Warning => "⚠️ 预警",
        Critical => "🚨 严重预警",
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Args::parse();
    let client = ApiClient::new(args.server);

    match args.command {
        Commands::Category { action } => match action {
            CategoryCommands::Create {
                name,
                parent,
                budget,
            } => {
                let req = CreateBudgetCategoryRequest {
                    name,
                    parent_id: parent,
                    budget_amount: budget,
                };
                let category = client.create_category(req).await?;
                println!("创建成功:");
                println!("{}", serde_json::to_string_pretty(&category)?);
            }
            CategoryCommands::List => {
                let categories = client.list_categories().await?;
                println!("{}", serde_json::to_string_pretty(&categories)?);
            }
            CategoryCommands::Tree => {
                let tree = client.get_category_tree().await?;
                println!("预算科目树:");
                print_category_tree(&tree, 0);
            }
            CategoryCommands::Update { id, name, budget } => {
                let req = UpdateBudgetCategoryRequest {
                    name,
                    budget_amount: budget,
                };
                let category = client.update_category(id, req).await?;
                println!("更新成功:");
                println!("{}", serde_json::to_string_pretty(&category)?);
            }
            CategoryCommands::Delete { id } => {
                client.delete_category(id).await?;
                println!("删除成功");
            }
        },
        Commands::Expense { action } => match action {
            ExpenseCommands::Create {
                category,
                amount,
                date,
                description,
            } => {
                let date = if let Some(d) = date {
                    d.parse()?
                } else {
                    chrono::Utc::now()
                };
                let req = CreateExpenseRequest {
                    category_id: category,
                    amount,
                    date,
                    description,
                };
                let expense = client.create_expense(req).await?;
                println!("支出记录成功:");
                println!("{}", serde_json::to_string_pretty(&expense)?);
            }
            ExpenseCommands::List => {
                let expenses = client.list_expenses().await?;
                println!("{}", serde_json::to_string_pretty(&expenses)?);
            }
        },
        Commands::Alert { action } => match action {
            AlertCommands::Get => {
                let config = client.get_alert_config().await?;
                println!("当前预警配置:");
                println!("  预警阈值: {}%", config.warning_threshold);
                println!("  严重预警阈值: {}%", config.critical_threshold);
            }
            AlertCommands::Set { warning, critical } => {
                let config = AlertConfig {
                    warning_threshold: warning,
                    critical_threshold: critical,
                };
                client.set_alert_config(config).await?;
                println!("预警配置已更新");
            }
        },
        Commands::Execution { action } => match action {
            ExecutionCommands::Summary {
                category,
                start,
                end,
            } => {
                let summaries = client
                    .get_execution_summary(category, start, end)
                    .await?;
                println!("执行情况汇总:");
                println!("{:-<80}", "");
                for s in summaries {
                    println!("科目: {}", s.category_name);
                    println!("  预算金额: {:.2}", s.budget_amount);
                    println!("  实际支出: {:.2}", s.actual_spent);
                    println!("  执行率: {:.2}%", s.execution_rate);
                    println!("  状态: {}", alert_level_to_string(s.alert_level));
                    println!("{:-<80}", "");
                }
            }
        },
    }

    Ok(())
}
