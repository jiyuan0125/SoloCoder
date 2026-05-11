mod handlers;

use std::sync::Arc;
use axum::{
    Router,
    routing::{get, post},
    http::StatusCode,
    Json,
    response::IntoResponse,
    extract::State,
};
use clap::Parser;
use chrono::Utc;
use tower_http::trace::TraceLayer;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};
use course_core::{
    CourseService,
    Course,
    Student,
    CourseId,
    StudentId,
    CompletedCourse,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value_t = 3000)]
    port: u16,

    #[arg(short = 'H', long, env = "HOST", default_value = "127.0.0.1")]
    host: String,
}

#[derive(Debug)]
struct AppError {
    code: StatusCode,
    message: String,
}

impl IntoResponse for AppError {
    fn into_response(self) -> axum::response::Response {
        (
            self.code,
            Json(serde_json::json!({
                "error": self.message
            })),
        ).into_response()
    }
}

impl<E: std::fmt::Display> From<E> for AppError {
    fn from(err: E) -> Self {
        AppError {
            code: StatusCode::INTERNAL_SERVER_ERROR,
            message: err.to_string(),
        }
    }
}

fn initialize_data(service: &CourseService) {
    service.add_course(Course {
        id: CourseId("CS101".to_string()),
        name: "C语言".to_string(),
        credits: 4,
        capacity: 100,
        prerequisites: vec![],
    });

    service.add_course(Course {
        id: CourseId("CS201".to_string()),
        name: "数据结构".to_string(),
        credits: 4,
        capacity: 80,
        prerequisites: vec![CourseId("CS101".to_string())],
    });

    service.add_course(Course {
        id: CourseId("CS301".to_string()),
        name: "操作系统".to_string(),
        credits: 4,
        capacity: 60,
        prerequisites: vec![CourseId("CS201".to_string())],
    });

    service.add_course(Course {
        id: CourseId("CS401".to_string()),
        name: "编译原理".to_string(),
        credits: 3,
        capacity: 40,
        prerequisites: vec![CourseId("CS301".to_string())],
    });

    service.add_course(Course {
        id: CourseId("CS102".to_string()),
        name: "Python程序设计".to_string(),
        credits: 3,
        capacity: 100,
        prerequisites: vec![],
    });

    service.add_course(Course {
        id: CourseId("MATH101".to_string()),
        name: "高等数学".to_string(),
        credits: 5,
        capacity: 150,
        prerequisites: vec![],
    });

    service.add_course(Course {
        id: CourseId("ENG101".to_string()),
        name: "大学英语".to_string(),
        credits: 3,
        capacity: 50,
        prerequisites: vec![],
    });

    service.add_student(Student {
        id: StudentId("S001".to_string()),
        name: "张三".to_string(),
        completed_courses: vec![
            CompletedCourse {
                course_id: CourseId("CS101".to_string()),
                score: 85.0,
            },
            CompletedCourse {
                course_id: CourseId("CS201".to_string()),
                score: 78.0,
            },
        ],
    });

    service.add_student(Student {
        id: StudentId("S002".to_string()),
        name: "李四".to_string(),
        completed_courses: vec![
            CompletedCourse {
                course_id: CourseId("CS101".to_string()),
                score: 90.0,
            },
            CompletedCourse {
                course_id: CourseId("CS201".to_string()),
                score: 88.0,
            },
            CompletedCourse {
                course_id: CourseId("CS301".to_string()),
                score: 82.0,
            },
        ],
    });

    service.add_student(Student {
        id: StudentId("S003".to_string()),
        name: "王五".to_string(),
        completed_courses: vec![
            CompletedCourse {
                course_id: CourseId("CS101".to_string()),
                score: 55.0,
            },
        ],
    });
}

#[tokio::main]
async fn main() {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "server=debug,tower_http=debug".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let args = Args::parse();

    let service = Arc::new(CourseService::new(Utc::now()));
    initialize_data(&service);

    let app = Router::new()
        .route("/courses", get(handlers::list_courses))
        .route("/courses/:id", get(handlers::get_course))
        .route("/students", get(handlers::list_students))
        .route("/students/:id", get(handlers::get_student))
        .route("/enroll", post(handlers::enroll))
        .route("/withdraw", post(handlers::withdraw))
        .route("/students/:id/enrollments", get(handlers::get_enrollments))
        .route("/students/:id/transcript", get(handlers::get_transcript))
        .with_state(service)
        .layer(TraceLayer::new_for_http());

    let addr = format!("{}:{}", args.host, args.port);
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    tracing::debug!("listening on {}", addr);

    axum::serve(listener, app).await.unwrap();
}
