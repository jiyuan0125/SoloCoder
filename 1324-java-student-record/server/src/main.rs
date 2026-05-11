use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use clap::Parser;
use serde::Serialize;
use std::collections::HashMap;
use std::sync::Arc;
use tower_http::cors::{Any, CorsLayer};
use student_record_core::*;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, default_value_t = 3000, env = "PORT")]
    port: u16,
    
    #[arg(short, long, default_value = "127.0.0.1", env = "HOST")]
    host: String,
}

#[derive(Clone)]
struct AppState {
    storage: Arc<InMemoryStorage>,
}

#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
}

struct AppError(StudentRecordError);

impl IntoResponse for AppError {
    fn into_response(self) -> axum::response::Response {
        let status = match &self.0 {
            StudentRecordError::StudentNotFound(_) => StatusCode::NOT_FOUND,
            StudentRecordError::MajorNotFound(_) => StatusCode::NOT_FOUND,
            StudentRecordError::CourseNotFound(_) => StatusCode::NOT_FOUND,
            StudentRecordError::AlreadyInTargetMajor => StatusCode::BAD_REQUEST,
            StudentRecordError::TransferInProgress => StatusCode::CONFLICT,
            StudentRecordError::InvalidInput(_) => StatusCode::BAD_REQUEST,
            StudentRecordError::InternalError(_) => StatusCode::INTERNAL_SERVER_ERROR,
        };
        
        (
            status,
            Json(ErrorResponse {
                error: self.0.to_string(),
            }),
        )
            .into_response()
    }
}

type ServerResult<T> = std::result::Result<T, AppError>;

impl From<StudentRecordError> for AppError {
    fn from(err: StudentRecordError) -> Self {
        AppError(err)
    }
}

async fn list_majors(State(state): State<AppState>) -> Json<Vec<Major>> {
    Json(state.storage.list_majors().await)
}

