from sqlalchemy import create_engine
from sqlalchemy.orm import declarative_base, sessionmaker
import os

DATABASE_URL = os.getenv("DATABASE_URL", "sqlite:///./grain_storage.db")

engine = create_engine(
    DATABASE_URL, connect_args={"check_same_thread": False} if "sqlite" in DATABASE_URL else {}
)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
Base = declarative_base()


class Database:
    @staticmethod
    def init_db():
        from . import models
        Base.metadata.create_all(bind=engine)

    @staticmethod
    def get_session():
        db = SessionLocal()
        try:
            yield db
        finally:
            db.close()


def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
