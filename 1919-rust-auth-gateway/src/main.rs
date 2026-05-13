use actix_web::{web, App, HttpResponse, HttpServer, Responder, http::StatusCode, error};
use actix_session::{Session, SessionMiddleware, storage::CookieSessionStore};
use actix_web::cookie::Key;
use chrono::{DateTime, Utc, Duration};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Mutex;
use uuid::Uuid;

const ACCESS_TOKEN_EXPIRE_MINUTES: i64 = 120;
const REFRESH_TOKEN_EXPIRE_DAYS: i64 = 7;
const AUTH_CODE_EXPIRE_SECONDS: i64 = 600;

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct Client {
    pub client_id: String,
    pub client_secret: String,
    pub created_at: DateTime<Utc>,
}

#[derive(Clone, Debug)]
pub struct AuthCode {
    pub code: String,
    pub client_id: String,
    pub redirect_uri: String,
    pub user_id: String,
    pub expires_at: DateTime<Utc>,
    pub used: bool,
}

#[derive(Clone, Debug, Serialize)]
pub struct AccessToken {
    pub token: String,
    pub client_id: String,
    pub user_id: String,
    pub expires_at: DateTime<Utc>,
}

#[derive(Clone, Debug)]
pub struct RefreshToken {
    pub token: String,
    pub client_id: String,
    pub user_id: String,
    pub expires_at: DateTime<Utc>,
    pub used: bool,
}

pub struct AppState {
    pub clients: Mutex<HashMap<String, Client>>,
    pub auth_codes: Mutex<HashMap<String, AuthCode>>,
    pub access_tokens: Mutex<HashMap<String, AccessToken>>,
    pub refresh_tokens: Mutex<HashMap<String, RefreshToken>>,
    pub users: Mutex<HashMap<String, String>>,
}

impl AppState {
    fn new() -> Self {
        let mut users = HashMap::new();
        users.insert("admin".to_string(), "admin123".to_string());
        users.insert("user".to_string(), "user123".to_string());
        
        AppState {
            clients: Mutex::new(HashMap::new()),
            auth_codes: Mutex::new(HashMap::new()),
            access_tokens: Mutex::new(HashMap::new()),
            refresh_tokens: Mutex::new(HashMap::new()),
            users: Mutex::new(users),
        }
    }
}

fn generate_random_string() -> String {
    Uuid::new_v4().to_string().replace("-", "")
}

#[derive(Deserialize)]
pub struct AuthorizeQuery {
    pub client_id: String,
    pub redirect_uri: String,
    pub response_type: String,
}

#[derive(Deserialize)]
pub struct LoginForm {
    pub username: String,
    pub password: String,
    pub client_id: String,
    pub redirect_uri: String,
}

#[derive(Deserialize)]
pub struct TokenRequest {
    pub grant_type: String,
    pub code: String,
    pub redirect_uri: String,
    pub client_id: String,
    pub client_secret: String,
}

#[derive(Deserialize)]
pub struct RefreshTokenRequest {
    pub refresh_token: String,
}

#[derive(Serialize)]
pub struct TokenResponse {
    pub access_token: String,
    pub token_type: String,
    pub expires_in: i64,
    pub refresh_token: String,
}

#[derive(Serialize)]
pub struct ClientResponse {
    pub client_id: String,
    pub client_secret: String,
}

fn json_error(err: &str, status: StatusCode) -> HttpResponse {
    HttpResponse::build(status)
        .content_type("application/json")
        .body(format!("{{\"error\": \"{}\"}}", err))
}

async fn register_client(data: web::Data<AppState>) -> impl Responder {
    let client_id = generate_random_string();
    let client_secret = generate_random_string();
    
    let client = Client {
        client_id: client_id.clone(),
        client_secret: client_secret.clone(),
        created_at: Utc::now(),
    };
    
    let mut clients = data.clients.lock().unwrap();
    clients.insert(client_id.clone(), client);
    
    HttpResponse::Ok().json(ClientResponse { client_id, client_secret })
}

