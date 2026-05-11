use axum::{
    extract::{Json, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Router,
};
use clap::Parser;
use medical_core::{
    CreatePatientState, Hospital, HospitalLevel, ItemCategory, MedicalItem, PatientYearlyState,
    PolicyConfig, SettlementRequest, SettlementResult,
};
use serde::Serialize;
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use tower_http::cors::CorsLayer;
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(name = "medical-insurance-server")]
struct Args {
    #[arg(long, env = "SERVER_PORT", default_value_t = 8080)]
    port: u16,

    #[arg(long, env = "SERVER_HOST", default_value = "0.0.0.0")]
    host: String,
}

struct AppState {
    hospitals: Mutex<HashMap<Uuid, Hospital>>,
    medical_items: Mutex<HashMap<String, MedicalItem>>,
    policy: PolicyConfig,
    patient_states: Mutex<HashMap<(String, u32), PatientYearlyState>>,
    settlement_history: Mutex<HashMap<Uuid, SettlementResult>>,
}

type SharedState = Arc<AppState>;

#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
}

fn init_seed_data() -> (HashMap<Uuid, Hospital>, HashMap<String, MedicalItem>) {
    let mut hospitals = HashMap::new();

    let h1 = Hospital {
        id: Uuid::new_v4(),
        name: "北京协和医院".to_string(),
        level: HospitalLevel::Level3,
    };
    hospitals.insert(h1.id, h1);

    let h2 = Hospital {
        id: Uuid::new_v4(),
        name: "北京市第二医院".to_string(),
        level: HospitalLevel::Level2,
    };
    hospitals.insert(h2.id, h2);

    let h3 = Hospital {
        id: Uuid::new_v4(),
        name: "朝阳区社区卫生服务中心".to_string(),
        level: HospitalLevel::Community,
    };
    hospitals.insert(h3.id, h3);

    let mut items = HashMap::new();

    items.insert(
        "MED001".to_string(),
        MedicalItem {
            id: "MED001".to_string(),
            name: "阿莫西林胶囊".to_string(),
            category: ItemCategory::ClassA,
            price_cents: 2500,
        },
    );

    items.insert(
        "MED002".to_string(),
        MedicalItem {
            id: "MED002".to_string(),
            name: "头孢克肟分散片".to_string(),
            category: ItemCategory::ClassB {
                self_pay_ratio_percent: 10,
            },
            price_cents: 4800,
        },
    );

    items.insert(
        "MED003".to_string(),
        MedicalItem {
            id: "MED003".to_string(),
            name: "进口靶向药物".to_string(),
            category: ItemCategory::ClassC,
            price_cents: 500000,
        },
    );

    items.insert(
        "TRT001".to_string(),
        MedicalItem {
            id: "TRT001".to_string(),
            name: "血常规检查".to_string(),
            category: ItemCategory::ClassA,
            price_cents: 1500,
        },
    );

    items.insert(
        "TRT002".to_string(),
        MedicalItem {
            id: "TRT002".to_string(),
            name: "CT扫描".to_string(),
            category: ItemCategory::ClassB {
                self_pay_ratio_percent: 20,
            },
            price_cents: 35000,
        },
    );

    items.insert(
        "TRT003".to_string(),
        MedicalItem {
            id: "TRT003".to_string(),
            name: "住院床位费(日)".to_string(),
            category: ItemCategory::ClassA,
            price_cents: 8000,
        },
    );

    (hospitals, items)
}

async fn list_hospitals(State(state): State<SharedState>) -> impl IntoResponse {
    let hospitals = state.hospitals.lock().unwrap();
    let list: Vec<_> = hospitals.values().cloned().collect();
    Json(list)
}

async fn list_medical_items(State(state): State<SharedState>) -> impl IntoResponse {
    let items = state.medical_items.lock().unwrap();
    let list: Vec<_> = items.values().cloned().collect();
    Json(list)
}

async fn get_policy(State(state): State<SharedState>) -> impl IntoResponse {
    Json(state.policy.clone())
}

