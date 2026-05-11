use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use chrono::NaiveDate;
use desk_core::{DeskType, Desk, Department, Employee, Reservation};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short, long, default_value = "http://localhost:8080")]
    server: String,
    
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Health,
    
    Department {
        #[command(subcommand)]
        action: DepartmentCommands,
    },
    
    Desk {
        #[command(subcommand)]
        action: DeskCommands,
    },
    
    Employee {
        #[command(subcommand)]
        action: EmployeeCommands,
    },
    
    Reservation {
        #[command(subcommand)]
        action: ReservationCommands,
    },
    
    Maintenance {
        #[command(subcommand)]
        action: MaintenanceCommands,
    },
}

#[derive(Subcommand, Debug)]
enum DepartmentCommands {
    Create {
        #[arg(short, long)]
        name: String,
        #[arg(short, long)]
        floor: i32,
    },
    List,
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
    Relocate {
        #[arg(short, long)]
        department_id: Uuid,
        #[arg(short, long)]
        target_floor: i32,
    },
}

#[derive(Subcommand, Debug)]
enum DeskCommands {
    Create {
        #[arg(short, long)]
        code: String,
        #[arg(short, long)]
        desk_type: DeskTypeArg,
        #[arg(short, long)]
        floor: i32,
    },
    List,
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
    Available {
        #[arg(short, long)]
        date: Option<NaiveDate>,
    },
}