async fn get_major(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> ServerResult<Json<Major>> {
    let major = state.storage.get_major(&id).await?;
    Ok(Json(major))
}

async fn list_students(State(state): State<AppState>) -> Json<Vec<Student>> {
    Json(state.storage.list_students().await)
}

async fn get_student(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> ServerResult<Json<Student>> {
    let student = state.storage.get_student(&id).await?;
    Ok(Json(student))
}

async fn get_student_credits(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> ServerResult<Json<CreditSummary>> {
    let summary = state.storage.get_student_credit_summary(&id).await?;
    Ok(Json(summary))
}

async fn get_student_gap(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> ServerResult<Json<GraduationGap>> {
    let gap = state.storage.get_student_graduation_gap(&id).await?;
    Ok(Json(gap))
}

async fn process_transfer(
    State(state): State<AppState>,
    Json(request): Json<TransferRequest>,
) -> ServerResult<Json<TransferResult>> {
    let storage = Arc::clone(&state.storage);
    let result = storage.process_transfer_request(request).await?;
    Ok(Json(result))
}

async fn create_major(
    State(state): State<AppState>,
    Json(major): Json<Major>,
) -> ServerResult<StatusCode> {
    state.storage.add_major(major).await?;
    Ok(StatusCode::CREATED)
}

async fn create_student(
    State(state): State<AppState>,
    Json(student): Json<Student>,
) -> ServerResult<StatusCode> {
    state.storage.add_student(student).await?;
    Ok(StatusCode::CREATED)
}

async fn seed_data(storage: Arc<InMemoryStorage>) {
    let mut cs_curriculum = HashMap::new();
    cs_curriculum.insert(
        "CS101".to_string(),
        Course {
            id: "CS101".to_string(),
            name: "计算机导论".to_string(),
            credit: 3.0,
            course_type: CourseType::Professional,
        },
    );
    cs_curriculum.insert(
        "CS201".to_string(),
        Course {
            id: "CS201".to_string(),
            name: "数据结构".to_string(),
            credit: 4.0,
            course_type: CourseType::Professional,
        },
    );
    cs_curriculum.insert(
        "CS301".to_string(),
        Course {
            id: "CS301".to_string(),
            name: "操作系统".to_string(),
            credit: 4.0,
            course_type: CourseType::Professional,
        },
    );

    let cs_major = Major {
        id: "CS".to_string(),
        name: "计算机科学与技术".to_string(),
        requirements: CreditRequirements {
            public: 40.0,
            professional: 60.0,
            elective: 20.0,
        },
        curriculum: cs_curriculum,
    };

    let mut ee_curriculum = HashMap::new();
    ee_curriculum.insert(
        "EE101".to_string(),
        Course {
            id: "EE101".to_string(),
            name: "电路原理".to_string(),
            credit: 3.0,
            course_type: CourseType::Professional,
        },
    );
    ee_curriculum.insert(
        "EE201".to_string(),
        Course {
            id: "EE201".to_string(),
            name: "模拟电子技术".to_string(),
            credit: 4.0,
            course_type: CourseType::Professional,
        },
    );
    ee_curriculum.insert(
        "EE301".to_string(),
        Course {
            id: "EE301".to_string(),
            name: "数字电子技术".to_string(),
            credit: 4.0,
            course_type: CourseType::Professional,
        },
    );

    let ee_major = Major {
        id: "EE".to_string(),
        name: "电子工程".to_string(),
        requirements: CreditRequirements {
            public: 35.0,
            professional: 65.0,
            elective: 20.0,
        },
        curriculum: ee_curriculum,
    };

    let _ = storage.add_major(cs_major).await;
    let _ = storage.add_major(ee_major).await;

    let student1 = Student {
        id: "2024001".to_string(),
        name: "张三".to_string(),
        enrollment_year: 2024,
        current_major_id: "CS".to_string(),
        completed_courses: vec![
            CompletedCourse {
                course_id: "MATH101".to_string(),
                course_name: "高等数学".to_string(),
                credit: 5.0,
                original_course_type: CourseType::Public,
            },
            CompletedCourse {
                course_id: "ENG101".to_string(),
                course_name: "大学英语".to_string(),
                credit: 4.0,
                original_course_type: CourseType::Public,
            },
            CompletedCourse {
                course_id: "CS101".to_string(),
                course_name: "计算机导论".to_string(),
                credit: 3.0,
                original_course_type: CourseType::Professional,
            },
            CompletedCourse {
                course_id: "CS201".to_string(),
                course_name: "数据结构".to_string(),
                credit: 4.0,
                original_course_type: CourseType::Professional,
            },
            CompletedCourse {
                course_id: "CS_ELEC1".to_string(),
                course_name: "人工智能导论".to_string(),
                credit: 3.0,
                original_course_type: CourseType::Elective,
            },
        ],
    };

    let student2 = Student {
        id: "2024002".to_string(),
        name: "李四".to_string(),
        enrollment_year: 2024,
        current_major_id: "EE".to_string(),
        completed_courses: vec![
            CompletedCourse {
                course_id: "MATH101".to_string(),
                course_name: "高等数学".to_string(),
                credit: 5.0,
                original_course_type: CourseType::Public,
            },
            CompletedCourse {
                course_id: "PHY101".to_string(),
                course_name: "大学物理".to_string(),
                credit: 4.0,
                original_course_type: CourseType::Public,
            },
            CompletedCourse {
                course_id: "EE101".to_string(),
                course_name: "电路原理".to_string(),
                credit: 3.0,
                original_course_type: CourseType::Professional,
            },
        ],
    };

    let _ = storage.add_student(student1).await;
    let _ = storage.add_student(student2).await;
    
    println!("已初始化示例数据：2个专业，2个学生");
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    
    let storage = InMemoryStorage::new();
    seed_data(Arc::clone(&storage)).await;
    
    let state = AppState { storage };
    
    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/api/majors", get(list_majors).post(create_major))
        .route("/api/majors/:id", get(get_major))
        .route("/api/students", get(list_students).post(create_student))
        .route("/api/students/:id", get(get_student))
        .route("/api/students/:id/credits", get(get_student_credits))
        .route("/api/students/:id/gap", get(get_student_gap))
        .route("/api/transfer", post(process_transfer))
        .layer(cors)
        .with_state(state);

    let addr = format!("{}:{}", args.host, args.port);
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    
    println!("学生学籍管理系统服务已启动，监听地址：{}", addr);
    println!("API 端点：");
    println!("  GET  /api/majors          - 列出所有专业");
    println!("  POST /api/majors          - 创建专业");
    println!("  GET  /api/majors/:id      - 获取专业详情");
    println!("  GET  /api/students        - 列出所有学生");
    println!("  POST /api/students        - 创建学生");
    println!("  GET  /api/students/:id    - 获取学生详情");
    println!("  GET  /api/students/:id/credits - 获取学生已修学分");
    println!("  GET  /api/students/:id/gap     - 获取学生毕业学分缺口");
    println!("  POST /api/transfer        - 提交转专业申请");
    
    axum::serve(listener, app).await.unwrap();
}
