from sqlalchemy.orm import Session
from sqlalchemy import and_
from typing import Optional, List
from datetime import datetime
from app.models import Config, Watch, ConfigStatus, Environment
from app.schemas import ConfigCreate, ConfigUpdate, WatchCreate


def get_config_by_key_env(db: Session, key: str, environment: Environment) -> Optional[Config]:
    return db.query(Config).filter(
        and_(
            Config.key == key,
            Config.environment == environment
        )
    ).order_by(Config.version.desc()).first()


def get_configs_by_environment(db: Session, environment: Environment) -> List[Config]:
    active_configs = db.query(Config).filter(
        and_(
            Config.environment == environment,
            Config.status == ConfigStatus.ACTIVE
        )
    ).all()
    
    latest_versions = {}
    for config in active_configs:
        if config.key not in latest_versions or config.version > latest_versions[config.key].version:
            latest_versions[config.key] = config
    
    return list(latest_versions.values())


def create_config(db: Session, config_data: ConfigCreate) -> Config:
    existing_draft = db.query(Config).filter(
        and_(
            Config.key == config_data.key,
            Config.environment == config_data.environment,
            Config.status == ConfigStatus.DRAFT
        )
    ).first()
    
    if existing_draft:
        existing_draft.value = config_data.value
        existing_draft.updated_at = datetime.utcnow()
        db.commit()
        db.refresh(existing_draft)
        return existing_draft
    
    active_config = db.query(Config).filter(
        and_(
            Config.key == config_data.key,
            Config.environment == config_data.environment,
            Config.status == ConfigStatus.ACTIVE
        )
    ).first()
    
    next_version = 1
    if active_config:
        next_version = active_config.version + 1
    else:
        max_version = db.query(Config).filter(
            and_(
                Config.key == config_data.key,
                Config.environment == config_data.environment
            )
        ).order_by(Config.version.desc()).first()
        if max_version:
            next_version = max_version.version + 1
    
    config = Config(
        key=config_data.key,
        environment=config_data.environment,
        value=config_data.value,
        version=next_version,
        status=ConfigStatus.DRAFT
    )
    db.add(config)
    db.commit()
    db.refresh(config)
    return config


def update_config(db: Session, key: str, environment: Environment, update_data: ConfigUpdate) -> Optional[Config]:
    config = get_config_by_key_env(db, key, environment)
    if config and config.status == ConfigStatus.DRAFT:
        config.value = update_data.value
        config.updated_at = datetime.utcnow()
        db.commit()
        db.refresh(config)
        return config
    return None


def publish_config(db: Session, key: str, environment: Environment) -> Optional[Config]:
    draft_config = db.query(Config).filter(
        and_(
            Config.key == key,
            Config.environment == environment,
            Config.status == ConfigStatus.DRAFT
        )
    ).first()
    
    if not draft_config:
        return None
    
    current_active = db.query(Config).filter(
        and_(
            Config.key == key,
            Config.environment == environment,
            Config.status == ConfigStatus.ACTIVE
        )
    ).first()
    
    if current_active:
        current_active.status = ConfigStatus.ARCHIVED
    
    draft_config.status = ConfigStatus.ACTIVE
    db.commit()
    db.refresh(draft_config)
    return draft_config


def rollback_config(db: Session, key: str, environment: Environment, version: int) -> Optional[Config]:
    target_config = db.query(Config).filter(
        and_(
            Config.key == key,
            Config.environment == environment,
            Config.version == version
        )
    ).first()
    
    if not target_config or target_config.status != ConfigStatus.ARCHIVED:
        return None
    
    current_active = db.query(Config).filter(
        and_(
            Config.key == key,
            Config.environment == environment,
            Config.status == ConfigStatus.ACTIVE
        )
    ).first()
    
    if current_active:
        current_active.status = ConfigStatus.ARCHIVED
    
    new_config = Config(
        key=target_config.key,
        environment=target_config.environment,
        value=target_config.value,
        version=current_active.version + 1 if current_active else target_config.version + 1,
        status=ConfigStatus.ACTIVE
    )
    db.add(new_config)
    db.commit()
    db.refresh(new_config)
    return new_config


def get_config_versions(db: Session, key: str, environment: Environment) -> List[Config]:
    return db.query(Config).filter(
        and_(
            Config.key == key,
            Config.environment == environment
        )
    ).order_by(Config.version.asc()).all()


def get_config_diff(db: Session, key: str, environment: Environment, from_version: int, to_version: int):
    from_config = db.query(Config).filter(
        and_(
            Config.key == key,
            Config.environment == environment,
            Config.version == from_version
        )
    ).first()
    
    to_config = db.query(Config).filter(
        and_(
            Config.key == key,
            Config.environment == environment,
            Config.version == to_version
        )
    ).first()
    
    if not from_config or not to_config:
        return None
    
    return {
        "from_version": from_version,
        "to_version": to_version,
        "key": key,
        "environment": environment,
        "from_value": from_config.value,
        "to_value": to_config.value,
        "changed": from_config.value != to_config.value
    }


def create_watch(db: Session, watch_data: WatchCreate) -> Watch:
    watch = Watch(
        key=watch_data.key,
        callback_url=watch_data.callback_url,
        environment=watch_data.environment
    )
    db.add(watch)
    db.commit()
    db.refresh(watch)
    return watch


def get_watch(db: Session, watch_id: int) -> Optional[Watch]:
    return db.query(Watch).filter(Watch.id == watch_id).first()


def activate_watch(db: Session, watch_id: int) -> Optional[Watch]:
    watch = get_watch(db, watch_id)
    if watch:
        watch.is_active = "active"
        watch.failure_count = 0
        watch.updated_at = datetime.utcnow()
        db.commit()
        db.refresh(watch)
    return watch