async fn authorize(
    query: web::Query<AuthorizeQuery>,
    session: Session,
    data: web::Data<AppState>,
) -> impl Responder {
    if query.response_type != "code" {
        return json_error("invalid_request", StatusCode::BAD_REQUEST);
    }
    
    let clients = data.clients.lock().unwrap();
    if !clients.contains_key(&query.client_id) {
        return json_error("invalid_client", StatusCode::BAD_REQUEST);
    }
    drop(clients);
    
    if let Some(user_id) = session.get::<String>("user_id").unwrap_or(None) {
        let auth_code = generate_random_string();
        let expires_at = Utc::now() + Duration::seconds(AUTH_CODE_EXPIRE_SECONDS);
        
        let code = AuthCode {
            code: auth_code.clone(),
            client_id: query.client_id.clone(),
            redirect_uri: query.redirect_uri.clone(),
            user_id,
            expires_at,
            used: false,
        };
        
        let mut codes = data.auth_codes.lock().unwrap();
        codes.insert(auth_code.clone(), code);
        
        let redirect_url = format!(
            "{}?code={}",
            query.redirect_uri,
            auth_code
        );
        
        return HttpResponse::Found()
            .append_header(("Location", redirect_url))
            .finish();
    }
    
    let login_html = format!(
        r#"<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>OAuth2 登录</title>
</head>
<body style="font-family: Arial, sans-serif; display: flex; justify-content: center; align-items: center; min-height: 100vh; margin: 0; background-color: #f5f5f5;">
    <div style="background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); width: 100%; max-width: 400px;">
        <h1 style="text-align: center; color: #333; margin-bottom: 1.5rem;">OAuth2 登录</h1>
        <form method="POST" action="/login">
            <input type="hidden" name="client_id" value="{}">
            <input type="hidden" name="redirect_uri" value="{}">
            <div style="margin-bottom: 1rem;">
                <label style="display: block; margin-bottom: 0.5rem; color: #555;">用户名</label>
                <input type="text" name="username" required style="width: 100%; padding: 0.75rem; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box;">
            </div>
            <div style="margin-bottom: 1.5rem;">
                <label style="display: block; margin-bottom: 0.5rem; color: #555;">密码</label>
                <input type="password" name="password" required style="width: 100%; padding: 0.75rem; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box;">
            </div>
            <button type="submit" style="width: 100%; padding: 0.75rem; background-color: #4CAF50; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 1rem;">登录</button>
        </form>
        <p style="margin-top: 1rem; text-align: center; color: #666; font-size: 0.9rem;">测试用户: admin/admin123 或 user/user123</p>
    </div>
</body>
</html>
"#,
        query.client_id,
        query.redirect_uri
    );
    
    HttpResponse::Ok()
        .content_type("text/html; charset=utf-8")
        .body(login_html)
}

async fn login(
    form: web::Form<LoginForm>,
    session: Session,
    data: web::Data<AppState>,
) -> impl Responder {
    let users = data.users.lock().unwrap();
    
    if let Some(stored_password) = users.get(&form.username) {
        if stored_password == &form.password {
            session.insert("user_id", &form.username).unwrap();
            drop(users);
            
            let auth_code = generate_random_string();
            let expires_at = Utc::now() + Duration::seconds(AUTH_CODE_EXPIRE_SECONDS);
            
            let code = AuthCode {
                code: auth_code.clone(),
                client_id: form.client_id.clone(),
                redirect_uri: form.redirect_uri.clone(),
                user_id: form.username.clone(),
                expires_at,
                used: false,
            };
            
            let mut codes = data.auth_codes.lock().unwrap();
            codes.insert(auth_code.clone(), code);
            
            let redirect_url = format!(
                "{}?code={}",
                form.redirect_uri,
                auth_code
            );
            
            return HttpResponse::Found()
                .append_header(("Location", redirect_url))
                .finish();
        }
    }
    
    HttpResponse::Unauthorized()
        .content_type("text/html; charset=utf-8")
        .body("<html><body><h1>登录失败</h1><p>用户名或密码错误</p><a href=\"#\" onclick=\"history.back()\">返回</a></body></html>")
}

