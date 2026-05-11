from sqlalchemy import create_engine, Column, Integer, String, Numeric, DateTime, ForeignKey, Enum as SAEnum
from sqlalchemy.orm import sessionmaker, declarative_base, relationship
from datetime import datetime

from src.core.models import (
    TodoStatus, TodoResolution, InspectionStatus, BatchStatus, RecipeStatus
)


SQLALCHEMY_DATABASE_URL = "sqlite:///./feed_factory.db"

engine = create_engine(
    SQLALCHEMY_DATABASE_URL, connect_args={"check_same_thread": False}
)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

Base = declarative_base()


class DBRecipe(Base):
    __tablename__ = "recipes"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, nullable=False)
    description = Column(String, nullable=True)
    status = Column(SAEnum(RecipeStatus), default=RecipeStatus.ACTIVE, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)

    raw_materials = relationship("DBRawMaterial", back_populates="recipe", cascade="all, delete-orphan")
    batches = relationship("DBProductionBatch", back_populates="recipe")


class DBRawMaterial(Base):
    __tablename__ = "raw_materials"

    id = Column(Integer, primary_key=True, index=True)
    recipe_id = Column(Integer, ForeignKey("recipes.id"), nullable=False)
    name = Column(String, nullable=False)
    percentage = Column(Numeric(5, 2), nullable=False)

    recipe = relationship("DBRecipe", back_populates="raw_materials")


class DBProductionBatch(Base):
    __tablename__ = "production_batches"

    id = Column(Integer, primary_key=True, index=True)
    recipe_id = Column(Integer, ForeignKey("recipes.id"), nullable=False)
    planned_quantity = Column(Numeric(12, 2), nullable=False)
    actual_quantity = Column(Numeric(12, 2), nullable=True)
    status = Column(SAEnum(BatchStatus), default=BatchStatus.PENDING, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, onupdate=datetime.utcnow, nullable=True)

    recipe = relationship("DBRecipe", back_populates="batches")
    inspections = relationship("DBQualityInspection", back_populates="batch", cascade="all, delete-orphan")
    todos = relationship("DBTodoItem", back_populates="batch", cascade="all, delete-orphan")


class DBQualityInspection(Base):
    __tablename__ = "quality_inspections"

    id = Column(Integer, primary_key=True, index=True)
    batch_id = Column(Integer, ForeignKey("production_batches.id"), nullable=False)
    status = Column(SAEnum(InspectionStatus), nullable=False)
    notes = Column(String, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)

    batch = relationship("DBProductionBatch", back_populates="inspections")


class DBTodoItem(Base):
    __tablename__ = "todo_items"

    id = Column(Integer, primary_key=True, index=True)
    batch_id = Column(Integer, ForeignKey("production_batches.id"), nullable=False)
    status = Column(SAEnum(TodoStatus), default=TodoStatus.PENDING, nullable=False)
    resolution = Column(SAEnum(TodoResolution), nullable=True)
    notes = Column(String, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    resolved_at = Column(DateTime, nullable=True)

    batch = relationship("DBProductionBatch", back_populates="todos")


def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


def init_db():
    Base.metadata.create_all(bind=engine)
