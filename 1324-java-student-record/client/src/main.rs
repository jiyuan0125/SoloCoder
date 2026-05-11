use clap::{Parser, Subcommand};
use student_record_core::{
    CreditSummary, CourseType, GraduationGap,
    Major, Student, TransferRequest, TransferResult,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, default_value = "http://127.0.0.1:3000", env = "SERVER_URL")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    ListMajors,
    GetMajor { id: String },
    ListStudents,
    GetStudent { id: String },
    GetStudentCredits { id: String },
    GetStudentGap { id: String },
    Transfer { student_id: String, target_major_id: String },
}

#[tokio::main]
async fn main() -> std::result::Result<(), Box<dyn std::error::Error>> {
    let args = Args::parse();
    let client = reqwest::Client::new();

    match args.command {
        Commands::ListMajors => {
            let url = format!("{}/api/majors", args.server);
            let response = client.get(&url).send().await?;
            let majors: Vec<Major> = response.json().await?;
            println!("专业列表：");
            for m in majors {
                println!("  ID: {}, 名称: {}", m.id, m.name);
                println!("    要求 - 公共课: {}, 专业课: {}, 选修课: {}",
                    m.requirements.public, m.requirements.professional, m.requirements.elective);
                println!("    课程数量: {}", m.curriculum.len());
            }
        }

        Commands::GetMajor { id } => {
            let url = format!("{}/api/majors/{}", args.server, id);
            let response = client.get(&url).send().await?;
            if response.status().is_success() {
                let major: Major = response.json().await?;
                println!("专业详情：");
                println!("  ID: {}, 名称: {}", major.id, major.name);
                println!("  毕业要求：");
                println!("    公共课: {} 学分", major.requirements.public);
                println!("    专业课: {} 学分", major.requirements.professional);
                println!("    选修课: {} 学分", major.requirements.elective);
                println!("  培养方案课程：");
                for course in major.curriculum.values() {
                    let type_str = match course.course_type {
                        CourseType::Public => "公共课",
                        CourseType::Professional => "专业课",
                        CourseType::Elective => "选修课",
                    };
                    println!("    {} - {} ({})", course.id, course.name, type_str);
                }
            } else {
                let error: serde_json::Value = response.json().await?;
                eprintln!("错误: {}", error["error"]);
            }
        }

        Commands::ListStudents => {
            let url = format!("{}/api/students", args.server);
            let response = client.get(&url).send().await?;
            let students: Vec<Student> = response.json().await?;
            println!("学生列表：");
            for s in students {
                println!("  学号: {}, 姓名: {}, 入学年份: {}, 当前专业: {}",
                    s.id, s.name, s.enrollment_year, s.current_major_id);
                println!("    已修课程: {} 门", s.completed_courses.len());
            }
        }

        Commands::GetStudent { id } => {
            let url = format!("{}/api/students/{}", args.server, id);
            let response = client.get(&url).send().await?;
            if response.status().is_success() {
                let student: Student = response.json().await?;
                println!("学生详情：");
                println!("  学号: {}", student.id);
                println!("  姓名: {}", student.name);
                println!("  入学年份: {}", student.enrollment_year);
                println!("  当前专业: {}", student.current_major_id);
                println!("  已修课程：");
                for course in &student.completed_courses {
                    let type_str = match course.original_course_type {
                        CourseType::Public => "公共课",
                        CourseType::Professional => "专业课",
                        CourseType::Elective => "选修课",
                    };
                    println!("    {} - {} ({}学分, {})",
                        course.course_id, course.course_name, course.credit, type_str);
                }
            } else {
                let error: serde_json::Value = response.json().await?;
                eprintln!("错误: {}", error["error"]);
            }
        }

        Commands::GetStudentCredits { id } => {
            let url = format!("{}/api/students/{}/credits", args.server, id);
            let response = client.get(&url).send().await?;
            if response.status().is_success() {
                let credits: CreditSummary = response.json().await?;
                println!("学生 {} 已修学分：", id);
                println!("  公共课: {} 学分", credits.public);
                println!("  专业课: {} 学分", credits.professional);
                println!("  选修课: {} 学分", credits.elective);
                println!("  合计: {} 学分", credits.public + credits.professional + credits.elective);
            } else {
                let error: serde_json::Value = response.json().await?;
                eprintln!("错误: {}", error["error"]);
            }
        }

        Commands::GetStudentGap { id } => {
            let url = format!("{}/api/students/{}/gap", args.server, id);
            let response = client.get(&url).send().await?;
            if response.status().is_success() {
                let gap: GraduationGap = response.json().await?;
                println!("学生 {} 毕业学分缺口：", id);
                println!("  公共课还需: {} 学分", gap.public);
                println!("  专业课还需: {} 学分", gap.professional);
                println!("  选修课还需: {} 学分", gap.elective);
                println!("  合计还需: {} 学分", gap.total);
                if gap.total == 0.0 {
                    println!("  恭喜！已满足毕业学分要求");
                }
            } else {
                let error: serde_json::Value = response.json().await?;
                eprintln!("错误: {}", error["error"]);
            }
        }

        Commands::Transfer { student_id, target_major_id } => {
            let url = format!("{}/api/transfer", args.server);
            let request = TransferRequest {
                student_id: student_id.clone(),
                target_major_id: target_major_id.clone(),
            };
            let response = client.post(&url).json(&request).send().await?;
            
            if response.status().is_success() {
                let result: TransferResult = response.json().await?;
                println!("转专业申请结果：");
                println!("  学生: {}", result.student_id);
                println!("  从专业: {} -> {}", result.from_major_id, result.to_major_id);
                println!("  状态: {}", if result.success { "成功" } else { "失败" });
                println!("  信息: {}", result.message);
                println!();
                println!("  认定学分：");
                println!("    公共课: {} 学分", result.transferred_credits.public);
                println!("    专业课: {} 学分", result.transferred_credits.professional);
                println!("    选修课: {} 学分", result.transferred_credits.elective);
                if result.excess_elective > 0.0 {
                    println!("    超出50%上限未认定的选修课: {} 学分", result.excess_elective);
                }
                println!();
                println!("  转专业后毕业缺口：");
                println!("    公共课还需: {} 学分", result.graduation_gap.public);
                println!("    专业课还需: {} 学分", result.graduation_gap.professional);
                println!("    选修课还需: {} 学分", result.graduation_gap.elective);
                println!("    合计还需: {} 学分", result.graduation_gap.total);
            } else {
                let error: serde_json::Value = response.json().await?;
                eprintln!("错误: {}", error["error"]);
            }
        }
    }

    Ok(())
}
