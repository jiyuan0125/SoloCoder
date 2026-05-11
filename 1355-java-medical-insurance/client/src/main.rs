use clap::{Parser, Subcommand};
use medical_core::{
    Hospital, HospitalLevel, InsuredStatus, ItemCategory, MedicalItem, PatientYearlyState,
    PolicyConfig, SettlementRequest, SettlementResult, VisitItem, VisitType,
};
use reqwest::Client;
use serde::Serialize;
use std::str::FromStr;
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(name = "medical-insurance-client")]
struct Args {
    #[arg(long, env = "SERVER_URL", default_value = "http://localhost:8080")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Hospitals,
    Items,
    Policy,
    PatientStates,
    CreatePatient {
        patient_id: String,
        policy_year: u32,
    },
    Settle {
        patient_id: String,
        hospital_id: String,
        #[arg(long = "type")]
        visit_type: String,
        #[arg(long = "insured")]
        insured_status: String,
        #[arg(long = "year", default_value_t = 2026)]
        policy_year: u32,
        #[arg(num_args = 1.., value_name = "ITEM_ID:QUANTITY")]
        items: Vec<String>,
    },
    Settlements,
}

fn format_cents(cents: u64) -> String {
    format!("{:.2}元", cents as f64 / 100.0)
}

fn parse_hospital_level(category: &ItemCategory) -> String {
    match category {
        ItemCategory::ClassA => "甲类".to_string(),
        ItemCategory::ClassB { self_pay_ratio_percent } => {
            format!("乙类(自负{}%)", self_pay_ratio_percent)
        }
        ItemCategory::ClassC => "丙类".to_string(),
    }
}

fn format_hospital_level(level: &HospitalLevel) -> String {
    match level {
        HospitalLevel::Community => "社区医院".to_string(),
        HospitalLevel::Level2 => "二级医院".to_string(),
        HospitalLevel::Level3 => "三级医院".to_string(),
    }
}

fn format_visit_type(vt: &VisitType) -> String {
    match vt {
        VisitType::Outpatient => "普通门诊".to_string(),
        VisitType::Inpatient => "住院".to_string(),
        VisitType::SpecialDiseaseOutpatient => "特殊病种门诊".to_string(),
    }
}

fn parse_visit_type(s: &str) -> Result<VisitType, String> {
    match s.to_lowercase().as_str() {
        "outpatient" | "门诊" => Ok(VisitType::Outpatient),
        "inpatient" | "住院" => Ok(VisitType::Inpatient),
        "special" | "special-disease" | "特病" => Ok(VisitType::SpecialDiseaseOutpatient),
        _ => Err(format!("无效就诊类型: {}", s)),
    }
}

fn parse_insured_status(s: &str) -> Result<InsuredStatus, String> {
    match s.to_lowercase().as_str() {
        "working" | "在职" => Ok(InsuredStatus::Working),
        "retired" | "退休" => Ok(InsuredStatus::Retired),
        _ => Err(format!("无效参保状态: {}", s)),
    }
}