async fn token(
    req: web::Json<TokenRequest>,
    data: web::Data<AppState>,
) -> impl Responder {
    if req.grant_type != "authorization_code" {
        return json_error("unsupported_grant_type", StatusCode::BAD_REQUEST);
    }
    
    let clients = data.clients.lock().unwrap();
    if let Some(client) = clients.get(&req.client_id) {
        if client.client_secret != req.client_secret {
            return json_error("invalid_client", StatusCode::BAD_REQUEST);
        }
    } else {
        return json_error("invalid_client", StatusCode::BAD_REQUEST);
    }
    drop(clients);
    
    let mut codes = data.auth_codes.lock().unwrap();
    if let Some(code) = codes.get_mut(&req.code) {
        if code.used || code.expires_at < Utc::now() || code.redirect_uri != req.redirect_uri {
            return json_error("invalid_grant", StatusCode::BAD_REQUEST);
        }
        
        code.used = true;
        let client_id = code.client_id.clone();
        let user_id = code.user_id.clone();
        drop(codes);
        
        let access_token = generate_random_string();
        let refresh_token = generate_random_string();
        
        let access_expires = Utc::now() + Duration::minutes(ACCESS_TOKEN_EXPIRE_MINUTES);
        let refresh_expires = Utc::now() + Duration::days(REFRESH_TOKEN_EXPIRE_DAYS);
        
        let at = AccessToken {
            token: access_token.clone(),
            client_id: client_id.clone(),
            user_id: user_id.clone(),
            expires_at: access_expires,
        };
        
        let rt = RefreshToken {
            token: refresh_token.clone(),
            client_id,
            user_id,
            expires_at: refresh_expires,
            used: false,
        };
        
        let mut access_tokens = data.access_tokens.lock().unwrap();
        access_tokens.insert(access_token.clone(), at);
        drop(access_tokens);
        
        let mut refresh_tokens = data.refresh_tokens.lock().unwrap();
        refresh_tokens.insert(refresh_token.clone(), rt);
        drop(refresh_tokens);
        
        return HttpResponse::Ok().json(TokenResponse {
            access_token,
            token_type: "Bearer".to_string(),
            expires_in: ACCESS_TOKEN_EXPIRE_MINUTES * 60,
            refresh_token,
        });
    }
    
    json_error("invalid_grant", StatusCode::BAD_REQUEST)
}

async fn refresh_token(
    req: web::Json<RefreshTokenRequest>,
    data: web::Data<AppState>,
) -> impl Responder {
    let mut refresh_tokens = data.refresh_tokens.lock().unwrap();
    
    if let Some(old_rt) = refresh_tokens.get_mut(&req.refresh_token) {
        if old_rt.used || old_rt.expires_at < Utc::now() {
            return json_error("invalid_grant", StatusCode::UNAUTHORIZED);
        }
        
        old_rt.used = true;
        let client_id = old_rt.client_id.clone();
        let user_id = old_rt.user_id.clone();
        drop(refresh_tokens);
        
        let access_token = generate_random_string();
        let refresh_token = generate_random_string();
        
        let access_expires = Utc::now() + Duration::minutes(ACCESS_TOKEN_EXPIRE_MINUTES);
        let refresh_expires = Utc::now() + Duration::days(REFRESH_TOKEN_EXPIRE_DAYS);
        
        let at = AccessToken {
            token: access_token.clone(),
            client_id: client_id.clone(),
            user_id: user_id.clone(),
            expires_at: access_expires,
        };
        
        let rt = RefreshToken {
            token: refresh_token.clone(),
            client_id,
            user_id,
            expires_at: refresh_expires,
            used: false,
        };
        
        let mut access_tokens = data.access_tokens.lock().unwrap();
        access_tokens.insert(access_token.clone(), at);
        drop(access_tokens);
        
        let mut refresh_tokens = data.refresh_tokens.lock().unwrap();
        refresh_tokens.insert(refresh_token.clone(), rt);
        drop(refresh_tokens);
        
        return HttpResponse::Ok().json(TokenResponse {
            access_token,
            token_type: "Bearer".to_string(),
            expires_in: ACCESS_TOKEN_EXPIRE_MINUTES * 60,
            refresh_token,
        });
    }
    
    json_error("invalid_grant", StatusCode::UNAUTHORIZED)
}

fn json_config(cfg: &mut web::ServiceConfig) {
    cfg.app_data(web::JsonConfig::default().error_handler(|err, _req| {
        error::InternalError::from_response(
            "",
            HttpResponse::BadRequest()
                .content_type("application/json")
                .body(format!("{{\"error\": \"{}\"}}", err)),
        )
        .into()
    }));
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let port: u16 = std::env::var("PORT")
        .ok()
        .and_then(|p| p.parse().ok())
        .unwrap_or(9202);
    
    let app_state = web::Data::new(AppState::new());
    let secret_key = Key::generate();
    
    println!("OAuth2 Gateway 启动在 http://127.0.0.1:{}", port);
    
    HttpServer::new(move || {
        App::new()
            .app_data(app_state.clone())
            .configure(json_config)
            .wrap(
                SessionMiddleware::builder(CookieSessionStore::default(), secret_key.clone())
                    .cookie_secure(false)
                    .build(),
            )
            .route("/clients", web::post().to(register_client))
            .route("/authorize", web::get().to(authorize))
            .route("/login", web::post().to(login))
            .route("/token", web::post().to(token))
            .route("/token/refresh", web::post().to(refresh_token))
    })
    .bind(("127.0.0.1", port))?
    .run()
    .await
}
