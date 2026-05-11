use anyhow::{Context, Result};
use chrono::NaiveDate;
use clap::{Parser, Subcommand};
use invoice_core::models::{CreateInvoiceRequest, Invoice, InvoiceStatus};
use reqwest::blocking::Client;
use rust_decimal::Decimal;
use serde::Deserialize;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(long, default_value = "http://127.0.0.1:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Create {
        #[arg(long)]
        invoice_code: String,
        #[arg(long)]
        invoice_number: String,
        #[arg(long)]
        issue_date: String,
        #[arg(long)]
        buyer_name: String,
        #[arg(long)]
        amount: Decimal,
        #[arg(long)]
        tax: Option<Decimal>,
        #[arg(long)]
        tax_rate: Option<Decimal>,
    },
    List,
    Get {
        #[arg(long)]
        id: String,
    },
    Search {
        #[arg(long)]
        buyer_name: Option<String>,
        #[arg(long)]
        start_date: Option<String>,
        #[arg(long)]
        end_date: Option<String>,
        #[arg(long)]
        status: Option<String>,
    },
    Void {
        #[arg(long)]
        id: String,
    },
    Red冲 {
        #[arg(long)]
        id: String,
    },
    Chain {
        #[arg(long)]
        id: String,
    },
}

#[derive(Debug, Deserialize)]
struct ErrorResponse {
    error: String,
}

fn print_invoice(inv: &Invoice) {
    println!("=====================================================================");
    println!("发票ID:     {}", inv.id);
    println!("发票代码:   {}", inv.invoice_code);
    println!("发票号码:   {}", inv.invoice_number);
    println!("开票日期:   {}", inv.issue_date);
    println!("购买方:     {}", inv.buyer_name);
    println!("金额:       {}", inv.amount);
    println!("税额:       {}", inv.tax);
    println!("税率:       {}%", inv.tax_rate * Decimal::from(100));
    println!("状态:       {}", status_display(inv.status));
    println!("是否红冲:   {}", if inv.is_red冲 { "是" } else { "否" });
    if let Some(oid) = inv.original_invoice_id {
        println!("原发票ID:   {}", oid);
    }
    if let Some(rid) = inv.red冲_invoice_id {
        println!("红冲发票ID: {}", rid);
    }
    println!("创建时间:   {}", inv.created_at.format("%Y-%m-%d %H:%M:%S"));
    println!("更新时间:   {}", inv.updated_at.format("%Y-%m-%d %H:%M:%S"));
    println!("=====================================================================");
}

fn status_display(status: InvoiceStatus) -> &'static str {
    match status {
        InvoiceStatus::Normal => "正常",
        InvoiceStatus::Voided => "已作废",
        InvoiceStatus::Red冲ed => "已红冲",
    }
}

fn handle_api_error(res: reqwest::blocking::Response) -> Result<()> {
    let status = res.status();
    if let Ok(err) = res.json::<ErrorResponse>() {
        anyhow::bail!("API错误 ({}): {}", status, err.error);
    }
    anyhow::bail!("API错误: {}", status);
}

fn main() -> Result<()> {
    let args = Args::parse();
    let client = Client::new();
    let base_url = args.server.trim_end_matches('/').to_string();

    match args.command {
        Commands::Create {
            invoice_code,
            invoice_number,
            issue_date,
            buyer_name,
            amount,
            tax,
            tax_rate,
        } => {
            let issue_date_parsed = NaiveDate::parse_from_str(&issue_date, "%Y-%m-%d")
                .with_context(|| format!("无效的日期格式: {}, 请使用 YYYY-MM-DD", issue_date))?;

            let req = CreateInvoiceRequest {
                invoice_code,
                invoice_number,
                issue_date: issue_date_parsed,
                buyer_name,
                amount,
                tax,
                tax_rate,
            };

            let res = client
                .post(format!("{}/invoices", base_url))
                .json(&req)
                .send()?;

            if !res.status().is_success() {
                return handle_api_error(res);
            }

            let inv: Invoice = res.json()?;
            println!("发票创建成功!");
            print_invoice(&inv);
        }

        Commands::List => {
            let res = client.get(format!("{}/invoices", base_url)).send()?;

            if !res.status().is_success() {
                return handle_api_error(res);
            }

            let invoices: Vec<Invoice> = res.json()?;
            if invoices.is_empty() {
                println!("没有找到发票记录");
            } else {
                println!("共找到 {} 张发票:", invoices.len());
                for inv in &invoices {
                    print_invoice(inv);
                }
            }
        }

        Commands::Get { id } => {
            let res = client.get(format!("{}/invoices/{}", base_url, id)).send()?;

            if !res.status().is_success() {
                return handle_api_error(res);
            }

            let inv: Invoice = res.json()?;
            print_invoice(&inv);
        }

        Commands::Search {
            buyer_name,
            start_date,
            end_date,
            status,
        } => {
            let mut url = format!("{}/invoices/search", base_url);
            let mut first = true;

            if let Some(bn) = buyer_name {
                url.push_str(if first { "?" } else { "&" });
                url.push_str(&format!("buyer_name={}", urlencoding::encode(&bn)));
                first = false;
            }

            if let Some(sd) = start_date {
                url.push_str(if first { "?" } else { "&" });
                url.push_str(&format!("start_date={}", sd));
                first = false;
            }

            if let Some(ed) = end_date {
                url.push_str(if first { "?" } else { "&" });
                url.push_str(&format!("end_date={}", ed));
                first = false;
            }

            if let Some(st) = status {
                url.push_str(if first { "?" } else { "&" });
                url.push_str(&format!("status={}", st));
            }

            let res = client.get(&url).send()?;

            if !res.status().is_success() {
                return handle_api_error(res);
            }

            let invoices: Vec<Invoice> = res.json()?;
            if invoices.is_empty() {
                println!("没有找到匹配的发票");
            } else {
                println!("共找到 {} 张匹配的发票:", invoices.len());
                for inv in &invoices {
                    print_invoice(inv);
                }
            }
        }

        Commands::Void { id } => {
            let res = client
                .post(format!("{}/invoices/{}/void", base_url, id))
                .send()?;

            if !res.status().is_success() {
                return handle_api_error(res);
            }

            let inv: Invoice = res.json()?;
            println!("发票作废成功!");
            print_invoice(&inv);
        }

        Commands::Red冲 { id } => {
            let res = client
                .post(format!("{}/invoices/{}/red冲", base_url, id))
                .send()?;

            if !res.status().is_success() {
                return handle_api_error(res);
            }

            let inv: Invoice = res.json()?;
            println!("发票红冲成功! 新生成的红冲发票:");
            print_invoice(&inv);
        }

        Commands::Chain { id } => {
            let res = client
                .get(format!("{}/invoices/{}/chain", base_url, id))
                .send()?;

            if !res.status().is_success() {
                return handle_api_error(res);
            }

            #[derive(Debug, Deserialize)]
            struct ChainResponse {
                original_invoice: Option<Invoice>,
                current_invoice: Invoice,
                red冲_invoice: Option<Invoice>,
            }

            let chain: ChainResponse = res.json()?;

            if let Some(orig) = chain.original_invoice {
                println!("=== 原始发票 ===");
                print_invoice(&orig);
            }

            println!("=== 当前发票 ===");
            print_invoice(&chain.current_invoice);

            if let Some(red冲) = chain.red冲_invoice {
                println!("=== 红冲发票 ===");
                print_invoice(&red冲);
            }
        }
    }

    Ok(())
}
