use clap::{Parser, Subcommand};
use app_core::models::{Course, EmployeeId, Notification, Registration};
use reqwest::blocking::Client;
use serde::{Deserialize, Serialize};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://127.0.0.1:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    #[command(about = "List all courses")]
    ListCourses,

    #[command(about = "Get course details")]
    GetCourse { id: String },

    #[command(about = "Create a new course")]
    CreateCourse {
        #[arg(short, long)]
        id: String,
        #[arg(short, long)]
        name: String,
        #[arg(short, long)]
        capacity: u32,
        #[arg(long)]
        start: u64,
        #[arg(long)]
        end: u64,
    },

    #[command(about = "Register for a course")]
    Register {
        course_id: String,
        employee_id: String,
    },

    #[command(about = "Cancel registration")]
    Cancel {
        course_id: String,
        employee_id: String,
    },

    #[command(about = "Confirm pending registration")]
    Confirm {
        course_id: String,
        employee_id: String,
    },

    #[command(about = "List registrations for a course")]
    ListRegistrations { course_id: String },

    #[command(about = "List waiting list for a course")]
    ListWaitingList { course_id: String },

    #[command(about = "List registrations for an employee")]
    MyRegistrations { employee_id: String },

    #[command(about = "List notifications for an employee")]
    MyNotifications { employee_id: String },

    #[command(about = "Mark notification as read")]
    MarkRead { notification_id: String },
}

#[derive(Debug, Deserialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    error: Option<String>,
}

#[derive(Debug, Serialize)]
struct CreateCourseRequest {
    id: String,
    name: String,
    capacity: u32,
    registration_start: u64,
    registration_end: u64,
}

#[derive(Debug, Serialize)]
struct RegisterRequest {
    employee_id: String,
}

