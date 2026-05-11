use anyhow::{Context, Result};
use chrono::Local;
use clap::{Parser, Subcommand};
use insurance_core::{
    Beneficiary, Insured, Policy, PolicyApplication, Product, ValidationResult,
};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(long, env = "INSURANCE_SERVER", default_value = "http://127.0.0.1:8080")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Products,
    Policies,
    GetPolicy {
        policy_number: String,
    },
    Apply {
        #[arg(long)]
        product_code: String,
        #[arg(long)]
        insured_name: String,
        #[arg(long)]
        insured_id: String,
        #[arg(long)]
        birth_date: String,
        #[arg(long)]
        occupation_level: u8,
        #[arg(long, default_value_t = false)]
        has_training: bool,
        #[arg(long)]
        amount: Decimal,
        #[arg(long)]
        beneficiary: Vec<String>,
    },
}

#[derive(Debug, Serialize)]
struct ValidateRequest {
    application: PolicyApplication,
}

#[derive(Debug, Deserialize)]
struct CreatePolicyResponse {
    success: bool,
    policy: Option<Policy>,
    validation_result: ValidationResult,
}

#[tokio::main]
async fn main() -> Result<()> {
    let cli = Cli::parse();

    let client = reqwest::Client::new();

    match cli.command {
        Commands::Products => {
            let products = list_products(&client, &cli.server).await?;
            println!("=== 产品列表 ===");
            for p in products {
                println!("代码: {}", p.code);
                println!("  名称: {}", p.name);
                println!("  类型: {}", p.product_type);
                println!("  基准保费: {} 元/万元保额", p.base_premium);
                println!();
            }
        }
        Commands::Policies => {
            let policies = list_policies(&client, &cli.server).await?;
            if policies.is_empty() {
                println!("暂无保单");
            } else {
                println!("=== 保单列表 ===");
                for p in policies {
                    print_policy(&p);
                }
            }
        }
        Commands::GetPolicy { policy_number } => {
            match get_policy(&client, &cli.server, &policy_number).await? {
                Some(policy) => print_policy(&policy),
                None => println!("保单不存在"),
            }
        }
        Commands::Apply {
            product_code,
            insured_name,
            insured_id,
            birth_date,
            occupation_level,
            has_training,
            amount,
            beneficiary,
        } => {
            let birth_date = chrono::NaiveDate::parse_from_str(&birth_date, "%Y-%m-%d")
                .context("出生日期格式错误，应为 YYYY-MM-DD")?;

            let beneficiaries = parse_beneficiaries(&beneficiary)?;

            let application = PolicyApplication {
                insured: Insured {
                    name: insured_name,
                    id_card: insured_id,
                    birth_date,
                    occupation_risk_level: occupation_level,
                    has_safety_training: has_training,
                },
                product_code,
                application_date: Local::now().date_naive(),
                beneficiaries,
                insured_amount: amount,
            };

            let resp = apply_policy(&client, &cli.server, &application).await?;

            if resp.success {
                println!("投保成功！");
                if let Some(policy) = resp.policy {
                    print_policy(&policy);
                }
            } else {
                println!("投保失败，原因如下:");
                for err in &resp.validation_result.errors {
                    println!("  - {}", err);
                }
            }
        }
    }

    Ok(())
}

fn parse_beneficiaries(args: &[String]) -> Result<Vec<Beneficiary>> {
    if args.is_empty() {
        anyhow::bail!("至少需要指定一个受益人");
    }

    let mut beneficiaries = Vec::new();
    for arg in args {
        let parts: Vec<&str> = arg.split(':').collect();
        if parts.len() != 3 {
            anyhow::bail!("受益人格式错误，应为: 姓名:身份证号:比例");
        }
        let ratio: Decimal = parts[2]
            .parse()
            .context("受益人比例必须是数字")?;
        beneficiaries.push(Beneficiary {
            name: parts[0].to_string(),
            id_card: parts[1].to_string(),
            ratio,
        });
    }
    Ok(beneficiaries)
}

fn print_policy(p: &Policy) {
    println!("保单号: {}", p.policy_number);
    println!("  产品: {} ({})", p.product.name, p.product.code);
    println!("  被保险人: {} ({})", p.insured.name, p.insured.id_card);
    println!("  投保日期: {}", p.application_date);
    println!("  保额: {} 元", p.insured_amount);
    println!("  保费: {} 元", p.premium);
    println!("  受益人:");
    for b in &p.beneficiaries {
        println!("    - {} ({}) {}%", b.name, b.id_card, b.ratio);
    }
    println!();
}

async fn list_products(client: &reqwest::Client, base_url: &str) -> Result<Vec<Product>> {
    let url = format!("{}/products", base_url);
    let resp = client.get(&url).send().await.context("请求失败")?;
    let products: Vec<Product> = resp.json().await.context("解析响应失败")?;
    Ok(products)
}

async fn list_policies(client: &reqwest::Client, base_url: &str) -> Result<Vec<Policy>> {
    let url = format!("{}/policies", base_url);
    let resp = client.get(&url).send().await.context("请求失败")?;
    let policies: Vec<Policy> = resp.json().await.context("解析响应失败")?;
    Ok(policies)
}

async fn get_policy(
    client: &reqwest::Client,
    base_url: &str,
    policy_number: &str,
) -> Result<Option<Policy>> {
    let url = format!("{}/policies/{}", base_url, policy_number);
    let resp = client.get(&url).send().await.context("请求失败")?;
    if resp.status().is_success() {
        let policy: Policy = resp.json().await.context("解析响应失败")?;
        Ok(Some(policy))
    } else {
        Ok(None)
    }
}

async fn apply_policy(
    client: &reqwest::Client,
    base_url: &str,
    application: &PolicyApplication,
) -> Result<CreatePolicyResponse> {
    let url = format!("{}/policies", base_url);
    let resp = client
        .post(&url)
        .json(application)
        .send()
        .await
        .context("请求失败")?;
    let result: CreatePolicyResponse = resp.json().await.context("解析响应失败")?;
    Ok(result)
}