#[derive(Subcommand, Debug)]
enum EmployeeCommands {
    Create {
        #[arg(short, long)]
        name: String,
        #[arg(short, long)]
        department_id: Uuid,
    },
    Batch {
        #[arg(short, long, num_args = 2..)]
        employees: Vec<String>,
    },
    List,
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
    Leave {
        #[arg(short, long)]
        id: Uuid,
    },
    Desk {
        #[arg(short, long)]
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum ReservationCommands {
    Create {
        #[arg(short, long)]
        desk_id: Uuid,
        #[arg(short, long)]
        employee_id: Uuid,
        #[arg(short, long)]
        date: NaiveDate,
    },
    List,
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
    Checkin {
        #[arg(short, long)]
        id: Uuid,
    },
    Cancel {
        #[arg(short, long)]
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum MaintenanceCommands {
    Expired,
    Pending,
}

#[derive(Debug, Clone, Copy)]
enum DeskTypeArg {
    Fixed,
    Shared,
    Visitor,
}

impl std::str::FromStr for DeskTypeArg {
    type Err = String;
    
    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_lowercase().as_str() {
            "fixed" => Ok(DeskTypeArg::Fixed),
            "shared" => Ok(DeskTypeArg::Shared),
            "visitor" => Ok(DeskTypeArg::Visitor),
            _ => Err(format!("无效的工位类型: {}", s)),
        }
    }
}

impl From<DeskTypeArg> for DeskType {
    fn from(arg: DeskTypeArg) -> Self {
        match arg {
            DeskTypeArg::Fixed => DeskType::Fixed,
            DeskTypeArg::Shared => DeskType::Shared,
            DeskTypeArg::Visitor => DeskType::Visitor,
        }
    }
}

#[derive(Debug, Serialize, Deserialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    error: Option<String>,
}

#[derive(Debug, Serialize)]
struct CreateDepartmentRequest {
    name: String,
    floor: i32,
}

#[derive(Debug, Serialize)]
struct CreateDeskRequest {
    code: String,
    desk_type: DeskType,
    floor: i32,
}

#[derive(Debug, Serialize)]
struct CreateEmployeeRequest {
    name: String,
    department_id: Uuid,
}

#[derive(Debug, Serialize)]
struct BatchCreateEmployeesRequest {
    employees: Vec<(String, Uuid)>,
}

#[derive(Debug, Serialize)]
struct ReserveDeskRequest {
    employee_id: Uuid,
    date: NaiveDate,
}

#[derive(Debug, Serialize)]
struct BatchRelocateRequest {
    department_id: Uuid,
    target_floor: i32,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    tracing_subscriber::fmt::init();
    
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server.trim_end_matches('/').to_string();
    
    match cli.command {
        Commands::Health => health_check(&client, &base_url).await?,
        
        Commands::Department { action } => {
            match action {
                DepartmentCommands::Create { name, floor } => {
                    create_department(&client, &base_url, name, floor).await?
                }
                DepartmentCommands::List => list_departments(&client, &base_url).await?,
                DepartmentCommands::Get { id } => get_department(&client, &base_url, id).await?,
                DepartmentCommands::Relocate { department_id, target_floor } => {
                    batch_relocate(&client, &base_url, department_id, target_floor).await?
                }
            }
        }
        
        Commands::Desk { action } => {
            match action {
                DeskCommands::Create { code, desk_type, floor } => {
                    create_desk(&client, &base_url, code, desk_type.into(), floor).await?
                }
                DeskCommands::List => list_desks(&client, &base_url).await?,
                DeskCommands::Get { id } => get_desk(&client, &base_url, id).await?,
                DeskCommands::Available { date } => {
                    list_available_shared_desks(&client, &base_url, date).await?
                }
            }
        }
        
        Commands::Employee { action } => {
            match action {
                EmployeeCommands::Create { name, department_id } => {
                    create_employee(&client, &base_url, name, department_id).await?
                }
                EmployeeCommands::Batch { employees } => {
                    batch_create_employees(&client, &base_url, employees).await?
                }
                EmployeeCommands::List => list_employees(&client, &base_url).await?,
                EmployeeCommands::Get { id } => get_employee(&client, &base_url, id).await?,
                EmployeeCommands::Leave { id } => employee_leave(&client, &base_url, id).await?,
                EmployeeCommands::Desk { id } => get_employee_desk(&client, &base_url, id).await?,
            }
        }
        
        Commands::Reservation { action } => {
            match action {
                ReservationCommands::Create { desk_id, employee_id, date } => {
                    reserve_desk(&client, &base_url, desk_id, employee_id, date).await?
                }
                ReservationCommands::List => list_reservations(&client, &base_url).await?,
                ReservationCommands::Get { id } => get_reservation(&client, &base_url, id).await?,
                ReservationCommands::Checkin { id } => check_in_reservation(&client, &base_url, id).await?,
                ReservationCommands::Cancel { id } => cancel_reservation(&client, &base_url, id).await?,
            }
        }
        
        Commands::Maintenance { action } => {
            match action {
                MaintenanceCommands::Expired => process_expired(&client, &base_url).await?,
                MaintenanceCommands::Pending => process_pending(&client, &base_url).await?,
            }
        }
    }
    
    Ok(())
}

async fn health_check(client: &Client, base_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let resp: ApiResponse<String> = client.get(format!("{}/api/health", base_url))
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn create_department(client: &Client, base_url: &str, name: String, floor: i32) -> Result<(), Box<dyn std::error::Error>> {
    let req = CreateDepartmentRequest { name, floor };
    let resp: ApiResponse<Department> = client.post(format!("{}/api/departments", base_url))
        .json(&req)
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn list_departments(client: &Client, base_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let resp: ApiResponse<Vec<Department>> = client.get(format!("{}/api/departments", base_url))
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn get_department(client: &Client, base_url: &str, id: Uuid) -> Result<(), Box<dyn std::error::Error>> {
    let resp: ApiResponse<Department> = client.get(format!("{}/api/departments/{}", base_url, id))
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn create_desk(client: &Client, base_url: &str, code: String, desk_type: DeskType, floor: i32) -> Result<(), Box<dyn std::error::Error>> {
    let req = CreateDeskRequest { code, desk_type, floor };
    let resp: ApiResponse<Desk> = client.post(format!("{}/api/desks", base_url))
        .json(&req)
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn list_desks(client: &Client, base_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let resp: ApiResponse<Vec<Desk>> = client.get(format!("{}/api/desks", base_url))
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn get_desk(client: &Client, base_url: &str, id: Uuid) -> Result<(), Box<dyn std::error::Error>> {
    let resp: ApiResponse<Desk> = client.get(format!("{}/api/desks/{}", base_url, id))
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn list_available_shared_desks(client: &Client, base_url: &str, date: Option<NaiveDate>) -> Result<(), Box<dyn std::error::Error>> {
    let mut url = format!("{}/api/desks/shared/available", base_url);
    if let Some(d) = date {
        url.push_str(&format!("?date={}", d));
    }
    
    let resp: ApiResponse<Vec<Desk>> = client.get(url)
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn create_employee(client: &Client, base_url: &str, name: String, department_id: Uuid) -> Result<(), Box<dyn std::error::Error>> {
    let req = CreateEmployeeRequest { name, department_id };
    let resp: ApiResponse<serde_json::Value> = client.post(format!("{}/api/employees", base_url))
        .json(&req)
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn batch_create_employees(client: &Client, base_url: &str, employees: Vec<String>) -> Result<(), Box<dyn std::error::Error>> {
    if employees.len() % 2 != 0 {
        println!("错误: 员工列表必须成对提供 (name, department_id)");
        return Ok(());
    }
    
    let mut emp_list = Vec::new();
    for chunk in employees.chunks(2) {
        let name = chunk[0].clone();
        let dept_id: Uuid = match chunk[1].parse() {
            Ok(id) => id,
            Err(e) => {
                println!("错误: 无效的部门ID {}: {}", chunk[1], e);
                return Ok(());
            }
        };
        emp_list.push((name, dept_id));
    }
    
    let req = BatchCreateEmployeesRequest { employees: emp_list };
    let resp: ApiResponse<Vec<(Employee, Desk)>> = client.post(format!("{}/api/employees/batch", base_url))
        .json(&req)
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn list_employees(client: &Client, base_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let resp: ApiResponse<Vec<Employee>> = client.get(format!("{}/api/employees", base_url))
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn get_employee(client: &Client, base_url: &str, id: Uuid) -> Result<(), Box<dyn std::error::Error>> {
    let resp: ApiResponse<Employee> = client.get(format!("{}/api/employees/{}", base_url, id))
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn employee_leave(client: &Client, base_url: &str, id: Uuid) -> Result<(), Box<dyn std::error::Error>> {
    let resp: ApiResponse<()> = client.put(format!("{}/api/employees/{}/leave", base_url, id))
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn get_employee_desk(client: &Client, base_url: &str, id: Uuid) -> Result<(), Box<dyn std::error::Error>> {
    let resp: ApiResponse<Option<Desk>> = client.get(format!("{}/api/employees/{}/desk", base_url, id))
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn reserve_desk(client: &Client, base_url: &str, desk_id: Uuid, employee_id: Uuid, date: NaiveDate) -> Result<(), Box<dyn std::error::Error>> {
    let req = ReserveDeskRequest { employee_id, date };
    let resp: ApiResponse<Reservation> = client.post(format!("{}/api/reservations/{}", base_url, desk_id))
        .json(&req)
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn list_reservations(client: &Client, base_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let resp: ApiResponse<Vec<Reservation>> = client.get(format!("{}/api/reservations", base_url))
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn get_reservation(client: &Client, base_url: &str, id: Uuid) -> Result<(), Box<dyn std::error::Error>> {
    let resp: ApiResponse<Reservation> = client.get(format!("{}/api/reservations/{}", base_url, id))
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn check_in_reservation(client: &Client, base_url: &str, id: Uuid) -> Result<(), Box<dyn std::error::Error>> {
    let resp: ApiResponse<()> = client.put(format!("{}/api/reservations/{}/checkin", base_url, id))
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn cancel_reservation(client: &Client, base_url: &str, id: Uuid) -> Result<(), Box<dyn std::error::Error>> {
    let resp: ApiResponse<()> = client.put(format!("{}/api/reservations/{}/cancel", base_url, id))
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn batch_relocate(client: &Client, base_url: &str, department_id: Uuid, target_floor: i32) -> Result<(), Box<dyn std::error::Error>> {
    let req = BatchRelocateRequest { department_id, target_floor };
    let resp: ApiResponse<Vec<(Uuid, Uuid)>> = client.post(format!("{}/api/departments/relocate", base_url))
        .json(&req)
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn process_expired(client: &Client, base_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let resp: ApiResponse<Vec<Reservation>> = client.post(format!("{}/api/maintenance/expired", base_url))
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

async fn process_pending(client: &Client, base_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let resp: ApiResponse<serde_json::Value> = client.post(format!("{}/api/maintenance/pending", base_url))
        .send()
        .await?
        .json()
        .await?;
    
    print_response(&resp);
    Ok(())
}

fn print_response<T: Serialize>(resp: &ApiResponse<T>) {
    if resp.success {
        if let Some(data) = &resp.data {
            println!("{}", serde_json::to_string_pretty(data).unwrap());
        } else {
            println!("操作成功");
        }
    } else {
        if let Some(err) = &resp.error {
            eprintln!("错误: {}", err);
        } else {
            eprintln!("操作失败");
        }
    }
}
