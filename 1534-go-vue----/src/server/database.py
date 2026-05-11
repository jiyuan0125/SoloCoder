import os
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from .config import settings


def ensure_data_dir():
    db_path = settings.database_url
    if db_path.startswith("sqlite:///"):
        file_path = db_path.replace("sqlite:///", "", 1)
        dir_name = os.path.dirname(file_path)
        if dir_name and not os.path.exists(dir_name):
            os.makedirs(dir_name, exist_ok=True)


ensure_data_dir()

engine = create_engine(
    settings.database_url,
    connect_args={"check_same_thread": False} if settings.database_url.startswith("sqlite") else {}
)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)


def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


def init_db():
    from core.models import Base
    Base.metadata.create_all(bind=engine)
