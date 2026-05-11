use anyhow::{Context, Result};
use clap::{Parser, Subcommand};
use process_route_core::models::*;
use reqwest::Client;
use serde::{Deserialize, Serialize};

#[derive(Parser, Debug)]
#[command(author, version, about = "工艺路线管理系统命令行客户端", long_about = None)]
struct Cli {
    #[arg(long, env = "PROCESS_ROUTE_SERVER", default_value = "http://localhost:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Route {
        #[command(subcommand)]
        action: RouteCommands,
    },
    Task {
        #[command(subcommand)]
        action: TaskCommands,
    },
    Demo,
}

#[derive(Subcommand, Debug)]
enum RouteCommands {
    Create {
        name: String,
    },
    List,
    Get {
        id: String,
    },
    Delete {
        id: String,
    },
    AddVersion {
        route_id: String,
        version: String,
        #[arg(long, value_parser = parse_processes_json)]
        processes: Vec<ProcessDefinition>,
    },
    CriticalPath {
        route_id: String,
        version: String,
    },
    Compare {
        route_id: String,
        v1: String,
        v2: String,
    },
}

#[derive(Subcommand, Debug)]
enum TaskCommands {
    Create {
        name: String,
        route_id: String,
        #[arg(long)]
        version: Option<String>,
    },
    List,
    Get {
        id: String,
    },
    Delete {
        id: String,
    },
    Available {
        task_id: String,
    },
    Start {
        task_id: String,
        process_id: String,
    },
    Complete {
        task_id: String,
        process_id: String,
    },
    Pause {
        task_id: String,
        process_id: String,
    },
    Resume {
        task_id: String,
        process_id: String,
    },
}

