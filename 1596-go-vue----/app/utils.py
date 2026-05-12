import secrets
import string
import io
import base64
from datetime import datetime, date, time, timedelta
from typing import Tuple, Optional, List
from cryptography.fernet import Fernet
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2HMAC
import qrcode
from app.config import settings


def generate_booking_no(prefix: str = "BK") -> str:
    timestamp = datetime.now().strftime("%Y%m%d%H%M%S")
    random_chars = ''.join(secrets.choice(string.ascii_uppercase + string.digits) for _ in range(4))
    return f"{prefix}{timestamp}{random_chars}"


def generate_payment_no() -> str:
    return generate_booking_no("PY")


def generate_refund_no() -> str:
    return generate_booking_no("RF")


def generate_class_no() -> str:
    return generate_booking_no("CL")


def generate_registration_no() -> str:
    return generate_booking_no("RG")


def get_qr_key() -> bytes:
    password = settings.QR_SECRET_KEY.encode()
    salt = b'gym-system-salt'
    kdf = PBKDF2HMAC(
        algorithm=hashes.SHA256(),
        length=32,
        salt=salt,
        iterations=100000,
    )
    key = base64.urlsafe_b64encode(kdf.derive(password))
    return key


def encrypt_qr_data(data: str) -> str:
    key = get_qr_key()
    f = Fernet(key)
    token = f.encrypt(data.encode())
    return base64.urlsafe_b64encode(token).decode()


def decrypt_qr_data(token_str: str) -> str:
    key = get_qr_key()
    f = Fernet(key)
    token = base64.urlsafe_b64decode(token_str.encode())
    return f.decrypt(token).decode()


def generate_qr_code(booking_no: str, user_id: int, venue_id: int, 
                     booking_date: date, start_time: time, end_time: time) -> Tuple[str, str]:
    data = f"{booking_no}|{user_id}|{venue_id}|{booking_date.isoformat()}|{start_time.isoformat()}|{end_time.isoformat()}"
    encrypted = encrypt_qr_data(data)
    
    qr = qrcode.QRCode(
        version=1,
        error_correction=qrcode.constants.ERROR_CORRECT_L,
        box_size=10,
        border=4,
    )
    qr.add_data(encrypted)
    qr.make(fit=True)
    
    img = qr.make_image(fill_color="black", back_color="white")
    buffer = io.BytesIO()
    img.save(buffer, format="PNG")
    img_base64 = base64.b64encode(buffer.getvalue()).decode()
    
    return encrypted, img_base64


def verify_qr_token(token: str) -> Optional[dict]:
    try:
        decrypted = decrypt_qr_data(token)
        parts = decrypted.split('|')
        if len(parts) != 6:
            return None
        return {
            "booking_no": parts[0],
            "user_id": int(parts[1]),
            "venue_id": int(parts[2]),
            "booking_date": date.fromisoformat(parts[3]),
            "start_time": time.fromisoformat(parts[4]),
            "end_time": time.fromisoformat(parts[5]),
        }
    except Exception:
        return None


def calculate_hours(start_time: time, end_time: time) -> float:
    start_dt = datetime.combine(date.today(), start_time)
    end_dt = datetime.combine(date.today(), end_time)
    if end_dt <= start_dt:
        end_dt += timedelta(days=1)
    delta = end_dt - start_dt
    return delta.total_seconds() / 3600.0


def is_time_overlap(slot1_start: time, slot1_end: time, 
                    slot2_start: time, slot2_end: time) -> bool:
    return not (slot1_end <= slot2_start or slot2_end <= slot1_start)


def is_continuous(slots: List[Tuple[time, time]]) -> bool:
    if len(slots) < 2:
        return False
    
    sorted_slots = sorted(slots, key=lambda x: x[0])
    
    for i in range(len(sorted_slots) - 1):
        current_end = sorted_slots[i][1]
        next_start = sorted_slots[i + 1][0]
        if current_end != next_start:
            return False
    
    return True


def calculate_refund_rate(booking_date: date, start_time: time) -> Tuple[float, str]:
    now = datetime.now()
    booking_datetime = datetime.combine(booking_date, start_time)
    
    if now >= booking_datetime:
        return 0.0, "活动已开始，不退款"
    
    delta = booking_datetime - now
    hours_before = delta.total_seconds() / 3600
    
    if hours_before >= 2:
        return 1.0, "提前2小时以上取消，全额退款"
    else:
        return settings.PARTIAL_REFUND_RATE, "2小时内取消，退款50%"


def round_robin_schedule(teams: List[int], start_date: date, 
                         start_time: time = time(9, 0), 
                         match_duration_hours: float = 2.0,
                         rest_days: int = 1) -> List[dict]:
    n = len(teams)
    if n < 2:
        return []
    
    if n % 2 != 0:
        teams = teams + [None]
        n += 1
    
    rounds = n - 1
    matches = []
    current_date = start_date
    match_no = 1
    
    for round_no in range(1, rounds + 1):
        round_matches = []
        
        half = n // 2
        for i in range(half):
            home = teams[i]
            away = teams[n - 1 - i]
            
            if home is not None and away is not None:
                round_matches.append({
                    "round_no": round_no,
                    "match_no": f"M{match_no:03d}",
                    "home_team_id": home,
                    "away_team_id": away,
                    "match_date": current_date,
                    "start_time": start_time,
                    "duration_hours": match_duration_hours,
                })
                match_no += 1
            
            teams = [teams[0]] + [teams[-1]] + teams[1:-1]
        
        current_date += timedelta(days=rest_days + 1)
    
    return matches