fn parse_item_spec(s: &str) -> Result<(String, u32), String> {
    let parts: Vec<_> = s.splitn(2, ':').collect();
    if parts.len() != 2 {
        return Err(format!("格式错误，应为 ITEM_ID:QUANTITY: {}", s));
    }
    let id = parts[0].to_string();
    let qty = u32::from_str(parts[1])
        .map_err(|e| format!("数量解析失败: {}", e))?;
    Ok((id, qty))
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    let client = Client::new();
    let base = args.server_url.trim_end_matches('/').to_string();

    match args.command {
        Commands::Hospitals => {
            let url = format!("{}/hospitals", base);
            let hospitals: Vec<Hospital> = client
                .get(&url)
                .send()
                .await
                .unwrap()
                .json()
                .await
                .unwrap();
            println!("\n医院列表:");
            for h in hospitals {
                println!(
                    "  [{}] {} - {}",
                    h.id,
                    h.name,
                    format_hospital_level(&h.level)
                );
            }
        }

        Commands::Items => {
            let url = format!("{}/medical-items", base);
            let items: Vec<MedicalItem> = client
                .get(&url)
                .send()
                .await
                .unwrap()
                .json()
                .await
                .unwrap();
            println!("\n药品/诊疗项目:");
            for item in items {
                println!(
                    "  [{}] {} - {} 单价: {}",
                    item.id,
                    item.name,
                    parse_hospital_level(&item.category),
                    format_cents(item.price_cents)
                );
            }
        }

        Commands::Policy => {
            let url = format!("{}/policy", base);
            let policy: PolicyConfig = client
                .get(&url)
                .send()
                .await
                .unwrap()
                .json()
                .await
                .unwrap();
            println!("\n医保政策配置:");
            println!("  门诊起付线: {}", format_cents(policy.outpatient_deductible_cents));
            println!("  门诊封顶线: {}", format_cents(policy.outpatient_cap_cents));
            println!("  住院起付线: {}", format_cents(policy.inpatient_deductible_cents));
            println!("  住院封顶线: {}", format_cents(policy.inpatient_cap_cents));
            println!("  退休人员报销加成: +{}%", policy.retired_bonus_percent);
            println!("\n报销比例表:");
            for (level, map) in policy.reimbursement_ratios {
                println!("  {}:", format_hospital_level(&level));
                for (vt, ratio) in map {
                    println!("    - {}: {}%", format_visit_type(&vt), ratio);
                }
            }
        }

        Commands::PatientStates => {
            let url = format!("{}/patient-states", base);
            let states: Vec<PatientYearlyState> = client
                .get(&url)
                .send()
                .await
                .unwrap()
                .json()
                .await
                .unwrap();
            if states.is_empty() {
                println!("\n暂无患者年度状态数据");
            } else {
                println!("\n患者年度状态:");
                for s in states {
                    println!("  患者: {}, 年度: {}", s.patient_id, s.policy_year);
                    println!(
                        "    门诊累计报销: {}",
                        format_cents(s.outpatient_reimbursed_cents)
                    );
                    println!(
                        "    门诊起付线已付: {}",
                        format_cents(s.outpatient_deductible_met_cents)
                    );
                    println!(
                        "    住院累计报销: {}",
                        format_cents(s.inpatient_reimbursed_cents)
                    );
                    println!(
                        "    住院起付线已付: {}",
                        format_cents(s.inpatient_deductible_met_cents)
                    );
                }
            }
        }

        Commands::CreatePatient { patient_id, policy_year } => {
            let url = format!("{}/patient-states", base);
            #[derive(Serialize)]
            struct Req {
                patient_id: String,
                policy_year: u32,
            }
            let resp = client
                .post(&url)
                .json(&Req { patient_id, policy_year })
                .send()
                .await
                .unwrap();
            if resp.status().is_success() {
                println!("\n患者年度状态已创建");
            } else {
                eprintln!("\n创建失败: {}", resp.status());
            }
        }

        Commands::Settle {
            patient_id,
            hospital_id,
            visit_type,
            insured_status,
            policy_year,
            items,
        } => {
            let hospital_id = Uuid::from_str(&hospital_id)
                .expect("医院ID格式错误");
            let vt = parse_visit_type(&visit_type).expect("就诊类型无效");
            let status = parse_insured_status(&insured_status).expect("参保状态无效");

            let mut visit_items = Vec::new();
            for spec in items {
                let (id, qty) = parse_item_spec(&spec).expect("项目格式错误");
                visit_items.push(VisitItem {
                    item_id: id,
                    quantity: qty,
                });
            }

            let request = SettlementRequest {
                patient_id,
                hospital_id,
                visit_type: vt,
                insured_status: status,
                items: visit_items,
                policy_year,
            };

            let url = format!("{}/settle", base);
            let resp = client.post(&url).json(&request).send().await.unwrap();

            if !resp.status().is_success() {
                let err_text = resp.text().await.unwrap();
                eprintln!("\n结算失败: {}", err_text);
                std::process::exit(1);
            }

            let result: SettlementResult = resp.json().await.unwrap();

            println!("\n========== 医保结算结果 ==========");
            println!("请求ID: {}", result.request_id);
            println!("患者ID: {}", result.patient_id);
            println!("就诊类型: {}", format_visit_type(&result.visit_type));
            println!();
            println!("总费用: {}", format_cents(result.total_cost_cents));
            println!("乙类先自付: {}", format_cents(result.total_self_pay_first_cents));
            println!("纳入报销范围: {}", format_cents(result.total_reimbursable_cents));
            println!();
            println!("医保报销: {}", format_cents(result.total_reimbursed_cents));
            println!("个人自付: {}", format_cents(result.total_final_self_pay_cents));
            println!();
            println!("年度门诊累计报销: {}", format_cents(result.year_to_date_outpatient_reimbursed_cents));
            println!("年度住院累计报销: {}", format_cents(result.year_to_date_inpatient_reimbursed_cents));
            println!();
            println!("---------- 明细 ----------");
            for detail in &result.details {
                println!(
                    "\n{} ({})\n  单价: {} x {} = {}\n  纳入报销: {} | 先自付: {}\n  医保报销: {} | 个人自付: {}",
                    detail.name,
                    parse_hospital_level(&detail.category),
                    format_cents(detail.unit_price_cents),
                    detail.quantity,
                    format_cents(detail.total_cost_cents),
                    format_cents(detail.reimbursable_cents),
                    format_cents(detail.self_pay_first_cents),
                    format_cents(detail.final_reimbursed_cents),
                    format_cents(detail.final_self_pay_cents)
                );
            }
        }

        Commands::Settlements => {
            let url = format!("{}/settlements", base);
            let list: Vec<SettlementResult> = client
                .get(&url)
                .send()
                .await
                .unwrap()
                .json()
                .await
                .unwrap();
            if list.is_empty() {
                println!("\n暂无结算记录");
            } else {
                println!("\n结算历史:");
                for r in list {
                    println!(
                        "\n[{}] 患者: {} | {}",
                        r.request_id,
                        r.patient_id,
                        format_visit_type(&r.visit_type)
                    );
                    println!(
                        "  总费用: {} | 报销: {} | 自付: {}",
                        format_cents(r.total_cost_cents),
                        format_cents(r.total_reimbursed_cents),
                        format_cents(r.total_final_self_pay_cents)
                    );
                }
            }
        }
    }
}