fn parse_processes_json(s: &str) -> Result<Vec<ProcessDefinition>, String> {
    serde_json::from_str(s).map_err(|e| e.to_string())
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateRouteRequest {
    name: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct AddVersionRequest {
    version: String,
    processes: Vec<ProcessDefinition>,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateTaskRequest {
    name: String,
    route_id: String,
    version: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
struct ProcessActionResponse {
    task: ProductionTask,
    warning: Option<Warning>,
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

    async fn create_route(&self, name: String) -> Result<ProcessRoute> {
        let url = format!("{}/routes", self.base_url);
        let resp = self
            .client
            .post(&url)
            .json(&CreateRouteRequest { name })
            .send()
            .await
            .context("请求失败")?;
        let status = resp.status();
        if !status.is_success() {
            let body: serde_json::Value = resp.json().await.unwrap_or_default();
            anyhow::bail!("API 错误 {}: {}", status, body);
        }
        let route: ProcessRoute = resp.json().await.context("解析响应失败")?;
        Ok(route)
    }

    async fn list_routes(&self) -> Result<Vec<ProcessRoute>> {
        let url = format!("{}/routes", self.base_url);
        let resp = self.client.get(&url).send().await.context("请求失败")?;
        let routes: Vec<ProcessRoute> = resp.json().await.context("解析响应失败")?;
        Ok(routes)
    }

    async fn get_route(&self, id: &str) -> Result<ProcessRoute> {
        let url = format!("{}/routes/{}", self.base_url, id);
        let resp = self.client.get(&url).send().await.context("请求失败")?;
        let status = resp.status();
        if !status.is_success() {
            let body: serde_json::Value = resp.json().await.unwrap_or_default();
            anyhow::bail!("API 错误 {}: {}", status, body);
        }
        let route: ProcessRoute = resp.json().await.context("解析响应失败")?;
        Ok(route)
    }

    async fn delete_route(&self, id: &str) -> Result<()> {
        let url = format!("{}/routes/{}", self.base_url, id);
        let resp = self.client.delete(&url).send().await.context("请求失败")?;
        let status = resp.status();
        if !status.is_success() {
            let body: serde_json::Value = resp.json().await.unwrap_or_default();
            anyhow::bail!("API 错误 {}: {}", status, body);
        }
        Ok(())
    }

    async fn add_version(
        &self,
        route_id: &str,
        version: String,
        processes: Vec<ProcessDefinition>,
    ) -> Result<ProcessRoute> {
        let url = format!("{}/routes/{}/versions", self.base_url, route_id);
        let resp = self
            .client
            .post(&url)
            .json(&AddVersionRequest { version, processes })
            .send()
            .await
            .context("请求失败")?;
        let status = resp.status();
        if !status.is_success() {
            let body: serde_json::Value = resp.json().await.unwrap_or_default();
            anyhow::bail!("API 错误 {}: {}", status, body);
        }
        let route: ProcessRoute = resp.json().await.context("解析响应失败")?;
        Ok(route)
    }

    async fn get_critical_path(&self, route_id: &str, version: &str) -> Result<CriticalPathInfo> {
        let url = format!(
            "{}/routes/{}/versions/{}/critical-path",
            self.base_url, route_id, version
        );
        let resp = self.client.get(&url).send().await.context("请求失败")?;
        let status = resp.status();
        if !status.is_success() {
            let body: serde_json::Value = resp.json().await.unwrap_or_default();
            anyhow::bail!("API 错误 {}: {}", status, body);
        }
        let info: CriticalPathInfo = resp.json().await.context("解析响应失败")?;
        Ok(info)
    }

    async fn compare_versions(
        &self,
        route_id: &str,
        v1: &str,
        v2: &str,
    ) -> Result<VersionComparison> {
        let url = format!("{}/routes/{}/compare/{}/{}", self.base_url, route_id, v1, v2);
        let resp = self.client.get(&url).send().await.context("请求失败")?;
        let status = resp.status();
        if !status.is_success() {
            let body: serde_json::Value = resp.json().await.unwrap_or_default();
            anyhow::bail!("API 错误 {}: {}", status, body);
        }
        let comparison: VersionComparison = resp.json().await.context("解析响应失败")?;
        Ok(comparison)
    }

    async fn create_task(
        &self,
        name: String,
        route_id: String,
        version: Option<String>,
    ) -> Result<ProductionTask> {
        let url = format!("{}/tasks", self.base_url);
        let resp = self
            .client
            .post(&url)
            .json(&CreateTaskRequest {
                name,
                route_id,
                version,
            })
            .send()
            .await
            .context("请求失败")?;
        let status = resp.status();
        if !status.is_success() {
            let body: serde_json::Value = resp.json().await.unwrap_or_default();
            anyhow::bail!("API 错误 {}: {}", status, body);
        }
        let task: ProductionTask = resp.json().await.context("解析响应失败")?;
        Ok(task)
    }

    async fn list_tasks(&self) -> Result<Vec<ProductionTask>> {
        let url = format!("{}/tasks", self.base_url);
        let resp = self.client.get(&url).send().await.context("请求失败")?;
        let tasks: Vec<ProductionTask> = resp.json().await.context("解析响应失败")?;
        Ok(tasks)
    }

    async fn get_task(&self, id: &str) -> Result<ProductionTask> {
        let url = format!("{}/tasks/{}", self.base_url, id);
        let resp = self.client.get(&url).send().await.context("请求失败")?;
        let status = resp.status();
        if !status.is_success() {
            let body: serde_json::Value = resp.json().await.unwrap_or_default();
            anyhow::bail!("API 错误 {}: {}", status, body);
        }
        let task: ProductionTask = resp.json().await.context("解析响应失败")?;
        Ok(task)
    }

    async fn delete_task(&self, id: &str) -> Result<()> {
        let url = format!("{}/tasks/{}", self.base_url, id);
        let resp = self.client.delete(&url).send().await.context("请求失败")?;
        let status = resp.status();
        if !status.is_success() {
            let body: serde_json::Value = resp.json().await.unwrap_or_default();
            anyhow::bail!("API 错误 {}: {}", status, body);
        }
        Ok(())
    }

    async fn get_available_processes(&self, task_id: &str) -> Result<Vec<ProcessInstance>> {
        let url = format!("{}/tasks/{}/available", self.base_url, task_id);
        let resp = self.client.get(&url).send().await.context("请求失败")?;
        let status = resp.status();
        if !status.is_success() {
            let body: serde_json::Value = resp.json().await.unwrap_or_default();
            anyhow::bail!("API 错误 {}: {}", status, body);
        }
        let processes: Vec<ProcessInstance> = resp.json().await.context("解析响应失败")?;
        Ok(processes)
    }

    async fn process_action(
        &self,
        task_id: &str,
        process_id: &str,
        action: &str,
    ) -> Result<ProcessActionResponse> {
        let url = format!(
            "{}/tasks/{}/processes/{}/{}",
            self.base_url, task_id, process_id, action
        );
        let resp = self.client.put(&url).send().await.context("请求失败")?;
        let status = resp.status();
        if !status.is_success() {
            let body: serde_json::Value = resp.json().await.unwrap_or_default();
            anyhow::bail!("API 错误 {}: {}", status, body);
        }
        let result: ProcessActionResponse = resp.json().await.context("解析响应失败")?;
        Ok(result)
    }
}

fn print_route(route: &ProcessRoute) {
    println!("路线 ID: {}", route.id);
    println!("名称: {}", route.name);
    println!("创建时间: {}", route.created_at);
    println!("版本数: {}", route.versions.len());
    for v in &route.versions {
        println!("  版本: {}, 工序数: {}", v.version, v.processes.len());
        for p in &v.processes {
            println!(
                "    {} ({}, {}分钟, 前置: {:?})",
                p.name, p.id, p.standard_time_minutes, p.predecessors
            );
        }
    }
}

fn print_task(task: &ProductionTask) {
    println!("任务 ID: {}", task.id);
    println!("名称: {}", task.name);
    println!("工艺路线: {} ({})", task.route_name, task.route_id);
    println!("版本: {}", task.version);
    println!("标准总工时: {} 分钟", task.total_time_minutes);
    println!("关键路径: {:?}", task.critical_path);
    println!("开始时间: {:?}", task.started_at);
    println!("完成时间: {:?}", task.completed_at);
    println!("工序状态:");
    for p in &task.processes {
        let critical_flag = if p.is_critical { "[关键]" } else { "" };
        println!(
            "  {} {}: {} (标准: {}分钟)",
            p.name, critical_flag, p.status, p.standard_time_minutes
        );
    }
}

#[tokio::main]
async fn main() -> Result<()> {
    let cli = Cli::parse();
    let api = ApiClient::new(cli.server);

    match cli.command {
        Commands::Route { action } => match action {
            RouteCommands::Create { name } => {
                let route = api.create_route(name).await?;
                println!("创建成功:");
                print_route(&route);
            }
            RouteCommands::List => {
                let routes = api.list_routes().await?;
                println!("共 {} 条工艺路线:", routes.len());
                for route in &routes {
                    println!("- {} ({}版本)", route.name, route.versions.len());
                }
            }
            RouteCommands::Get { id } => {
                let route = api.get_route(&id).await?;
                print_route(&route);
            }
            RouteCommands::Delete { id } => {
                api.delete_route(&id).await?;
                println!("已删除路线 {}", id);
            }
            RouteCommands::AddVersion {
                route_id,
                version,
                processes,
            } => {
                let route = api.add_version(&route_id, version, processes).await?;
                println!("版本添加成功:");
                print_route(&route);
            }
            RouteCommands::CriticalPath { route_id, version } => {
                let info = api.get_critical_path(&route_id, &version).await?;
                println!("关键路径总工时: {} 分钟", info.total_time_minutes);
                println!("关键路径: {:?}", info.path);
            }
            RouteCommands::Compare { route_id, v1, v2 } => {
                let cmp = api.compare_versions(&route_id, &v1, &v2).await?;
                println!("版本 {} vs {}", cmp.version1, cmp.version2);
                println!("总工时: {} vs {} 分钟", cmp.total_time1, cmp.total_time2);
                println!("差异: {} 分钟", cmp.time_difference);
                println!("关键路径1: {:?}", cmp.critical_path1);
                println!("关键路径2: {:?}", cmp.critical_path2);
            }
        },
        Commands::Task { action } => match action {
            TaskCommands::Create {
                name,
                route_id,
                version,
            } => {
                let task = api.create_task(name, route_id, version).await?;
                println!("任务创建成功:");
                print_task(&task);
            }
            TaskCommands::List => {
                let tasks = api.list_tasks().await?;
                println!("共 {} 个生产任务:", tasks.len());
                for task in &tasks {
                    let status = if task.completed_at.is_some() {
                        "已完成"
                    } else if task.started_at.is_some() {
                        "进行中"
                    } else {
                        "未开始"
                    };
                    println!(
                        "- {} ({}, 版本{}, {})",
                        task.name, task.route_name, task.version, status
                    );
                }
            }
            TaskCommands::Get { id } => {
                let task = api.get_task(&id).await?;
                print_task(&task);
            }
            TaskCommands::Delete { id } => {
                api.delete_task(&id).await?;
                println!("已删除任务 {}", id);
            }
            TaskCommands::Available { task_id } => {
                let processes = api.get_available_processes(&task_id).await?;
                println!("可开始的工序 ({}个):", processes.len());
                for p in &processes {
                    let critical = if p.is_critical { "[关键]" } else { "" };
                    println!("  - {} {} ({}分钟)", p.name, critical, p.standard_time_minutes);
                }
            }
            TaskCommands::Start { task_id, process_id } => {
                let result = api.process_action(&task_id, &process_id, "start").await?;
                println!("工序已开始");
                print_task(&result.task);
            }
            TaskCommands::Complete { task_id, process_id } => {
                let result = api
                    .process_action(&task_id, &process_id, "complete")
                    .await?;
                println!("工序已完成");
                print_task(&result.task);
            }
            TaskCommands::Pause { task_id, process_id } => {
                let result = api.process_action(&task_id, &process_id, "pause").await?;
                println!("工序已暂停");
                if let Some(w) = result.warning {
                    println!("⚠️  警告: {}", w.message);
                }
                print_task(&result.task);
            }
            TaskCommands::Resume { task_id, process_id } => {
                let result = api
                    .process_action(&task_id, &process_id, "resume")
                    .await?;
                println!("工序已恢复");
                print_task(&result.task);
            }
        },
        Commands::Demo => {
            run_demo(&api).await?;
        }
    }

    Ok(())
}

async fn run_demo(api: &ApiClient) -> Result<()> {
    println!("=== 工艺路线管理系统演示 ===");
    println!();

    println!("1. 创建工艺路线 '产品A装配线'");
    let route = api.create_route("产品A装配线".to_string()).await?;
    println!("路线 ID: {}", route.id);
    println!();

    println!("2. 添加版本 v1.0");
    let processes_v1 = vec![
        ProcessDefinition::with_id("A", "零件准备", 30, vec![]),
        ProcessDefinition::with_id("B", "主体装配", 60, vec!["A".to_string()]),
        ProcessDefinition::with_id("C", "电子测试", 20, vec!["B".to_string()]),
        ProcessDefinition::with_id("D", "外观检查", 15, vec!["B".to_string()]),
        ProcessDefinition::with_id("E", "包装", 10, vec!["C".to_string(), "D".to_string()]),
    ];
    let route = api
        .add_version(&route.id, "v1.0".to_string(), processes_v1)
        .await?;
    println!("版本 v1.0 添加成功");
    println!();

    println!("3. 计算 v1.0 关键路径");
    let cp = api.get_critical_path(&route.id, "v1.0").await?;
    println!("总工时: {} 分钟", cp.total_time_minutes);
    println!("关键路径: {:?}", cp.path);
    println!();

    println!("4. 添加版本 v2.0（优化了主体装配）");
    let processes_v2 = vec![
        ProcessDefinition::with_id("A", "零件准备", 30, vec![]),
        ProcessDefinition::with_id("B", "主体装配", 40, vec!["A".to_string()]),
        ProcessDefinition::with_id("C", "电子测试", 20, vec!["B".to_string()]),
        ProcessDefinition::with_id("D", "外观检查", 15, vec!["B".to_string()]),
        ProcessDefinition::with_id("E", "包装", 10, vec!["C".to_string(), "D".to_string()]),
    ];
    let _route = api
        .add_version(&route.id, "v2.0".to_string(), processes_v2)
        .await?;
    println!();

    println!("5. 对比 v1.0 和 v2.0");
    let cmp = api.compare_versions(&route.id, "v1.0", "v2.0").await?;
    println!(
        "总工时: {} -> {} (差异: {}分钟)",
        cmp.total_time1, cmp.total_time2, cmp.time_difference
    );
    println!();

    println!("6. 创建生产任务（使用 v2.0）");
    let task = api
        .create_task(
            "生产订单 #2026-001".to_string(),
            route.id.clone(),
            Some("v2.0".to_string()),
        )
        .await?;
    println!("任务 ID: {}", task.id);
    println!("关键路径: {:?}", task.critical_path);
    println!();

    println!("7. 查看可开始的工序");
    let available = api.get_available_processes(&task.id).await?;
    for p in &available {
        println!(
            "  - {} ({}分钟, {})",
            p.name,
            p.standard_time_minutes,
            if p.is_critical { "关键" } else { "非关键" }
        );
    }
    println!();

    println!("8. 开始零件准备 (A)");
    let result = api.process_action(&task.id, "A", "start").await?;
    let process_a = result
        .task
        .processes
        .iter()
        .find(|p| p.definition_id == "A")
        .unwrap();
    println!("状态: {}", process_a.status);
    println!();

    println!("9. 完成零件准备 (A)");
    let result = api.process_action(&task.id, "A", "complete").await?;
    let process_a = result
        .task
        .processes
        .iter()
        .find(|p| p.definition_id == "A")
        .unwrap();
    println!("状态: {}", process_a.status);
    println!();

    println!("10. 开始主体装配 (B) - 关键工序");
    let result = api.process_action(&task.id, "B", "start").await?;
    let process_b = result
        .task
        .processes
        .iter()
        .find(|p| p.definition_id == "B")
        .unwrap();
    println!(
        "状态: {} (关键: {})",
        process_b.status, process_b.is_critical
    );
    println!();

    println!("11. 暂停关键工序 B");
    let result = api.process_action(&task.id, "B", "pause").await?;
    if let Some(w) = result.warning {
        println!("⚠️  警告: {}", w.message);
    }
    println!();

    println!("12. 恢复工序 B");
    let _ = api.process_action(&task.id, "B", "resume").await?;
    println!("已恢复");
    println!();

    println!("=== 演示完成 ===");

    Ok(())
}
