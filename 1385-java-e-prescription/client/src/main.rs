use clap::{Parser, Subcommand};
use eprescription_core::{
    CreatePrescriptionRequest, DispensePrescriptionRequest, MedicineRecordRequest,
    ReviewPrescriptionRequest,
};
use reqwest::Client;
use std::error::Error;
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://localhost:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Doctors,
    Patients,
    Medicines,
    Prescriptions,
    Prescription {
        id: String,
    },
    CreatePrescription {
        doctor_id: String,
        doctor_name: String,
        patient_id: String,
        patient_name: String,
        medicines: String,
    },
    ReviewPrescription {
        id: String,
        approved: bool,
        notes: Option<String>,
    },
    DispensePrescription {
        id: String,
    },
    DoctorStats,
    MedicineStats,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let args = Args::parse();
    let client = Client::new();

    match args.command {
        Commands::Doctors => {
            let doctors: serde_json::Value = client
                .get(&format!("{}/doctors", args.server_url))
                .send()
                .await?
                .json()
                .await?;
            println!("{}", serde_json::to_string_pretty(&doctors)?);
        }
        Commands::Patients => {
            let patients: serde_json::Value = client
                .get(&format!("{}/patients", args.server_url))
                .send()
                .await?
                .json()
                .await?;
            println!("{}", serde_json::to_string_pretty(&patients)?);
        }
        Commands::Medicines => {
            let medicines: serde_json::Value = client
                .get(&format!("{}/medicines", args.server_url))
                .send()
                .await?
                .json()
                .await?;
            println!("{}", serde_json::to_string_pretty(&medicines)?);
        }
        Commands::Prescriptions => {
            let prescriptions: serde_json::Value = client
                .get(&format!("{}/prescriptions", args.server_url))
                .send()
                .await?
                .json()
                .await?;
            println!("{}", serde_json::to_string_pretty(&prescriptions)?);
        }
        Commands::Prescription { id } => {
            let prescription: serde_json::Value = client
                .get(&format!("{}/prescriptions/{}", args.server_url, id))
                .send()
                .await?
                .json()
                .await?;
            println!("{}", serde_json::to_string_pretty(&prescription)?);
        }
        Commands::CreatePrescription {
            doctor_id,
            doctor_name,
            patient_id,
            patient_name,
            medicines,
        } => {
            let medicine_records: Vec<MedicineRecordRequest> = serde_json::from_str(&medicines)?;

            let request = CreatePrescriptionRequest {
                doctor_id: Uuid::parse_str(&doctor_id)?,
                doctor_name,
                patient_id: Uuid::parse_str(&patient_id)?,
                patient_name,
                medicines: medicine_records,
            };

            let response = client
                .post(&format!("{}/prescriptions", args.server_url))
                .json(&request)
                .send()
                .await?;

            let result: serde_json::Value = response.json().await?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        Commands::ReviewPrescription { id, approved, notes } => {
            let request = ReviewPrescriptionRequest {
                prescription_id: Uuid::parse_str(&id)?,
                approved,
                notes,
            };

            let response = client
                .post(&format!("{}/prescriptions/review", args.server_url))
                .json(&request)
                .send()
                .await?;

            let result: serde_json::Value = response.json().await?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        Commands::DispensePrescription { id } => {
            let request = DispensePrescriptionRequest {
                prescription_id: Uuid::parse_str(&id)?,
            };

            let response = client
                .post(&format!("{}/prescriptions/dispense", args.server_url))
                .json(&request)
                .send()
                .await?;

            let result: serde_json::Value = response.json().await?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        Commands::DoctorStats => {
            let stats: serde_json::Value = client
                .get(&format!("{}/doctors/stats", args.server_url))
                .send()
                .await?
                .json()
                .await?;
            println!("{}", serde_json::to_string_pretty(&stats)?);
        }
        Commands::MedicineStats => {
            let stats: serde_json::Value = client
                .get(&format!("{}/medicines/stats", args.server_url))
                .send()
                .await?
                .json()
                .await?;
            println!("{}", serde_json::to_string_pretty(&stats)?);
        }
    }

    Ok(())
}
