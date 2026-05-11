use clap::{Parser, Subcommand};
use chrono::Utc;
use reqwest::blocking::Client;
use uuid::Uuid;

use lab_report_core::{Gender, Report, ReportInput, TestInput, TestResultValue};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "LAB_REPORT_SERVER", default_value = "http://localhost:8080")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Health,
    
    Create {
        #[arg(short, long)]
        name: String,
        
        #[arg(short, long)]
        gender: String,
        
        #[arg(short, long)]
        age: u32,
        
        #[arg(short, long, value_delimiter = ',')]
        tests: Vec<String>,
        
        #[arg(short, long)]
        doctor_notes: Option<String>,
        
        #[arg(short, long)]
        suggestion: Option<String>,
    },
    
    List,
    
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
    
    Search {
        #[arg(short, long)]
        name: Option<String>,
        
        #[arg(long)]
        start: Option<String>,
        
        #[arg(long)]
        end: Option<String>,
    },
    
    UpdateDoctor {
        #[arg(short, long)]
        id: Uuid,
        
        #[arg(short, long)]
        notes: Option<String>,
        
        #[arg(short, long)]
        suggestion: Option<String>,
    },
    
    Delete {
        #[arg(short, long)]
        id: Uuid,
    },
}

fn parse_test_result(s: &str) -> Result<(String, TestResultValue, Option<String>), String> {
    let parts: Vec<&str> = s.split(':').collect();
    if parts.len() < 2 {
        return Err(format!("格式错误: {}，应为 项目名:数值[:单位] 或 项目名:结果文字", s));
    }
    
    let name = parts[0].trim().to_string();
    let value_str = parts[1].trim();
    
    let value = if let Ok(num) = value_str.parse::<f64>() {
        TestResultValue::Quantitative(num)
    } else {
        TestResultValue::Qualitative(value_str.to_string())
    };
    
    let unit = if parts.len() > 2 {
        Some(parts[2].trim().to_string())
    } else {
        None
    };
    
    Ok((name, value, unit))
}

fn print_report(report: &Report) {
    println!("========================================");
    println!("报告ID: {}", report.id);
    println!("报告日期: {}", report.report_date);
    println!("----------------------------------------");
    println!("患者信息:");
    println!("  姓名: {}", report.patient.name);
    println!("  性别: {}", report.patient.gender.as_str());
    println!("  年龄: {}岁", report.patient.age);
    println!("----------------------------------------");
    println!("检验项目:");
    println!("{:<20} {:<15} {:<20} {}", "项目名称", "结果", "参考范围", "异常");
    println!("{}", "-".repeat(70));
    
    for item in &report.items {
        let result_str = match &item.value {
            TestResultValue::Quantitative(v) => {
                let unit = item.unit.as_deref().unwrap_or("");
                format!("{}{}", v, unit)
            }
            TestResultValue::Qualitative(s) => s.clone(),
        };
        
        let ref_text = item.reference_text.as_deref().unwrap_or("-");
        let ab_str = item.abnormality.as_ref().map(|a| a.as_str()).unwrap_or("");
        
        println!("{:<20} {:<15} {:<20} {}", item.name, result_str, ref_text, ab_str);
    }
    
    println!("----------------------------------------");
    println!("结论: {}", report.conclusion);
    
    if let Some(notes) = &report.doctor_notes {
        println!("医生备注: {}", notes);
    }
    if let Some(suggestion) = &report.diagnosis_suggestion {
        println!("诊断建议: {}", suggestion);
    }
    println!("========================================");
    println!();
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Args::parse();
    let client = Client::new();
    let base_url = args.server.trim_end_matches('/');
    
    match args.command {
        Commands::Health => {
            let resp = client.get(format!("{}/health", base_url)).send()?;
            println!("{}", resp.text()?);
        }
        
        Commands::Create { name, gender, age, tests, doctor_notes, suggestion } => {
            let gender_enum = match gender.to_lowercase().as_str() {
                "male" | "男" | "m" => Gender::Male,
                "female" | "女" | "f" => Gender::Female,
                _ => return Err("性别必须为 male/男 或 female/女".into()),
            };
            
            let mut test_results = Vec::new();
            for test_str in &tests {
                let (name, value, unit) = parse_test_result(test_str)?;
                test_results.push(TestInput { name, value, unit });
            }
            
            let input = ReportInput {
                patient_name: name,
                gender: gender_enum,
                age,
                report_date: Some(Utc::now().date_naive()),
                test_results,
                doctor_notes,
                diagnosis_suggestion: suggestion,
            };
            
            let resp = client
                .post(format!("{}/reports", base_url))
                .json(&input)
                .send()?;
            
            let report: Report = resp.json()?;
            println!("报告创建成功!");
            print_report(&report);
        }
        
        Commands::List => {
            let resp = client.get(format!("{}/reports", base_url)).send()?;
            let reports: Vec<Report> = resp.json()?;
            println!("共找到 {} 份报告:", reports.len());
            for report in reports {
                print_report(&report);
            }
        }
        
        Commands::Get { id } => {
            let resp = client.get(format!("{}/reports/{}", base_url, id)).send()?;
            let report: Option<Report> = resp.json()?;
            match report {
                Some(r) => print_report(&r),
                None => println!("未找到报告 ID: {}", id),
            }
        }
        
        Commands::Search { name, start, end } => {
            let mut url = format!("{}/reports/search?", base_url);
            let mut first = true;
            
            if let Some(n) = &name {
                if !first { url.push('&'); }
                url.push_str(&format!("name={}", n));
                first = false;
            }
            if let Some(s) = &start {
                if !first { url.push('&'); }
                url.push_str(&format!("start_date={}", s));
                first = false;
            }
            if let Some(e) = &end {
                if !first { url.push('&'); }
                url.push_str(&format!("end_date={}", e));
            }
            
            let resp = client.get(url).send()?;
            let reports: Vec<Report> = resp.json()?;
            println!("搜索结果: 找到 {} 份报告", reports.len());
            for report in reports {
                print_report(&report);
            }
        }
        
        Commands::UpdateDoctor { id, notes, suggestion } => {
            let body = serde_json::json!({
                "doctor_notes": notes,
                "diagnosis_suggestion": suggestion,
            });
            
            let resp = client
                .put(format!("{}/reports/{}/doctor", base_url, id))
                .json(&body)
                .send()?;
            
            let report: Option<Report> = resp.json()?;
            match report {
                Some(r) => {
                    println!("医生信息已更新!");
                    print_report(&r);
                }
                None => println!("未找到报告 ID: {}", id),
            }
        }
        
        Commands::Delete { id } => {
            let resp = client.delete(format!("{}/reports/{}", base_url, id)).send()?;
            if resp.status().is_success() {
                println!("报告已删除 ID: {}", id);
            } else {
                println!("删除失败，状态码: {}", resp.status());
            }
        }
    }
    
    Ok(())
}