async fn create_patient_state(
    State(state): State<SharedState>,
    Json(payload): Json<CreatePatientState>,
) -> impl IntoResponse {
    let mut states = state.patient_states.lock().unwrap();
    let key = (payload.patient_id.clone(), payload.policy_year);
    if !states.contains_key(&key) {
        states.insert(
            key,
            PatientYearlyState {
                patient_id: payload.patient_id,
                policy_year: payload.policy_year,
                ..Default::default()
            },
        );
    }
    StatusCode::CREATED
}

async fn get_patient_state(
    State(state): State<SharedState>,
) -> impl IntoResponse {
    let states = state.patient_states.lock().unwrap();
    let list: Vec<_> = states.values().cloned().collect();
    Json(list)
}

async fn settle(
    State(state): State<SharedState>,
    Json(request): Json<SettlementRequest>,
) -> impl IntoResponse {
    let hospitals = state.hospitals.lock().unwrap();
    let medical_items = state.medical_items.lock().unwrap();

    let mut patient_states = state.patient_states.lock().unwrap();
    let key = (request.patient_id.clone(), request.policy_year);
    let patient_state = patient_states
        .entry(key)
        .or_insert_with(|| PatientYearlyState {
            patient_id: request.patient_id.clone(),
            policy_year: request.policy_year,
            ..Default::default()
        });

    let ctx = medical_core::SettlementContext {
        hospitals: &hospitals,
        medical_items: &medical_items,
        policy: &state.policy,
    };

    match medical_core::process_settlement(&ctx, &request, patient_state) {
        Ok(result) => {
            let mut history = state.settlement_history.lock().unwrap();
            history.insert(result.request_id, result.clone());
            (StatusCode::OK, Json(result)).into_response()
        }
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: e.to_string(),
            }),
        )
            .into_response(),
    }
}

async fn list_settlements(State(state): State<SharedState>) -> impl IntoResponse {
    let history = state.settlement_history.lock().unwrap();
    let list: Vec<_> = history.values().cloned().collect();
    Json(list)
}

#[tokio::main]
async fn main() {
    let args = Args::parse();

    let (hospitals, medical_items) = init_seed_data();

    println!("初始化医院数据:");
    for (id, h) in &hospitals {
        println!(
            "  - {} ({:?}) [ID: {}]",
            h.name, h.level, id
        );
    }

    println!("\n初始化药品/诊疗项目:");
    for (id, item) in &medical_items {
        println!(
            "  - {} ({:?}) 价格: {:.2}元 [ID: {}]",
            item.name,
            item.category,
            item.price_cents as f64 / 100.0,
            id
        );
    }

    let policy = PolicyConfig::default();
    println!("\n政策配置:");
    println!("  门诊起付线: {:.2}元", policy.outpatient_deductible_cents as f64 / 100.0);
    println!("  门诊封顶线: {:.2}元", policy.outpatient_cap_cents as f64 / 100.0);
    println!("  住院起付线: {:.2}元", policy.inpatient_deductible_cents as f64 / 100.0);
    println!("  住院封顶线: {:.2}元", policy.inpatient_cap_cents as f64 / 100.0);
    println!("  退休人员报销比例加成: +{}%", policy.retired_bonus_percent);

    let state = Arc::new(AppState {
        hospitals: Mutex::new(hospitals),
        medical_items: Mutex::new(medical_items),
        policy,
        patient_states: Mutex::new(HashMap::new()),
        settlement_history: Mutex::new(HashMap::new()),
    });

    let app = Router::new()
        .route("/hospitals", get(list_hospitals))
        .route("/medical-items", get(list_medical_items))
        .route("/policy", get(get_policy))
        .route("/patient-states", post(create_patient_state).get(get_patient_state))
        .route("/settle", post(settle))
        .route("/settlements", get(list_settlements))
        .with_state(state)
        .layer(CorsLayer::permissive());

    let addr = format!("{}:{}", args.host, args.port);
    println!("\n服务启动于: http://{}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