fn main() {
    let args = Args::parse();
    let client = Client::new();
    let base_url = args.server.trim_end_matches('/').to_string();

    match args.command {
        Commands::ListCourses => {
            let resp = client.get(&format!("{}/courses", base_url))
                .send()
                .expect("Failed to send request");
            
            let result: ApiResponse<Vec<Course>> = resp.json().expect("Failed to parse response");
            
            if result.success {
                let courses = result.data.unwrap();
                println!("Total courses: {}", courses.len());
                for course in courses {
                    print_course(&course);
                }
            } else {
                println!("Error: {}", result.error.unwrap());
            }
        }
        Commands::GetCourse { id } => {
            let resp = client.get(&format!("{}/courses/{}", base_url, id))
                .send()
                .expect("Failed to send request");
            
            let result: ApiResponse<Course> = resp.json().expect("Failed to parse response");
            
            if result.success {
                if let Some(course) = result.data {
                    print_course(&course);
                }
            } else {
                println!("Error: {}", result.error.unwrap());
            }
        }
        Commands::CreateCourse { id, name, capacity, start, end } => {
            let req = CreateCourseRequest {
                id,
                name,
                capacity,
                registration_start: start,
                registration_end: end,
            };
            
            let resp = client.post(&format!("{}/courses", base_url))
                .json(&req)
                .send()
                .expect("Failed to send request");
            
            let result: ApiResponse<Course> = resp.json().expect("Failed to parse response");
            
            if result.success {
                println!("Course created successfully!");
                if let Some(course) = result.data {
                    print_course(&course);
                }
            } else {
                println!("Error: {}", result.error.unwrap());
            }
        }
        Commands::Register { course_id, employee_id } => {
            let req = RegisterRequest { employee_id };
            
            let resp = client.post(&format!("{}/courses/{}/register", base_url, course_id))
                .json(&req)
                .send()
                .expect("Failed to send request");
            
            let result: ApiResponse<Registration> = resp.json().expect("Failed to parse response");
            
            if result.success {
                println!("Registration successful!");
                if let Some(reg) = result.data {
                    print_registration(&reg);
                }
            } else {
                println!("Error: {}", result.error.unwrap());
            }
        }
        Commands::Cancel { course_id, employee_id } => {
            let resp = client.delete(&format!("{}/courses/{}/registrations/{}", base_url, course_id, employee_id))
                .send()
                .expect("Failed to send request");
            
            let result: ApiResponse<()> = resp.json().expect("Failed to parse response");
            
            if result.success {
                println!("Registration cancelled successfully!");
            } else {
                println!("Error: {}", result.error.unwrap());
            }
        }
        Commands::Confirm { course_id, employee_id } => {
            let resp = client.put(&format!("{}/courses/{}/registrations/{}/confirm", base_url, course_id, employee_id))
                .send()
                .expect("Failed to send request");
            
            let result: ApiResponse<Registration> = resp.json().expect("Failed to parse response");
            
            if result.success {
                println!("Registration confirmed!");
                if let Some(reg) = result.data {
                    print_registration(&reg);
                }
            } else {
                println!("Error: {}", result.error.unwrap());
            }
        }
        Commands::ListRegistrations { course_id } => {
            let resp = client.get(&format!("{}/courses/{}/registrations", base_url, course_id))
                .send()
                .expect("Failed to send request");
            
            let result: ApiResponse<Vec<Registration>> = resp.json().expect("Failed to parse response");
            
            if result.success {
                let registrations = result.data.unwrap();
                println!("Total registrations: {}", registrations.len());
                for reg in registrations {
                    print_registration(&reg);
                }
            } else {
                println!("Error: {}", result.error.unwrap());
            }
        }
        Commands::ListWaitingList { course_id } => {
            let resp = client.get(&format!("{}/courses/{}/waiting-list", base_url, course_id))
                .send()
                .expect("Failed to send request");
            
            let result: ApiResponse<Vec<EmployeeId>> = resp.json().expect("Failed to parse response");
            
            if result.success {
                let waiting_list = result.data.unwrap();
                println!("Waiting list ({} people):", waiting_list.len());
                for (idx, emp) in waiting_list.iter().enumerate() {
                    println!("  {}. {}", idx + 1, emp.0);
                }
            } else {
                println!("Error: {}", result.error.unwrap());
            }
        }
        Commands::MyRegistrations { employee_id } => {
            let resp = client.get(&format!("{}/employees/{}/registrations", base_url, employee_id))
                .send()
                .expect("Failed to send request");
            
            let result: ApiResponse<Vec<Registration>> = resp.json().expect("Failed to parse response");
            
            if result.success {
                let registrations = result.data.unwrap();
                println!("My registrations ({}):", registrations.len());
                for reg in registrations {
                    print_registration(&reg);
                }
            } else {
                println!("Error: {}", result.error.unwrap());
            }
        }
        Commands::MyNotifications { employee_id } => {
            let resp = client.get(&format!("{}/employees/{}/notifications", base_url, employee_id))
                .send()
                .expect("Failed to send request");
            
            let result: ApiResponse<Vec<Notification>> = resp.json().expect("Failed to parse response");
            
            if result.success {
                let notifications = result.data.unwrap();
                println!("Notifications ({}):", notifications.len());
                for notif in notifications {
                    print_notification(&notif);
                }
            } else {
                println!("Error: {}", result.error.unwrap());
            }
        }
        Commands::MarkRead { notification_id } => {
            let resp = client.put(&format!("{}/notifications/{}/read", base_url, notification_id))
                .send()
                .expect("Failed to send request");
            
            let result: ApiResponse<()> = resp.json().expect("Failed to parse response");
            
            if result.success {
                println!("Notification marked as read!");
            } else {
                println!("Error: {}", result.error.unwrap());
            }
        }
    }
}

fn print_course(course: &Course) {
    println!("\n=== Course {} ===", course.id.0);
    println!("  Name: {}", course.name);
    println!("  Capacity: {}", course.capacity);
    println!("  Registration Start: {}", course.registration_start);
    println!("  Registration End: {}", course.registration_end);
}

fn print_registration(reg: &Registration) {
    println!("\n--- Registration ---");
    println!("  Employee: {}", reg.employee_id.0);
    println!("  Course: {}", reg.course_id.0);
    println!("  Status: {:?}", reg.status);
    println!("  Created: {}", reg.created_at);
    println!("  Updated: {}", reg.updated_at);
}

fn print_notification(notif: &Notification) {
    println!("\n--- Notification {} ---", notif.id);
    println!("  Employee: {}", notif.employee_id.0);
    println!("  Course: {}", notif.course_id.0);
    println!("  Message: {}", notif.message);
    println!("  Read: {}", if notif.read { "Yes" } else { "No" });
    println!("  Created: {}", notif.created_at);
}
