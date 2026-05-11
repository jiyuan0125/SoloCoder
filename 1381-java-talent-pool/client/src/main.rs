use std::fmt::Write;

use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use talent_core::{
    Department, Employee, Assessment, SuccessionPlan, GridDistribution,
    GridZone,
};

#[derive(Parser, Debug)]
#[command(author, version, about = "梯队人才管理系统命令行客户端", long_about = None)]
struct Args {
    #[arg(short, long, env = "TALENT_SERVER", default_value = "http://127.0.0.1:3000")]
    server: String,
    
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Department {
        #[command(subcommand)]
        action: DepartmentCommands,
    },
    Employee {
        #[command(subcommand)]
        action: EmployeeCommands,
    },
    Assessment {
        #[command(subcommand)]
        action: AssessmentCommands,
    },
    Succession {
        #[command(subcommand)]
        action: SuccessionCommands,
    },
    Grid {
        #[command(subcommand)]
        action: GridCommands,
    },
    Warnings,
}

#[derive(Subcommand, Debug)]
enum DepartmentCommands {
    Create {
        #[arg(short, long)]
        name: String,
    },
    List,
}

#[derive(Subcommand, Debug)]
enum EmployeeCommands {
    Create {
        #[arg(short, long)]
        name: String,
        #[arg(short, long)]
        number: String,
        #[arg(short, long)]
        department_id: uuid::Uuid,
        #[arg(short, long)]
        position: String,
    },
    List {
        #[arg(short, long)]
        department_id: Option<uuid::Uuid>,
    },
}

