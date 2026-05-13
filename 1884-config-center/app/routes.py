from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from typing import List, Optional
from app.database import get_db
from app.models import Config, ConfigVersion
from app.schemas import ConfigCreate, ConfigResponse, ConfigVersionResponse
from app.notifier import notify_watchers

router = APIRouter()


def _get_snapshot(db: Session, project: str, env: str, global_version: int) -> dict:
    all_versions = db.query(ConfigVersion).join(Config).filter(
        Config.project == project,
        Config.env == env,
        ConfigVersion.id <= global_version
    ).order_by(ConfigVersion.created_at.asc()).all()

    snapshot = {}
    for ver in all_versions:
        if ver.value is not None:
            snapshot[ver.config.key] = ver.value
        elif ver.config.key in snapshot:
            del snapshot[ver.config.key]
    return snapshot


@router.get("/configs/{project}/{env}/diff")
def get_diff(
    project: str,
    env: str,
    v1: int = Query(..., description="First global version number (ConfigVersion.id)"),
    v2: int = Query(..., description="Second global version number (ConfigVersion.id)"),
    db: Session = Depends(get_db)
):
    v1_state = _get_snapshot(db, project, env, v1)
    v2_state = _get_snapshot(db, project, env, v2)

    added = []
    removed = []
    modified = []

    all_keys = set(v1_state.keys()) | set(v2_state.keys())

    for key in all_keys:
        in_v1 = key in v1_state
        in_v2 = key in v2_state

        if in_v1 and not in_v2:
            removed.append({
                "key": key,
                "old_value": v1_state[key]
            })
        elif not in_v1 and in_v2:
            added.append({
                "key": key,
                "new_value": v2_state[key]
            })
        elif v1_state[key] != v2_state[key]:
            modified.append({
                "key": key,
                "old_value": v1_state[key],
                "new_value": v2_state[key]
            })

    return {
        "project": project,
        "env": env,
        "v1": v1,
        "v2": v2,
        "added": added,
        "removed": removed,
        "modified": modified
    }


@router.put("/configs/{project}/{env}/{key}", response_model=ConfigResponse)
async def put_config(
    project: str,
    env: str,
    key: str,
    body: ConfigCreate,
    db: Session = Depends(get_db)
):
    from app.models import Watch

    watches = db.query(Watch).filter(
        Watch.project == project,
        Watch.env == env,
        Watch.key == key,
        Watch.status == "active"
    ).all()
    watches_data = [
        {"id": w.id, "callback_url": w.callback_url, "failed_count": w.failed_count}
        for w in watches
    ]

    config = db.query(Config).filter(
        Config.project == project,
        Config.env == env,
        Config.key == key
    ).first()

    if config:
        if config.current_value == body.value:
            return ConfigResponse(
                project=config.project,
                env=config.env,
                key=config.key,
                value=config.current_value,
                version=config.current_version
            )
        new_version = config.current_version + 1
        config.current_value = body.value
        config.current_version = new_version
        operation = "update"
    else:
        new_version = 1
        config = Config(
            project=project,
            env=env,
            key=key,
            current_value=body.value,
            current_version=new_version
        )
        db.add(config)
        db.flush()
        operation = "create"

    version_record = ConfigVersion(
        config_id=config.id,
        version=new_version,
        value=body.value,
        operation=operation
    )
    db.add(version_record)
    db.commit()
    db.refresh(config)

    results = await notify_watchers(watches_data, project, env, key, body.value, new_version)
    if results:
        for result in results:
            watch = db.query(Watch).filter(Watch.id == result["id"]).first()
            if watch:
                watch.failed_count = result["failed_count"]
                watch.status = result["status"]
        db.commit()

    return ConfigResponse(
        project=config.project,
        env=config.env,
        key=config.key,
        value=config.current_value,
        version=config.current_version
    )


@router.get("/configs/{project}/{env}/{key}/history", response_model=List[ConfigVersionResponse])
def get_config_history(
    project: str,
    env: str,
    key: str,
    db: Session = Depends(get_db)
):
    config = db.query(Config).filter(
        Config.project == project,
        Config.env == env,
        Config.key == key
    ).first()

    if not config:
        raise HTTPException(status_code=404, detail="Config not found")

    versions = db.query(ConfigVersion).filter(
        ConfigVersion.config_id == config.id
    ).order_by(ConfigVersion.version.asc()).all()

    return [
        ConfigVersionResponse(
            version=v.version,
            value=v.value,
            operation=v.operation,
            created_at=v.created_at
        ) for v in versions
    ]


@router.post("/configs/{project}/{env}/{key}/rollback", response_model=ConfigResponse)
async def rollback_config(
    project: str,
    env: str,
    key: str,
    version: int = Query(..., description="Target version to rollback to"),
    db: Session = Depends(get_db)
):
    from app.models import Watch

    watches = db.query(Watch).filter(
        Watch.project == project,
        Watch.env == env,
        Watch.key == key,
        Watch.status == "active"
    ).all()
    watches_data = [
        {"id": w.id, "callback_url": w.callback_url, "failed_count": w.failed_count}
        for w in watches
    ]

    config = db.query(Config).filter(
        Config.project == project,
        Config.env == env,
        Config.key == key
    ).first()

    if not config:
        raise HTTPException(status_code=404, detail="Config not found")

    target_version = db.query(ConfigVersion).filter(
        ConfigVersion.config_id == config.id,
        ConfigVersion.version == version
    ).first()

    if not target_version:
        raise HTTPException(status_code=404, detail="Target version not found")

    new_version = config.current_version + 1
    config.current_value = target_version.value
    config.current_version = new_version

    version_record = ConfigVersion(
        config_id=config.id,
        version=new_version,
        value=target_version.value,
        operation=f"rollback-to-{version}"
    )
    db.add(version_record)
    db.commit()
    db.refresh(config)

    results = await notify_watchers(watches_data, project, env, key, target_version.value, new_version)
    if results:
        for result in results:
            watch = db.query(Watch).filter(Watch.id == result["id"]).first()
            if watch:
                watch.failed_count = result["failed_count"]
                watch.status = result["status"]
        db.commit()

    return ConfigResponse(
        project=config.project,
        env=config.env,
        key=config.key,
        value=config.current_value,
        version=config.current_version
    )


@router.get("/configs/{project}/{env}/{key}", response_model=ConfigResponse)
def get_config(
    project: str,
    env: str,
    key: str,
    db: Session = Depends(get_db)
):
    config = db.query(Config).filter(
        Config.project == project,
        Config.env == env,
        Config.key == key
    ).first()

    if not config:
        raise HTTPException(status_code=404, detail="Config not found")

    return ConfigResponse(
        project=config.project,
        env=config.env,
        key=config.key,
        value=config.current_value,
        version=config.current_version
    )


@router.delete("/configs/{project}/{env}/{key}")
def delete_config(
    project: str,
    env: str,
    key: str,
    db: Session = Depends(get_db)
):
    config = db.query(Config).filter(
        Config.project == project,
        Config.env == env,
        Config.key == key
    ).first()

    if not config:
        raise HTTPException(status_code=404, detail="Config not found")

    db.delete(config)
    db.commit()
    return {"status": "deleted"}