#[derive(Subcommand, Debug)]
enum AssessmentCommands {
    Create {
        #[arg(short, long)]
        employee_id: uuid::Uuid,
        #[arg(short, long)]
        year: i32,
        #[arg(short, long)]
        performance: u8,
        #[arg(short, long)]
        potential: u8,
    },
    List {
        #[arg(short, long)]
        employee_id: uuid::Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum SuccessionCommands {
    Create {
        #[arg(short, long)]
        position: String,
        #[arg(short, long)]
        department_id: uuid::Uuid,
        #[arg(short, long)]
        key: bool,
    },
    List {
        #[arg(short, long)]
        department_id: Option<uuid::Uuid>,
    },
    AddCandidate {
        #[arg(short, long)]
        plan_id: uuid::Uuid,
        #[arg(short, long)]
        employee_id: uuid::Uuid,
        #[arg(short, long)]
        year: Option<i32>,
        #[arg(short, long)]
        notes: Option<String>,
    },
}

#[derive(Subcommand, Debug)]
enum GridCommands {
    Distribution {
        #[arg(short, long)]
        year: i32,
        #[arg(short, long)]
        department_id: Option<uuid::Uuid>,
    },
    Employees {
        #[arg(short, long)]
        year: i32,
        #[arg(short, long)]
        zone: String,
        #[arg(short, long)]
        department_id: Option<uuid::Uuid>,
    },
}

#[derive(Debug, Serialize)]
struct CreateDepartmentRequest {
    name: String,
}

#[derive(Debug, Serialize)]
struct CreateEmployeeRequest {
    name: String,
    employee_number: String,
    department_id: uuid::Uuid,
    position: String,
}

#[derive(Debug, Serialize)]
struct CreateAssessmentRequest {
    year: i32,
    performance: u8,
    potential: u8,
}

#[derive(Debug, Serialize)]
struct CreateSuccessionPlanRequest {
    position: String,
    department_id: uuid::Uuid,
    is_key_position: bool,
}

#[derive(Debug, Serialize)]
struct AddCandidateRequest {
    employee_id: uuid::Uuid,
    assessment_year: Option<i32>,
    notes: Option<String>,
}

#[derive(Debug, Deserialize)]
struct ErrorResponse {
    error: String,
}

#[derive(Debug, Deserialize)]
struct EmployeeWithAssessment {
    employee: Employee,
    assessment: Assessment,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Args::parse();
    let client = Client::new();
    let base_url = args.server.trim_end_matches('/').to_string();
    
    match args.command {
        Commands::Department { action } => handle_department(&client, &base_url, action).await?,
        Commands::Employee { action } => handle_employee(&client, &base_url, action).await?,
        Commands::Assessment { action } => handle_assessment(&client, &base_url, action).await?,
        Commands::Succession { action } => handle_succession(&client, &base_url, action).await?,
        Commands::Grid { action } => handle_grid(&client, &base_url, action).await?,
        Commands::Warnings => handle_warnings(&client, &base_url).await?,
    }
    
    Ok(())
}

async fn handle_department(
    client: &Client,
    base_url: &str,
    action: DepartmentCommands,
) -> Result<(), Box<dyn std::error::Error>> {
    match action {
        DepartmentCommands::Create { name } => {
            let req = CreateDepartmentRequest { name };
            let resp = client
                .post(&format!("{}/api/departments", base_url))
                .json(&req)
                .send()
                .await?;
            
            if resp.status().is_success() {
                let dept: Department = resp.json().await?;
                println!("创建部门成功:");
                print_department(&dept);
            } else {
                let err: ErrorResponse = resp.json().await?;
                eprintln!("错误: {}", err.error);
            }
        }
        DepartmentCommands::List => {
            let resp = client
                .get(&format!("{}/api/departments", base_url))
                .send()
                .await?;
            
            let depts: Vec<Department> = resp.json().await?;
            println!("部门列表:");
            println!("{}", "=".repeat(60));
            for dept in depts {
                print_department(&dept);
                println!();
            }
        }
    }
    Ok(())
}

async fn handle_employee(
    client: &Client,
    base_url: &str,
    action: EmployeeCommands,
) -> Result<(), Box<dyn std::error::Error>> {
    match action {
        EmployeeCommands::Create { name, number, department_id, position } => {
            let req = CreateEmployeeRequest {
                name,
                employee_number: number,
                department_id,
                position,
            };
            let resp = client
                .post(&format!("{}/api/employees", base_url))
                .json(&req)
                .send()
                .await?;
            
            if resp.status().is_success() {
                let emp: Employee = resp.json().await?;
                println!("创建员工成功:");
                print_employee(&emp);
            } else {
                let err: ErrorResponse = resp.json().await?;
                eprintln!("错误: {}", err.error);
            }
        }
        EmployeeCommands::List { department_id } => {
            let mut url = format!("{}/api/employees", base_url);
            if let Some(id) = department_id {
                url.push_str(&format!("?department_id={}", id));
            }
            let resp = client.get(&url).send().await?;
            let employees: Vec<Employee> = resp.json().await?;
            
            println!("员工列表:");
            println!("{}", "=".repeat(80));
            for emp in employees {
                print_employee(&emp);
                println!();
            }
        }
    }
    Ok(())
}

async fn handle_assessment(
    client: &Client,
    base_url: &str,
    action: AssessmentCommands,
) -> Result<(), Box<dyn std::error::Error>> {
    match action {
        AssessmentCommands::Create { employee_id, year, performance, potential } => {
            let req = CreateAssessmentRequest {
                year,
                performance,
                potential,
            };
            let resp = client
                .post(&format!("{}/api/employees/{}/assessments", base_url, employee_id))
                .json(&req)
                .send()
                .await?;
            
            if resp.status().is_success() {
                let assessment: Assessment = resp.json().await?;
                println!("创建评估成功:");
                print_assessment(&assessment);
            } else {
                let err: ErrorResponse = resp.json().await?;
                eprintln!("错误: {}", err.error);
            }
        }
        AssessmentCommands::List { employee_id } => {
            let resp = client
                .get(&format!("{}/api/employees/{}/assessments", base_url, employee_id))
                .send()
                .await?;
            let assessments: Vec<Assessment> = resp.json().await?;
            
            println!("员工评估历史:");
            println!("{}", "=".repeat(80));
            for assessment in assessments {
                print_assessment(&assessment);
                println!();
            }
        }
    }
    Ok(())
}

async fn handle_succession(
    client: &Client,
    base_url: &str,
    action: SuccessionCommands,
) -> Result<(), Box<dyn std::error::Error>> {
    match action {
        SuccessionCommands::Create { position, department_id, key } => {
            let req = CreateSuccessionPlanRequest {
                position,
                department_id,
                is_key_position: key,
            };
            let resp = client
                .post(&format!("{}/api/succession-plans", base_url))
                .json(&req)
                .send()
                .await?;
            
            if resp.status().is_success() {
                let plan: SuccessionPlan = resp.json().await?;
                println!("创建继任计划成功:");
                print_succession_plan(&plan);
            } else {
                let err: ErrorResponse = resp.json().await?;
                eprintln!("错误: {}", err.error);
            }
        }
        SuccessionCommands::List { department_id } => {
            let mut url = format!("{}/api/succession-plans", base_url);
            if let Some(id) = department_id {
                url.push_str(&format!("?department_id={}", id));
            }
            let resp = client.get(&url).send().await?;
            let plans: Vec<SuccessionPlan> = resp.json().await?;
            
            println!("继任计划列表:");
            println!("{}", "=".repeat(80));
            for plan in plans {
                print_succession_plan(&plan);
                println!();
            }
        }
        SuccessionCommands::AddCandidate { plan_id, employee_id, year, notes } => {
            let req = AddCandidateRequest {
                employee_id,
                assessment_year: year,
                notes,
            };
            let resp = client
                .post(&format!("{}/api/succession-plans/{}/candidates", base_url, plan_id))
                .json(&req)
                .send()
                .await?;
            
            if resp.status().is_success() {
                let plan: SuccessionPlan = resp.json().await?;
                println!("添加候选人成功:");
                print_succession_plan(&plan);
            } else {
                let err: ErrorResponse = resp.json().await?;
                eprintln!("错误: {}", err.error);
            }
        }
    }
    Ok(())
}

fn parse_grid_zone(s: &str) -> Result<GridZone, String> {
    match s.to_uppercase().as_str() {
        "A1" => Ok(GridZone::A1),
        "A2" => Ok(GridZone::A2),
        "A3" => Ok(GridZone::A3),
        "B1" => Ok(GridZone::B1),
        "B2" => Ok(GridZone::B2),
        "B3" => Ok(GridZone::B3),
        "C1" => Ok(GridZone::C1),
        "C2" => Ok(GridZone::C2),
        "C3" => Ok(GridZone::C3),
        _ => Err(format!("无效的九宫格区域: {}. 有效值: A1, A2, A3, B1, B2, B3, C1, C2, C3", s)),
    }
}

async fn handle_grid(
    client: &Client,
    base_url: &str,
    action: GridCommands,
) -> Result<(), Box<dyn std::error::Error>> {
    match action {
        GridCommands::Distribution { year, department_id } => {
            let mut url = format!("{}/api/grid/distribution?year={}", base_url, year);
            if let Some(id) = department_id {
                url.push_str(&format!("&department_id={}", id));
            }
            let resp = client.get(&url).send().await?;
            let distribution: GridDistribution = resp.json().await?;
            
            print_grid_distribution(&distribution);
        }
        GridCommands::Employees { year, zone, department_id } => {
            let zone = parse_grid_zone(&zone)?;
            let mut url = format!("{}/api/grid/employees?year={}&zone={:?}", base_url, year, zone);
            if let Some(id) = department_id {
                url.push_str(&format!("&department_id={}", id));
            }
            let resp = client.get(&url).send().await?;
            let results: Vec<EmployeeWithAssessment> = resp.json().await?;
            
            println!("{} 区域的员工列表 ({}年):", zone.name(), year);
            println!("{}", "=".repeat(80));
            for result in results {
                print_employee(&result.employee);
                print_assessment(&result.assessment);
                println!();
            }
        }
    }
    Ok(())
}

async fn handle_warnings(
    client: &Client,
    base_url: &str,
) -> Result<(), Box<dyn std::error::Error>> {
    let resp = client
        .get(&format!("{}/api/warnings", base_url))
        .send()
        .await?;
    let warnings: Vec<String> = resp.json().await?;
    
    if warnings.is_empty() {
        println!("没有预警信息，所有关键岗位都有足够的继任候选人。");
    } else {
        println!("继任计划预警:");
        println!("{}", "=".repeat(80));
        for (i, warning) in warnings.iter().enumerate() {
            println!("{}. {}", i + 1, warning);
        }
    }
    Ok(())
}

fn print_department(dept: &Department) {
    println!("ID:   {}", dept.id);
    println!("名称: {}", dept.name);
}

fn print_employee(emp: &Employee) {
    println!("ID:         {}", emp.id);
    println!("姓名:       {}", emp.name);
    println!("工号:       {}", emp.employee_number);
    println!("部门ID:     {}", emp.department_id);
    println!("职位:       {}", emp.position);
}

fn print_assessment(assessment: &Assessment) {
    let zone = assessment.zone;
    println!("评估ID:     {}", assessment.id);
    println!("年份:       {}", assessment.year);
    println!("绩效:       P{}", assessment.performance.value());
    println!("潜力:       L{}", assessment.potential.value());
    println!("九宫格:     {:?} ({})", assessment.zone, zone.name());
    println!("含义:       {}", zone.description());
}

fn print_succession_plan(plan: &SuccessionPlan) {
    println!("计划ID:     {}", plan.id);
    println!("岗位:       {}", plan.position);
    println!("部门ID:     {}", plan.department_id);
    println!("关键岗位:   {}", if plan.is_key_position { "是" } else { "否" });
    println!("候选人数:   {}", plan.candidates.len());
    if let Some(warning) = plan.warning_message() {
        println!("⚠️  预警:     {}", warning);
    }
    if !plan.candidates.is_empty() {
        println!("候选人:");
        for (i, candidate) in plan.candidates.iter().enumerate() {
            println!("  {}. 员工ID: {}", i + 1, candidate.employee_id);
            println!("     九宫格: {:?} ({})", candidate.current_zone, candidate.current_zone.name());
            println!("     评估年: {}", candidate.assessment_year);
            if let Some(notes) = &candidate.notes {
                println!("     备注:   {}", notes);
            }
        }
    }
}

fn print_grid_distribution(distribution: &GridDistribution) {
    let counts = &distribution.counts;
    
    let mut output = String::new();
    writeln!(&mut output, "九宫格人才分布 ({}年):", distribution.year).unwrap();
    writeln!(&mut output, "{}", "=".repeat(70)).unwrap();
    writeln!(&mut output, "总人数: {}", counts.total).unwrap();
    writeln!(&mut output).unwrap();
    
    writeln!(&mut output, "┌──────────────────────────────────────────────────────────────────────┐").unwrap();
    writeln!(&mut output, "│                      高潜力              中潜力              低潜力    │").unwrap();
    writeln!(&mut output, "┌────────────┬─────────────────────┬─────────────────────┬─────────────────────┐").unwrap();
    writeln!(&mut output, "│            │ A1: 明星员工         │ A2: 高绩效稳定型     │ A3: 绩效贡献者        │").unwrap();
    writeln!(&mut output, "│ 高绩效     │      {:>3} 人         │      {:>3} 人         │      {:>3} 人         │", counts.a1, counts.a2, counts.a3).unwrap();
    writeln!(&mut output, "├────────────┼─────────────────────┼─────────────────────┼─────────────────────┤").unwrap();
    writeln!(&mut output, "│            │ B1: 高潜力成长型     │ B2: 中坚力量         │ B3: 待发展稳定型      │").unwrap();
    writeln!(&mut output, "│ 中绩效     │      {:>3} 人         │      {:>3} 人         │      {:>3} 人         │", counts.b1, counts.b2, counts.b3).unwrap();
    writeln!(&mut output, "├────────────┼─────────────────────┼─────────────────────┼─────────────────────┤").unwrap();
    writeln!(&mut output, "│            │ C1: 待观察高潜型     │ C2: 待改进型         │ C3: 需优化型          │").unwrap();
    writeln!(&mut output, "│ 低绩效     │      {:>3} 人         │      {:>3} 人         │      {:>3} 人         │", counts.c1, counts.c2, counts.c3).unwrap();
    writeln!(&mut output, "└────────────┴─────────────────────┴─────────────────────┴─────────────────────┘").unwrap();
    
    println!("{}", output);
}
